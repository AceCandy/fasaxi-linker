package api

import (
	"fmt"
	"os"

	"github.com/fasaxi-linker/servergo/internal/task"
)

// Helper to validate paths in task
func validatePathMapping(t task.Task) error {
	for _, mapping := range t.PathsMapping {
		// Check Source
		if _, err := os.Stat(mapping.Source); os.IsNotExist(err) {
			return fmt.Errorf("源路径不存在: %s", mapping.Source)
		}

		// Check Dest
		if _, err := os.Stat(mapping.Dest); os.IsNotExist(err) {
			return fmt.Errorf("目标路径不存在: %s", mapping.Dest)
		}
	}
	return nil
}
