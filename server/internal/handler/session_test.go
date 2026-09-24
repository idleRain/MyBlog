package handler

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

// jsonUnmarshalResponse 解析统一响应信封的业务码字段。
func jsonUnmarshalResponse(recorder *httptest.ResponseRecorder, target any) error {
	return json.Unmarshal(recorder.Body.Bytes(), target)
}

// sessionUserService 会话场景测试替身，显式实现接口全部方法，
// 避免内嵌空接口在接口演进时以 nil panic 暴露的运行时脆性。
// RefreshToken 与 Logout 具有行为，用于断言旋转签发与撤销链路。
type sessionUserService struct {
	service.UserService

	pair        *service.TokenPair
	refreshErr  error
	logoutCalls [][2]string
	logoutErr   error
}

func (f *sessionUserService) RefreshToken(refreshToken string) (*service.TokenPair, error) {
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}
	return f.pair, nil
}

func (f *sessionUserService) Logout(accessToken, refreshToken string) error {
	if f.logoutErr != nil {
		return f.logoutErr
	}
	f.logoutCalls = append(f.logoutCalls, [2]string{accessToken, refreshToken})
	return nil
}

func (f *sessionUserService) CreateUser(req *domain.CreateUserRequest) (*domain.User, error) {
	return nil, nil
}

func (f *sessionUserService) UpdateUser(req *domain.UpdateUserRequest) (*domain.User, error) {
	return nil, nil
}

func (f *sessionUserService) GetUserByID(id uint) (*domain.User, error) {
	return nil, nil
}

func (f *sessionUserService) GetUserList(page, pageSize int, keyword string) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (f *sessionUserService) DeleteUser(id uint) error {
	return nil
}

func (f *sessionUserService) Login(username, password string) (*service.LoginResponse, error) {
	return nil, nil
}

func (f *sessionUserService) GetProfile(userID uint) (*domain.User, error) {
	return nil, nil
}

func (f *sessionUserService) UpdateProfile(userID uint, req *service.UpdateProfileRequest) (*domain.User, error) {
	return nil, nil
}

func (f *sessionUserService) ChangePassword(userID uint, req *service.ChangePasswordRequest) error {
	return nil
}

func (f *sessionUserService) CanUserManageRole(managerRole, targetRole string) bool {
	return false
}

func (f *sessionUserService) ValidateRoleTransition(currentRole, newRole string) error {
	return nil
}

// newSessionTestHandler 组装会话测试处理器。
func newSessionTestHandler(fake *sessionUserService) UserHandlerInterface {
	return NewUserHandler(fake, testSessionCookieConfig())
}

// performSessionRequest 以给定请求体与 Cookie 执行一次 POST /auth/session。
func performSessionRequest(handler UserHandlerInterface, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/auth/session", handler.CreateSession)
	router.POST("/api/auth/logout", handler.Logout)

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/session", reader)
	request.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// cookieByName 从响应中按名称取 Set-Cookie，未找到时返回 nil。
func cookieByName(t *testing.T, recorder *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("响应缺少名为 %s 的 Set-Cookie", name)
	return nil
}

// TestCreateSessionWithBodyToken 登录后首次建立会话：请求体携带刷新令牌，响应写入会话 Cookie。
func TestCreateSessionWithBodyToken(t *testing.T) {
	fake := &sessionUserService{pair: &service.TokenPair{
		AccessToken:  "access-token-value",
		RefreshToken: "refresh-token-value",
		ExpiresIn:    900,
	}}
	handler := newSessionTestHandler(fake)

	recorder := performSessionRequest(handler, `{"refreshToken":"login-issued-refresh"}`)

	if recorder.Code != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d，期望 200", recorder.Code)
	}

	accessCookie := cookieByName(t, recorder, domain.SessionAccessTokenCookie)
	if accessCookie.Value != "access-token-value" {
		t.Errorf("访问令牌 Cookie 值 = %s，期望签发的访问令牌", accessCookie.Value)
	}
	if !accessCookie.HttpOnly {
		t.Error("访问令牌 Cookie 必须为 HttpOnly，脚本不可读取")
	}
	if accessCookie.Path != SessionCookiePath {
		t.Errorf("访问令牌 Cookie 路径 = %s，期望 %s", accessCookie.Path, SessionCookiePath)
	}
	if accessCookie.MaxAge != testSessionCookieConfig().AccessMaxAge {
		t.Errorf("访问令牌 Cookie 存活秒数 = %d，期望与配置一致", accessCookie.MaxAge)
	}

	refreshCookie := cookieByName(t, recorder, domain.SessionRefreshTokenCookie)
	if refreshCookie.Value != "refresh-token-value" {
		t.Errorf("刷新令牌 Cookie 值 = %s，期望签发的刷新令牌", refreshCookie.Value)
	}
	if refreshCookie.MaxAge != testSessionCookieConfig().RefreshMaxAge {
		t.Errorf("刷新令牌 Cookie 存活秒数 = %d，期望与配置一致", refreshCookie.MaxAge)
	}
}

// TestCreateSessionWithCookie 续期场景：请求体省略，刷新令牌由浏览器 Cookie 携带。
func TestCreateSessionWithCookie(t *testing.T) {
	fake := &sessionUserService{pair: &service.TokenPair{
		AccessToken:  "rotated-access",
		RefreshToken: "rotated-refresh",
		ExpiresIn:    900,
	}}
	handler := newSessionTestHandler(fake)

	existingRefresh := &http.Cookie{Name: domain.SessionRefreshTokenCookie, Value: "stale-refresh"}
	recorder := performSessionRequest(handler, "", existingRefresh)

	if recorder.Code != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d，期望 200", recorder.Code)
	}
	cookieByName(t, recorder, domain.SessionAccessTokenCookie)
	cookieByName(t, recorder, domain.SessionRefreshTokenCookie)
}

// TestCreateSessionWithoutToken 无请求体且无 Cookie 时拒绝建立会话。
func TestCreateSessionWithoutToken(t *testing.T) {
	handler := newSessionTestHandler(&sessionUserService{})

	recorder := performSessionRequest(handler, "")

	if recorder.Code != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d，期望统一信封固定 200", recorder.Code)
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := jsonUnmarshalResponse(recorder, &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Code != response.CodeInvalid {
		t.Errorf("响应业务码 = %d，期望 %d", body.Code, response.CodeInvalid)
	}
}

// TestCreateSessionWithInvalidToken 刷新令牌无效时返回业务码 401 且不写入 Cookie。
func TestCreateSessionWithInvalidToken(t *testing.T) {
	fake := &sessionUserService{refreshErr: errors.New("刷新令牌验证失败")}
	handler := newSessionTestHandler(fake)

	recorder := performSessionRequest(handler, `{"refreshToken":"revoked-token"}`)

	if len(recorder.Result().Cookies()) != 0 {
		t.Error("会话建立失败时不得写入任何 Cookie")
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := jsonUnmarshalResponse(recorder, &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Code != response.CodeAuth {
		t.Errorf("响应业务码 = %d，期望 %d", body.Code, response.CodeAuth)
	}
}

// TestLogoutClearsSessionCookies Cookie 会话登出：撤销请求中携带的令牌对并指示浏览器清除 Cookie。
func TestLogoutClearsSessionCookies(t *testing.T) {
	fake := &sessionUserService{}
	handler := newSessionTestHandler(fake)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/auth/logout", handler.Logout)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(""))
	request.AddCookie(&http.Cookie{Name: domain.SessionAccessTokenCookie, Value: "cookie-access"})
	request.AddCookie(&http.Cookie{Name: domain.SessionRefreshTokenCookie, Value: "cookie-refresh"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if len(fake.logoutCalls) != 1 {
		t.Fatalf("服务层撤销调用次数 = %d，期望 1", len(fake.logoutCalls))
	}
	if fake.logoutCalls[0] != [2]string{"cookie-access", "cookie-refresh"} {
		t.Errorf("撤销的令牌对 = %v，期望会话 Cookie 中的访问与刷新令牌", fake.logoutCalls[0])
	}

	expiredAccess := cookieByName(t, recorder, domain.SessionAccessTokenCookie)
	if expiredAccess.MaxAge != -1 {
		t.Errorf("清除访问令牌 Cookie 的 MaxAge = %d，期望 -1 表示立即过期", expiredAccess.MaxAge)
	}
	expiredRefresh := cookieByName(t, recorder, domain.SessionRefreshTokenCookie)
	if expiredRefresh.MaxAge != -1 {
		t.Errorf("清除刷新令牌 Cookie 的 MaxAge = %d，期望 -1 表示立即过期", expiredRefresh.MaxAge)
	}
}

// TestLogoutWithoutAnyToken 无 Header 令牌且无会话 Cookie 时拒绝登出。
func TestLogoutWithoutAnyToken(t *testing.T) {
	fake := &sessionUserService{}
	handler := newSessionTestHandler(fake)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/auth/logout", handler.Logout)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader("")))

	var body struct {
		Code int `json:"code"`
	}
	if err := jsonUnmarshalResponse(recorder, &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Code != response.CodeInvalid {
		t.Errorf("响应业务码 = %d，期望 %d", body.Code, response.CodeInvalid)
	}
}
