package task

import (
	"encoding/json"
	"fmt"

	"github.com/fasaxi-linker/servergo/pkg/core"
)

// Helper to get options for a task, considering its linked config
func (s *Service) getTaskOptions(t Task) core.Options {
	// Priority 1: Use task's own fields if they exist (snapshot mode)
	// This is the new behavior - tasks store their own config fields
	if len(t.Include) > 0 || len(t.Exclude) > 0 {
		// Task has its own config fields, use them directly
		return t.ToCoreOptions()
	}

	// Priority 2: Fallback to config lookup for legacy tasks or tasks without synced fields
	// Reload configs to ensure we have the latest version from disk
	_, configs, err := s.store.Load()

	var lookupConfigs []Config
	if err == nil {
		lookupConfigs = configs
	} else {
		// Log error
		fmt.Printf("⚠️ 加载配置失败: %v。将使用默认配置。\n", err)
	}

	if t.ConfigID != 0 || t.Config != "" {
		// Try to find the config in the fresh list
		var foundConfig *Config
		for _, c := range lookupConfigs {
			if t.ConfigID != 0 && c.ID == t.ConfigID {
				foundConfig = &c
				break
			}
			if foundConfig == nil && t.Config != "" && c.Name == t.Config {
				foundConfig = &c // fallback by name for legacy data
			}
		}

		if foundConfig != nil {
			// Parse config detail
			var rc RuntimeConfig
			if err := json.Unmarshal([]byte(foundConfig.Detail), &rc); err == nil {
				return t.ToCoreOptionsWithConfig(&rc)
			}
			fmt.Printf("❌ 解析配置失败 %s(%d) (%s): %v. 使用默认配置。\n", foundConfig.Name, foundConfig.ID, t.Name, err)
		} else {
			fmt.Printf("⚠️ 未找到配置 %s(%d) (%s). 使用默认配置。\n", t.Config, t.ConfigID, t.Name)
		}
	}
	return t.ToCoreOptions()
}

// GetOptions 获取任务最终使用的配置（优先使用 config_id 关联最新 configs）
func (s *Service) GetOptions(taskID int) (core.Options, error) {
	s.mu.RLock()
	task, ok := s.tasksMap[taskID]
	s.mu.RUnlock()
	if !ok {
		return core.Options{}, fmt.Errorf("task %d not found", taskID)
	}
	return s.getTaskOptions(task), nil
}
