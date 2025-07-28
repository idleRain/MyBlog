// Package middleware RBAC权限控制中间件
package middleware

import (
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequirePermission 权限验证中间件
func RequirePermission(jwtService service.JWTService, userRepo repository.UserRepository, rbacService service.RBACService, permissions ...service.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查权限
		hasPermission := false
		for _, permission := range permissions {
			if rbacService.HasPermission(user.Role, permission) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, "权限不足，无法访问该资源")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireAllPermissions 要求拥有所有权限的中间件
func RequireAllPermissions(jwtService service.JWTService, userRepo repository.UserRepository, rbacService service.RBACService, permissions ...service.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查是否拥有所有权限
		if !rbacService.HasAllPermissions(user.Role, permissions...) {
			response.Forbidden(c, "权限不足，缺少必要权限")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireRoleLevel 要求最低角色级别的中间件
func RequireRoleLevel(jwtService service.JWTService, userRepo repository.UserRepository, rbacService service.RBACService, minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查角色级别
		if !rbacService.IsRoleHigherThan(user.Role, minRole) && user.Role != minRole {
			response.Forbidden(c, "权限不足，需要更高角色权限")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireSuperAdmin 要求超级管理员权限的中间件
func RequireSuperAdmin(jwtService service.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查是否为超级管理员
		if user.Role != string(service.RoleSuperAdmin) {
			response.Forbidden(c, "权限不足，需要超级管理员权限")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireAdminOrAbove 要求管理员或更高权限的中间件
func RequireAdminOrAbove(jwtService service.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查是否为管理员或超级管理员
		if !service.IsAdminRole(user.Role) {
			response.Forbidden(c, "权限不足，需要管理员权限")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireEditorOrAbove 要求编辑者或更高权限的中间件
func RequireEditorOrAbove(jwtService service.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 检查是否为编辑者或更高权限
		if !service.IsEditorOrAbove(user.Role) {
			response.Forbidden(c, "权限不足，需要编辑者或更高权限")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// RequireOwnershipOrAdmin 要求资源所有权或管理员权限的中间件
// 用于检查用户是否可以操作特定资源（自己的资源或管理员权限）
func RequireOwnershipOrAdmin(jwtService service.JWTService, userRepo repository.UserRepository, getResourceOwnerID func(*gin.Context) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 如果是管理员，直接通过
		if service.IsAdminRole(user.Role) {
			setUserContext(c, user)
			c.Next()
			return
		}

		// 检查资源所有权
		ownerID, err := getResourceOwnerID(c)
		if err != nil {
			response.InternalError(c, "无法验证资源所有权")
			c.Abort()
			return
		}

		if user.ID != ownerID {
			response.Forbidden(c, "权限不足，只能操作自己的资源")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// CanManageUserRole 检查是否可以管理指定角色用户的中间件
func CanManageUserRole(jwtService service.JWTService, userRepo repository.UserRepository, rbacService service.RBACService, getTargetRole func(*gin.Context) (string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证JWT令牌
		user, err := validateUserFromToken(c, jwtService, userRepo)
		if err != nil {
			return // 错误已在validateUserFromToken中处理
		}

		// 获取目标用户角色
		targetRole, err := getTargetRole(c)
		if err != nil {
			response.InternalError(c, "无法获取目标用户角色")
			c.Abort()
			return
		}

		// 检查是否可以管理该角色的用户
		if !rbacService.CanManageUser(user.Role, targetRole) {
			response.Forbidden(c, "权限不足，无法管理该角色的用户")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserContext(c, user)
		c.Next()
	}
}

// validateUserFromToken 从JWT令牌验证用户身份
func validateUserFromToken(c *gin.Context, jwtService service.JWTService, userRepo repository.UserRepository) (*repository.User, error) {
	token := c.GetHeader("Authorization")

	if token == "" {
		response.Unauthorized(c, "未提供认证令牌")
		c.Abort()
		return nil, fmt.Errorf("no token")
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
		return nil, err
	}

	// 从数据库查询用户信息
	user, err := userRepo.GetByID(claims.UserID)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		c.Abort()
		return nil, err
	}

	// 验证用户状态
	if user.Status != 1 {
		response.Forbidden(c, "用户已被禁用")
		c.Abort()
		return nil, fmt.Errorf("user disabled")
	}

	// 验证角色有效性
	rbacService := service.NewRBACService()
	if !rbacService.IsValidRole(user.Role) {
		response.Forbidden(c, "用户角色无效")
		c.Abort()
		return nil, fmt.Errorf("invalid role")
	}

	return user, nil
}

// setUserContext 设置用户上下文信息
func setUserContext(c *gin.Context, user *repository.User) {
	c.Set("userID", user.ID)
	c.Set("userRole", user.Role)
	c.Set("username", user.Username)
	c.Set("user", user)
	c.Set("isAdmin", service.IsAdminRole(user.Role))
	c.Set("isEditor", service.IsEditorOrAbove(user.Role))
}

// GetCurrentUser 从上下文获取当前用户信息
func GetCurrentUser(c *gin.Context) (*repository.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*repository.User); ok {
			return u, true
		}
	}
	return nil, false
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(uint); ok {
			return id, true
		}
	}
	return 0, false
}

// GetCurrentUserRole 从上下文获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	if userRole, exists := c.Get("userRole"); exists {
		if role, ok := userRole.(string); ok {
			return role, true
		}
	}
	return "", false
}
