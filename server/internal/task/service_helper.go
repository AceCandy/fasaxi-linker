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

	if t.Config != "" {
		// Try to find the config in the fresh list
		var foundConfig *Config
		for _, c := range lookupConfigs {
			if c.Name == t.Config {
				foundConfig = &c
				break
			}
		}

		if foundConfig != nil {
			// Parse config detail
			var rc RuntimeConfig
			if err := json.Unmarshal([]byte(foundConfig.Detail), &rc); err == nil {
				// Success (removed debug log)
				// fmt.Printf("DEBUG: Applying config %s to task %s\n", t.Config, t.Name)
				return t.ToCoreOptionsWithConfig(&rc)
			} else {
				fmt.Printf("❌ 解析配置失败 %s (%s): %v. 使用默认配置。\n", t.Config, t.Name, err)
			}
		} else {
			fmt.Printf("⚠️ 未找到配置 %s (%s). 使用默认配置。\n", t.Config, t.Name)
		}
	}
	return t.ToCoreOptions()
}
