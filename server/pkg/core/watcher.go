package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

// Watcher manages file watching using rjeczalik/notify
// This library provides native recursive watching support:
// - macOS: FSEvents (very fast, no per-directory overhead)
// - Linux: inotify with automatic recursive handling
// - Windows: ReadDirectoryChangesW
type Watcher struct {
	events   chan notify.EventInfo
	options  Options
	done     chan bool
	logger   func(string, string) // type, message
	mu       sync.Mutex
	isClosed bool
	memCache sync.Map // L1 Memory Cache
}

// NewWatcher creates a watcher
func NewWatcher(opts Options, logger func(string, string)) (*Watcher, error) {
	// Create a buffered channel for events
	events := make(chan notify.EventInfo, 1000)

	return &Watcher{
		events:  events,
		options: opts,
		done:    make(chan bool),
		logger:  logger,
	}, nil
}

// Start begins watching (async initialization for large directories)
func (w *Watcher) Start() error {
	taskName := w.options.Name
	w.logger("INFO", fmt.Sprintf("🚀 [%s] 监听服务启动中...", taskName))

	// Quick validation first (synchronous)
	var validPaths []struct {
		src   string
		dests []string
	}

	for src, dests := range w.options.PathsMapping {
		// 1. Check source existence
		if _, err := os.Stat(src); err != nil {
			w.logger("ERROR", fmt.Sprintf("❌ 源路径不存在: %s", src))
			continue
		}

		// 2. Check destination(s) existence
		destOk := true
		for _, dest := range dests {
			if _, err := os.Stat(dest); err != nil {
				w.logger("ERROR", fmt.Sprintf("❌ 目标路径不存在: %s", dest))
				destOk = false
				break
			}
		}
		if destOk {
			validPaths = append(validPaths, struct {
				src   string
				dests []string
			}{src, dests})
		}
	}

	if len(validPaths) == 0 {
		return fmt.Errorf("没有可用的监听路径")
	}

	// Start event loop immediately (we'll receive events as watchers are added)
	go w.eventLoop()

	// ASYNC: Add watchers in background
	go func() {
		startTime := time.Now()
		for _, p := range validPaths {
			watchPath := filepath.Join(p.src, "...")
			if err := notify.Watch(watchPath, w.events, notify.Create, notify.Write, notify.Rename); err != nil {
				w.logger("ERROR", fmt.Sprintf("❌ 无法监听路径 %s: %v", p.src, err))
			} else {
				w.logger("INFO", fmt.Sprintf("🩺 路径[%s] => %v 已就绪", p.src, p.dests))
			}
		}
		elapsed := time.Since(startTime)
		w.logger("INFO", fmt.Sprintf("✅ [%s] 所有路径监听就绪 (耗时 %.1f 秒)", taskName, elapsed.Seconds()))
	}()

	// Return immediately - service is ready to receive events
	w.logger("INFO", fmt.Sprintf("✅ [%s] 监听服务已就绪 (后台初始化中...)", taskName))
	return nil
}

// Stop stops watching
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.isClosed {
		w.isClosed = true
		notify.Stop(w.events)
		close(w.done)
	}
}

func (w *Watcher) eventLoop() {
	// Debounce buffer
	pendingEvents := make(map[string]struct{})
	var debounceTimer *time.Timer
	const debounceInterval = 500 * time.Millisecond

	processEvents := func() {
		w.mu.Lock()
		paths := make([]string, 0, len(pendingEvents))
		for p := range pendingEvents {
			paths = append(paths, p)
		}
		// Clear map
		pendingEvents = make(map[string]struct{})
		w.mu.Unlock()

		for _, p := range paths {
			w.handleAdd(p)
		}
	}

	for {
		select {
		case event, ok := <-w.events:
			if !ok {
				return
			}

			path := event.Path()

			// Buffer all file events for debouncing
			w.mu.Lock()
			pendingEvents[path] = struct{}{}
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceInterval, processEvents)
			w.mu.Unlock()

		case <-w.done:
			w.mu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			w.mu.Unlock()
			return
		}
	}
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
		// 1. L1 Memory Cache Check
		if _, ok := w.memCache.Load(path); ok {
			return
		}

		cache := NewCache()
		cache.SetTaskID(w.options.TaskID) // Set task ID for cache isolation

		// 2. L2 DB Cache Check
		if has, _ := cache.Has(path, true); has {
			w.logger("WARN", fmt.Sprintf("⚠️ 跳过(已缓存): %s", path))
			// Add to memory cache
			w.memCache.Store(path, struct{}{})
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
			cache := NewCache()
			cache.SetTaskID(w.options.TaskID) // Set task ID for cache isolation

			// Add file to cache
			if err := cache.Add([]string{path}); err != nil {
				w.logger("ERROR", fmt.Sprintf("❌ 写入缓存失败: %v", err))
			} else {
				w.logger("INFO", fmt.Sprintf("💾 已加入缓存: %s", path))
				// Add to memory cache
				w.memCache.Store(path, struct{}{})
			}
		}
	}
}

// RemoveFromCache removes specific files from memory cache
func (w *Watcher) RemoveFromCache(files []string) {
	for _, f := range files {
		w.memCache.Delete(f)
	}
}

// ClearCache clears the entire memory cache
func (w *Watcher) ClearCache() {
	w.memCache.Range(func(key, value interface{}) bool {
		w.memCache.Delete(key)
		return true
	})
}
