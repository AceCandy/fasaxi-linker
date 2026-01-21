package api

import (
	"net/http"
	"os"
	"path/filepath"

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
		config.GET("/related-tasks", h.GetConfigRelatedTasks)
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

	// Serve static files (frontend)
	staticPath := os.Getenv("STATIC_PATH")
	if staticPath == "" {
		staticPath = "./static"
	}

	// Check if static directory exists
	if _, err := os.Stat(staticPath); err == nil {
		// Serve static assets (js, css, images, etc.)
		r.Static("/assets", filepath.Join(staticPath, "assets"))

		// Serve index.html for all non-API routes (SPA fallback)
		r.NoRoute(func(c *gin.Context) {
			// Skip API routes
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}
			c.File(filepath.Join(staticPath, "index.html"))
		})
	}

	return r
}
