package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	FilePath string
	mu       sync.RWMutex
}

var (
	storeInstance *Store
	storeOnce     sync.Once
)

// GetSharedStore returns a singleton instance of Store
func GetSharedStore() *Store {
	storeOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		dbDir := filepath.Join(homeDir, ".hlink") // Match hlinkHomeDir
		_ = os.MkdirAll(dbDir, 0755)

		// Use db.json like the TS version
		storeInstance = &Store{
			FilePath: filepath.Join(dbDir, "db.json"),
		}
	})
	return storeInstance
}

func NewStore() *Store {
	return GetSharedStore()
}

// Config represents the configuration with metadata and content
type Config struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Detail      string `json:"detail"` // Direct storage of config content instead of external file path
}

type DBWrapper struct {
	Tasks   []Task   `json:"tasks"`
	Configs []Config `json:"configs"`
}

func (s *Store) Load() ([]Task, []Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.FilePath)
	if os.IsNotExist(err) {
		return []Task{}, []Config{}, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var wrapper DBWrapper
	// If it fails to unmarshal as wrapper, maybe it's array? (No, assuming DB format)
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// Fallback: try array if user started fresh with my previous code?
		// But I haven't released yet, so assume db.json is standard.
		return []Task{}, []Config{}, nil
	}
	return wrapper.Tasks, wrapper.Configs, nil
}

func (s *Store) Save(tasks []Task, configs []Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wrapper := DBWrapper{
		Tasks:   tasks,
		Configs: configs,
	}

	updatedData, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.FilePath, updatedData, 0644)
}
