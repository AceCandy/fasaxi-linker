package task

import (
	"encoding/json"
	"fmt"

	"github.com/fasaxi-linker/servergo/pkg/core"
)

// Helper to get options for a task, considering its linked config
func (s *Service) getTaskOptions(t Task) core.Options {
	// Reload configs to ensure we have the latest version from disk
	// This is important because ConfigService might have updated them
	_, configs, err := s.store.Load()

	var lookupConfigs []Config
	if err == nil {
		lookupConfigs = configs
		// Update local cache as well
		s.mu.Lock()
		s.configs = configs
		s.mu.Unlock()
	} else {
		// Log error if critical, or skip if safe
		fmt.Printf("⚠️ 加载配置失败: %v。使用缓存配置。\n", err)
		s.mu.RLock()
		lookupConfigs = s.configs
		s.mu.RUnlock()
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
func (s *Service) GetOptions(name string) (core.Options, error) {
	s.mu.RLock()
	task, ok := s.tasksMap[name]
	s.mu.RUnlock()
	if !ok {
		return core.Options{}, fmt.Errorf("task %s not found", name)
	}
	return s.getTaskOptions(task), nil
}
