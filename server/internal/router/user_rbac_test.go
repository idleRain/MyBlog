package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// testBearerPrefix 测试用 Bearer 认证头前缀，与真实身份解析器的解析规则一致。
const testBearerPrefix = "Bearer "

// 编译期断言替身完整实现 handler 接口，接口新增方法时在此暴露而非运行时 panic。
var _ handler.UserHandlerInterface = (*recordingUserHandler)(nil)

// 编译期断言替身完整实现身份解析接口，与真实实现同构演进。
var _ middleware.IdentityProvider = (*stubIdentity)(nil)

// recordingUserHandler 用户处理器测试替身，显式实现接口全部方法，
// 避免内嵌空接口在接口演进时以 nil panic 暴露的运行时脆性。
// 仅 GetUserByID 具有行为，用于断言权限放行后业务链路真实触达。
type recordingUserHandler struct {
	handlerCalled *bool
}

func (h *recordingUserHandler) CreateUser(c *gin.Context) {}

func (h *recordingUserHandler) UpdateUser(c *gin.Context) {}

func (h *recordingUserHandler) GetUserByID(c *gin.Context) {
	*h.handlerCalled = true
	response.Success(c, gin.H{"id": uint(1)})
}

func (h *recordingUserHandler) GetUserList(c *gin.Context) {}

func (h *recordingUserHandler) DeleteUser(c *gin.Context) {}

func (h *recordingUserHandler) Login(c *gin.Context) {}

func (h *recordingUserHandler) CreateSession(c *gin.Context) {}

func (h *recordingUserHandler) RefreshToken(c *gin.Context) {}

func (h *recordingUserHandler) Logout(c *gin.Context) {}

func (h *recordingUserHandler) GetProfile(c *gin.Context) {}

func (h *recordingUserHandler) UpdateProfile(c *gin.Context) {}

func (h *recordingUserHandler) ChangePassword(c *gin.Context) {}

// stubTokenService 令牌服务测试替身，显式实现接口全部方法。
// /users/get 经权限中间件完成认证，不再触达令牌服务，
// 但构造 UserRoutes 组合根签名要求提供该依赖。
type stubTokenService struct{}

func (s *stubTokenService) GenerateTokenPair(user *domain.User) (*service.TokenPair, error) {
	return nil, nil
}

func (s *stubTokenService) ValidateAccessToken(tokenString string) (*service.TokenIdentity, error) {
	return nil, nil
}

func (s *stubTokenService) ValidateRefreshToken(tokenString string) (*service.TokenIdentity, error) {
	return nil, nil
}

func (s *stubTokenService) RefreshAccessToken(refreshTokenString string) (*service.TokenPair, error) {
	return nil, nil
}

func (s *stubTokenService) RevokeToken(tokenString string) error { return nil }

func (s *stubTokenService) RevokeUserTokens(userID uint) error { return nil }

// stubIdentity 按令牌串映射角色名固定角色的身份解析测试替身。
// 空令牌模拟未认证请求，与真实实现一致地写入 401 响应并中止请求链。
type stubIdentity struct{}

func (s *stubIdentity) Resolve(c *gin.Context) (*domain.User, error) {
	role := strings.TrimPrefix(c.GetHeader("Authorization"), testBearerPrefix)
	if role == "" {
		response.Unauthorized(c, "未提供认证令牌")
		c.Abort()
		return nil, errors.New("no token")
	}
	return &domain.User{ID: uint(1), Username: "tester", Role: role, Status: domain.UserStatusActive}, nil
}

// newUserRoutesUnderTest 组装仅包含用户路由模块的被测路由。
// RBAC 服务使用真实实现，使权限断言落在与生产一致的内置权限表上。
func newUserRoutesUnderTest(handlerCalled *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	routes := NewUserRoutes(
		&recordingUserHandler{handlerCalled: handlerCalled},
		&stubTokenService{},
		&stubIdentity{},
		service.NewRBACService(),
	)
	router := gin.New()
	routes.RegisterRoutes(router.Group("/api"))
	return router
}

// performUsersGet 以指定令牌请求 /users/get 并返回响应记录器。
func performUsersGet(router *gin.Engine, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/users/get", strings.NewReader(`{"id":1}`))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", testBearerPrefix+token)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// parseResponseCode 解析统一响应信封的业务码字段，HTTP 状态码固定为 200。
func parseResponseCode(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v，body=%s", err, recorder.Body.String())
	}
	return body.Code
}

// TestUsersGetRequiresUserListPermission 表驱动验证 /users/get 的权限门装配。
// 修复前该端点仅挂基础认证，任意登录用户可查任意用户的 email 与生日；
// 修复后与 /users/list 同权，仅持有用户列表权限的管理角色可访问。
func TestUsersGetRequiresUserListPermission(t *testing.T) {
	testCases := []struct {
		name                string
		token               string
		expectCode          int
		expectHandlerCalled bool
	}{
		{
			name:                "未提供令牌时返回 401",
			token:               "",
			expectCode:          response.CodeAuth,
			expectHandlerCalled: false,
		},
		{
			name:                "普通用户缺少用户列表权限时返回 403",
			token:               string(service.RoleUser),
			expectCode:          response.CodeForbid,
			expectHandlerCalled: false,
		},
		{
			name:                "编辑者仅持有用户查看权限时返回 403",
			token:               string(service.RoleEditor),
			expectCode:          response.CodeForbid,
			expectHandlerCalled: false,
		},
		{
			name:                "管理员持有用户列表权限时放行",
			token:               string(service.RoleAdmin),
			expectCode:          response.CodeSuccess,
			expectHandlerCalled: true,
		},
		{
			name:                "超级管理员持有用户列表权限时放行",
			token:               string(service.RoleSuperAdmin),
			expectCode:          response.CodeSuccess,
			expectHandlerCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			router := newUserRoutesUnderTest(&handlerCalled)
			recorder := performUsersGet(router, tc.token)

			if recorder.Code != http.StatusOK {
				t.Errorf("HTTP 状态码 = %d，期望统一信封固定 200", recorder.Code)
			}
			if handlerCalled != tc.expectHandlerCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", handlerCalled, tc.expectHandlerCalled)
			}
			if code := parseResponseCode(t, recorder); code != tc.expectCode {
				t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
			}
		})
	}
}
