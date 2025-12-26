package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// System
	r.GET("/api/version", h.Version)
	r.GET("/api/update", h.Update)

	// Define /api group first
	api := r.Group("/api")

	// System (move under api group if needed, but previously defined as absolute paths on r)
	// r.GET("/api/version") works fine too. but consistency.
	// Let's stick with existing system defines for now.

	// Config
	config := api.Group("/config")
	{
		config.GET("/list", h.GetConfigList)
		config.GET("/default", h.GetConfigDefault)
		config.GET("/", h.GetConfig)
		config.GET("/detail", h.GetConfigDetail)
		config.POST("/", h.AddConfig)
		config.PUT("/", h.UpdateConfig)
		config.DELETE("/", h.DeleteConfig)
	}

	// Task
	t := api.Group("/task")
	{
		t.GET("/list", h.GetTaskList)
		t.GET("/", h.GetTask)
		t.POST("/", h.CreateTask)
		t.PUT("/", h.UpdateTask)
		t.DELETE("/", h.DeleteTask)

		t.GET("/run", h.RunTask)

		t.POST("/watch/start", h.StartWatch)
		t.POST("/watch/stop", h.StopWatch)
		t.GET("/watch/status", h.GetWatchStatus)

		t.GET("/log", h.GetTaskLog)
		t.DELETE("/log", h.ClearTaskLog)
	}

	// Cache
	cache := api.Group("/cache")
	{
		cache.GET("/", h.GetCache)
		cache.PUT("/", h.UpdateCache)
		cache.GET("/log", h.GetCacheLog)
		cache.DELETE("/log", h.ClearCacheLog)
	}

	return r
}
