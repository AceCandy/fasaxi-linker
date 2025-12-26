package core

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// Cache manages the list of processed files to avoid duplicates
type Cache struct {
	FilePath string
	mu       sync.RWMutex
}

// NewCache creates a new Cache instance
func NewCache(filePath string) *Cache {
	return &Cache{
		FilePath: filePath,
	}
}

// Read reads the cache from file
func (c *Cache) Read() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.FilePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var files []string
	if err := json.Unmarshal(data, &files); err != nil {
		return []string{}, nil // Return empty on error or handle? JS seems permissive.
	}
	return files, nil
}

// Write writes the cache to file
func (c *Cache) Write(files []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Make unique
	unique := make(map[string]bool)
	var result []string
	for _, f := range files {
		if !unique[f] {
			unique[f] = true
			result = append(result, f)
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.FilePath, data, 0644)
}

// Add adds new files to cache
func (c *Cache) Add(newFiles []string) error {
	current, err := c.Read()
	if err != nil {
		return err
	}
	return c.Write(append(current, newFiles...))
}

// Has checks if a file is in cache
func (c *Cache) Has(file string, ignoreCase bool) (bool, error) {
	current, err := c.Read()
	if err != nil {
		return false, err
	}

	if ignoreCase {
		file = strings.ToLower(file)
		for _, f := range current {
			if strings.ToLower(f) == file {
				return true, nil
			}
		}
	} else {
		for _, f := range current {
			if f == file {
				return true, nil
			}
		}
	}
	return false, nil
}
