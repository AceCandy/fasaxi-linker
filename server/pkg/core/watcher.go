package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher manages file watching
type Watcher struct {
	internal *fsnotify.Watcher
	options  Options
	done     chan bool
	logger   func(string, string) // type, message
	mu       sync.Mutex
	isClosed bool
}

// NewWatcher creates a watcher
func NewWatcher(opts Options, logger func(string, string)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		internal: w,
		options:  opts,
		done:     make(chan bool),
		logger:   logger,
	}, nil
}

// Start begins watching
func (w *Watcher) Start() error {
	taskName := w.options.Name
	w.logger("INFO", fmt.Sprintf("🚀 [%s] 监听服务启动中...", taskName))

	// Collect path errors by category
	var failedSrcs []string
	var failedDests []string
	var otherErrors []string
	failedCount := 0

	// Add paths
	for src, dests := range w.options.PathsMapping {
		// 1. Check source existence
		if _, err := os.Stat(src); err != nil {
			failedSrcs = append(failedSrcs, src)
			w.logger("ERROR", fmt.Sprintf("❌ 源路径不存在: %s", src))
			failedCount++
			continue
		}

		// 2. Check destination(s) existence
		destOk := true
		for _, dest := range dests {
			if _, err := os.Stat(dest); err != nil {
				failedDests = append(failedDests, dest)
				w.logger("ERROR", fmt.Sprintf("❌ 目标路径不存在: %s", dest))
				destOk = false
				break
			}
		}
		if !destOk {
			failedCount++
			continue
		}

		// 3. Add to watcher
		if err := w.AddRecursive(src); err != nil {
			otherErrors = append(otherErrors, fmt.Sprintf("%s: %v", src, err))
			w.logger("ERROR", fmt.Sprintf("❌ 无法监听路径 %s: %v", src, err))
			failedCount++
		} else {
			w.logger("INFO", fmt.Sprintf("🩺 路径[%s] => %v 正在监听中...", src, dests))
		}
	}

	// If any path failed, return formatted error
	if failedCount > 0 {
		var msgs []string
		if len(failedSrcs) > 0 {
			msgs = append(msgs, fmt.Sprintf("源路径无法监听 (%s)", strings.Join(failedSrcs, "、")))
		}
		if len(failedDests) > 0 {
			msgs = append(msgs, fmt.Sprintf("目标路径不存在:(%s)", strings.Join(failedDests, "、")))
		}
		if len(otherErrors) > 0 {
			msgs = append(msgs, fmt.Sprintf("其他异常: %s", strings.Join(otherErrors, "; ")))
		}
		return fmt.Errorf("监听失败: %s", strings.Join(msgs, "; "))
	}

	go w.eventLoop()
	w.logger("INFO", fmt.Sprintf("✅ [%s] 监听服务已就绪", taskName))
	return nil
}

// Stop stops watching
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.isClosed {
		w.isClosed = true
		w.internal.Close()
		close(w.done)
	}
}

func (w *Watcher) eventLoop() {
	// Debounce mechanics could be added here
	for {
		select {
		case event, ok := <-w.internal.Events:
			if !ok {
				return
			}

			// Handle Create / Write / Rename
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) != 0 {
				w.handleAdd(event.Name)
			}

			// If directory created, watch it
			if event.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.AddRecursive(event.Name)
				}
			}

		case err, ok := <-w.internal.Errors:
			if !ok {
				return
			}
			w.logger("ERROR", fmt.Sprintf("❌ 监听错误: %v", err))
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) AddRecursive(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip excluded directories early if possible?
			// fsnotify needs to watch dir to see events inside.
			return w.internal.Add(p)
		}
		return nil
	})
}

func (w *Watcher) handleAdd(path string) {
	// Ignore if directory (logic focuses on files)
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return
	}

	// Check Supported
	supported := Supported(path, w.options.Include, w.options.Exclude)
	if !supported {
		return
	}

	// Check Cache
	if w.options.OpenCache {
		// Get cache path
		hlinkHome := os.Getenv("HLINK_HOME")
		if hlinkHome == "" {
			homeDir, _ := os.UserHomeDir()
			hlinkHome = filepath.Join(homeDir, ".hlink")
		}
		cachePath := filepath.Join(hlinkHome, "cache-array.json")
		cache := NewCache(cachePath)

		// Check if file is already in cache
		if has, _ := cache.Has(path, true); has {
			w.logger("WARN", fmt.Sprintf("⚠️ 跳过(已缓存): %s", path))
			return
		}
	}

	// Find Source Root for this file
	var sourceRoot string
	for src := range w.options.PathsMapping {
		if strings.HasPrefix(path, src) {
			sourceRoot = src
			break
		}
	}
	if sourceRoot == "" {
		return
	}

	dests := w.options.PathsMapping[sourceRoot]
	for _, dest := range dests {
		targetDir, err := GetOriginalDestPath(path, sourceRoot, dest, w.options.KeepDirStruct, w.options.MkdirIfSingle)
		if err != nil {
			w.logger("ERROR", fmt.Sprintf("❌ 计算目标路径失败: %v", err))
			continue
		}

		finalTarget, err := Link(path, targetDir)
		linkSuccess := true
		if err != nil {
			if strings.Contains(err.Error(), "file exists") {
				// File already exists, but should still be added to cache
				w.logger("WARN", fmt.Sprintf("⚠️ 文件已存在: %s → %s", path, finalTarget))
			} else {
				w.logger("ERROR", fmt.Sprintf("❌ 硬链失败: %v", err))
				linkSuccess = false
			}
		} else {
			w.logger("SUCCEED", fmt.Sprintf("✅ 硬链成功: %s → %s", path, finalTarget))
		}

		// Add to cache if enabled and processing was successful or file exists
		if w.options.OpenCache && linkSuccess {
			// Get cache path
			hlinkHome := os.Getenv("HLINK_HOME")
			if hlinkHome == "" {
				homeDir, _ := os.UserHomeDir()
				hlinkHome = filepath.Join(homeDir, ".hlink")
			}
			cachePath := filepath.Join(hlinkHome, "cache-array.json")
			cache := NewCache(cachePath)

			// Add file to cache
			if err := cache.Add([]string{path}); err != nil {
				w.logger("ERROR", fmt.Sprintf("❌ 写入缓存失败: %v", err))
			} else {
				w.logger("INFO", fmt.Sprintf("💾 已加入缓存: %s", path))
			}
		}
	}
}
