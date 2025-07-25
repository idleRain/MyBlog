package middleware

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 移除 Bearer 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 验证访问令牌
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("userID", claims.UserID)
		// username已从 JWT 中移除，如需使用请从数据库查询

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "" {
			// 移除 Bearer 前缀
			if strings.HasPrefix(token, "Bearer ") {
				token = token[7:]
			}

			// 验证访问令牌
			if claims, err := jwtService.ValidateAccessToken(token); err == nil {
				c.Set("userID", claims.UserID)
				// username已从 JWT 中移除
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先验证基本认证
		token := c.GetHeader("Authorization")

		if token == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 验证管理员权限
		if !isAdmin(claims.UserID) {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		// username已从 JWT 中移除
		c.Set("isAdmin", true)

		c.Next()
	}
}

// isAdmin 检查是否为管理员（简单实现，实际应该从数据库查询用户角色）
func isAdmin(userID uint) bool {
	// 这里可以根据用户ID查询数据库确定是否为管理员
	// 暂时简单判断，实际应该有更完善的权限系统
	return userID == 1 // 假设 ID为1的用户是管理员
}
