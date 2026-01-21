package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fasaxi-linker/servergo/internal/db"
	"github.com/jackc/pgx/v5"
)

type Store struct {
	mu sync.RWMutex
}

var (
	storeInstance *Store
	storeOnce     sync.Once
)

func GetSharedStore() *Store {
	storeOnce.Do(func() {
		storeInstance = &Store{}
	})
	return storeInstance
}

func NewStore() *Store {
	return GetSharedStore()
}

type Config struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

type DBWrapper struct {
	Tasks   []Task   `json:"tasks"`
	Configs []Config `json:"configs"`
}

func (s *Store) Load() ([]Task, []Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return nil, nil, fmt.Errorf("database connection pool is not initialized")
	}

	tasks, err := s.loadTasks(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	configs, err := s.loadConfigs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configs: %w", err)
	}

	return tasks, configs, nil
}

func (s *Store) loadTasks(ctx context.Context) ([]Task, error) {
	pool := db.GetPool()

	query := `
		SELECT id, name, type, paths_mapping, include_patterns, exclude_patterns,
		       save_mode, open_cache, mkdir_if_single, delete_dir, keep_dir_struct,
		       schedule_type, schedule_value, reverse, config, config_id, is_watching, watch_error
		FROM tasks
		ORDER BY name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var pathsMappingJSON, includeJSON, excludeJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Type, &pathsMappingJSON, &includeJSON, &excludeJSON,
			&t.SaveMode, &t.OpenCache, &t.MkdirIfSingle, &t.DeleteDir, &t.KeepDirStruct,
			&t.ScheduleType, &t.ScheduleValue, &t.Reverse, &t.Config, &t.ConfigID, &t.IsWatching, &t.WatchError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		if err := json.Unmarshal(pathsMappingJSON, &t.PathsMapping); err != nil {
			return nil, fmt.Errorf("failed to unmarshal paths_mapping: %w", err)
		}

		if err := json.Unmarshal(includeJSON, &t.Include); err != nil {
			return nil, fmt.Errorf("failed to unmarshal include_patterns: %w", err)
		}

		if err := json.Unmarshal(excludeJSON, &t.Exclude); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exclude_patterns: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Store) loadConfigs(ctx context.Context) ([]Config, error) {
	pool := db.GetPool()

	query := `
		SELECT id, name, detail
		FROM configs
		ORDER BY name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []Config
	for rows.Next() {
		var c Config
		var detailJSON []byte

		err := rows.Scan(&c.ID, &c.Name, &detailJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config row: %w", err)
		}

		c.Detail = string(detailJSON)
		configs = append(configs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return configs, nil
}

// GetTaskIDByName retrieves the task ID for a given task name
func (s *Store) GetTaskIDByName(taskName string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return 0, fmt.Errorf("database connection pool is not initialized")
	}

	var taskID int
	query := `SELECT id FROM tasks WHERE name = $1`
	err := pool.QueryRow(ctx, query, taskName).Scan(&taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to get task ID for name %s: %w", taskName, err)
	}

	return taskID, nil
}

func (s *Store) Save(tasks []Task, configs []Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM tasks"); err != nil {
		return fmt.Errorf("failed to delete existing tasks: %w", err)
	}

	if _, err := tx.Exec(ctx, "DELETE FROM configs"); err != nil {
		return fmt.Errorf("failed to delete existing configs: %w", err)
	}

	// Insert configs first to resolve IDs
	configIDByName := make(map[string]int)
	configNameByID := make(map[int]string)
	for i := range configs {
		if err := s.insertConfig(ctx, tx, &configs[i]); err != nil {
			return fmt.Errorf("failed to insert config %s: %w", configs[i].Name, err)
		}
		configIDByName[configs[i].Name] = configs[i].ID
		configNameByID[configs[i].ID] = configs[i].Name
	}

	// Ensure sequence is up-to-date after manual ID inserts
	if _, err := tx.Exec(ctx, "SELECT setval(pg_get_serial_sequence('configs', 'id'), COALESCE((SELECT MAX(id) FROM configs), 0))"); err != nil {
		return fmt.Errorf("failed to sync configs sequence: %w", err)
	}

	// Insert tasks with resolved config IDs/names
	for i := range tasks {
		configID := tasks[i].ConfigID
		configName := tasks[i].Config

		if configID == 0 && tasks[i].Config != "" {
			if id, ok := configIDByName[tasks[i].Config]; ok {
				configID = id
			}
		}

		if configName == "" && configID != 0 {
			if name, ok := configNameByID[configID]; ok {
				configName = name
			}
		}

		if configID != 0 {
			tasks[i].ConfigID = configID
		}
		if configName != "" {
			tasks[i].Config = configName
		}

		if err := s.insertTask(ctx, tx, tasks[i], configID, configName); err != nil {
			return fmt.Errorf("failed to insert task %s: %w", tasks[i].Name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Store) insertTask(ctx context.Context, tx pgx.Tx, t Task, configID int, configName string) error {
	pathsMappingJSON, err := json.Marshal(t.PathsMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal paths_mapping: %w", err)
	}

	includeJSON, err := json.Marshal(t.Include)
	if err != nil {
		return fmt.Errorf("failed to marshal include: %w", err)
	}

	excludeJSON, err := json.Marshal(t.Exclude)
	if err != nil {
		return fmt.Errorf("failed to marshal exclude: %w", err)
	}

	query := `
		INSERT INTO tasks (
			name, type, paths_mapping, include_patterns, exclude_patterns,
			save_mode, open_cache, mkdir_if_single, delete_dir, keep_dir_struct,
			schedule_type, schedule_value, reverse, config, config_id, is_watching, watch_error,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, CURRENT_TIMESTAMP)
	`

	_, err = tx.Exec(ctx, query,
		t.Name, t.Type, pathsMappingJSON, includeJSON, excludeJSON,
		t.SaveMode, t.OpenCache, t.MkdirIfSingle, t.DeleteDir, t.KeepDirStruct,
		t.ScheduleType, t.ScheduleValue, t.Reverse, configName, configID, t.IsWatching, t.WatchError,
	)

	return err
}

func (s *Store) insertConfig(ctx context.Context, tx pgx.Tx, c *Config) error {
	query := `
		INSERT INTO configs (id, name, detail, updated_at)
		VALUES (COALESCE(NULLIF($1, 0), nextval(pg_get_serial_sequence('configs', 'id'))), $2, $3::jsonb, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
			detail = EXCLUDED.detail,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`

	var id int
	if err := tx.QueryRow(ctx, query, c.ID, c.Name, c.Detail).Scan(&id); err != nil {
		return err
	}
	c.ID = id
	return nil
}
