package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fasaxi-linker/servergo/internal/db"
)

// Store manages cache data in PostgreSQL
type Store struct{}

// CacheEntry represents a cached file with its creation time
type CacheEntry struct {
	FilePath  string    `json:"filePath"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetByTaskID returns cached file paths for a specific task
func (s *Store) GetByTaskID(taskID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	query := `SELECT file_path FROM cache_files WHERE task_id = $1 ORDER BY created_at`
	rows, err := pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cache files: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return nil, fmt.Errorf("failed to scan cache file: %w", err)
		}
		files = append(files, filePath)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// Has checks if a file path exists in cache for a specific task
func (s *Store) Has(taskID int, filePath string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return false, fmt.Errorf("database connection pool is not initialized")
	}

	query := `SELECT EXISTS(SELECT 1 FROM cache_files WHERE task_id = $1 AND file_path = $2)`

	var exists bool
	err := pool.QueryRow(ctx, query, taskID, filePath).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check cache: %w", err)
	}

	return exists, nil
}

// Add adds new file paths to cache for a specific task (batch insert for performance)
func (s *Store) Add(taskID int, filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}

	// Increase timeout for large batches (1 minute per 10000 files)
	timeout := time.Duration(len(filePaths)/10000+1) * time.Minute
	if timeout > 10*time.Minute {
		timeout = 10 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	// Use batch insert with UNNEST for better performance
	const batchSize = 5000
	for i := 0; i < len(filePaths); i += batchSize {
		end := i + batchSize
		if end > len(filePaths) {
			end = len(filePaths)
		}
		batch := filePaths[i:end]

		// Build values for batch insert
		values := make([]string, len(batch))
		args := make([]interface{}, len(batch)+1)
		args[0] = taskID

		for j, path := range batch {
			values[j] = fmt.Sprintf("($1, $%d)", j+2)
			args[j+1] = path
		}

		query := fmt.Sprintf(
			`INSERT INTO cache_files (task_id, file_path) VALUES %s ON CONFLICT (task_id, file_path) DO NOTHING`,
			strings.Join(values, ", "),
		)

		if _, err := pool.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to batch insert cache files: %w", err)
		}
	}

	return nil
}

// Clear removes all cache entries
func (s *Store) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	query := `DELETE FROM cache_files`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}

// ClearByTaskID removes all cache entries for a specific task
func (s *Store) ClearByTaskID(taskID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	query := `DELETE FROM cache_files WHERE task_id = $1`
	if _, err := pool.Exec(ctx, query, taskID); err != nil {
		return fmt.Errorf("failed to clear cache for task %d: %w", taskID, err)
	}

	return nil
}

// ClearByTaskIDWithSearch removes cache entries matching search condition
func (s *Store) ClearByTaskIDWithSearch(taskID int, search string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	query := `DELETE FROM cache_files WHERE task_id = $1 AND file_path ILIKE $2`
	searchPattern := "%" + search + "%"
	if _, err := pool.Exec(ctx, query, taskID, searchPattern); err != nil {
		return fmt.Errorf("failed to clear cache for task %d with search: %w", taskID, err)
	}

	return nil
}

// Remove removes specific file paths from cache (for a specific task)
func (s *Store) Remove(taskID int, filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(filePaths))
	args := make([]interface{}, len(filePaths)+1)
	args[0] = taskID
	for i, path := range filePaths {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = path
	}

	query := fmt.Sprintf(`DELETE FROM cache_files WHERE task_id = $1 AND file_path IN (%s)`, strings.Join(placeholders, ","))
	if _, err := pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to remove cache files: %w", err)
	}

	return nil
}

// GetByTaskIDPaged returns cached file paths for a specific task with pagination and search
func (s *Store) GetByTaskIDPaged(taskID, page, pageSize int, search string) ([]CacheEntry, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return nil, 0, fmt.Errorf("database connection pool is not initialized")
	}

	offset := (page - 1) * pageSize
	var query string
	var args []interface{}
	var countQuery string
	var countArgs []interface{}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = `SELECT file_path, created_at FROM cache_files WHERE task_id = $1 AND file_path ILIKE $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = []interface{}{taskID, searchPattern, pageSize, offset}

		countQuery = `SELECT COUNT(*) FROM cache_files WHERE task_id = $1 AND file_path ILIKE $2`
		countArgs = []interface{}{taskID, searchPattern}
	} else {
		query = `SELECT file_path, created_at FROM cache_files WHERE task_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{taskID, pageSize, offset}

		countQuery = `SELECT COUNT(*) FROM cache_files WHERE task_id = $1`
		countArgs = []interface{}{taskID}
	}

	// Get total count
	var total int
	if err := pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count cache files: %w", err)
	}

	// Get data
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query cache files: %w", err)
	}
	defer rows.Close()

	var files []CacheEntry
	for rows.Next() {
		var entry CacheEntry
		if err := rows.Scan(&entry.FilePath, &entry.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan cache file: %w", err)
		}
		files = append(files, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
