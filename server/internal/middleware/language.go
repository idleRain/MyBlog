package middleware

import (
	"MyBlog/internal/config"
	"MyBlog/internal/domain"

	"github.com/gin-gonic/gin"
)

// AcceptLanguageHeader 语言协商请求头名称，遵循 HTTP 标准 Accept-Language 语义。
const AcceptLanguageHeader = "Accept-Language"

// LanguageMiddlewareFromConfig 按配置构建语言协商中间件，解析 Accept-Language 写入请求上下文。
// 白名单外的语言标签回退到配置的缺省语言，管理端全量包标识原样放行；
// 中间件只做解析与写入，不做响应中断，业务 handler 始终能读到明确的语言标识。
func LanguageMiddlewareFromConfig(cfg config.I18NConfig) gin.HandlerFunc {
	supported := domain.ParseLanguages(cfg.SupportedLanguages)
	fallback := domain.DefaultLanguage
	if configured := domain.Language(cfg.DefaultLanguage); configured != "" {
		fallback = configured
	}

	return func(c *gin.Context) {
		language := domain.ParseLanguage(c.GetHeader(AcceptLanguageHeader), supported, fallback)
		c.Set(domain.LanguageContextKey, language)
		c.Next()
	}
}
