package task

import (
	"fmt"
)

func (s *Service) RestartWatch(name string) error {
	// Check if watching
	s.wMu.RLock()
	_, ok := s.watchers[name]
	s.wMu.RUnlock()

	if !ok {
		return nil
	}

	logger := GetLogger(name)
	fmt.Printf("⚠️ 正在重启监听: %s (配置变更)\n", name)

	// Stop
	if err := s.StopWatch(name); err != nil {
		return err
	}

	// Start (will reload config)
	return s.StartWatch(name, logger)
}
