package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// corsVaryHeaderName 告知缓存层响应随 Origin 变化的 Vary 头名称。
const corsVaryHeaderName = "Vary"

// corsVaryHeaderValue Vary 头的 Origin 值，白名单内外请求的 CORS 头集合不同，
// 中间层缓存禁止跨 Origin 复用响应。
const corsVaryHeaderValue = "Origin"

// CORSConfig CORS 跨域配置，来源为 config.yaml 的 cors 节。
type CORSConfig struct {
	AllowedOrigins   []string // 允许跨域的 Origin 白名单，精确匹配，空列表表示拒绝所有跨域
	AllowedMethods   []string // 允许的 HTTP 方法，POST-Only 规范下仅 POST 与预检 OPTIONS
	AllowedHeaders   []string // 允许跨域携带的请求头
	AllowCredentials bool     // 是否允许跨域携带凭证
}

// CORSWithConfig 白名单化 CORS 中间件，仅对白名单内 Origin 回显 CORS 响应头。
// 白名单外与未携带 Origin 的请求不返回任何 CORS 头，生产同源网关形态下白名单可为空。
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := slices.Contains(config.AllowedOrigins, origin)

		// 无论是否放行都声明响应随 Origin 变化，避免中间层缓存串用 CORS 头。
		c.Header(corsVaryHeaderName, corsVaryHeaderValue)

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		// 预检请求统一 204，白名单外请求因缺少 Allow-Origin 头在浏览器侧被拒绝。
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
