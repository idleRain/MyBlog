package domain

// 会话 Cookie 名称常量。
// Cookie 承载服务端签发的不透明令牌，令牌串本身不含可解码的身份信息，
// 身份与生命周期仍由服务端令牌表唯一权威维护，Cookie 仅是传输通道。
const (
	// SessionAccessTokenCookie 承载访问令牌的 Cookie 名称，前缀用于避免与其他应用同名冲突。
	SessionAccessTokenCookie = "mb_access_token"
	// SessionRefreshTokenCookie 承载刷新令牌的 Cookie 名称，仅认证端点消费。
	SessionRefreshTokenCookie = "mb_refresh_token"
)
