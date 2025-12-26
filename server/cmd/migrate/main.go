package main
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fasaxi-linker/servergo/internal/task"
)

// OldConfig represents the old config structure with external file path
type OldConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ConfigPath  string `json:"configPath"`
}

// OldDBWrapper represents the old database structure
type OldDBWrapper struct {
	Tasks   []task.Task   `json:"tasks"`
	Configs []OldConfig   `json:"configs"`
}

// NewConfig represents the new config structure with embedded detail
type NewConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
}

// NewDBWrapper represents the new database structure
type NewDBWrapper struct {
	Tasks   []task.Task   `json:"tasks"`
	Configs []NewConfig   `json:"configs"`
}

func main() {
	// Check if migration is needed
	dbPath := filepath.Join(os.Getenv("HOME"), ".hlink", "db.json")
	
	// Read current database
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read database file: %v", err)
	}

	// Try to parse as old format
	var oldDB OldDBWrapper
	if err := json.Unmarshal(data, &oldDB); err != nil {
		// If parsing fails with configPath field, assume it's already migrated
		var newDB NewDBWrapper
		if err := json.Unmarshal(data, &newDB); err == nil {
			if len(newDB.Configs) > 0 && newDB.Configs[0].Detail != "" {
				fmt.Println("Database is already migrated.")
				return
			}
		}
		log.Fatalf("Failed to parse database: %v", err)
	}

	fmt.Printf("Starting migration for %d configs...\n", len(oldDB.Configs))

	// Migrate configs
	newConfigs := make([]NewConfig, len(oldDB.Configs))
	for i, oldConfig := range oldDB.Configs {
		// Read the external file
		detail, err := ioutil.ReadFile(oldConfig.ConfigPath)
		if err != nil {
			fmt.Printf("Warning: Failed to read config file %s: %v\n", oldConfig.ConfigPath, err)
			detail = []byte("export default {}") // Default empty config
		} else {
			fmt.Printf("Migrated config %s from %s\n", oldConfig.Name, oldConfig.ConfigPath)
		}

		newConfigs[i] = NewConfig{
			Name:        oldConfig.Name,
			Description: oldConfig.Description,
			Detail:      string(detail),
		}
	}

	// Create new database structure
	newDB := NewDBWrapper{
		Tasks:   oldDB.Tasks,
		Configs: newConfigs,
	}

	// Save migrated database
	newData, err := json.MarshalIndent(newDB, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal new database: %v", err)
	}

	// Create backup
	backupPath := dbPath + ".backup"
	if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}
	fmt.Printf("Created backup: %s\n", backupPath)

	// Write new database
	if err := ioutil.WriteFile(dbPath, newData, 0644); err != nil {
		log.Fatalf("Failed to write new database: %v", err)
	}

	fmt.Printf("Migration completed successfully!\n")
	fmt.Printf("Migrated %d configs to new format.\n", len(newConfigs))

	// Optionally, remove old config files after successful migration
	removedCount := 0
	for _, oldConfig := range oldDB.Configs {
		if oldConfig.ConfigPath != "" {
			if err := os.Remove(oldConfig.ConfigPath); err != nil {
				fmt.Printf("Warning: Failed to remove old config file %s: %v\n", oldConfig.ConfigPath, err)
			} else {
				fmt.Printf("Removed old config file: %s\n", oldConfig.ConfigPath)
				removedCount++
			}
		}
	}
	if removedCount > 0 {
		fmt.Printf("Removed %d old config files.\n", removedCount)
	}
}
