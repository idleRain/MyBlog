package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 移除 Bearer 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 验证令牌（这里是示例，实际应该验证 JWT 或其他令牌）
		if !validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		userID := getUserIDFromToken(token)
		c.Set("userID", userID)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "" {
			// 移除 Bearer 前缀
			if strings.HasPrefix(token, "Bearer ") {
				token = token[7:]
			}

			// 验证令牌
			if validateToken(token) {
				userID := getUserIDFromToken(token)
				c.Set("userID", userID)
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先验证基本认证
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		if !validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 验证管理员权限
		if !isAdmin(token) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		userID := getUserIDFromToken(token)
		c.Set("userID", userID)
		c.Set("isAdmin", true)

		c.Next()
	}
}

// validateToken 验证令牌（示例实现）
func validateToken(token string) bool {
	// TODO: 实现真实的令牌验证逻辑
	// 这里只是示例，实际应该验证 JWT 或查询数据库
	return token != ""
}

// getUserIDFromToken 从令牌中获取用户ID（示例实现）
func getUserIDFromToken(token string) uint {
	// TODO: 实现真实的用户ID提取逻辑
	// 这里只是示例
	return 1
}

// isAdmin 检查是否为管理员（示例实现）
func isAdmin(token string) bool {
	// TODO: 实现真实的管理员验证逻辑
	// 这里只是示例
	return false
}
