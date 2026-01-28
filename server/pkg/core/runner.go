package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Run executes the main linking task with concurrent processing
func Run(opts Options, logger func(string, string)) (Stats, error) {
	fmt.Println("DEBUG: Run() started")

	stats := Stats{
		FailFiles: make(map[string][]string),
	}

	// Load Cache if enabled
	var cache *Cache
	if opts.OpenCache {
		cache = NewCache()
		cache.SetTaskID(opts.TaskID)
		fmt.Printf("DEBUG: Cache ENABLED for task %s (ID=%d)\n", opts.Name, opts.TaskID)
	} else {
		fmt.Printf("DEBUG: Cache DISABLED for task %s (ID=%d)\n", opts.Name, opts.TaskID)
	}

	var newCachedFiles []string
	var mu sync.Mutex

	// Collect all files first
	fmt.Println("DEBUG: Starting file collection...")
	var allFiles []fileJob
	fileCount := 0

	for src, dests := range opts.PathsMapping {
		fmt.Printf("DEBUG: Walking source: %s\n", src)
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if info.IsDir() {
				return nil
			}

			fileCount++
			if fileCount%1000 == 0 {
				fmt.Printf("DEBUG: Scanned %d files...\n", fileCount)
			}

			// Check Supported
			if !Supported(path, opts.Include, opts.Exclude) {
				return nil
			}

			// Check Cache (skip for now to speed up)
			if opts.OpenCache && cache != nil {
				has, _ := cache.Has(path)
				if has {
					if logger != nil {
						logger("WARN", fmt.Sprintf("⚠️ 跳过(已缓存): %s", path))
					}
					return nil
				}
			}

			allFiles = append(allFiles, fileJob{
				path:  path,
				src:   src,
				dests: dests,
			})

			return nil
		})

		if err != nil {
			fmt.Printf("DEBUG: Walk error: %v\n", err)
			return stats, err
		}
	}

	fmt.Printf("DEBUG: Collected %d files to process\n", len(allFiles))

	if len(allFiles) == 0 {
		fmt.Println("DEBUG: No files to process")
		return stats, nil
	}

	// Process files concurrently
	numWorkers := runtime.NumCPU() * 2
	if numWorkers > 16 {
		numWorkers = 16
	}
	if len(allFiles) < numWorkers {
		numWorkers = len(allFiles)
	}

	fmt.Printf("DEBUG: Starting %d workers for %d files\n", numWorkers, len(allFiles))

	jobs := make(chan fileJob, len(allFiles))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				processFile(job, opts, cache, logger, &stats, &newCachedFiles, &mu)
			}
		}(i)
	}

	// Send jobs
	for _, job := range allFiles {
		jobs <- job
	}
	close(jobs)

	fmt.Println("DEBUG: Waiting for workers to complete...")
	wg.Wait()
	fmt.Println("DEBUG: All workers completed")

	// Update Cache
	if opts.OpenCache && cache != nil && len(newCachedFiles) > 0 {
		if logger != nil {
			logger("INFO", fmt.Sprintf("💾 已加入缓存: %d 个文件", len(newCachedFiles)))
		}
		_ = cache.Add(newCachedFiles)
	}

	fmt.Printf("DEBUG: Run() completed. Success: %d, Fail: %d\n", stats.SuccessCount, stats.FailCount)
	return stats, nil
}

type fileJob struct {
	path  string
	src   string
	dests []string
}

func processFile(job fileJob, opts Options, cache *Cache, logger func(string, string), stats *Stats, newCachedFiles *[]string, mu *sync.Mutex) {
	var linkSuccess bool
	var anySuccess bool

	for _, dest := range job.dests {
		targetDir, err := GetOriginalDestPath(job.path, job.src, dest, opts.KeepDirStruct, opts.MkdirIfSingle)
		if err != nil {
			mu.Lock()
			stats.FailFiles["Path Calc Error"] = append(stats.FailFiles["Path Calc Error"], job.path)
			stats.FailCount++
			mu.Unlock()
			continue
		}

		targetFile, err := Link(job.path, targetDir)
		if err != nil {
			if strings.Contains(err.Error(), "file exists") {
				if logger != nil {
					logger("WARN", fmt.Sprintf("⚠️ 文件已存在: %s → %s", job.path, targetFile))
				}
				linkSuccess = true
			} else {
				mu.Lock()
				stats.FailFiles[err.Error()] = append(stats.FailFiles[err.Error()], job.path+" -> "+targetDir)
				stats.FailCount++
				mu.Unlock()
				if logger != nil {
					logger("ERROR", fmt.Sprintf("❌ 硬链失败: %s → %s (%v)", job.path, targetDir, err))
				}
				continue
			}
		} else {
			if logger != nil {
				logger("SUCCEED", fmt.Sprintf("✅ 硬链成功: %s → %s", job.path, targetFile))
			}
			linkSuccess = true
			anySuccess = true
		}
	}

	mu.Lock()
	if anySuccess || linkSuccess {
		stats.SuccessCount++
		if opts.OpenCache {
			*newCachedFiles = append(*newCachedFiles, job.path)
		}
	}
	mu.Unlock()
}
