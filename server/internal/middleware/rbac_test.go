package middleware

import (
	"errors"
	"testing"

	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// TestRequirePermission 表驱动验证权限中间件的放行与拦截分支。
// 统一响应信封固定 HTTP 200，权限拦截以业务码 403 表达，故断言响应体 code 字段。
func TestRequirePermission(t *testing.T) {
	testCases := []struct {
		name         string
		identity     IdentityProvider
		rbac         service.RBACService
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "身份解析失败时链路中止",
			identity:     &fakeIdentityProvider{err: errors.New("no token")},
			rbac:         &fakeRBACService{hasPermission: true},
			expectCalled: false,
		},
		{
			name:         "拥有权限时放行",
			identity:     &fakeIdentityProvider{user: newTestUser("user", 1)},
			rbac:         &fakeRBACService{hasPermission: true},
			expectCalled: true,
		},
		{
			name:         "缺少权限时返回 403",
			identity:     &fakeIdentityProvider{user: newTestUser("user", 1)},
			rbac:         &fakeRBACService{hasPermission: false},
			expectCalled: false,
			expectCode:   response.CodeForbid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			router := newMiddlewareRouter(RequirePermission(tc.identity, tc.rbac, service.PermissionUserList), &handlerCalled)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if tc.expectCode != 0 {
				if code := responseCode(t, recorder); code != tc.expectCode {
					t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
				}
			}
		})
	}
}

// TestRequireAllPermissions 表驱动验证全权限中间件。
func TestRequireAllPermissions(t *testing.T) {
	testCases := []struct {
		name         string
		rbac         service.RBACService
		expectCalled bool
	}{
		{
			name:         "拥有全部权限时放行",
			rbac:         &fakeRBACService{hasAll: true},
			expectCalled: true,
		},
		{
			name:         "缺少任一权限时返回 403",
			rbac:         &fakeRBACService{hasAll: false},
			expectCalled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			router := newMiddlewareRouter(
				RequireAllPermissions(&fakeIdentityProvider{user: newTestUser("admin", 1)}, tc.rbac, service.PermissionUserRead, service.PermissionUserUpdate),
				&handlerCalled,
			)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if !tc.expectCalled {
				if code := responseCode(t, recorder); code != response.CodeForbid {
					t.Errorf("拦截时响应业务码 = %d，期望 %d", code, response.CodeForbid)
				}
			}
		})
	}
}

// TestRequireRoleLevel 表驱动验证最低角色级别中间件。
func TestRequireRoleLevel(t *testing.T) {
	testCases := []struct {
		name         string
		roleHigher   bool
		userRole     string
		expectCalled bool
	}{
		{
			name:         "角色级别满足要求时放行",
			roleHigher:   true,
			userRole:     "admin",
			expectCalled: true,
		},
		{
			name:         "角色级别不足时返回 403",
			roleHigher:   false,
			userRole:     "user",
			expectCalled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			router := newMiddlewareRouter(
				RequireRoleLevel(&fakeIdentityProvider{user: newTestUser(tc.userRole, 1)}, &fakeRBACService{roleHigher: tc.roleHigher}, "editor"),
				&handlerCalled,
			)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if !tc.expectCalled {
				if code := responseCode(t, recorder); code != response.CodeForbid {
					t.Errorf("拦截时响应业务码 = %d，期望 %d", code, response.CodeForbid)
				}
			}
		})
	}
}

// TestRoleLevelGates 表驱动验证按角色层级放行的三个中间件。
// 角色分别为 superadmin/admin/editor/user，覆盖放行与拦截的层级边界。
func TestRoleLevelGates(t *testing.T) {
	testCases := []struct {
		name         string
		build        func(identity IdentityProvider) gin.HandlerFunc
		userRole     string
		expectCalled bool
	}{
		{
			name:         "超级管理员中间件放行 superadmin",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireSuperAdmin(identity) },
			userRole:     "superadmin",
			expectCalled: true,
		},
		{
			name:         "超级管理员中间件拦截 admin",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireSuperAdmin(identity) },
			userRole:     "admin",
			expectCalled: false,
		},
		{
			name:         "管理员及以上中间件放行 admin",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireAdminOrAbove(identity) },
			userRole:     "admin",
			expectCalled: true,
		},
		{
			name:         "管理员及以上中间件放行 superadmin",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireAdminOrAbove(identity) },
			userRole:     "superadmin",
			expectCalled: true,
		},
		{
			name:         "管理员及以上中间件拦截 editor",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireAdminOrAbove(identity) },
			userRole:     "editor",
			expectCalled: false,
		},
		{
			name:         "编辑及以上中间件放行 editor",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireEditorOrAbove(identity) },
			userRole:     "editor",
			expectCalled: true,
		},
		{
			name:         "编辑及以上中间件拦截 user",
			build:        func(identity IdentityProvider) gin.HandlerFunc { return RequireEditorOrAbove(identity) },
			userRole:     "user",
			expectCalled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 角色层级中间件仅依赖包级角色判定函数，不注入 RBACService 实例。
			handlerCalled := false
			router := newMiddlewareRouter(tc.build(&fakeIdentityProvider{user: newTestUser(tc.userRole, 1)}), &handlerCalled)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if !tc.expectCalled {
				if code := responseCode(t, recorder); code != response.CodeForbid {
					t.Errorf("拦截时响应业务码 = %d，期望 %d", code, response.CodeForbid)
				}
			}
		})
	}
}

// TestRequireOwnershipOrAdmin 表驱动验证资源所有权或管理员放行中间件。
func TestRequireOwnershipOrAdmin(t *testing.T) {
	resourceOwnerID := uint(42)

	testCases := []struct {
		name         string
		userRole     string
		userID       uint
		ownerErr     error
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "管理员直接放行",
			userRole:     "admin",
			userID:       1,
			expectCalled: true,
		},
		{
			name:         "资源所有者放行",
			userRole:     "user",
			userID:       resourceOwnerID,
			expectCalled: true,
		},
		{
			name:         "非所有者返回 403",
			userRole:     "user",
			userID:       1,
			expectCalled: false,
			expectCode:   response.CodeForbid,
		},
		{
			name:         "所有权查询失败返回 500",
			userRole:     "user",
			userID:       1,
			ownerErr:     errors.New("owner query failed"),
			expectCalled: false,
			expectCode:   response.CodeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := newTestUser(tc.userRole, 1)
			user.ID = tc.userID
			handlerCalled := false
			middleware := RequireOwnershipOrAdmin(&fakeIdentityProvider{user: user}, func(c *gin.Context) (uint, error) {
				return resourceOwnerID, tc.ownerErr
			})
			router := newMiddlewareRouter(middleware, &handlerCalled)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if tc.expectCode != 0 {
				if code := responseCode(t, recorder); code != tc.expectCode {
					t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
				}
			}
		})
	}
}

// TestCanManageUserRole 表驱动验证角色管理权限中间件。
func TestCanManageUserRole(t *testing.T) {
	testCases := []struct {
		name         string
		canManage    bool
		targetErr    error
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "具备管理目标角色权限时放行",
			canManage:    true,
			expectCalled: true,
		},
		{
			name:         "不具备管理权限时返回 403",
			canManage:    false,
			expectCalled: false,
			expectCode:   response.CodeForbid,
		},
		{
			name:         "目标角色读取失败返回 500",
			targetErr:    errors.New("target role read failed"),
			expectCalled: false,
			expectCode:   response.CodeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			middleware := CanManageUserRole(
				&fakeIdentityProvider{user: newTestUser("admin", 1)},
				&fakeRBACService{canManage: tc.canManage},
				func(c *gin.Context) (string, error) { return "editor", tc.targetErr },
			)
			router := newMiddlewareRouter(middleware, &handlerCalled)
			recorder := performTestRequest(router)

			if handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectCalled)
			}
			if tc.expectCode != 0 {
				if code := responseCode(t, recorder); code != tc.expectCode {
					t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
				}
			}
		})
	}
}
