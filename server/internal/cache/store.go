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

// GetAll returns all cached file paths (for backward compatibility, returns all tasks)
func (s *Store) GetAll() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	query := `SELECT file_path FROM cache_files ORDER BY created_at`
	rows, err := pool.Query(ctx, query)
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

// GetByTaskName returns cached file paths for a specific task
func (s *Store) GetByTaskName(taskName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	query := `SELECT file_path FROM cache_files WHERE task_name = $1 ORDER BY created_at`
	rows, err := pool.Query(ctx, query, taskName)
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
func (s *Store) Has(taskName, filePath string, ignoreCase bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return false, fmt.Errorf("database connection pool is not initialized")
	}

	var query string
	var args []interface{}

	if ignoreCase {
		query = `SELECT EXISTS(SELECT 1 FROM cache_files WHERE task_name = $1 AND LOWER(file_path) = LOWER($2))`
		args = []interface{}{taskName, filePath}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM cache_files WHERE task_name = $1 AND file_path = $2)`
		args = []interface{}{taskName, filePath}
	}

	var exists bool
	err := pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check cache: %w", err)
	}

	return exists, nil
}

// Add adds new file paths to cache for a specific task
func (s *Store) Add(taskName string, filePaths []string) error {
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

	query := `INSERT INTO cache_files (task_name, file_path) VALUES ($1, $2) ON CONFLICT (task_name, file_path) DO NOTHING`

	for _, filePath := range filePaths {
		if _, err := tx.Exec(ctx, query, taskName, filePath); err != nil {
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

// ClearByTaskName removes all cache entries for a specific task
func (s *Store) ClearByTaskName(taskName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	query := `DELETE FROM cache_files WHERE task_name = $1`
	if _, err := pool.Exec(ctx, query, taskName); err != nil {
		return fmt.Errorf("failed to clear cache for task %s: %w", taskName, err)
	}

	return nil
}

// Remove removes specific file paths from cache (for a specific task)
func (s *Store) Remove(taskName string, filePaths []string) error {
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
	args[0] = taskName
	for i, path := range filePaths {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = path
	}

	query := fmt.Sprintf(`DELETE FROM cache_files WHERE task_name = $1 AND file_path IN (%s)`, strings.Join(placeholders, ","))
	if _, err := pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to remove cache files: %w", err)
	}

	return nil
}

// GetAllAsJSON returns all cached files as JSON string
func (s *Store) GetAllAsJSON() (string, error) {
	files, err := s.GetAll()
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "[]", nil
	}

	// Build JSON manually to match expected format
	var builder strings.Builder
	builder.WriteString("[\n")
	for i, file := range files {
		builder.WriteString("  \"")
		builder.WriteString(strings.ReplaceAll(file, "\"", "\\\""))
		builder.WriteString("\"")
		if i < len(files)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("]")

	return builder.String(), nil
}

// SetFromJSON replaces all cache entries from JSON string
// Deprecated: This method is for backward compatibility only.
// In the new task-isolated design, use Add(taskName, filePaths) instead.
func (s *Store) SetFromJSON(jsonContent string) error {
	// Parse JSON array
	jsonContent = strings.TrimSpace(jsonContent)
	if jsonContent == "" || jsonContent == "[]" {
		return s.Clear()
	}

	// Simple JSON array parser (assumes valid JSON)
	jsonContent = strings.TrimPrefix(jsonContent, "[")
	jsonContent = strings.TrimSuffix(jsonContent, "]")
	jsonContent = strings.TrimSpace(jsonContent)

	if jsonContent == "" {
		return s.Clear()
	}

	// Split by comma and extract strings
	var filePaths []string
	parts := strings.Split(jsonContent, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		if part != "" {
			filePaths = append(filePaths, part)
		}
	}

	// Clear and add new entries
	if err := s.Clear(); err != nil {
		return err
	}

	// Use a default task name for backward compatibility
	// This is not ideal but maintains API compatibility
	return s.Add("legacy", filePaths)
}
