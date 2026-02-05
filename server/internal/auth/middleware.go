package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID 用户 ID 上下文键
	ContextKeyUserID = "userId"
	// ContextKeyUsername 用户名上下文键
	ContextKeyUsername = "username"
)

// AuthMiddleware 创建认证中间件
func AuthMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":      false,
				"errorMessage": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":      false,
				"errorMessage": "无效的认证格式，请使用 Bearer token",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			message := "认证失败"
			if err == ErrTokenExpired {
				message = "认证令牌已过期，请重新登录"
			} else if err == ErrInvalidToken {
				message = "无效的认证令牌"
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":      false,
				"errorMessage": message,
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)

		c.Next()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	id, ok := userID.(int)
	return id, ok
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(ContextKeyUsername)
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}
