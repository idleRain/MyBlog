// Package service RBAC权限管理服务
package service

// Role 角色定义
type Role string

const (
	// RoleSuperAdmin 超级管理员 - 拥有系统最高权限
	RoleSuperAdmin Role = "superadmin"
	// RoleAdmin 管理员 - 可管理内容和普通用户
	RoleAdmin Role = "admin"
	// RoleEditor 编辑者 - 可发布和管理文章
	RoleEditor Role = "editor"
	// RoleUser 普通用户 - 只读权限，评论
	RoleUser Role = "user"
)

// Permission 权限定义
type Permission string

const (
	// 系统管理权限
	PermissionSystemConfig Permission = "system:config" // 系统配置管理
	PermissionSystemLogs   Permission = "system:logs"   // 系统日志查看
	PermissionSystemStats  Permission = "system:stats"  // 系统统计信息

	// 用户管理权限
	PermissionUserCreate Permission = "user:create" // 创建用户
	PermissionUserRead   Permission = "user:read"   // 查看用户信息
	PermissionUserUpdate Permission = "user:update" // 更新用户信息
	PermissionUserDelete Permission = "user:delete" // 删除用户
	PermissionUserList   Permission = "user:list"   // 用户列表

	// 文章管理权限
	PermissionArticleCreate  Permission = "article:create"  // 创建文章
	PermissionArticleRead    Permission = "article:read"    // 查看文章
	PermissionArticleUpdate  Permission = "article:update"  // 更新文章
	PermissionArticleDelete  Permission = "article:delete"  // 删除文章
	PermissionArticleList    Permission = "article:list"    // 文章列表
	PermissionArticlePublish Permission = "article:publish" // 发布文章
	PermissionArticleManage  Permission = "article:manage"  // 管理所有文章

	// 分类标签权限
	PermissionCategoryManage Permission = "category:manage" // 分类管理
	PermissionTagManage      Permission = "tag:manage"      // 标签管理

	// 评论管理权限
	PermissionCommentCreate   Permission = "comment:create"   // 发表评论
	PermissionCommentRead     Permission = "comment:read"     // 查看评论
	PermissionCommentUpdate   Permission = "comment:update"   // 更新评论
	PermissionCommentDelete   Permission = "comment:delete"   // 删除评论
	PermissionCommentModerate Permission = "comment:moderate" // 评论审核

	// 文件管理权限
	PermissionFileUpload Permission = "file:upload" // 文件上传
	PermissionFileRead   Permission = "file:read"   // 文件查看
	PermissionFileDelete Permission = "file:delete" // 文件删除
)

// RolePermissions 角色权限映射
var RolePermissions = map[Role][]Permission{
	RoleSuperAdmin: {
		// 超级管理员拥有所有权限
		PermissionSystemConfig, PermissionSystemLogs, PermissionSystemStats,
		PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete, PermissionUserList,
		PermissionArticleCreate, PermissionArticleRead, PermissionArticleUpdate, PermissionArticleDelete, PermissionArticleList, PermissionArticlePublish, PermissionArticleManage,
		PermissionCategoryManage, PermissionTagManage,
		PermissionCommentCreate, PermissionCommentRead, PermissionCommentUpdate, PermissionCommentDelete, PermissionCommentModerate,
		PermissionFileUpload, PermissionFileRead, PermissionFileDelete,
	},
	RoleAdmin: {
		// 管理员权限（除系统配置外的大部分权限）
		PermissionSystemLogs, PermissionSystemStats,
		PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete, PermissionUserList,
		PermissionArticleCreate, PermissionArticleRead, PermissionArticleUpdate, PermissionArticleDelete, PermissionArticleList, PermissionArticlePublish, PermissionArticleManage,
		PermissionCategoryManage, PermissionTagManage,
		PermissionCommentCreate, PermissionCommentRead, PermissionCommentUpdate, PermissionCommentDelete, PermissionCommentModerate,
		PermissionFileUpload, PermissionFileRead, PermissionFileDelete,
	},
	RoleEditor: {
		// 编辑者权限（主要是内容管理）
		PermissionUserRead, // 可以查看用户信息但不能管理
		PermissionArticleCreate, PermissionArticleRead, PermissionArticleUpdate, PermissionArticleDelete, PermissionArticleList, PermissionArticlePublish,
		PermissionCommentCreate, PermissionCommentRead, PermissionCommentUpdate, PermissionCommentDelete,
		PermissionFileUpload, PermissionFileRead,
	},
	RoleUser: {
		// 普通用户权限（基础读写权限）
		PermissionArticleRead,
		PermissionCommentCreate, PermissionCommentRead, PermissionCommentUpdate, // 只能更新自己的评论
	},
}

// RoleHierarchy 角色层级定义（数字越大权限越高）
var RoleHierarchy = map[Role]int{
	RoleUser:       1,
	RoleEditor:     2,
	RoleAdmin:      3,
	RoleSuperAdmin: 4,
}

// RBACService RBAC权限管理服务接口
type RBACService interface {
	// HasPermission 检查用户是否有指定权限
	HasPermission(userRole string, permission Permission) bool
	// HasAnyPermission 检查用户是否有任意一个权限
	HasAnyPermission(userRole string, permissions ...Permission) bool
	// HasAllPermissions 检查用户是否有所有权限
	HasAllPermissions(userRole string, permissions ...Permission) bool
	// GetUserPermissions 获取用户的所有权限
	GetUserPermissions(userRole string) []Permission
	// IsRoleHigherThan 检查角色A是否比角色B权限更高
	IsRoleHigherThan(roleA, roleB string) bool
	// IsValidRole 检查是否为有效角色
	IsValidRole(role string) bool
	// CanManageUser 检查是否可以管理指定角色的用户
	CanManageUser(managerRole, targetRole string) bool
}

type rbacService struct{}

// NewRBACService 创建RBAC权限管理服务实例
func NewRBACService() RBACService {
	return &rbacService{}
}

// HasPermission 检查用户是否有指定权限
func (s *rbacService) HasPermission(userRole string, permission Permission) bool {
	role := Role(userRole)
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查用户是否有任意一个权限
func (s *rbacService) HasAnyPermission(userRole string, permissions ...Permission) bool {
	for _, permission := range permissions {
		if s.HasPermission(userRole, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查用户是否有所有权限
func (s *rbacService) HasAllPermissions(userRole string, permissions ...Permission) bool {
	for _, permission := range permissions {
		if !s.HasPermission(userRole, permission) {
			return false
		}
	}
	return true
}

// GetUserPermissions 获取用户的所有权限
func (s *rbacService) GetUserPermissions(userRole string) []Permission {
	role := Role(userRole)
	permissions, exists := RolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// IsRoleHigherThan 检查角色A是否比角色B权限更高
func (s *rbacService) IsRoleHigherThan(roleA, roleB string) bool {
	levelA, existsA := RoleHierarchy[Role(roleA)]
	levelB, existsB := RoleHierarchy[Role(roleB)]

	if !existsA || !existsB {
		return false
	}

	return levelA > levelB
}

// IsValidRole 检查是否为有效角色
func (s *rbacService) IsValidRole(role string) bool {
	_, exists := RoleHierarchy[Role(role)]
	return exists
}

// CanManageUser 检查是否可以管理指定角色的用户
func (s *rbacService) CanManageUser(managerRole, targetRole string) bool {
	// 超级管理员可以管理所有用户
	if managerRole == string(RoleSuperAdmin) {
		return true
	}

	// 管理员可以管理编辑者和普通用户，但不能管理超级管理员和其他管理员
	if managerRole == string(RoleAdmin) {
		return targetRole == string(RoleEditor) || targetRole == string(RoleUser)
	}

	// 其他角色不能管理用户
	return false
}

// GetRoleDisplayName 获取角色显示名称
func GetRoleDisplayName(role string) string {
	switch Role(role) {
	case RoleSuperAdmin:
		return "超级管理员"
	case RoleAdmin:
		return "管理员"
	case RoleEditor:
		return "编辑者"
	case RoleUser:
		return "用户"
	default:
		return "未知角色"
	}
}

// GetAllRoles 获取所有可用角色
func GetAllRoles() []Role {
	return []Role{RoleUser, RoleEditor, RoleAdmin, RoleSuperAdmin}
}

// IsAdminRole 检查是否为管理员级别角色
func IsAdminRole(role string) bool {
	return role == string(RoleAdmin) || role == string(RoleSuperAdmin)
}

// IsEditorOrAbove 检查是否为编辑者及以上角色
func IsEditorOrAbove(role string) bool {
	level, exists := RoleHierarchy[Role(role)]
	if !exists {
		return false
	}
	return level >= RoleHierarchy[RoleEditor]
}
