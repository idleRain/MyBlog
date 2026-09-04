package config

import (
	"testing"
)

// TestLoadParsesRBACSection 验证 config.yaml 的 rbac 节可被正确解析并通过校验。
func TestLoadParsesRBACSection(t *testing.T) {
	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.RBAC.RoleHierarchy["superadmin"] != 4 {
		t.Error("superadmin 层级应为 4")
	}
	if len(cfg.RBAC.RolePermissions["superadmin"]) == 0 {
		t.Error("superadmin 应配置至少一条权限")
	}
	if len(cfg.RBAC.RolePermissions["user"]) == 0 {
		t.Error("user 角色应配置基础权限")
	}
}
