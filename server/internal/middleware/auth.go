package middleware

import (
	"MyBlog/internal/repository"
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
func AdminAuth(jwtService service.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
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

		// 从数据库查询用户信息验证管理员权限
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}

		// 验证用户状态
		if user.Status != 1 {
			response.Forbidden(c, "用户已被禁用")
			c.Abort()
			return
		}

		// 验证管理员权限
		if !isAdminRole(user.Role) {
			response.Forbidden(c, "权限不足，需要管理员权限")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", user.Role)
		c.Set("isAdmin", true)

		c.Next()
	}
}

// isAdminRole 检查角色是否为管理员级别
func isAdminRole(role string) bool {
	return role == "admin" || role == "superadmin"
}

// RequireRole 角色验证中间件
func RequireRole(jwtService service.JWTService, userRepo repository.UserRepository, allowedRoles ...string) gin.HandlerFunc {
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

		// 从数据库查询用户信息
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}

		// 验证用户状态
		if user.Status != 1 {
			response.Forbidden(c, "用户已被禁用")
			c.Abort()
			return
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

		c.Set("userID", claims.UserID)
		c.Set("userRole", user.Role)

		c.Next()
	}
}
