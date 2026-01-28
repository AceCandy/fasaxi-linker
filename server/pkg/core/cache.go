package core

import (
	"github.com/fasaxi-linker/servergo/internal/cache"
)

// Cache manages the list of processed files to avoid duplicates
// Now uses PostgreSQL instead of file storage
type Cache struct {
	store  *cache.Store
	taskID int
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	return &Cache{
		store:  &cache.Store{},
		taskID: 0,
	}
}

// SetTaskID sets the task ID for this cache instance.
func (c *Cache) SetTaskID(taskID int) {
	c.taskID = taskID
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
