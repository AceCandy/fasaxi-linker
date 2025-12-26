package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	c.JSON(http.StatusOK, APIResponse{ // Frontend likely expects 200 with success:false? Fetch wrapper throws if !success.
		Success:      false,
		ErrorMessage: err.Error(),
	})
}

func ErrorMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, APIResponse{
		Success:      false,
		ErrorMessage: msg,
	})
}
