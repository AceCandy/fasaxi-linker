package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/fasaxi-linker/servergo/internal/task"
)

type Service struct {
	store      *task.Store
	tasks      []task.Task
	configs    []task.Config
	configsMap map[string]task.Config
	mu         sync.RWMutex
}

func NewService() (*Service, error) {
	store := task.GetSharedStore()
	tasks, configs, err := store.Load()
	if err != nil {
		return nil, err
	}

	s := &Service{
		store:   store,
		tasks:   tasks,
		configs: configs,
	}
	s.rebuildMap()
	return s, nil
}

func (s *Service) rebuildMap() {
	s.configsMap = make(map[string]task.Config)
	for _, c := range s.configs {
		s.configsMap[c.Name] = c
	}
}

// Reload reloads data from store to sync with TaskService changes
func (s *Service) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tasks, configs, err := s.store.Load()
	if err != nil {
		return err
	}
	s.tasks = tasks
	s.configs = configs
	s.rebuildMap()
	return nil
}

func (s *Service) save() error {
	// Re-read tasks to avoid overwriting changes from TaskService
	// There is still a race condition here if TaskService saves between Load and Save.
	// In a real app, Store should be a singleton with internal locking for the whole DB.
	// Current architecture with multiple services having their own store instance is flawed for shared file.
	// Quick fix: TaskService and ConfigService should probably share the same Store pointer passed from main.
	// But refactoring that now is too much.
	// Let's just try to be safe: Load -> Update Configs -> Save.

	tasks, _, err := s.store.Load()
	if err == nil {
		s.tasks = tasks // update local tasks
	}

	return s.store.Save(s.tasks, s.configs)
}

func (s *Service) GetAll() []task.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Reload() // Optionally reload first?
	res := make([]task.Config, len(s.configs))
	copy(res, s.configs)
	return res
}

func (s *Service) Get(name string) (task.Config, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.configsMap[name]
	if !ok {
		return task.Config{}, "", false
	}

	// Return config with JSON detail
	return c, c.Detail, true
}

// GetParsed returns the parsed configuration as JSON
func (s *Service) GetParsed(name string) (ParsedConfig, bool) {
	c, _, ok := s.Get(name)
	if !ok {
		return ParsedConfig{}, false
	}

	// First, try to detect if this is old format by checking for nested structure
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(c.Detail), &rawConfig); err != nil {
		return ParsedConfig{}, false
	}

	config := ParsedConfig{
		KeepDirStruct: true, // defaults
		OpenCache:     true,
		MkdirIfSingle: false,
		DeleteDir:     false,
	}

	// Check if this is old format (nested include/exclude with exts)
	isOldFormat := false
	if include, ok := rawConfig["include"].(map[string]interface{}); ok {
		if _, hasExts := include["exts"]; hasExts {
			isOldFormat = true
		} else if _, hasExts := include["Exts"]; hasExts {
			isOldFormat = true
		}
	}

	if isOldFormat {
		// Parse old format
		if include, ok := rawConfig["include"].(map[string]interface{}); ok {
			if exts, ok := include["exts"].([]interface{}); ok {
				for _, ext := range exts {
					if s, ok := ext.(string); ok {
						config.Include = append(config.Include, s)
					}
				}
			} else if exts, ok := include["Exts"].([]interface{}); ok {
				// Handle capitalized field names
				for _, ext := range exts {
					if s, ok := ext.(string); ok {
						config.Include = append(config.Include, s)
					}
				}
			}
		}

		if exclude, ok := rawConfig["exclude"].(map[string]interface{}); ok {
			if exts, ok := exclude["exts"].([]interface{}); ok {
				for _, ext := range exts {
					if s, ok := ext.(string); ok {
						config.Exclude = append(config.Exclude, s)
					}
				}
			} else if exts, ok := exclude["Exts"].([]interface{}); ok {
				// Handle capitalized field names
				for _, ext := range exts {
					if s, ok := ext.(string); ok {
						config.Exclude = append(config.Exclude, s)
					}
				}
			}
		}
	} else {
		// Try parsing as new flat format
		if err := json.Unmarshal([]byte(c.Detail), &config); err != nil {
			return ParsedConfig{}, false
		}

		// Parse Include array from rawConfig
		if include, ok := rawConfig["include"].([]interface{}); ok {
			config.Include = []string{}
			for _, pattern := range include {
				if s, ok := pattern.(string); ok {
					config.Include = append(config.Include, s)
				}
			}
		}

		// Parse Exclude array from rawConfig
		if exclude, ok := rawConfig["exclude"].([]interface{}); ok {
			config.Exclude = []string{}
			for _, pattern := range exclude {
				if s, ok := pattern.(string); ok {
					config.Exclude = append(config.Exclude, s)
				}
			}
		}
	}

	// Parse boolean fields (for both formats)
	if v, ok := rawConfig["keepDirStruct"].(bool); ok {
		config.KeepDirStruct = v
	}
	if v, ok := rawConfig["openCache"].(bool); ok {
		config.OpenCache = v
	}
	if v, ok := rawConfig["mkdirIfSingle"].(bool); ok {
		config.MkdirIfSingle = v
	}
	if v, ok := rawConfig["deleteDir"].(bool); ok {
		config.DeleteDir = v
	}

	return config, true
}

// ConvertJSToJSON converts JavaScript configuration to JSON format
func (s *Service) ConvertJSToJSON(jsConfig string) (string, error) {
	// Parse the JavaScript configuration
	parsed, err := s.parseConfigDetail(jsConfig)
	if err != nil {
		return "", err
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// parseConfigDetail parses JavaScript configuration string
func (s *Service) parseConfigDetail(detail string) (*ParsedConfig, error) {
	config := &ParsedConfig{
		KeepDirStruct: true, // default values
		OpenCache:     true,
		MkdirIfSingle: false,
		DeleteDir:     false,
	}

	// Remove newlines and tabs for easier regex matching
	detailOneLine := strings.ReplaceAll(detail, "\n", " ")
	detailOneLine = strings.ReplaceAll(detailOneLine, "\t", " ")

	// Try to parse new format first (flat arrays)
	newIncludeRe := regexp.MustCompile(`include:\s*\[(.*?)\]`)
	if matches := newIncludeRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		patternsStr := strings.TrimSpace(matches[1])
		if patternsStr != "" {
			patterns := strings.Split(patternsStr, ",")
			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				pattern = strings.Trim(pattern, `'"`)
				if pattern != "" {
					config.Include = append(config.Include, pattern)
				}
			}
		}
	}

	newExcludeRe := regexp.MustCompile(`exclude:\s*\[(.*?)\]`)
	if matches := newExcludeRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		patternsStr := strings.TrimSpace(matches[1])
		if patternsStr != "" {
			patterns := strings.Split(patternsStr, ",")
			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				pattern = strings.Trim(pattern, `'"`)
				if pattern != "" {
					config.Exclude = append(config.Exclude, pattern)
				}
			}
		}
	}

	// If new format didn't match, try old format (nested exts)
	if len(config.Include) == 0 && len(config.Exclude) == 0 {
		// Parse exclude exts (old format)
		excludeRe := regexp.MustCompile(`exclude:\s*\{[^}]*exts:\s*\[(.*?)\]`)
		if matches := excludeRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
			extsStr := strings.TrimSpace(matches[1])
			if extsStr != "" {
				exts := strings.Split(extsStr, ",")
				for _, ext := range exts {
					ext = strings.TrimSpace(ext)
					ext = strings.Trim(ext, `'"`)
					if ext != "" {
						if strings.HasPrefix(ext, "!") {
							ext = ext[1:]
						}
						config.Exclude = append(config.Exclude, ext)
					}
				}
			}
		}

		// Parse include exts (old format)
		includeRe := regexp.MustCompile(`include:\s*\{[^}]*exts:\s*\[(.*?)\]`)
		if matches := includeRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
			extsStr := strings.TrimSpace(matches[1])
			if extsStr != "" {
				exts := strings.Split(extsStr, ",")
				for _, ext := range exts {
					ext = strings.TrimSpace(ext)
					ext = strings.Trim(ext, `'"`)
					if ext != "" {
						if strings.HasPrefix(ext, "!") {
							ext = ext[1:]
						}
						config.Include = append(config.Include, ext)
					}
				}
			}
		}
	}

	// Parse boolean values
	keepDirStructRe := regexp.MustCompile(`keepDirStruct:\s*(true|false)`)
	if matches := keepDirStructRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		config.KeepDirStruct = matches[1] == "true"
	}

	openCacheRe := regexp.MustCompile(`openCache:\s*(true|false)`)
	if matches := openCacheRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		config.OpenCache = matches[1] == "true"
	}

	mkdirIfSingleRe := regexp.MustCompile(`mkdirIfSingle:\s*(true|false)`)
	if matches := mkdirIfSingleRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		config.MkdirIfSingle = matches[1] == "true"
	}

	deleteDirRe := regexp.MustCompile(`deleteDir:\s*(true|false)`)
	if matches := deleteDirRe.FindStringSubmatch(detailOneLine); len(matches) > 1 {
		config.DeleteDir = matches[1] == "true"
	}

	return config, nil
}

func (s *Service) Add(c task.Config, detail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.configsMap[c.Name]; ok {
		return fmt.Errorf("config %s already exists", c.Name)
	}

	// First, try to unmarshal the detail as a JSON string (escaped)
	var jsonStr string
	if err := json.Unmarshal([]byte(detail), &jsonStr); err == nil {
		// detail is an escaped JSON string
		detail = jsonStr
	}

	// Parse the detail to ensure it's valid
	var config ParsedConfig
	if err := json.Unmarshal([]byte(detail), &config); err == nil {
		// Valid JSON, re-marshal to ensure consistent format
		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}
		c.Detail = string(jsonBytes)
	} else {
		// Convert JavaScript detail to JSON format
		jsonDetail, err := s.ConvertJSToJSON(detail)
		if err != nil {
			return fmt.Errorf("failed to convert config to JSON: %v", err)
		}
		c.Detail = jsonDetail
	}

	s.configs = append(s.configs, c)
	s.rebuildMap()
	return s.save()
}

func (s *Service) Update(prevName string, c task.Config, detail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.configsMap[prevName]
	if !ok {
		return fmt.Errorf("config %s does not exist", prevName)
	}

	// First, try to unmarshal the detail as a JSON string (escaped)
	var jsonStr string
	if err := json.Unmarshal([]byte(detail), &jsonStr); err == nil {
		// detail is an escaped JSON string
		detail = jsonStr
	}

	// Parse the detail to ensure it's valid
	var config ParsedConfig
	if err := json.Unmarshal([]byte(detail), &config); err == nil {
		// Valid JSON, re-marshal to ensure consistent format
		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}
		existing.Detail = string(jsonBytes)
	} else {
		// Convert JavaScript detail to JSON format
		jsonDetail, err := s.ConvertJSToJSON(detail)
		if err != nil {
			return fmt.Errorf("failed to convert config to JSON: %v", err)
		}
		existing.Detail = jsonDetail
	}

	existing.Name = c.Name
	existing.Description = c.Description

	// If name changed, check collision
	if prevName != c.Name {
		if _, ok := s.configsMap[c.Name]; ok {
			return fmt.Errorf("config %s already exists", c.Name)
		}
	}

	// Update list
	for i, conf := range s.configs {
		if conf.Name == prevName {
			s.configs[i] = existing
			break
		}
	}
	s.rebuildMap()
	return s.save()
}

func (s *Service) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.configsMap[name]; !ok {
		return fmt.Errorf("config %s not found", name)
	}

	// Remove from list
	var newConfigs []task.Config
	for _, c := range s.configs {
		if c.Name != name {
			newConfigs = append(newConfigs, c)
		}
	}
	s.configs = newConfigs
	s.rebuildMap()
	return s.save()
}
