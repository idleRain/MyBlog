package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// newIdentityProvider 构造带可配置替身的 身份解析器。
func newIdentityProvider(jwt *fakeTokenService, repo *fakeUserRepository) *tokenIdentityProvider {
	return &tokenIdentityProvider{tokenService: jwt, userRepo: repo}
}

// resolveViaRequest 以指定认证头执行身份解析，返回解析结果、响应业务码与是否中断。
// 统一响应信封固定 HTTP 200，认证失败以业务码 401/403 表达。
func resolveViaRequest(provider IdentityProvider, header string) (*domain.User, error, int, bool) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	if header != "" {
		ctx.Request.Header.Set("Authorization", header)
	}

	user, err := provider.Resolve(ctx)
	code := 0
	if recorder.Body.Len() > 0 {
		var body struct {
			Code int `json:"code"`
		}
		_ = json.Unmarshal(recorder.Body.Bytes(), &body)
		code = body.Code
	}
	return user, err, code, ctx.IsAborted()
}

// TestResolveNoToken 验证缺失令牌头时返回业务码 401 并中断链路。
func TestResolveNoToken(t *testing.T) {
	provider := newIdentityProvider(&fakeTokenService{}, &fakeUserRepository{})

	_, err, code, aborted := resolveViaRequest(provider, "")

	if err == nil {
		t.Error("缺失令牌时应返回错误")
	}
	if code != response.CodeAuth {
		t.Errorf("响应业务码 = %d，期望 %d", code, response.CodeAuth)
	}
	if !aborted {
		t.Error("缺失令牌时应中断请求链")
	}
}

// TestResolveInvalidToken 验证令牌校验失败时返回业务码 401。
func TestResolveInvalidToken(t *testing.T) {
	provider := newIdentityProvider(
		&fakeTokenService{err: errors.New("token invalid")},
		&fakeUserRepository{},
	)

	_, err, code, _ := resolveViaRequest(provider, "Bearer bad-token")

	if err == nil {
		t.Error("令牌无效时应返回错误")
	}
	if code != response.CodeAuth {
		t.Errorf("响应业务码 = %d，期望 %d", code, response.CodeAuth)
	}
}

// TestResolveUserNotFound 验证用户不存在时返回业务码 401。
func TestResolveUserNotFound(t *testing.T) {
	provider := newIdentityProvider(
		&fakeTokenService{claims: &service.TokenIdentity{UserID: 1}},
		&fakeUserRepository{err: errors.New("user not found")},
	)

	_, err, code, _ := resolveViaRequest(provider, "Bearer valid-token")

	if err == nil {
		t.Error("用户不存在时应返回错误")
	}
	if code != response.CodeAuth {
		t.Errorf("响应业务码 = %d，期望 %d", code, response.CodeAuth)
	}
}

// TestResolveDisabledUser 验证用户被禁用时返回业务码 403。
func TestResolveDisabledUser(t *testing.T) {
	provider := newIdentityProvider(
		&fakeTokenService{claims: &service.TokenIdentity{UserID: 1}},
		&fakeUserRepository{user: newTestUser("user", 0)},
	)

	_, err, code, _ := resolveViaRequest(provider, "Bearer valid-token")

	if err == nil {
		t.Error("禁用用户应返回错误")
	}
	if code != response.CodeForbid {
		t.Errorf("响应业务码 = %d，期望 %d", code, response.CodeForbid)
	}
}

// TestResolveInvalidRole 验证用户角色无效时返回业务码 403。
func TestResolveInvalidRole(t *testing.T) {
	provider := newIdentityProvider(
		&fakeTokenService{claims: &service.TokenIdentity{UserID: 1}},
		&fakeUserRepository{user: newTestUser("hacker", 1)},
	)

	_, err, code, _ := resolveViaRequest(provider, "Bearer valid-token")

	if err == nil {
		t.Error("角色无效时应返回错误")
	}
	if code != response.CodeForbid {
		t.Errorf("响应业务码 = %d，期望 %d", code, response.CodeForbid)
	}
}

// TestResolveSuccess 验证有效令牌与正常用户时返回用户且不写错误响应。
func TestResolveSuccess(t *testing.T) {
	activeUser := newTestUser("admin", 1)
	provider := newIdentityProvider(
		&fakeTokenService{claims: &service.TokenIdentity{UserID: 1}},
		&fakeUserRepository{user: activeUser},
	)

	user, err, code, aborted := resolveViaRequest(provider, "Bearer valid-token")

	if err != nil {
		t.Errorf("正常用户不应返回错误: %v", err)
	}
	if user != activeUser {
		t.Error("应返回与仓储一致的完整用户")
	}
	if code != 0 {
		t.Errorf("正常解析不应写错误响应，业务码 = %d", code)
	}
	if aborted {
		t.Error("正常解析不应中断请求链")
	}
}

// TestResolveTrimsBearerPrefix 验证裸令牌同样可解析，前缀剥离逻辑不误伤原值。
func TestResolveTrimsBearerPrefix(t *testing.T) {
	activeUser := newTestUser("user", 1)
	provider := newIdentityProvider(
		&fakeTokenService{claims: &service.TokenIdentity{UserID: 1}},
		&fakeUserRepository{user: activeUser},
	)

	_, err, _, _ := resolveViaRequest(provider, strings.TrimPrefix("Bearer valid-token", bearerTokenPrefix))

	if err != nil {
		t.Errorf("裸令牌解析不应返回错误: %v", err)
	}
}
