package middleware

import (
	"MyBlog/internal/config"
	"MyBlog/internal/domain"

	"github.com/gin-gonic/gin"
)

// AcceptLanguageHeader 语言协商请求头名称，遵循 HTTP 标准 Accept-Language 语义。
const AcceptLanguageHeader = "Accept-Language"

// LanguagePolicyFromConfig 将配置节转换为语言协商策略。
// 配置值经归一化保证白名单成员与解析结果同形态可比，缺省项缺失时回退 domain 默认策略。
func LanguagePolicyFromConfig(cfg config.I18NConfig) domain.LanguagePolicy {
	policy := domain.DefaultLanguagePolicy
	if supported := domain.ParseLanguages(cfg.SupportedLanguages); len(supported) > 0 {
		policy.Supported = supported
	}
	if configured := domain.Language(cfg.DefaultLanguage); configured != "" {
		policy.Default = configured
	}
	return policy
}

// LanguageMiddlewareFromConfig 按配置构建语言协商中间件，解析 Accept-Language 写入请求上下文。
// 白名单外的语言标签回退到配置的缺省语言，管理端全量包标识原样放行；
// 中间件只做解析与写入，不做响应中断，业务 handler 始终能读到明确的语言标识。
func LanguageMiddlewareFromConfig(cfg config.I18NConfig) gin.HandlerFunc {
	policy := LanguagePolicyFromConfig(cfg)

	return func(c *gin.Context) {
		language := domain.ParseLanguage(c.GetHeader(AcceptLanguageHeader), policy.Supported, policy.Default)
		c.Set(domain.LanguageContextKey, language)
		c.Next()
	}
}
