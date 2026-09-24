package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// fakeIdentityProvider 可配置的 IdentityProvider 测试替身，err 非空时模拟身份解析失败。
type fakeIdentityProvider struct {
	user *domain.User
	err  error
}

func (f *fakeIdentityProvider) Resolve(c *gin.Context) (*domain.User, error) {
	if f.err != nil {
		// 与真实实现一致：身份解析失败时已写入 401 响应并调用 Abort 中断请求链。
		response.Unauthorized(c, f.err.Error())
		c.Abort()
		return nil, f.err
	}
	return f.user, nil
}

// fakeRBACService 可配置的 RBACService 测试替身，仅覆盖 RBAC 中间件消费的方法。
type fakeRBACService struct {
	service.RBACService
	hasPermission bool
	hasAll        bool
	roleHigher    bool
	canManage     bool
}

func (f *fakeRBACService) HasPermission(userRole string, permission service.Permission) bool {
	return f.hasPermission
}

func (f *fakeRBACService) HasAllPermissions(userRole string, permissions ...service.Permission) bool {
	return f.hasAll
}

func (f *fakeRBACService) IsRoleHigherThan(roleA, roleB string) bool {
	return f.roleHigher
}

func (f *fakeRBACService) CanManageUser(managerRole, targetRole string) bool {
	return f.canManage
}

// fakeTokenService 可配置的 tokenService 测试替身，仅覆盖令牌校验方法。
type fakeTokenService struct {
	service.TokenServiceInterface
	identity *service.TokenIdentity
	err      error
}

func (f *fakeTokenService) ValidateAccessToken(tokenString string) (*service.TokenIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identity, nil
}

// fakeUserRepository 可配置的 UserRepository 测试替身，仅覆盖身份解析消费的查询方法。
type fakeUserRepository struct {
	repository.UserRepository
	user *domain.User
	err  error
}

func (f *fakeUserRepository) GetByID(id uint) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

// newTestUser 构造指定角色与状态的测试用户，ID 固定为 1。
func newTestUser(role string, status int) *domain.User {
	return &domain.User{ID: 1, Username: "tester", Role: role, Status: status}
}

// newMiddlewareRouter 组装 gin 路由：挂载给定中间件与标记 handler。
// handlerCalled 用于断言中间件放行后业务 handler 是否执行。
func newMiddlewareRouter(middleware gin.HandlerFunc, handlerCalled *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", middleware, func(c *gin.Context) {
		*handlerCalled = true
		c.Status(http.StatusOK)
	})
	return router
}

// newAuthRouter 组装认证类中间件路由，handler 内执行自定义断言并标记执行。
func newAuthRouter(middleware gin.HandlerFunc, assert func(*gin.Context)) (*gin.Engine, *bool) {
	gin.SetMode(gin.TestMode)
	handlerCalled := false
	router := gin.New()
	router.POST("/test", middleware, func(c *gin.Context) {
		handlerCalled = true
		if assert != nil {
			assert(c)
		}
		c.Status(http.StatusOK)
	})
	return router, &handlerCalled
}

// performTestRequest 执行一次 POST /test 请求并返回响应记录器。
func performTestRequest(router *gin.Engine) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// buildRequestWithHeader 构造带指定认证头的 POST /test 请求。
func buildRequestWithHeader(header string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	if header != "" {
		request.Header.Set("Authorization", header)
	}
	return request
}

// serve 使用路由执行给定请求并返回响应记录器。
func serve(router *gin.Engine, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// responseCode 解析统一响应信封的业务码字段，HTTP 状态码固定为 200。
func responseCode(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v，body=%s", err, recorder.Body.String())
	}
	return body.Code
}
