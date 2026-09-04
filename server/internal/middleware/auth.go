package middleware

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// bearerTokenPrefix 认证令牌的 Bearer 前缀，用于从请求头解析令牌
const bearerTokenPrefix = "Bearer "

// Auth 认证中间件
func Auth(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 移除 Bearer 前缀，无前缀时保持原值
		token = strings.TrimPrefix(token, bearerTokenPrefix)

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
			// 移除 Bearer 前缀，无前缀时保持原值
			token = strings.TrimPrefix(token, bearerTokenPrefix)

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
func AdminAuth(identity IdentityProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := identity.Resolve(c)
		if err != nil {
			return // 错误已在 Resolve 中处理
		}

		// 验证管理员权限
		if !service.IsAdminRole(user.Role) {
			response.Forbidden(c, "权限不足，需要管理员权限")
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.Set("isAdmin", true)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func RequireRole(identity IdentityProvider, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := identity.Resolve(c)
		if err != nil {
			return // 错误已在 Resolve 中处理
		}

		// 验证角色权限
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if user.Role == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}
