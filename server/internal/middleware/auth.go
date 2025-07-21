package middleware

import (
	"MyBlog/internal/service"
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

// validateToken 验证JWT令牌
func validateToken(token string) bool {
	_, err := service.ValidateToken(token)
	return err == nil
}

// getUserIDFromToken 从JWT令牌中获取用户ID
func getUserIDFromToken(token string) uint {
	claims, err := service.ValidateToken(token)
	if err != nil {
		return 0
	}
	return claims.UserID
}

// isAdmin 检查是否为管理员（简单实现，实际应该从数据库查询用户角色）
func isAdmin(token string) bool {
	claims, err := service.ValidateToken(token)
	if err != nil {
		return false
	}
	// 这里可以根据用户ID查询数据库确定是否为管理员
	// 暂时简单判断，实际应该有更完善的权限系统
	return claims.UserID == 1 // 假设ID为1的用户是管理员
}
