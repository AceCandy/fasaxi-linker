package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fasaxi-linker/servergo/pkg/core"
)

type Service struct {
	store    *Store
	tasks    []Task
	tasksMap map[string]Task
	mu       sync.RWMutex
	configs  []Config

	// Configs (cached but managed by config service usually, or we share store?)
	// Ideally we separate ConfigService, but Store is shared.
	// For now, Task Service only touches Tasks.
	// PROBLEM: Store.Save now needs Configs.
	// We need to keep Configs in memory in TaskService or fetch them.
	// Easier: Just load fresh every time we save? Or cache.
	watchers map[string]*core.Watcher
	wMu      sync.RWMutex
}

func NewService() (*Service, error) {
	store := NewStore()
	tasks, configs, err := store.Load()
	if err != nil {
		return nil, err
	}

	s := &Service{
		store:    store,
		tasks:    tasks,
		configs:  configs,
		tasksMap: make(map[string]Task),
		watchers: make(map[string]*core.Watcher),
	}
	s.rebuildMap()

	// Restore watching state
	if err := s.restoreWatchState(); err != nil {
		// Log warning but don't fail startup
		fmt.Printf("Warning: RESTORE WATCH STATE FAILED: %v\n", err)
	}

	return s, nil
}

func (s *Service) rebuildMap() {
	s.tasksMap = make(map[string]Task)
	for _, t := range s.tasks {
		s.tasksMap[t.Name] = t
	}
}

func (s *Service) GetAll() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Task, len(s.tasks))
	copy(result, s.tasks)
	return result
}

func (s *Service) Get(name string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasksMap[name]
	return t, ok
}

// GetConfigByName returns the configuration by name
func (s *Service) GetConfigByName(name string) (Config, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, config := range s.configs {
		if config.Name == name {
			return config, true
		}
	}
	return Config{}, false
}

// Helper to save tasks preserving configs
func (s *Service) saveTasks() error {
	// Re-read configs to avoid overwriting changes from ConfigService
	_, configs, err := s.store.Load()
	if err == nil {
		s.configs = configs // update local configs
	}
	return s.store.Save(s.tasks, s.configs)
}

func (s *Service) Add(t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasksMap[t.Name]; ok {
		return fmt.Errorf("task %s already exists", t.Name)
	}

	s.tasks = append(s.tasks, t)
	s.rebuildMap()
	return s.saveTasks()
}

func (s *Service) Update(prevName string, t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasksMap[prevName]; !ok {
		return fmt.Errorf("task %s does not exist", prevName)
	}

	for i, existing := range s.tasks {
		if existing.Name == prevName {
			s.tasks[i] = t
			break
		}
	}
	s.rebuildMap()
	return s.saveTasks()
}

func (s *Service) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	var newTasks []Task
	for _, t := range s.tasks {
		if t.Name == name {
			found = true
			continue
		}
		newTasks = append(newTasks, t)
	}

	if !found {
		return fmt.Errorf("task %s not found", name)
	}

	s.tasks = newTasks
	s.rebuildMap()
	return s.saveTasks()
}

// Watcher methods
func (s *Service) StartWatch(name string, logger func(string, string)) error {
	s.mu.RLock()
	task, ok := s.tasksMap[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("task not found")
	}

	s.wMu.Lock()
	defer s.wMu.Unlock()

	if _, ok := s.watchers[name]; ok {
		return nil // Already watching
	}

	opts := s.getTaskOptions(task)

	w, err := core.NewWatcher(opts, logger)
	if err != nil {
		return err
	}

	if err := w.Start(); err != nil {
		return err
	}

	s.watchers[name] = w

	// Save watch state
	go func() {
		if err := s.saveWatchState(); err != nil {
			fmt.Printf("⚠️ 保存监听状态失败: %v\n", err)
		}
	}()

	return nil
}

func (s *Service) StartWatchWithOptions(name string, logger func(string, string), opts core.Options) error {
	s.mu.RLock()
	_, ok := s.tasksMap[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("task not found")
	}

	s.wMu.Lock()
	defer s.wMu.Unlock()

	if _, ok := s.watchers[name]; ok {
		return nil // Already watching
	}

	w, err := core.NewWatcher(opts, logger)
	if err != nil {
		// Update task state: set error message
		if task, ok := s.Get(name); ok {
			task.IsWatching = false
			task.WatchError = err.Error()
			s.Update(name, task)
		}
		return err
	}

	if err := w.Start(); err != nil {
		// Update task state: set error message
		if task, ok := s.Get(name); ok {
			task.IsWatching = false
			task.WatchError = err.Error()
			s.Update(name, task)
		}
		return err
	}

	s.watchers[name] = w

	// Success: clear error and set watching state
	if task, ok := s.Get(name); ok {
		task.IsWatching = true
		task.WatchError = ""
		s.Update(name, task)
	}

	// Save watch state
	go func() {
		if err := s.saveWatchState(); err != nil {
			fmt.Printf("⚠️ 保存监听状态失败: %v\n", err)
		}
	}()

	return nil
}

func (s *Service) StopWatch(name string) error {
	s.wMu.Lock()
	defer s.wMu.Unlock()

	w, ok := s.watchers[name]
	if !ok {
		return nil
	}

	// Stop watcher in a goroutine to avoid blocking
	go w.Stop()
	delete(s.watchers, name)

	// Clear watch state and error
	if task, ok := s.Get(name); ok {
		task.IsWatching = false
		task.WatchError = ""
		s.Update(name, task)
	}

	// Save watch state in a goroutine to avoid blocking
	go func() {
		if err := s.saveWatchState(); err != nil {
			fmt.Printf("⚠️ 保存监听状态失败: %v\n", err)
		}
	}()

	return nil
}

func (s *Service) IsWatching(name string) bool {
	s.wMu.RLock()
	defer s.wMu.RUnlock()
	_, ok := s.watchers[name]
	return ok
}

// saveWatchState saves the current watching state to disk
func (s *Service) saveWatchState() error {
	s.wMu.RLock()
	watchingTasks := make([]string, 0, len(s.watchers))
	for name := range s.watchers {
		watchingTasks = append(watchingTasks, name)
	}
	s.wMu.RUnlock()

	// Get hlink home directory
	hlinkHome := os.Getenv("HLINK_HOME")
	if hlinkHome == "" {
		homeDir, _ := os.UserHomeDir()
		hlinkHome = filepath.Join(homeDir, ".hlink")
	}
	watchStateFile := filepath.Join(hlinkHome, "watch-state.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(watchStateFile), 0755); err != nil {
		return fmt.Errorf("failed to create watch state directory: %v", err)
	}

	// Write watching tasks to file
	data, err := json.MarshalIndent(watchingTasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal watch state: %v", err)
	}

	return os.WriteFile(watchStateFile, data, 0644)
}

// restoreWatchState restores watching state from disk and starts watchers
func (s *Service) restoreWatchState() error {
	// Get hlink home directory
	hlinkHome := os.Getenv("HLINK_HOME")
	if hlinkHome == "" {
		homeDir, _ := os.UserHomeDir()
		hlinkHome = filepath.Join(homeDir, ".hlink")
	}
	watchStateFile := filepath.Join(hlinkHome, "watch-state.json")

	// Check if watch state file exists
	if _, err := os.Stat(watchStateFile); os.IsNotExist(err) {
		// No saved state, that's fine
		return nil
	}

	// Read watch state file
	data, err := os.ReadFile(watchStateFile)
	if err != nil {
		return fmt.Errorf("failed to read watch state file: %v", err)
	}

	var watchingTasks []string
	if err := json.Unmarshal(data, &watchingTasks); err != nil {
		return fmt.Errorf("failed to unmarshal watch state: %v", err)
	}

	// Restart watching for each task
	for _, taskName := range watchingTasks {
		// Check if task still exists
		s.mu.RLock()
		task, ok := s.tasksMap[taskName]
		s.mu.RUnlock()

		if !ok {
			// Task no longer exists, skip
			continue
		}

		// Start watching
		s.wMu.Lock()
		if _, alreadyWatching := s.watchers[taskName]; !alreadyWatching {
			opts := s.getTaskOptions(task)
			logger := GetLogger(taskName)
			w, err := core.NewWatcher(opts, logger)
			if err != nil {
				fmt.Printf("❌ 创建任务监听器失败 %s: %v\n", taskName, err)
				s.wMu.Unlock()
				continue
			}

			if err := w.Start(); err != nil {
				fmt.Printf("❌ 启动任务监听失败 %s: %v\n", taskName, err)
				w.Stop()
				delete(s.watchers, taskName)
				s.wMu.Unlock()

				// Update task state: set error message
				if task, ok := s.Get(taskName); ok {
					task.IsWatching = false
					task.WatchError = err.Error()
					s.Update(taskName, task)
				}
				continue
			}

			s.watchers[taskName] = w
			s.wMu.Unlock()

			// Success: clear error and set watching state
			if task, ok := s.Get(taskName); ok {
				task.IsWatching = true
				task.WatchError = ""
				s.Update(taskName, task)
			}

			fmt.Printf("✅ 已恢复任务监听: %s\n", taskName)
		}
	}

	return nil
}

// SyncConfigToTasks syncs configuration fields to all tasks that use the given config ID
func (s *Service) SyncConfigToTasks(configID int, configName string, configDetail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse config detail
	var rc RuntimeConfig
	if err := json.Unmarshal([]byte(configDetail), &rc); err != nil {
		return fmt.Errorf("failed to parse config detail: %v", err)
	}

	// Find all tasks that use this config
	updated := false
	for i, task := range s.tasks {
		if task.ConfigID == configID {
			// Sync config name and fields to task
			s.tasks[i].Config = configName
			s.tasks[i].Include = rc.Include
			s.tasks[i].Exclude = rc.Exclude
			s.tasks[i].KeepDirStruct = rc.KeepDirStruct
			s.tasks[i].OpenCache = rc.OpenCache
			s.tasks[i].MkdirIfSingle = rc.MkdirIfSingle
			s.tasks[i].DeleteDir = rc.DeleteDir
			updated = true
		}
	}

	if updated {
		s.rebuildMap()
		return s.saveTasks()
	}

	return nil
}
