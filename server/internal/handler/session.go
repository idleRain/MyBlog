package handler

import (
	"net/http"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// SessionRequest 会话建立请求体，刷新令牌可选提交。
// 上限 512 字符为保护性约束，不透明令牌实际长度为 32 个字符。
type SessionRequest struct {
	RefreshToken string `json:"refreshToken" binding:"omitempty,max=512"`
}

// SessionCookiePath 会话 Cookie 的作用路径。
// 访问令牌经 /api 前缀的业务端点消费，刷新令牌仅认证端点消费，
// 统一限定在 /api 前缀下以缩小 Cookie 的暴露面。
const SessionCookiePath = "/api"

// SessionCookieConfig 会话 Cookie 的传输属性，由组合根从令牌配置换算注入。
type SessionCookieConfig struct {
	// AccessMaxAge 访问令牌 Cookie 的最大存活秒数，与令牌表中的有效期保持一致。
	AccessMaxAge int
	// RefreshMaxAge 刷新令牌 Cookie 的最大存活秒数，与令牌表中的有效期保持一致。
	RefreshMaxAge int
	// Secure 标记 Cookie 是否仅经 HTTPS 传输，生产环境必须开启。
	Secure bool
}

// sessionCookieSameSite 会话 Cookie 的同站策略。
// 采用 Lax 在缓解跨站伪造的同时保留外部链接跳转后的会话识别；
// 业务接口全部为 POST 且同源部署，跨站 POST 不会携带 Cookie。
const sessionCookieSameSite = http.SameSiteLaxMode

// CreateSession 建立会话 POST /api/auth/session
// 凭刷新令牌旋转签发新令牌对并写入 HttpOnly Cookie。
// 登录后首次建立会话时刷新令牌经请求体提交，续期场景由浏览器自动携带 Cookie，请求体可省略。
func (h *UserHandler) CreateSession(c *gin.Context) {
	refreshToken := h.resolveRefreshToken(c)
	if refreshToken == "" {
		response.BadRequest(c, "缺少刷新令牌")
		return
	}

	tokenPair, err := h.userService.RefreshToken(refreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	h.setSessionCookies(c, tokenPair)
	response.SuccessWithMessage(c, "会话已建立", gin.H{"expiresIn": tokenPair.ExpiresIn})
}

// resolveRefreshToken 依序从请求体与会话 Cookie 读取刷新令牌。
// 请求体省略或为空时不视为错误，继续从 Cookie 读取以支持续期场景。
func (h *UserHandler) resolveRefreshToken(c *gin.Context) string {
	var req SessionRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}

	cookieToken, err := c.Cookie(domain.SessionRefreshTokenCookie)
	if err != nil {
		return ""
	}
	return cookieToken
}

// setSessionCookies 把新令牌对写入 HttpOnly Cookie。
// 浏览器随后自动随同源请求携带，脚本无法读取，降低令牌被窃取的暴露面。
func (h *UserHandler) setSessionCookies(c *gin.Context, pair *service.TokenPair) {
	accessCookie := http.Cookie{
		Name:     domain.SessionAccessTokenCookie,
		Value:    pair.AccessToken,
		Path:     SessionCookiePath,
		MaxAge:   h.sessionCookie.AccessMaxAge,
		HttpOnly: true,
		Secure:   h.sessionCookie.Secure,
		SameSite: sessionCookieSameSite,
	}
	refreshCookie := http.Cookie{
		Name:     domain.SessionRefreshTokenCookie,
		Value:    pair.RefreshToken,
		Path:     SessionCookiePath,
		MaxAge:   h.sessionCookie.RefreshMaxAge,
		HttpOnly: true,
		Secure:   h.sessionCookie.Secure,
		SameSite: sessionCookieSameSite,
	}

	http.SetCookie(c.Writer, &accessCookie)
	http.SetCookie(c.Writer, &refreshCookie)
}

// clearSessionCookies 写入同名的过期空值 Cookie，指示浏览器移除本地会话。
// 清除时的 Path 与各属性必须与写入时完全一致，浏览器才视为同一 Cookie。
func (h *UserHandler) clearSessionCookies(c *gin.Context) {
	expiredAccess := http.Cookie{
		Name:     domain.SessionAccessTokenCookie,
		Value:    "",
		Path:     SessionCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.sessionCookie.Secure,
		SameSite: sessionCookieSameSite,
	}
	expiredRefresh := http.Cookie{
		Name:     domain.SessionRefreshTokenCookie,
		Value:    "",
		Path:     SessionCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.sessionCookie.Secure,
		SameSite: sessionCookieSameSite,
	}

	http.SetCookie(c.Writer, &expiredAccess)
	http.SetCookie(c.Writer, &expiredRefresh)
}

// sessionCookiesToken 读取会话 Cookie 中的令牌对，任一 Cookie 不存在时返回空值。
func sessionCookiesToken(c *gin.Context) (accessToken string, refreshToken string) {
	accessToken, _ = c.Cookie(domain.SessionAccessTokenCookie)
	refreshToken, _ = c.Cookie(domain.SessionRefreshTokenCookie)
	return accessToken, refreshToken
}
