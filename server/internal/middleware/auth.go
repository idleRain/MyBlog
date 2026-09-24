package middleware

import (
	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// bearerTokenPrefix 认证令牌的 Bearer 前缀，用于从请求头解析令牌
const bearerTokenPrefix = "Bearer "

// resolveAccessToken 依序从 Authorization 头与会话 Cookie 读取访问令牌。
// 双轨过渡期两个通道并存，Header 优先兼容存量客户端，Cookie 是浏览器的默认通道。
func resolveAccessToken(c *gin.Context) string {
	headerToken := strings.TrimPrefix(c.GetHeader("Authorization"), bearerTokenPrefix)
	if headerToken != "" {
		return headerToken
	}

	cookieToken, err := c.Cookie(domain.SessionAccessTokenCookie)
	if err != nil {
		return ""
	}
	return cookieToken
}

// Auth 认证中间件
func Auth(tokenService service.TokenServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := resolveAccessToken(c)

		if token == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 校验访问令牌
		identity, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("userID", identity.UserID)
		// username 已从令牌中移除，如需使用请从数据库查询

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth(tokenService service.TokenServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := resolveAccessToken(c)

		if token != "" {
			// 校验访问令牌
			if identity, err := tokenService.ValidateAccessToken(token); err == nil {
				c.Set("userID", identity.UserID)
				// username 已从令牌中移除，如需使用请从数据库查询
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
