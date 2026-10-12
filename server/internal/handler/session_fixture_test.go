package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// fakeSessionUserService 会话场景测试替身，仅覆盖 RefreshToken 与 Logout 方法。
type fakeSessionUserService struct {
	service.UserService
	tokenPair  *service.TokenPair
	refreshErr error
	logoutErr  error
}

func (f *fakeSessionUserService) RefreshToken(refreshToken string) (*service.TokenPair, error) {
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}
	return f.tokenPair, nil
}

func (f *fakeSessionUserService) Logout(accessToken, refreshToken string) error {
	return f.logoutErr
}

// TestSessionCreatedMatchesFixture 验证会话续期响应与契约金样本一致，构成契约锁 ①。
// 令牌对经 Set-Cookie 下发，响应体只承载访问令牌有效期，Cookie 属性由 session_test.go 断言。
func TestSessionCreatedMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserHandler(&fakeSessionUserService{tokenPair: &service.TokenPair{
		AccessToken:  "3f2a9c1e4b7d8056a1c3e5f70b2d4689",
		RefreshToken: "8b1d4e7a2c9f0356b8e1a4d7c0f32695",
		ExpiresIn:    900,
	}}, testSessionCookieConfig())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/auth/session",
		strings.NewReader(`{"refreshToken":"8b1d4e7a2c9f0356b8e1a4d7c0f32695"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateSession(ctx)

	assertResponseMatchesFixture(t, recorder, "session.created.json")
}

// TestLogoutSuccessMatchesFixture 验证登出响应与契约金样本一致，构成契约锁 ①。
// 撤销请求经 Authorization 头提交，响应体无 data 字段，锁定「空数据成功响应不输出 data」的契约。
func TestLogoutSuccessMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserHandler(&fakeSessionUserService{}, testSessionCookieConfig())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout",
		strings.NewReader(`{"refreshToken":"8b1d4e7a2c9f0356b8e1a4d7c0f32695"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("Authorization", "Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689")

	handler.Logout(ctx)

	assertResponseMatchesFixture(t, recorder, "logout.success.json")
}
