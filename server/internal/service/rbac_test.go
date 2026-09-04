package service

import (
	"testing"
)

// TestLoadRBACConfigOverridesDefaults 验证配置加载覆盖默认权限表，并保持权限判定生效。
func TestLoadRBACConfigOverridesDefaults(t *testing.T) {
	savedPermissions := RolePermissions
	savedHierarchy := RoleHierarchy
	defer func() {
		RolePermissions = savedPermissions
		RoleHierarchy = savedHierarchy
	}()

	hierarchy := map[string]int{
		"superadmin": 4,
		"admin":      3,
		"editor":     2,
		"user":       1,
	}
	permissions := map[string][]string{
		"superadmin": {"system:config", "article:read"},
		"user":       {"article:read"},
	}

	LoadRBACConfig(hierarchy, permissions)

	svc := NewRBACService()
	if !svc.HasPermission("superadmin", PermissionSystemConfig) {
		t.Error("配置加载后超级管理员应拥有 system:config 权限")
	}
	if svc.HasPermission("user", PermissionSystemConfig) {
		t.Error("普通用户不应拥有 system:config 权限")
	}
	if !svc.HasPermission("user", PermissionArticleRead) {
		t.Error("user 角色配置了 article:read，应命中")
	}
	if !svc.IsValidRole("editor") {
		t.Error("editor 角色应在层级配置中有效")
	}
}
