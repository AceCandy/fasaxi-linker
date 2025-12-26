package task

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GetLogger returns a logger function that writes to task log file and stdout
func GetLogger(taskName string) func(level, msg string) {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".hlink", "logs")
	_ = os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", taskName))

	return func(level, msg string) {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)

		// Write to file
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			_, _ = f.WriteString(logLine)
		} else {
			fmt.Printf("Error writing to log file: %v\n", err)
		}

		// Also print to stdout
		fmt.Print(logLine)
	}
}

// GetLogContent reads the log file content
func GetLogContent(taskName string) string {
	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, ".hlink", "logs", fmt.Sprintf("%s.log", taskName))
	
	content, err := os.ReadFile(logFile)
	if err != nil {
		return ""
	}
	return string(content)
}

// ClearLog clears the log file
func ClearLog(taskName string) error {
	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, ".hlink", "logs", fmt.Sprintf("%s.log", taskName))
	return os.WriteFile(logFile, []byte{}, 0644)
}
