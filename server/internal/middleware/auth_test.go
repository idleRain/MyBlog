package middleware

import (
	"errors"
	"net/http"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// TestAuth 表驱动验证认证中间件的令牌校验分支。
// 统一响应信封固定 HTTP 200，认证失败以业务码 401 表达，故断言响应体 code 字段。
func TestAuth(t *testing.T) {
	testCases := []struct {
		name         string
		header       string
		tokenErr     error
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "缺少令牌头时返回 401",
			header:       "",
			expectCalled: false,
			expectCode:   response.CodeAuth,
		},
		{
			name:         "令牌无效时返回 401",
			header:       "Bearer invalid-token",
			tokenErr:     errors.New("token invalid"),
			expectCalled: false,
			expectCode:   response.CodeAuth,
		},
		{
			name:         "令牌有效时放行并写入用户 ID",
			header:       "Bearer valid-token",
			expectCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService := &fakeTokenService{identity: &service.TokenIdentity{UserID: 7}, err: tc.tokenErr}
			router, handlerCalled := newAuthRouter(Auth(tokenService), func(c *gin.Context) {
				if got, exists := c.Get("userID"); tc.expectCalled {
					if !exists || got != uint(7) {
						t.Errorf("上下文 userID = %v，期望 7", got)
					}
				}
			})

			request := buildRequestWithHeader(tc.header)
			recorder := serve(router, request)

			if *handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", *handlerCalled, tc.expectCalled)
			}
			if tc.expectCode != 0 {
				if code := responseCode(t, recorder); code != tc.expectCode {
					t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
				}
			}
		})
	}
}

// TestOptionalAuth 表驱动验证可选认证中间件：无令牌与无效令牌均放行，有效令牌写入上下文。
func TestOptionalAuth(t *testing.T) {
	testCases := []struct {
		name                string
		header              string
		tokenErr            error
		expectAuthenticated bool
	}{
		{
			name:                "无令牌时放行且不标记认证",
			header:              "",
			expectAuthenticated: false,
		},
		{
			name:                "无效令牌时放行且不标记认证",
			header:              "Bearer bad",
			tokenErr:            errors.New("token invalid"),
			expectAuthenticated: false,
		},
		{
			name:                "有效令牌时放行并标记认证",
			header:              "Bearer good",
			expectAuthenticated: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService := &fakeTokenService{identity: &service.TokenIdentity{UserID: 3}, err: tc.tokenErr}
			router, handlerCalled := newAuthRouter(OptionalAuth(tokenService), func(c *gin.Context) {
				_, exists := c.Get("authenticated")
				if exists != tc.expectAuthenticated {
					t.Errorf("上下文 authenticated 存在状态 = %v，期望 %v", exists, tc.expectAuthenticated)
				}
			})

			recorder := serve(router, buildRequestWithHeader(tc.header))

			if !*handlerCalled {
				t.Error("可选认证不应拦截请求，业务 handler 未执行")
			}
			if recorder.Code != http.StatusOK {
				t.Errorf("HTTP 状态码 = %d，期望 200", recorder.Code)
			}
		})
	}
}

// TestAdminAuth 表驱动验证管理员认证中间件。
func TestAdminAuth(t *testing.T) {
	testCases := []struct {
		name         string
		identityErr  error
		userRole     string
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "身份解析失败时链路中止",
			identityErr:  errors.New("no token"),
			expectCalled: false,
		},
		{
			name:         "非管理员角色返回 403",
			userRole:     "user",
			expectCalled: false,
			expectCode:   response.CodeForbid,
		},
		{
			name:         "管理员角色放行",
			userRole:     "admin",
			expectCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			identity := &fakeIdentityProvider{user: newTestUser(tc.userRole, 1), err: tc.identityErr}
			handlerCalled := false
			router := newMiddlewareRouter(AdminAuth(identity), &handlerCalled)
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

// TestRequireRole 表驱动验证角色匹配中间件。
func TestRequireRole(t *testing.T) {
	testCases := []struct {
		name         string
		identityErr  error
		userRole     string
		allowedRoles []string
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "身份解析失败时链路中止",
			identityErr:  errors.New("no token"),
			allowedRoles: []string{"admin"},
			expectCalled: false,
		},
		{
			name:         "角色匹配时放行",
			userRole:     "editor",
			allowedRoles: []string{"editor", "admin"},
			expectCalled: true,
		},
		{
			name:         "角色不匹配时返回 403",
			userRole:     "user",
			allowedRoles: []string{"admin"},
			expectCalled: false,
			expectCode:   response.CodeForbid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			identity := &fakeIdentityProvider{user: newTestUser(tc.userRole, 1), err: tc.identityErr}
			handlerCalled := false
			router := newMiddlewareRouter(RequireRole(identity, tc.allowedRoles...), &handlerCalled)
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

// TestAuthCookieFallback 表驱动验证会话 Cookie 双轨：Header 优先，Cookie 为浏览器默认通道。
func TestAuthCookieFallback(t *testing.T) {
	testCases := []struct {
		name         string
		header       string
		cookieToken  string
		expectCalled bool
		expectCode   int
	}{
		{
			name:         "仅携带会话 Cookie 时放行",
			cookieToken:  "cookie-access-token",
			expectCalled: true,
		},
		{
			name:         "Header 与 Cookie 并存时优先 Header",
			header:       "Bearer header-token",
			cookieToken:  "cookie-access-token",
			expectCalled: true,
		},
		{
			name:         "两通道均无令牌时返回 401",
			expectCalled: false,
			expectCode:   response.CodeAuth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService := &fakeTokenService{identity: &service.TokenIdentity{UserID: 7}}
			router, handlerCalled := newAuthRouter(Auth(tokenService), nil)

			request := buildRequestWithHeader(tc.header)
			if tc.cookieToken != "" {
				request.AddCookie(&http.Cookie{Name: domain.SessionAccessTokenCookie, Value: tc.cookieToken})
			}
			recorder := serve(router, request)

			if *handlerCalled != tc.expectCalled {
				t.Errorf("业务 handler 执行状态 = %v，期望 %v", *handlerCalled, tc.expectCalled)
			}
			if tc.expectCode != 0 {
				if code := responseCode(t, recorder); code != tc.expectCode {
					t.Errorf("响应业务码 = %d，期望 %d", code, tc.expectCode)
				}
			}
		})
	}
}
