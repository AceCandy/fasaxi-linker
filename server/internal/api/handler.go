package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/fasaxi-linker/servergo/internal/cache"
	"github.com/fasaxi-linker/servergo/internal/config"
	"github.com/fasaxi-linker/servergo/internal/logs"
	"github.com/fasaxi-linker/servergo/internal/task"
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
		"id":     conf.ID,
		"name":   conf.Name,
		"detail": detail,
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
		ID     int         `json:"id"`
		Name   string      `json:"name"`
		Detail interface{} `json:"detail"`
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
		Name:   body.Name,
		Detail: "", // This will be set in the Add method
	}

	if err := h.ConfigService.Add(conf, detailStr); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	var body struct {
		ID     int         `json:"id"`
		Name   string      `json:"name"`
		Detail interface{} `json:"detail"`
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

	if existingConfig.Name == body.Name && IsDetailEqual {
		// No changes
		Success(c, true)
		return
	}

	conf := task.Config{
		ID:     body.ID,
		Name:   body.Name,
		Detail: "", // This will be set in the Update method
	}

	if err := h.ConfigService.UpdateByID(body.ID, conf, detailStr); err != nil {
		Error(c, err)
		return
	}

	// Sync config name and fields to all related tasks
	if err := h.Service.SyncConfigToTasks(body.ID, body.Name, detailStr); err != nil {
		fmt.Printf("Warning: Failed to sync config to tasks: %v\n", err)
		// Don't fail the request, just log the warning
	}

	// Check for watching tasks that use this config and restart them
	tasks := h.Service.GetAll()
	for _, t := range tasks {
		if t.ConfigID == body.ID || (t.ConfigID == 0 && t.Config == existingConfig.Name) {
			if h.Service.IsWatching(t.ID) {
				go func(taskID int, taskName string) {
					task.GetLogger(taskID)("WARN", fmt.Sprintf("⚠️ 正在重启监听: %s (配置-%d变更)\n", taskName, body.ID))
					if err := h.Service.RestartWatch(taskID); err != nil {
						fmt.Printf("Failed to restart task %s: %v\n", taskName, err)
					}
				}(t.ID, t.Name)
			}
		}
	}

	Success(c, true)
}

func (h *Handler) GetConfigRelatedTasks(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ErrorMsg(c, "Config id is required")
		return
	}

	// Get config to check if it exists
	_, _, ok := h.ConfigService.GetByID(id)
	if !ok {
		ErrorMsg(c, "Config not found")
		return
	}

	// Find all tasks that use this config
	tasks := h.Service.GetAll()
	var relatedTasks []string
	for _, t := range tasks {
		if t.ConfigID == id {
			relatedTasks = append(relatedTasks, t.Name)
		}
	}

	Success(c, relatedTasks)
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
			IsWatching: h.Service.IsWatching(t.ID),
		}
	}

	Success(c, result)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	t, ok := h.Service.Get(taskID)
	if !ok {
		ErrorMsg(c, "Task not found")
		return
	}
	Success(c, gin.H{ // Frontend expects mixed object
		"id":            t.ID,
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
		"isWatching":    h.Service.IsWatching(taskID),
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
		if cfg, detail, ok := h.ConfigService.GetByID(t.ConfigID); ok {
			t.Config = cfg.Name
			// Sync config fields to task on creation
			var rc struct {
				Include       []string `json:"include"`
				Exclude       []string `json:"exclude"`
				KeepDirStruct bool     `json:"keepDirStruct"`
				OpenCache     bool     `json:"openCache"`
				MkdirIfSingle bool     `json:"mkdirIfSingle"`
				DeleteDir     bool     `json:"deleteDir"`
			}
			if err := json.Unmarshal([]byte(detail), &rc); err == nil {
				t.Include = rc.Include
				t.Exclude = rc.Exclude
				t.KeepDirStruct = rc.KeepDirStruct
				t.OpenCache = rc.OpenCache
				t.MkdirIfSingle = rc.MkdirIfSingle
				t.DeleteDir = rc.DeleteDir
			}
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
		TaskID int `json:"taskId"`
		task.Task
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	if body.TaskID <= 0 {
		ErrorMsg(c, "taskId is required")
		return
	}

	// Resolve config name by ID (association by ID)
	if body.Task.ConfigID > 0 {
		if cfg, detail, ok := h.ConfigService.GetByID(body.Task.ConfigID); ok {
			body.Task.Config = cfg.Name
			// Sync config fields to task when config ID changes
			var rc struct {
				Include       []string `json:"include"`
				Exclude       []string `json:"exclude"`
				KeepDirStruct bool     `json:"keepDirStruct"`
				OpenCache     bool     `json:"openCache"`
				MkdirIfSingle bool     `json:"mkdirIfSingle"`
				DeleteDir     bool     `json:"deleteDir"`
			}
			if err := json.Unmarshal([]byte(detail), &rc); err == nil {
				body.Task.Include = rc.Include
				body.Task.Exclude = rc.Exclude
				body.Task.KeepDirStruct = rc.KeepDirStruct
				body.Task.OpenCache = rc.OpenCache
				body.Task.MkdirIfSingle = rc.MkdirIfSingle
				body.Task.DeleteDir = rc.DeleteDir
			}
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
	existingTask, ok := h.Service.Get(body.TaskID)
	if !ok {
		ErrorMsg(c, "Task not found")
		return
	}

	// Dirty check: If nothing changed, return success immediately
	body.Task.ID = existingTask.ID
	if existingTask.Name == body.Task.Name {
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
	wasWatching := h.Service.IsWatching(body.TaskID)
	if wasWatching {
		// Stop the old watcher (especially important if name changes)
		// We use a temporary log to indicate restart
		task.GetLogger(body.TaskID)("WARN", fmt.Sprintf("⚠️ 正在重启监听:%s (任务变更)\n", existingTask.Name))
		if err := h.Service.StopWatch(body.TaskID); err != nil {
			fmt.Printf("Warning: Failed to stop watcher for update: %v\n", err)
		}
	}

	if err := h.Service.Update(body.TaskID, body.Task); err != nil {
		Error(c, err)
		return
	}

	// Restart watcher if it was running
	if wasWatching {
		// Start watcher with new name (in case it changed)
		go func(taskID int) {
			logger := task.GetLogger(taskID)
			// Small delay to ensure DB save completes? Usually not needed with locks.
			if err := h.Service.StartWatch(taskID, logger); err != nil {
				errMsg := fmt.Sprintf("任务更新后重启监听失败: %v", err)
				logger("ERROR", errMsg)
				fmt.Println(errMsg)
			}
		}(body.TaskID)
	}
	Success(c, true)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	if err := h.Service.Delete(taskID); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

// === Run (Async) ===

func (h *Handler) RunTask(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "taskId parameter is required"})
		return
	}

	opts, err := h.Service.GetOptions(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Task not found"})
		return
	}

	// Check if already running
	if task.IsRunning(taskID) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务正在执行中", "running": true})
		return
	}

	// Start async execution
	if err := task.StartRun(taskID, opts); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "任务已开始执行", "running": true})
}

func (h *Handler) StopRun(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "taskId parameter is required"})
		return
	}

	if err := task.StopRun(taskID); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "任务已停止"})
}

func (h *Handler) GetRunStatus(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "taskId parameter is required"})
		return
	}

	running := task.IsRunning(taskID)
	c.JSON(http.StatusOK, gin.H{"running": running})
}

// === Watch ===

func (h *Handler) StartWatch(c *gin.Context) {
	var body struct {
		TaskID int `json:"taskId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	if body.TaskID <= 0 {
		ErrorMsg(c, "taskId is required")
		return
	}

	logger := task.GetLogger(body.TaskID)

	// Start watcher with latest configs (resolved inside service)
	if err := h.Service.StartWatch(body.TaskID, logger); err != nil {
		Error(c, err)
		return
	}
	Success(c, true)
}

func (h *Handler) StopWatch(c *gin.Context) {
	var body struct {
		TaskID int `json:"taskId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, err)
		return
	}

	if body.TaskID <= 0 {
		ErrorMsg(c, "taskId is required")
		return
	}

	if err := h.Service.StopWatch(body.TaskID); err != nil {
		Error(c, err)
		return
	}

	// Add log entry for stop in a goroutine to avoid blocking
	taskName := fmt.Sprintf("%d", body.TaskID)
	if t, ok := h.Service.Get(body.TaskID); ok {
		taskName = t.Name
	}
	go func(taskID int, name string) {
		task.GetLogger(taskID)("WARN", fmt.Sprintf("⛔ [%s] 监听服务已停止", name))
	}(body.TaskID, taskName)

	Success(c, true)
}

func (h *Handler) GetWatchStatus(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}
	Success(c, h.Service.IsWatching(taskID))
}

func (h *Handler) GetLogFiles(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	files, err := task.GetLogFiles(taskID)
	if err != nil {
		ErrorMsg(c, fmt.Sprintf("读取日志文件列表失败: %v", err))
		return
	}

	Success(c, files)
}

func (h *Handler) GetTaskLog(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	filename := c.Query("file")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "200")
	levelFilter := c.Query("level")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 200
	}
	if pageSize > 1000 {
		pageSize = 1000 // Limit max page size
	}

	// Return structured log entries
	logEntries, total, err := task.GetLogEntries(taskID, filename, page, pageSize, levelFilter, search)
	if err != nil {
		ErrorMsg(c, fmt.Sprintf("读取日志失败: %v", err))
		return
	}

	if logEntries == nil {
		logEntries = []logs.LogEntry{} // Return empty array instead of null
	}

	Success(c, gin.H{
		"list":  logEntries,
		"total": total,
		"file":  filename,
	})
}

func (h *Handler) ClearTaskLog(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}
	filename := c.Query("file")

	if err := task.ClearLog(taskID, filename); err != nil {
		ErrorMsg(c, fmt.Sprintf("清空日志失败: %v", err))
		return
	}
	Success(c, true)
}

// === Cache ===

func (h *Handler) GetCache(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	cacheStore := &cache.Store{}

	if _, ok := h.Service.Get(taskID); !ok {
		ErrorMsg(c, "任务不存在")
		return
	}

	// Get cache for specific task with pagination
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	files, total, err := cacheStore.GetByTaskIDPaged(taskID, page, pageSize, search)
	if err != nil {
		ErrorMsg(c, fmt.Sprintf("读取缓存失败: %v", err))
		return
	}

	// Return JSON object with list and total
	if files == nil {
		files = []cache.CacheEntry{}
	}

	Success(c, gin.H{
		"list":  files,
		"total": total,
	})
}

func (h *Handler) UpdateCache(c *gin.Context) {
	ErrorMsg(c, "UpdateCache API is deprecated and no longer supported")
}

func (h *Handler) DeleteCache(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	// Get files from query string (supports multiple values: ?files=a&files=b)
	files := c.QueryArray("files")
	if len(files) == 0 {
		ErrorMsg(c, "files parameter is required")
		return
	}

	if err := h.Service.RemoveCache(taskID, files); err != nil {
		ErrorMsg(c, fmt.Sprintf("删除缓存失败: %v", err))
		return
	}

	Success(c, true)
}

func (h *Handler) GetCacheLog(c *gin.Context) {
	// Deprecated: Cache logs are now part of task logs (file-based)
	// Return empty for backward compatibility
	Success(c, "")
}

func (h *Handler) ClearCacheLog(c *gin.Context) {
	// Deprecated: Cache logs are now part of task logs (file-based)
	// Return success for backward compatibility
	Success(c, true)
}

func (h *Handler) ClearTaskCache(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		ErrorMsg(c, "taskId parameter is required")
		return
	}

	if err := h.Service.ClearCache(taskID); err != nil {
		ErrorMsg(c, fmt.Sprintf("清空缓存失败: %v", err))
		return
	}

	Success(c, true)
}
