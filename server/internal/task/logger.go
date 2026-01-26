package task

import (
	"fmt"

	"github.com/fasaxi-linker/servergo/internal/logs"
)

var logStore = &logs.Store{}

// GetLogger returns a logger function that writes to database and stdout
func GetLogger(taskName string) func(level, msg string) {
	// Resolve task name to ID
	store := GetSharedStore()
	taskID, err := store.GetTaskIDByName(taskName)
	if err != nil {
		// Fallback: log error but continue
		fmt.Printf("Warning: failed to resolve task name %s to ID: %v\n", taskName, err)
		taskID = 0
	}

	return func(level, msg string) {
		// Write to database if we have a valid task ID
		if taskID > 0 {
			if err := logStore.Add(taskID, level, msg); err != nil {
				fmt.Printf("Error writing to log database: %v\n", err)
			}
		}

		// Also print to stdout
		fmt.Printf("[%s] %s\n", level, msg)
	}
}

// GetLogEntries reads the log entries from database
func GetLogEntries(taskName string, offset, limit int) []logs.LogEntry {
	// Resolve task name to ID
	store := GetSharedStore()
	taskID, err := store.GetTaskIDByName(taskName)
	if err != nil {
		fmt.Printf("Error resolving task name to ID: %v\n", err)
		return nil
	}

	entries, err := logStore.GetStructByTaskID(taskID, offset, limit)
	if err != nil {
		fmt.Printf("Error reading log from database: %v\n", err)
		return nil
	}
	return entries
}

// GetLogContent reads the log content from database (Legacy string format)
func GetLogContent(taskName string) string {
	// Resolve task name to ID
	store := GetSharedStore()
	taskID, err := store.GetTaskIDByName(taskName)
	if err != nil {
		fmt.Printf("Error resolving task name to ID: %v\n", err)
		return ""
	}

	content, err := logStore.GetByTaskID(taskID)
	if err != nil {
		fmt.Printf("Error reading log from database: %v\n", err)
		return ""
	}
	return content
}

// ClearLog clears the log for a specific task
func ClearLog(taskName string) error {
	// Resolve task name to ID
	store := GetSharedStore()
	taskID, err := store.GetTaskIDByName(taskName)
	if err != nil {
		return fmt.Errorf("failed to resolve task name to ID: %w", err)
	}

	return logStore.Clear(taskID)
}
