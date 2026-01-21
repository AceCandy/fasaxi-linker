package core

import (
	"context"
	"fmt"
	"time"

	"github.com/fasaxi-linker/servergo/internal/cache"
	"github.com/fasaxi-linker/servergo/internal/db"
)

// Cache manages the list of processed files to avoid duplicates
// Now uses PostgreSQL instead of file storage
type Cache struct {
	store    *cache.Store
	taskID   int
	taskName string
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	return &Cache{
		store:    &cache.Store{},
		taskID:   0,
		taskName: "",
	}
}

// SetTaskName sets the task name for this cache instance and resolves it to task ID
func (c *Cache) SetTaskName(taskName string) {
	c.taskName = taskName
	// Resolve task name to ID directly using db pool to avoid import cycle
	if taskID, err := getTaskIDByName(taskName); err == nil {
		c.taskID = taskID
	}
}

// getTaskIDByName retrieves the task ID for a given task name
func getTaskIDByName(taskName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return 0, fmt.Errorf("database connection pool is not initialized")
	}

	var taskID int
	query := `SELECT id FROM tasks WHERE name = $1`
	err := pool.QueryRow(ctx, query, taskName).Scan(&taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to get task ID for name %s: %w", taskName, err)
	}

	return taskID, nil
}

// Read reads the cache from database for this task
func (c *Cache) Read() ([]string, error) {
	return c.store.GetByTaskID(c.taskID)
}

// Write writes the cache to database (replaces all entries for this task)
func (c *Cache) Write(files []string) error {
	// Make unique
	unique := make(map[string]bool)
	var result []string
	for _, f := range files {
		if !unique[f] {
			unique[f] = true
			result = append(result, f)
		}
	}

	// Clear this task's cache and add new entries
	if err := c.store.ClearByTaskID(c.taskID); err != nil {
		return err
	}

	return c.store.Add(c.taskID, result)
}

// Add adds new files to cache for this task
func (c *Cache) Add(newFiles []string) error {
	return c.store.Add(c.taskID, newFiles)
}

// Has checks if a file is in cache for this task
func (c *Cache) Has(file string, ignoreCase bool) (bool, error) {
	return c.store.Has(c.taskID, file, ignoreCase)
}
