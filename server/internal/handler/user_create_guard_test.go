package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// operatorRoleContextKey 认证中间件写入操作者角色的上下文键。
const operatorRoleContextKey = "userRole"

// createUserGuardUserService 创建用户角色守卫用例的测试替身。
// 内嵌 sessionUserService 复用其余接口实现，仅覆盖需要观察与配置的方法。
type createUserGuardUserService struct {
	sessionUserService

	createCalls int
	createUser  *domain.User
	manageRole  func(managerRole, targetRole string) bool
}

// CreateUser 记录服务层被触达的次数，供断言拦截路径未调用服务层。
func (f *createUserGuardUserService) CreateUser(req *domain.CreateUserRequest) (*domain.User, error) {
	f.createCalls++
	return f.createUser, nil
}

// CanUserManageRole 按用例注入的判定规则返回结果，用于分别覆盖拦截与放行路径。
func (f *createUserGuardUserService) CanUserManageRole(managerRole, targetRole string) bool {
	return f.manageRole(managerRole, targetRole)
}

// adminManageableRoles 与 RBAC 服务的管理范围一致，管理员只能管理编辑者与普通用户。
func adminManageableRoles(managerRole, targetRole string) bool {
	if managerRole != string(service.RoleAdmin) {
		return false
	}
	return targetRole == string(service.RoleEditor) || targetRole == string(service.RoleUser)
}

// operatorRoleMiddleware 模拟认证层向上下文写入操作者角色。
func operatorRoleMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(operatorRoleContextKey, role)
	}
}

// performCreateUserRequest 在带操作者角色的路由上下文中执行一次创建用户请求。
func performCreateUserRequest(
	userService service.UserService,
	setOperatorRole gin.HandlerFunc,
	body string,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/users/create", setOperatorRole, NewUserHandler(userService, testSessionCookieConfig()).CreateUser)

	request := httptest.NewRequest(http.MethodPost, "/api/users/create", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// TestCreateUserRejectsSuperAdminAssignment 管理员提交超级管理员角色时拒绝创建，服务层不会被调用。
func TestCreateUserRejectsSuperAdminAssignment(t *testing.T) {
	fake := &createUserGuardUserService{
		manageRole: adminManageableRoles,
		createUser: &domain.User{ID: 1},
	}

	recorder := performCreateUserRequest(
		fake,
		operatorRoleMiddleware(string(service.RoleAdmin)),
		`{"username":"intruder","email":"intruder@example.com","password":"secret123","role":"superadmin"}`,
	)

	assertResponseCode(t, recorder, response.CodeForbid)
	if fake.createCalls != 0 {
		t.Errorf("服务层创建调用次数 = %d，期望拦截路径不调用服务层", fake.createCalls)
	}
}

// TestCreateUserAllowsEditorAssignment 管理员的角色可分配范围包含编辑者时请求放行至服务层。
func TestCreateUserAllowsEditorAssignment(t *testing.T) {
	fake := &createUserGuardUserService{
		manageRole: adminManageableRoles,
		createUser: &domain.User{ID: 2, Username: "editor", Role: string(service.RoleEditor)},
	}

	recorder := performCreateUserRequest(
		fake,
		operatorRoleMiddleware(string(service.RoleAdmin)),
		`{"username":"editor","email":"editor@example.com","password":"secret123","role":"editor"}`,
	)

	assertResponseCode(t, recorder, response.CodeSuccess)
	if fake.createCalls != 1 {
		t.Errorf("服务层创建调用次数 = %d，期望 1", fake.createCalls)
	}
}

// TestCreateUserWithoutOperatorRole 上下文缺少操作者角色时返回未授权，服务层不会被调用。
func TestCreateUserWithoutOperatorRole(t *testing.T) {
	fake := &createUserGuardUserService{
		manageRole: adminManageableRoles,
		createUser: &domain.User{ID: 3},
	}

	recorder := performCreateUserRequest(
		fake,
		func(*gin.Context) {},
		`{"username":"nobody","email":"nobody@example.com","password":"secret123","role":"user"}`,
	)

	assertResponseCode(t, recorder, response.CodeAuth)
	if fake.createCalls != 0 {
		t.Errorf("服务层创建调用次数 = %d，期望拦截路径不调用服务层", fake.createCalls)
	}
}

// TestCreateUserWithoutRoleSkipsAssignmentGuard 请求体未提交角色时由服务层取默认角色，不进入分配校验。
func TestCreateUserWithoutRoleSkipsAssignmentGuard(t *testing.T) {
	fake := &createUserGuardUserService{
		manageRole: func(managerRole, targetRole string) bool { return false },
		createUser: &domain.User{ID: 4, Username: "plain", Role: string(service.RoleUser)},
	}

	recorder := performCreateUserRequest(
		fake,
		operatorRoleMiddleware(string(service.RoleAdmin)),
		`{"username":"plain","email":"plain@example.com","password":"secret123"}`,
	)

	assertResponseCode(t, recorder, response.CodeSuccess)
	if fake.createCalls != 1 {
		t.Errorf("服务层创建调用次数 = %d，期望 1", fake.createCalls)
	}
}
