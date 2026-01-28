package task

import (
	"fmt"
)

func (s *Service) RestartWatch(taskID int) error {
	// Check if watching
	s.wMu.RLock()
	_, ok := s.watchers[taskID]
	s.wMu.RUnlock()

	if !ok {
		return nil
	}

	task, ok := s.Get(taskID)
	if !ok {
		return fmt.Errorf("task %d not found", taskID)
	}

	logger := GetLogger(taskID)
	fmt.Printf("⚠️ 正在重启监听: %s (配置变更)\n", task.Name)

	// Stop
	if err := s.StopWatch(taskID); err != nil {
		return err
	}

	// Start (will reload config)
	return s.StartWatch(taskID, logger)
}
