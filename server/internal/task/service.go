package task

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fasaxi-linker/servergo/pkg/core"
)

type Service struct {
	store    *Store
	tasks    []Task
	tasksMap map[string]Task
	mu       sync.RWMutex
	// configs  []Config // Removed: redundant cache

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
	tasks, _, err := store.Load() // Ignore configs
	if err != nil {
		return nil, err
	}

	s := &Service{
		store:    store,
		tasks:    tasks,
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

func (s *Service) Add(t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasksMap[t.Name]; ok {
		return fmt.Errorf("task %s already exists", t.Name)
	}

	// 使用单任务插入，获取数据库分配的 ID
	id, err := s.store.AddTask(t)
	if err != nil {
		return err
	}
	t.ID = id

	s.tasks = append(s.tasks, t)
	s.rebuildMap()
	return nil
}

func (s *Service) Update(prevName string, t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tasksMap[prevName]
	if !ok {
		return fmt.Errorf("task %s does not exist", prevName)
	}

	// 保留原任务的 ID，确保更新时使用相同的 ID
	t.ID = existing.ID

	// 使用单任务更新
	if err := s.store.UpdateTask(t); err != nil {
		return err
	}

	for i, ex := range s.tasks {
		if ex.Name == prevName {
			s.tasks[i] = t
			break
		}
	}
	s.rebuildMap()
	return nil
}

func (s *Service) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tasksMap[name]
	if !ok {
		return fmt.Errorf("task %s not found", name)
	}

	// 使用单任务删除
	if err := s.store.DeleteTask(existing.ID); err != nil {
		return err
	}

	var newTasks []Task
	for _, t := range s.tasks {
		if t.Name == name {
			continue
		}
		newTasks = append(newTasks, t)
	}

	s.tasks = newTasks
	s.rebuildMap()
	return nil
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

	// Success: update watching state in DB
	if task, ok := s.Get(name); ok {
		task.IsWatching = true
		task.WatchError = ""
		s.Update(name, task)
	}

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

	return nil
}

func (s *Service) IsWatching(name string) bool {
	s.wMu.RLock()
	defer s.wMu.RUnlock()
	_, ok := s.watchers[name]
	return ok
}

// restoreWatchState restores watching state from DB and starts watchers
func (s *Service) restoreWatchState() error {
	s.mu.RLock()
	// Make a copy to avoid holding lock while starting watchers
	tasksCopy := make([]Task, len(s.tasks))
	copy(tasksCopy, s.tasks)
	s.mu.RUnlock()

	for _, task := range tasksCopy {
		if task.IsWatching {
			// Try to start watcher
			// We use StartWatch logic but we need to handle failure carefully
			// Since IsWatching is already true in DB, we don't need to update it if success.
			// But if failure, we MUST update it to false.

			// Shortcut: directly recreate watcher to avoid extra DB update on success,
			// but update DB on failure.

			logger := GetLogger(task.Name)
			opts := s.getTaskOptions(task)

			s.wMu.Lock()
			if _, ok := s.watchers[task.Name]; ok {
				s.wMu.Unlock()
				continue
			}

			w, err := core.NewWatcher(opts, logger)
			if err != nil {
				fmt.Printf("❌ [Restore] 创建监听器失败 %s: %v\n", task.Name, err)
				s.wMu.Unlock()
				// Update DB to reflect failure
				task.IsWatching = false
				task.WatchError = err.Error()
				s.Update(task.Name, task)
				continue
			}

			if err := w.Start(); err != nil {
				fmt.Printf("❌ [Restore] 启动监听失败 %s: %v\n", task.Name, err)
				w.Stop()
				s.wMu.Unlock()
				// Update DB to reflect failure
				task.IsWatching = false
				task.WatchError = err.Error()
				s.Update(task.Name, task)
				continue
			}

			s.watchers[task.Name] = w
			s.wMu.Unlock()

			fmt.Printf("✅ [Restore] 已恢复监听: %s\n", task.Name)
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
		// Save updated tasks one by one to avoid full delete
		for _, task := range s.tasks {
			if task.ConfigID == configID {
				if err := s.store.UpdateTask(task); err != nil {
					return fmt.Errorf("failed to update task %s: %w", task.Name, err)
				}
			}
		}
		return nil
	}

	return nil
}
