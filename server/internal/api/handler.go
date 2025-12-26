package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fasaxi-linker/servergo/internal/config"
	"github.com/fasaxi-linker/servergo/internal/task"
	"github.com/fasaxi-linker/servergo/pkg/core"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service       *task.Service
	ConfigService *config.Service
}

func NewHandler() (*Handler, error) {
	s, err := task.NewService()
	if err != nil {
		return nil, err
	}
	cs, err := config.NewService()
	if err != nil {
		return nil, err
	}
	return &Handler{
		Service:       s,
		ConfigService: cs,
	}, nil
}

// ... existing System methods ...

// === Config ===

func (h *Handler) GetConfigDefault(c *gin.Context) {
	// Return default config template, hardcoded for now or from file
	defaultConfig := `/**
 * @type {import('@hlink/core').IConfig}
 */
export default {
  // Add your config here
}`
	Success(c, defaultConfig)
}

func (h *Handler) GetConfigList(c *gin.Context) {
	h.ConfigService.Reload() // Ensure fresh data
	Success(c, h.ConfigService.GetAll())
}

func (h *Handler) GetConfig(c *gin.Context) {
	name := c.Query("name")
	conf, detail, ok := h.ConfigService.Get(name)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

	Success(c, gin.H{
		"name":        conf.Name,
		"description": conf.Description,
		"detail":      detail,
	})
}

func (h *Handler) GetConfigDetail(c *gin.Context) {
	// Parse the config content and return the configuration object
	name := c.Query("name")
	if name == "" {
		ErrorMsg(c, "Config name is required")
		return
	}

	// Get parsed config from config service
	config, ok := h.ConfigService.GetParsed(name)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

	Success(c, config)
}

func (h *Handler) AddConfig(c *gin.Context) {
	var body struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Detail      interface{} `json:"detail"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	// Convert detail to JSON string
	var detailStr string
	switch detail := body.Detail.(type) {
	case string:
		detailStr = detail
	case map[string]interface{}, []interface{}:
		detailBytes, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			ErrorMsg(c, "Failed to marshal detail to JSON")
			return
		}
		detailStr = string(detailBytes)
	default:
		// Try to marshal as JSON anyway
		detailBytes, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			ErrorMsg(c, "Invalid detail format, expected string or JSON object")
			return
		}
		detailStr = string(detailBytes)
	}

	conf := task.Config{
		Name:        body.Name,
		Description: body.Description,
		Detail:      "", // This will be set in the Add method
	}

	if err := h.ConfigService.Add(conf, detailStr); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	var body struct {
		PreName     string      `json:"preName"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Detail      interface{} `json:"detail"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	// Convert detail to JSON string
	var detailStr string
	switch detail := body.Detail.(type) {
	case string:
		detailStr = detail
	case map[string]interface{}, []interface{}:
		detailBytes, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			ErrorMsg(c, "Failed to marshal detail to JSON")
			return
		}
		detailStr = string(detailBytes)
	default:
		// Try to marshal as JSON anyway
		detailBytes, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			ErrorMsg(c, "Invalid detail format, expected string or JSON object")
			return
		}
		detailStr = string(detailBytes)
	}

	// Dirty check
	existingConfig, existingDetail, ok := h.ConfigService.Get(body.PreName)
	if ok {
		// Compare fields
		// For Detail which is JSON, we parse both and compare objects to ignore formatting differences
		var IsDetailEqual bool
		if existingDetail == detailStr {
			IsDetailEqual = true
		} else {
			var v1, v2 interface{}
			err1 := json.Unmarshal([]byte(existingDetail), &v1)
			err2 := json.Unmarshal([]byte(detailStr), &v2)
			if err1 == nil && err2 == nil {
				IsDetailEqual = reflect.DeepEqual(v1, v2)
			}
		}

		if body.PreName == body.Name &&
			existingConfig.Description == body.Description &&
			IsDetailEqual {
			// No changes
			Success(c, true)
			return
		}
	}

	conf := task.Config{
		Name:        body.Name,
		Description: body.Description,
		Detail:      "", // This will be set in the Update method
	}

	if err := h.ConfigService.Update(body.PreName, conf, detailStr); err != nil {
		Error(c, err)
		return
	}

	// Check for watching tasks that use this config and restart them
	tasks := h.Service.GetAll()
	for _, t := range tasks {
		// If task uses the updated config (check against PreName as that's what the task currently holds)
		// Or Check if t.Config == body.Name if we assume rename works?
		// But in Update(), we update ConfigService, but Tasks still hold the old name string if it was renamed.
		// If PreName == Name (no rename), then t.Config == PreName is correct.
		// If PreName != Name (rename), the task link is effectively broken until task is updated,
		// BUT we should still try to restart if it matches PreName because the config content changed.
		// However, if we renamed it, the new start will try to load `t.Config` (which is PreName).
		// But `PreName` no longer exists in ConfigService (it was renamed to Name).
		// So checking for restart here handles the "Update Content" case perfectly.
		// For "Rename" case, restarting might fail to find config, reverting to default.
		// That is acceptable behavior for now (user should update task too).
		if t.Config == body.PreName {
			if h.Service.IsWatching(t.Name) {
				go func(taskName string) {
					task.GetLogger(taskName)("WARN", fmt.Sprintf("⚠️ 正在重启监听: %s (配置-%s变更)\n", taskName, body.PreName))
					if err := h.Service.RestartWatch(taskName); err != nil {
						fmt.Printf("Failed to restart task %s: %v\n", taskName, err)
					}
				}(t.Name)
			}
		}
	}

	Success(c, true)
}

func (h *Handler) DeleteConfig(c *gin.Context) {
	name := c.Query("name")
	if err := h.ConfigService.Delete(name); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

// ... existing Task methods ...

// === System ===

func (h *Handler) Version(c *gin.Context) {
	// Mock version for now
	Success(c, gin.H{
		"tag":        "stable",
		"version":    "0.0.1-go",
		"needUpdate": false,
	})
}

func (h *Handler) Update(c *gin.Context) {
	// No-op for Go version initially, or implement self-update
	Success(c, true)
}

// === Task ===

func (h *Handler) GetTaskList(c *gin.Context) {
	tasks := h.Service.GetAll()
	// Frontend expects isWatching field
	type TaskWithStatus struct {
		task.Task
		IsWatching bool `json:"isWatching"`
	}

	result := make([]TaskWithStatus, len(tasks))
	for i, t := range tasks {
		result[i] = TaskWithStatus{
			Task:       t,
			IsWatching: h.Service.IsWatching(t.Name),
		}
	}

	Success(c, result)
}

func (h *Handler) GetTask(c *gin.Context) {
	name := c.Query("name")
	t, ok := h.Service.Get(name)
	if !ok {
		ErrorMsg(c, "Task not found")
		return
	}
	Success(c, gin.H{ // Frontend expects mixed object
		"name":          t.Name,
		"type":          t.Type,
		"pathsMapping":  t.PathsMapping,
		"include":       t.Include,
		"exclude":       t.Exclude,
		"saveMode":      t.SaveMode,
		"openCache":     t.OpenCache,
		"mkdirIfSingle": t.MkdirIfSingle,
		"deleteDir":     t.DeleteDir,
		"keepDirStruct": t.KeepDirStruct,
		"scheduleType":  t.ScheduleType,
		"scheduleValue": t.ScheduleValue,
		"reverse":       t.Reverse,
		"config":        t.Config, // This should be the config name, not the whole task object
		"isWatching":    h.Service.IsWatching(name),
	})
}

func (h *Handler) CreateTask(c *gin.Context) {
	var t task.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		Error(c, err)
		return
	}

	// Validate paths
	if err := validatePathMapping(t); err != nil {
		ErrorMsg(c, err.Error())
		return
	}

	if err := h.Service.Add(t); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	var body struct {
		PreName string `json:"preName"`
		task.Task
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	// Validate paths
	if err := validatePathMapping(body.Task); err != nil {
		ErrorMsg(c, err.Error())
		return
	}

	// Check if task exists to prevent error later, and also for dirty check
	existingTask, ok := h.Service.Get(body.PreName)
	if !ok {
		ErrorMsg(c, "Task not found")
		return
	}

	// Dirty check: If nothing changed, return success immediately
	if body.PreName == body.Task.Name {
		// Normalize slices for comparison (nil vs empty)
		t1 := existingTask
		t2 := body.Task

		if t1.PathsMapping == nil {
			t1.PathsMapping = []task.PathMapping{}
		}
		if t2.PathsMapping == nil {
			t2.PathsMapping = []task.PathMapping{}
		}
		if t1.Include == nil {
			t1.Include = []string{}
		}
		if t2.Include == nil {
			t2.Include = []string{}
		}
		if t1.Exclude == nil {
			t1.Exclude = []string{}
		}
		if t2.Exclude == nil {
			t2.Exclude = []string{}
		}

		if reflect.DeepEqual(t1, t2) {
			Success(c, true)
			return
		}
	}

	// Check if task is currently watching before update
	wasWatching := h.Service.IsWatching(body.PreName)
	if wasWatching {
		// Stop the old watcher (especially important if name changes)
		// We use a temporary log to indicate restart
		task.GetLogger(body.PreName)("WARN", fmt.Sprintf("⚠️ 正在重启监听:%s (任务变更)\n", body.PreName))
		if err := h.Service.StopWatch(body.PreName); err != nil {
			fmt.Printf("Warning: Failed to stop watcher for update: %v\n", err)
		}
	}

	if err := h.Service.Update(body.PreName, body.Task); err != nil {
		Error(c, err)
		return
	}

	// Restart watcher if it was running
	if wasWatching {
		// Start watcher with new name (in case it changed)
		go func(name string) {
			logger := task.GetLogger(name)
			// Small delay to ensure DB save completes? Usually not needed with locks.
			if err := h.Service.StartWatch(name, logger); err != nil {
				errMsg := fmt.Sprintf("任务更新后重启监听失败: %v", err)
				logger("ERROR", errMsg)
				fmt.Println(errMsg)
			}
		}(body.Task.Name)
	}
	Success(c, true)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	name := c.Query("name")
	if err := h.Service.Delete(name); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

// === Run (SSE) ===

func (h *Handler) RunTask(c *gin.Context) {
	name := c.Query("name")
	// alive := c.Query("alive") // '0' or '1'

	t, ok := h.Service.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Logger callback that sends SSE events
	logChan := make(chan gin.H)

	go func() {
		defer close(logChan)

		// Get associated config as JSON
		var opts core.Options
		if t.Config != "" {
			// Get parsed config from config service
			if parsedConfig, ok := h.ConfigService.GetParsed(t.Config); ok {
				opts = t.ToCoreOptionsWithConfig(&parsedConfig)
			} else {
				opts = t.ToCoreOptions()
			}
		} else {
			opts = t.ToCoreOptions()
		}

		if t.Type == "prune" {
			// Prune logic
			files, err := core.GetPruneFiles(opts)
			if err != nil {
				logChan <- gin.H{"status": "failed", "type": "prune", "output": fmt.Sprintf("Error: %v", err)}
				return
			}
			if len(files) == 0 {
				logChan <- gin.H{"status": "succeed", "type": "prune", "output": "没有找到需要修剪的硬链"}
			} else {
				// Prune requires confirmation usually. Frontend expects list of files?
				// TaskSDK.run logic for prune:
				// It waits for confirmation if files found.
				// My core implementation doesn't support interactive confirmation yet via Runner.
				// But let's simulate:
				// Send "confirm" status?
				// TaskSDK: output: '...WARN...confirm...'
				logChan <- gin.H{"status": "ongoing", "type": "prune", "output": fmt.Sprintf("Wait implement prune confirm.")}
			}

		} else {
			// Main logic
			// Get file logger
			fileLogger := task.GetLogger(name)

			_, err := core.Run(opts, func(level, msg string) {
				// 1. Emit to SSE
				logChan <- gin.H{
					"status": "ongoing",
					"type":   "main",
					"output": fmt.Sprintf("[%s] %s", level, msg),
				}

				// 2. Write to log file
				fileLogger(level, msg)
			})
			if err != nil {
				errMsg := err.Error()
				logChan <- gin.H{"status": "failed", "type": "main", "output": errMsg}
				fileLogger("ERROR", errMsg)
			} else {
				logChan <- gin.H{"status": "succeed", "type": "main", "output": "Done"}
				fileLogger("SUCCEED", "Task Completed")
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		msg, ok := <-logChan
		if !ok {
			return false
		}
		c.SSEvent("message", msg)
		return true
	})
}

// === Watch ===

func (h *Handler) StartWatch(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	// Get task
	t, ok := h.Service.Get(body.Name)
	if !ok {
		ErrorMsg(c, "Task not found")
		return
	}

	// Get associated config
	var opts core.Options
	if t.Config != "" {
		if parsedConfig, ok := h.ConfigService.GetParsed(t.Config); ok {
			opts = t.ToCoreOptionsWithConfig(&parsedConfig)
		} else {
			opts = t.ToCoreOptions()
		}
	} else {
		opts = t.ToCoreOptions()
	}

	logger := task.GetLogger(body.Name)

	// Start watcher with options
	if err := h.Service.StartWatchWithOptions(body.Name, logger, opts); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

func (h *Handler) StopWatch(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	if err := h.Service.StopWatch(body.Name); err != nil {
		Error(c, err)
		return
	}

	// Add log entry for stop in a goroutine to avoid blocking
	go func() {
		task.GetLogger(body.Name)("WARN", fmt.Sprintf("⛔ [%s] 监听服务已停止", body.Name))
	}()

	Success(c, true)
}

func (h *Handler) GetWatchStatus(c *gin.Context) {
	name := c.Query("name")
	Success(c, h.Service.IsWatching(name))
}

func (h *Handler) GetTaskLog(c *gin.Context) {
	name := c.Query("name")
	Success(c, task.GetLogContent(name))
}

func (h *Handler) ClearTaskLog(c *gin.Context) {
	name := c.Query("name")
	_ = task.ClearLog(name)
	Success(c, true)
}

// === Cache ===

func (h *Handler) GetCache(c *gin.Context) {
	// Get cache file path
	cachePath := h.getCachePath()

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if os.IsNotExist(err) {
		Success(c, "[]")
		return
	}
	if err != nil {
		ErrorMsg(c, fmt.Sprintf("读取缓存文件失败: %v", err))
		return
	}

	Success(c, string(data))
}

func (h *Handler) UpdateCache(c *gin.Context) {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	cachePath := h.getCachePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		Error(c, err)
		return
	}

	// Write cache content
	if err := os.WriteFile(cachePath, []byte(body.Content), 0644); err != nil {
		Error(c, err)
		return
	}

	Success(c, true)
}

func (h *Handler) GetCacheLog(c *gin.Context) {
	logPath := h.getCacheLogPath()

	data, err := os.ReadFile(logPath)
	if os.IsNotExist(err) {
		Success(c, "")
		return
	}
	if err != nil {
		ErrorMsg(c, fmt.Sprintf("读取日志文件失败: %v", err))
		return
	}

	Success(c, string(data))
}

func (h *Handler) ClearCacheLog(c *gin.Context) {
	logPath := h.getCacheLogPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		Error(c, err)
		return
	}

	// Clear log file
	if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
		Error(c, err)
		return
	}

	Success(c, true)
}

// Helper methods for cache paths

func (h *Handler) getCachePath() string {
	// Get hlink home directory
	hlinkHome := os.Getenv("HLINK_HOME")
	if hlinkHome == "" {
		homeDir, _ := os.UserHomeDir()
		hlinkHome = filepath.Join(homeDir, ".hlink")
	}
	return filepath.Join(hlinkHome, "cache-array.json")
}

func (h *Handler) getCacheLogPath() string {
	hlinkHome := os.Getenv("HLINK_HOME")
	if hlinkHome == "" {
		homeDir, _ := os.UserHomeDir()
		hlinkHome = filepath.Join(homeDir, ".hlink")
	}
	return filepath.Join(hlinkHome, "serve.log")
}
