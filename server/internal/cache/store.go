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
func (s *Store) Has(taskID int, filePath string, ignoreCase bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return false, fmt.Errorf("database connection pool is not initialized")
	}

	var query string
	var args []interface{}

	if ignoreCase {
		query = `SELECT EXISTS(SELECT 1 FROM cache_files WHERE task_id = $1 AND LOWER(file_path) = LOWER($2))`
		args = []interface{}{taskID, filePath}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM cache_files WHERE task_id = $1 AND file_path = $2)`
		args = []interface{}{taskID, filePath}
	}

	var exists bool
	err := pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check cache: %w", err)
	}

	return exists, nil
}

// Add adds new file paths to cache for a specific task
func (s *Store) Add(taskID int, filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO cache_files (task_id, file_path) VALUES ($1, $2) ON CONFLICT (task_id, file_path) DO NOTHING`

	for _, filePath := range filePaths {
		if _, err := tx.Exec(ctx, query, taskID, filePath); err != nil {
			return fmt.Errorf("failed to insert cache file %s: %w", filePath, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
