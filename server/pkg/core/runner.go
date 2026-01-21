package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Run executes the main linking task
func Run(opts Options, logger func(string, string)) (Stats, error) {
	stats := Stats{
		FailFiles: make(map[string][]string),
	}

	// Load Cache if enabled
	var cache *Cache
	if opts.OpenCache {
		cache = NewCache()
		cache.SetTaskName(opts.Name) // Set task name for cache isolation
		if logger != nil {
			logger("INFO", fmt.Sprintf("Cache enabled for task: %s", opts.Name))
		}
	} else {
		if logger != nil {
			logger("INFO", "Cache disabled")
		}
	}

	var newCachedFiles []string
	var mu sync.Mutex

	for src, dests := range opts.PathsMapping {
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Record error but continue
				mu.Lock()
				stats.FailFiles["Access Error"] = append(stats.FailFiles["Access Error"], path)
				stats.FailCount++
				mu.Unlock()
				return nil
			}

			if info.IsDir() {
				// Should we skip excluded directories to save time?
				// Simple doublestar match on dir path?
				// For now, continue walking.
				return nil
			}

			// Check Supported
			if !Supported(path, opts.Include, opts.Exclude) {
				if logger != nil {
					logger("DEBUG", fmt.Sprintf("File not supported by filters: %s", path))
				}
				return nil
			}

			// Check Cache
			if opts.OpenCache && cache != nil {
				has, _ := cache.Has(path, true) // Ignore case for safety
				if has {
					if logger != nil {
						logger("INFO", fmt.Sprintf("Skip cached file: %s", path))
					}
					return nil
				}
			}

			// Process Link
			// Handle multiple destinations
			var linkSuccess bool
			var anySuccess bool
			
			for _, dest := range dests {
				targetDir, err := GetOriginalDestPath(path, src, dest, opts.KeepDirStruct, opts.MkdirIfSingle)
				if err != nil {
					mu.Lock()
					stats.FailFiles["Path Calc Error"] = append(stats.FailFiles["Path Calc Error"], path)
					stats.FailCount++
					mu.Unlock()
					continue
				}

				targetFile, err := Link(path, targetDir)
				if err != nil {
					// Check if it's "file exists"
					if strings.Contains(err.Error(), "file exists") {
						if logger != nil {
							logger("INFO", fmt.Sprintf("File already exists, skipping: %s", targetFile))
						}
						linkSuccess = true // File exists, consider as processed
					} else {
						mu.Lock()
						stats.FailFiles[err.Error()] = append(stats.FailFiles[err.Error()], path+" -> "+targetDir)
						stats.FailCount++
						mu.Unlock()
						if logger != nil {
							logger("ERROR", fmt.Sprintf("Failed: %s -> %s (%v)", path, targetDir, err))
						}
						continue
					}
				} else {
					// Success
					// fmt.Printf("Linked: %s -> %s\n", path, targetFile)
					if logger != nil {
						logger("SUCCEED", fmt.Sprintf("Linked: %s -> %s", path, targetFile))
					}
					linkSuccess = true
					anySuccess = true
					if logger != nil {
						logger("DEBUG", fmt.Sprintf("Set linkSuccess=true for %s", path))
					}
				}
			}

			

			mu.Lock()
			if anySuccess || linkSuccess {
				stats.SuccessCount++
				// Add to cache if processing was successful
				if opts.OpenCache {
					newCachedFiles = append(newCachedFiles, path)
					if logger != nil {
						logger("DEBUG", fmt.Sprintf("Added to cache: %s (total: %d)", path, len(newCachedFiles)))
					}
				}
			} else {
				if logger != nil {
					logger("DEBUG", fmt.Sprintf("Not adding to cache: %s (anySuccess=%v, linkSuccess=%v)", path, anySuccess, linkSuccess))
				}
			}
			mu.Unlock()

			return nil
		})

		if err != nil {
			return stats, err
		}
	}

	// Update Cache
	if opts.OpenCache && cache != nil {
		if len(newCachedFiles) > 0 {
			if logger != nil {
				logger("INFO", fmt.Sprintf("Adding %d files to cache", len(newCachedFiles)))
			}
			_ = cache.Add(newCachedFiles)
		} else {
			if logger != nil {
				logger("INFO", "No new files to add to cache")
			}
		}
	}

	return stats, nil
}
