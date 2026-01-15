package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/fasaxi-linker/servergo/internal/cache"
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
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ErrorMsg(c, "Config id is required")
		return
	}
	conf, detail, ok := h.ConfigService.GetByID(id)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

	Success(c, gin.H{
		"id":          conf.ID,
		"name":        conf.Name,
		"description": conf.Description,
		"detail":      detail,
	})
}

func (h *Handler) GetConfigDetail(c *gin.Context) {
	// Parse the config content and return the configuration object
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ErrorMsg(c, "Config id is required")
		return
	}

	// Get parsed config from config service
	config, ok := h.ConfigService.GetParsedByID(id)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

	Success(c, config)
}

func (h *Handler) AddConfig(c *gin.Context) {
	var body struct {
		ID          int         `json:"id"`
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
		ID          int         `json:"id"`
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
	existingConfig, existingDetail, ok := h.ConfigService.GetByID(body.ID)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

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

	if existingConfig.Name == body.Name &&
		existingConfig.Description == body.Description &&
		IsDetailEqual {
		// No changes
		Success(c, true)
		return
	}

	conf := task.Config{
		ID:          body.ID,
		Name:        body.Name,
		Description: body.Description,
		Detail:      "", // This will be set in the Update method
	}

	if err := h.ConfigService.UpdateByID(body.ID, conf, detailStr); err != nil {
		Error(c, err)
		return
	}

	// Check for watching tasks that use this config and restart them
	tasks := h.Service.GetAll()
	for _, t := range tasks {
		if t.ConfigID == body.ID || (t.ConfigID == 0 && t.Config == existingConfig.Name) {
			if h.Service.IsWatching(t.Name) {
				go func(taskName string) {
					task.GetLogger(taskName)("WARN", fmt.Sprintf("⚠️ 正在重启监听: %s (配置-%d变更)\n", taskName, body.ID))
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
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ErrorMsg(c, "Config id is required")
		return
	}

	if err := h.ConfigService.Delete(id); err != nil {
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
		"config":        t.Config,   // config name (display)
		"configId":      t.ConfigID, // association id
		"isWatching":    h.Service.IsWatching(name),
	})
}

func (h *Handler) CreateTask(c *gin.Context) {
	var t task.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		Error(c, err)
		return
	}

	// Resolve config name by ID (association by ID)
	if t.ConfigID > 0 {
		if cfg, _, ok := h.ConfigService.GetByID(t.ConfigID); ok {
			t.Config = cfg.Name
		} else {
			ErrorMsg(c, "Config not found")
			return
		}
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

	// Resolve config name by ID (association by ID)
	if body.Task.ConfigID > 0 {
		if cfg, _, ok := h.ConfigService.GetByID(body.Task.ConfigID); ok {
			body.Task.Config = cfg.Name
		} else {
			ErrorMsg(c, "Config not found")
			return
		}
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

	opts, err := h.Service.GetOptions(name)
	if err != nil {
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

		if opts.Type == "prune" {
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

	logger := task.GetLogger(body.Name)

	// Start watcher with latest configs (resolved inside service)
	if err := h.Service.StartWatch(body.Name, logger); err != nil {
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
	taskName := c.Query("taskName")
	cacheStore := &cache.Store{}

	var jsonContent string
	var err error

	if taskName != "" {
		// Get cache for specific task
		files, err := cacheStore.GetByTaskName(taskName)
		if err != nil {
			ErrorMsg(c, fmt.Sprintf("读取缓存失败: %v", err))
			return
		}

		// Build JSON manually
		if len(files) == 0 {
			jsonContent = "[]"
		} else {
			var builder strings.Builder
			builder.WriteString("[\n")
			for i, file := range files {
				builder.WriteString("  \"")
				builder.WriteString(strings.ReplaceAll(file, "\"", "\\\""))
				builder.WriteString("\"")
				if i < len(files)-1 {
					builder.WriteString(",")
				}
				builder.WriteString("\n")
			}
			builder.WriteString("]")
			jsonContent = builder.String()
		}
	} else {
		// Get all cache (backward compatibility)
		jsonContent, err = cacheStore.GetAllAsJSON()
		if err != nil {
			ErrorMsg(c, fmt.Sprintf("读取缓存失败: %v", err))
			return
		}
	}

	Success(c, jsonContent)
}

func (h *Handler) UpdateCache(c *gin.Context) {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	cacheStore := &cache.Store{}

	// Update cache from JSON string
	if err := cacheStore.SetFromJSON(body.Content); err != nil {
		Error(c, err)
		return
	}

	Success(c, true)
}

func (h *Handler) GetCacheLog(c *gin.Context) {
	// Cache log is now stored in task_logs table
	// Return empty for backward compatibility
	Success(c, "")
}

func (h *Handler) ClearCacheLog(c *gin.Context) {
	// Cache log is now stored in task_logs table
	// Return success for backward compatibility
	Success(c, true)
}
