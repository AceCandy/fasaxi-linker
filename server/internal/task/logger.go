package task

import (
	"fmt"

	"github.com/fasaxi-linker/servergo/internal/logs"
)

var logStore = &logs.Store{}

// GetLogger returns a logger function that writes to database and stdout
func GetLogger(taskName string) func(level, msg string) {
	return func(level, msg string) {
		// Write to database
		if err := logStore.Add(taskName, level, msg); err != nil {
			fmt.Printf("Error writing to log database: %v\n", err)
		}

		// Also print to stdout
		fmt.Printf("[%s] %s\n", level, msg)
	}
}

// GetLogContent reads the log content from database
func GetLogContent(taskName string) string {
	content, err := logStore.GetByTaskName(taskName)
	if err != nil {
		fmt.Printf("Error reading log from database: %v\n", err)
		return ""
	}
	return content
}

// ClearLog clears the log for a specific task
func ClearLog(taskName string) error {
	return logStore.Clear(taskName)
}
