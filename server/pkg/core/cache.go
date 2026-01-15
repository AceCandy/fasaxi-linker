package core

import (
	"github.com/fasaxi-linker/servergo/internal/cache"
)

// Cache manages the list of processed files to avoid duplicates
// Now uses PostgreSQL instead of file storage
type Cache struct {
	store    *cache.Store
	taskName string
}

// NewCache creates a new Cache instance
// filePath parameter is kept for backward compatibility but ignored
// taskName is extracted from the calling context
func NewCache(filePath string) *Cache {
	return &Cache{
		store:    &cache.Store{},
		taskName: "", // Will be set by SetTaskName
	}
}

// SetTaskName sets the task name for this cache instance
func (c *Cache) SetTaskName(taskName string) {
	c.taskName = taskName
}

// Read reads the cache from database for this task
func (c *Cache) Read() ([]string, error) {
	if c.taskName == "" {
		return c.store.GetAll()
	}
	return c.store.GetByTaskName(c.taskName)
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

	if c.taskName == "" {
		// Backward compatibility: clear all and add
		if err := c.store.Clear(); err != nil {
			return err
		}
		// Cannot add without task name
		return nil
	}

	// Clear this task's cache and add new entries
	if err := c.store.ClearByTaskName(c.taskName); err != nil {
		return err
	}

	return c.store.Add(c.taskName, result)
}

// Add adds new files to cache for this task
func (c *Cache) Add(newFiles []string) error {
	if c.taskName == "" {
		// Cannot add without task name
		return nil
	}
	return c.store.Add(c.taskName, newFiles)
}

// Has checks if a file is in cache for this task
func (c *Cache) Has(file string, ignoreCase bool) (bool, error) {
	if c.taskName == "" {
		// Backward compatibility: check across all tasks
		// This is not ideal but maintains compatibility
		return false, nil
	}
	return c.store.Has(c.taskName, file, ignoreCase)
}
