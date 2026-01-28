package task

import (
	"fmt"
	"sync"

	"github.com/fasaxi-linker/servergo/internal/logs"
)

// activeLoggers stores currently active file loggers
var activeLoggers = make(map[int]*logs.FileLogger)
var loggersMu sync.RWMutex

// GetLogger returns a logger function that writes to file and stdout
// This is the default logger for watch mode
func GetLogger(taskID int) func(level, msg string) {
	return GetLoggerWithType(taskID, logs.ExecWatch)
}

// GetLoggerWithType returns a logger function with specified execution type
func GetLoggerWithType(taskID int, execType logs.ExecutionType) func(level, msg string) {
	loggersMu.Lock()
	defer loggersMu.Unlock()

	// Check if logger already exists for this task
	if logger, ok := activeLoggers[taskID]; ok {
		// For watch mode, reuse existing logger
		if execType == logs.ExecWatch {
			return func(level, msg string) {
				logger.Log(level, msg)
				fmt.Printf("[%s] %s\n", level, msg)
			}
		}
		// For run/cron, close existing and create new
		logger.Close()
		delete(activeLoggers, taskID)
	}

	// Create new file logger
	logger, err := logs.NewFileLogger(taskID, execType)
	if err != nil {
		fmt.Printf("Error creating file logger: %v\n", err)
		// Return a fallback logger that only prints to stdout
		return func(level, msg string) {
			fmt.Printf("[%s] %s\n", level, msg)
		}
	}

	activeLoggers[taskID] = logger

	return func(level, msg string) {
		logger.Log(level, msg)
		fmt.Printf("[%s] %s\n", level, msg)
	}
}

// CloseLogger closes the logger for a specific task
func CloseLogger(taskID int) {
	loggersMu.Lock()
	defer loggersMu.Unlock()

	if logger, ok := activeLoggers[taskID]; ok {
		logger.Close()
		delete(activeLoggers, taskID)
	}
}

// GetLogFiles returns the list of log files for a task
func GetLogFiles(taskID int) ([]logs.LogFileInfo, error) {
	return logs.GetLogFiles(taskID)
}

// GetLogEntries reads the log entries from file with pagination and filtering
func GetLogEntries(taskID int, filename string, page, pageSize int, levelFilter, search string) ([]logs.LogEntry, int, error) {
	// If no filename specified, use the latest
	if filename == "" {
		var err error
		filename, err = logs.GetLatestLogFile(taskID)
		if err != nil {
			return nil, 0, err
		}
		if filename == "" {
			return []logs.LogEntry{}, 0, nil
		}
	}

	return logs.ReadLogFile(taskID, filename, page, pageSize, levelFilter, search)
}

// ClearLog clears log files for a specific task
func ClearLog(taskID int, filename string) error {
	if taskID <= 0 {
		return fmt.Errorf("invalid task id")
	}

	// Also close active logger if clearing all
	if filename == "" {
		CloseLogger(taskID)
	}

	return logs.ClearLogFile(taskID, filename)
}
