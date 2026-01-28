package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fasaxi-linker/servergo/internal/logs"
	"github.com/fasaxi-linker/servergo/pkg/core"
)

// RunState represents the running state of a task
type RunState struct {
	TaskID    int                `json:"taskId"`
	StartTime time.Time          `json:"startTime"`
	Cancel    context.CancelFunc `json:"-"`
}

// RunManager manages running task instances
type RunManager struct {
	running map[int]*RunState
	mu      sync.RWMutex
}

var runManager = &RunManager{
	running: make(map[int]*RunState),
}

// IsRunning checks if a task is currently running
func IsRunning(taskID int) bool {
	runManager.mu.RLock()
	defer runManager.mu.RUnlock()
	_, ok := runManager.running[taskID]
	return ok
}

// StartRun starts a task asynchronously
func StartRun(taskID int, opts core.Options) error {
	runManager.mu.Lock()
	if _, ok := runManager.running[taskID]; ok {
		runManager.mu.Unlock()
		return fmt.Errorf("任务正在执行中")
	}

	ctx, cancel := context.WithCancel(context.Background())
	runManager.running[taskID] = &RunState{
		TaskID:    taskID,
		StartTime: time.Now(),
		Cancel:    cancel,
	}
	runManager.mu.Unlock()

	// Start async execution
	go func() {
		defer func() {
			runManager.mu.Lock()
			delete(runManager.running, taskID)
			runManager.mu.Unlock()
		}()

		// Get file logger
		fileLogger := GetLoggerWithType(taskID, logs.ExecRun)
		fileLogger("INFO", "🚀 任务开始执行...")

		// Run with context for cancellation support
		stats, err := runWithContext(ctx, opts, func(level, msg string) {
			fileLogger(level, msg)
		})

		if ctx.Err() == context.Canceled {
			fileLogger("WARN", "⚠️ 任务已被手动停止")
			return
		}

		if err != nil {
			fileLogger("ERROR", fmt.Sprintf("❌ 任务失败: %s", err.Error()))
		} else {
			fileLogger("SUCCEED", fmt.Sprintf("✅ 任务完成 (成功: %d, 失败: %d)", stats.SuccessCount, stats.FailCount))
		}

		// Close log file
		CloseLogger(taskID)
	}()

	return nil
}

// StopRun stops a running task
func StopRun(taskID int) error {
	runManager.mu.Lock()
	defer runManager.mu.Unlock()

	state, ok := runManager.running[taskID]
	if !ok {
		return fmt.Errorf("任务未在执行")
	}

	state.Cancel()
	return nil
}

// runWithContext wraps core.Run with context cancellation support
func runWithContext(ctx context.Context, opts core.Options, logger func(string, string)) (core.Stats, error) {
	// Check if already cancelled
	select {
	case <-ctx.Done():
		return core.Stats{}, ctx.Err()
	default:
	}

	// Run the task (core.Run doesn't support context yet, but we can check periodically)
	return core.Run(opts, func(level, msg string) {
		// Check cancellation before each log
		select {
		case <-ctx.Done():
			return
		default:
			if logger != nil {
				logger(level, msg)
			}
		}
	})
}
