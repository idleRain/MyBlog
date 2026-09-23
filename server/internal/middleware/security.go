package middleware

import (
	"MyBlog/internal/config"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 频率限制配置
	RateLimit struct {
		Enabled        bool          `json:"enabled"`
		MaxRequests    int           `json:"max_requests"`
		Window         time.Duration `json:"window"`
		UserMaxRequest int           `json:"user_max_requests"`
		UserWindow     time.Duration `json:"user_window"`
	} `json:"rate_limit"`

	// 安全头配置
	SecurityHeaders struct {
		Enabled               bool   `json:"enabled"`
		ContentSecurityPolicy string `json:"content_security_policy"`
		XFrameOptions         string `json:"x_frame_options"`
		XContentTypeOptions   string `json:"x_content_type_options"`
		ReferrerPolicy        string `json:"referrer_policy"`
		StrictTransportSec    string `json:"strict_transport_security"`
	} `json:"security_headers"`

	// 输入验证配置
	InputValidation struct {
		Enabled           bool     `json:"enabled"`
		MaxRequestSize    int64    `json:"max_request_size"`
		BlockedPatterns   []string `json:"blocked_patterns"`
		AllowedUserAgents []string `json:"allowed_user_agents"`
		BlockedUserAgents []string `json:"blocked_user_agents"`
	} `json:"input_validation"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimit: struct {
			Enabled        bool          `json:"enabled"`
			MaxRequests    int           `json:"max_requests"`
			Window         time.Duration `json:"window"`
			UserMaxRequest int           `json:"user_max_requests"`
			UserWindow     time.Duration `json:"user_window"`
		}{
			Enabled:        true,
			MaxRequests:    100,
			Window:         time.Minute,
			UserMaxRequest: 300,
			UserWindow:     time.Minute,
		},
		SecurityHeaders: struct {
			Enabled               bool   `json:"enabled"`
			ContentSecurityPolicy string `json:"content_security_policy"`
			XFrameOptions         string `json:"x_frame_options"`
			XContentTypeOptions   string `json:"x_content_type_options"`
			ReferrerPolicy        string `json:"referrer_policy"`
			StrictTransportSec    string `json:"strict_transport_security"`
		}{
			Enabled:               true,
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:",
			XFrameOptions:         "SAMEORIGIN",
			XContentTypeOptions:   "nosniff",
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			StrictTransportSec:    "max-age=31536000; includeSubDomains",
		},
		InputValidation: struct {
			Enabled           bool     `json:"enabled"`
			MaxRequestSize    int64    `json:"max_request_size"`
			BlockedPatterns   []string `json:"blocked_patterns"`
			AllowedUserAgents []string `json:"allowed_user_agents"`
			BlockedUserAgents []string `json:"blocked_user_agents"`
		}{
			Enabled:        true,
			MaxRequestSize: 10 * 1024 * 1024, // 10MB
			// 阻止模式与 SecurityMiddlewareFromConfig 共用 getDefaultBlockedPatterns 单一来源。
			BlockedPatterns:   getDefaultBlockedPatterns(),
			AllowedUserAgents: []string{},
			// BlockedUserAgents 为子串匹配语义，配置项必须选取明确的工具或脚本特征。
			// 禁止收录 bot、crawler、spider 等泛化词，否则 Googlebot、Baiduspider、bingbot
			// 等搜索引擎爬虫会被整体 403，直接损害 SSR 博客的搜索收录，见体检项 BE-04。
			BlockedUserAgents: []string{
				"curl",
				"wget",
				"python-requests",
			},
		},
	}
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc {
	// 编译正则表达式
	var blockedPatterns []*regexp.Regexp
	for _, pattern := range config.InputValidation.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// 阻止模式无法编译时静默跳过，会让 WAF 在没有任何告警的情况下失效。
			// 模式集是代码常量，编译失败意味着程序缺陷，必须在启动期暴露。
			panic(fmt.Sprintf("WAF 阻止模式 %q 无法编译: %v", pattern, err))
		}
		blockedPatterns = append(blockedPatterns, re)
	}

	// 创建速率限制器
	var ipLimiter, userLimiter *RateLimiter
	if config.RateLimit.Enabled {
		ipLimiter = NewRateLimiter(config.RateLimit.MaxRequests, config.RateLimit.Window)
		userLimiter = NewRateLimiter(config.RateLimit.UserMaxRequest, config.RateLimit.UserWindow)
	}

	return func(c *gin.Context) {
		// 1. 频率限制检查
		if config.RateLimit.Enabled {
			clientIP := c.ClientIP()

			// IP 级别限制
			if !ipLimiter.Allow(clientIP) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    429,
					"message": "请求过于频繁，请稍后再试",
					"data":    nil,
				})
				c.Abort()
				return
			}

			// 用户级别限制（如果已认证）
			if userID, exists := c.Get("userID"); exists {
				userKey := fmt.Sprintf("user:%v", userID)
				if !userLimiter.Allow(userKey) {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"code":    429,
						"message": "用户请求过于频繁，请稍后再试",
						"data":    nil,
					})
					c.Abort()
					return
				}
			}
		}

		// 2. 安全头设置
		if config.SecurityHeaders.Enabled {
			c.Header("Content-Security-Policy", config.SecurityHeaders.ContentSecurityPolicy)
			c.Header("X-Frame-Options", config.SecurityHeaders.XFrameOptions)
			c.Header("X-Content-Type-Options", config.SecurityHeaders.XContentTypeOptions)
			c.Header("Referrer-Policy", config.SecurityHeaders.ReferrerPolicy)
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("X-DNS-Prefetch-Control", "off")
			c.Header("X-Download-Options", "noopen")
			c.Header("X-Permitted-Cross-Domain-Policies", "none")

			// HTTPS 环境下设置 HSTS
			if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
				c.Header("Strict-Transport-Security", config.SecurityHeaders.StrictTransportSec)
			}
		}

		// 3. 输入验证
		if config.InputValidation.Enabled {
			// 请求体体积在读取层强制约束。
			// Content-Length 在分块传输编码下为 -1，仅凭该头部无法拦截超大请求体，
			// 而后续的恶意内容扫描会把请求体整体读入内存，因此必须由 MaxBytesReader 兜住上限。
			if config.InputValidation.MaxRequestSize > 0 {
				c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.InputValidation.MaxRequestSize)
			}

			// 检查 User-Agent
			userAgent := c.GetHeader("User-Agent")
			if userAgent != "" {
				// 检查是否在阻止列表中
				for _, blocked := range config.InputValidation.BlockedUserAgents {
					if strings.Contains(strings.ToLower(userAgent), strings.ToLower(blocked)) {
						c.JSON(http.StatusForbidden, gin.H{
							"code":    403,
							"message": "请求被拒绝",
							"data":    nil,
						})
						c.Abort()
						return
					}
				}

				// 如果设置了允许列表，检查是否在允许列表中
				if len(config.InputValidation.AllowedUserAgents) > 0 {
					allowed := false
					for _, allowedUA := range config.InputValidation.AllowedUserAgents {
						if strings.Contains(strings.ToLower(userAgent), strings.ToLower(allowedUA)) {
							allowed = true
							break
						}
					}
					if !allowed {
						c.JSON(http.StatusForbidden, gin.H{
							"code":    403,
							"message": "请求被拒绝",
							"data":    nil,
						})
						c.Abort()
						return
					}
				}
			}

			// 检查请求参数中的恶意模式，请求体超限单独映射为 413。
			if err := validateRequest(c, blockedPatterns); err != nil {
				if errors.Is(err, errRequestTooLarge) {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{
						"code":    413,
						"message": "请求体过大",
						"data":    nil,
					})
					c.Abort()
					return
				}

				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "请求包含非法内容",
					"data":    nil,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// errRequestTooLarge 请求体超过配置上限的哨兵错误，供调用方映射为 413。
var errRequestTooLarge = errors.New("请求体超过大小上限")

// validateRequest 验证请求内容
func validateRequest(c *gin.Context, patterns []*regexp.Regexp) error {
	// 检查 URL 参数
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if containsMaliciousContent(key+value, patterns) {
				return fmt.Errorf("malicious content in URL parameter")
			}
		}
	}

	// 检查 JSON 请求体
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// 读取请求体，超限错误向上传递以便映射为 413。
			body, err := c.GetRawData()
			if err != nil {
				return errRequestTooLarge
			}
			if len(body) > 0 {
				bodyStr := string(body)
				if containsMaliciousContent(bodyStr, patterns) {
					return fmt.Errorf("malicious content in JSON body")
				}
				// 重新设置请求体，以便后续处理器可以读取
				c.Request.Body = io.NopCloser(strings.NewReader(bodyStr))
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
			strings.Contains(contentType, "multipart/form-data") {
			// 检查表单参数
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if containsMaliciousContent(key+value, patterns) {
							return fmt.Errorf("malicious content in form parameter")
						}
					}
				}
			}
		}
	}

	// 检查请求头
	suspiciousHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"Referer",
		"Origin",
	}

	for _, header := range suspiciousHeaders {
		value := c.GetHeader(header)
		if value != "" && containsMaliciousContent(value, patterns) {
			return fmt.Errorf("malicious content in header")
		}
	}

	return nil
}

// containsMaliciousContent 检查内容是否包含恶意模式
func containsMaliciousContent(content string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	// 将白名单转换为 map 以提高查找效率
	whitelistMap := make(map[string]bool)
	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 如果白名单为空，允许所有IP
		if len(whitelistMap) == 0 {
			c.Next()
			return
		}

		// 检查IP是否在白名单中
		if !whitelistMap[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "访问被拒绝",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityMiddlewareFromConfig 从配置文件创建安全中间件
func SecurityMiddlewareFromConfig(cfg *config.Config) gin.HandlerFunc {
	securityConfig := &SecurityConfig{
		RateLimit: struct {
			Enabled        bool          `json:"enabled"`
			MaxRequests    int           `json:"max_requests"`
			Window         time.Duration `json:"window"`
			UserMaxRequest int           `json:"user_max_requests"`
			UserWindow     time.Duration `json:"user_window"`
		}{
			Enabled:        cfg.Security.RateLimit.Enabled,
			MaxRequests:    cfg.Security.RateLimit.MaxRequests,
			Window:         time.Duration(cfg.Security.RateLimit.WindowMinutes) * time.Minute,
			UserMaxRequest: cfg.Security.RateLimit.UserMaxRequests,
			UserWindow:     time.Duration(cfg.Security.RateLimit.UserWindowMinutes) * time.Minute,
		},
		SecurityHeaders: struct {
			Enabled               bool   `json:"enabled"`
			ContentSecurityPolicy string `json:"content_security_policy"`
			XFrameOptions         string `json:"x_frame_options"`
			XContentTypeOptions   string `json:"x_content_type_options"`
			ReferrerPolicy        string `json:"referrer_policy"`
			StrictTransportSec    string `json:"strict_transport_security"`
		}{
			Enabled:               cfg.Security.SecurityHeaders.Enabled,
			ContentSecurityPolicy: cfg.Security.SecurityHeaders.ContentSecurityPolicy,
			XFrameOptions:         cfg.Security.SecurityHeaders.XFrameOptions,
			XContentTypeOptions:   cfg.Security.SecurityHeaders.XContentTypeOptions,
			ReferrerPolicy:        cfg.Security.SecurityHeaders.ReferrerPolicy,
			StrictTransportSec:    cfg.Security.SecurityHeaders.StrictTransportSecurity,
		},
		InputValidation: struct {
			Enabled           bool     `json:"enabled"`
			MaxRequestSize    int64    `json:"max_request_size"`
			BlockedPatterns   []string `json:"blocked_patterns"`
			AllowedUserAgents []string `json:"allowed_user_agents"`
			BlockedUserAgents []string `json:"blocked_user_agents"`
		}{
			Enabled:           cfg.Security.InputValidation.Enabled,
			MaxRequestSize:    int64(cfg.Security.InputValidation.MaxRequestSizeMB) * 1024 * 1024,
			BlockedPatterns:   getDefaultBlockedPatterns(),
			AllowedUserAgents: []string{},
			BlockedUserAgents: cfg.Security.InputValidation.BlockedUserAgents,
		},
	}

	return SecurityMiddleware(securityConfig)
}

// AdminSecurityMiddlewareFromConfig 从配置文件创建管理员安全中间件
func AdminSecurityMiddlewareFromConfig(cfg *config.Config) gin.HandlerFunc {
	// 显式关闭管理员接口加固时直接放行，不叠加任何额外限制。
	// 原实现在此处回退到硬编码的更严格配置，与开关语义相反。
	if !cfg.Security.AdminSecurity.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	securityConfig := &SecurityConfig{
		RateLimit: struct {
			Enabled        bool          `json:"enabled"`
			MaxRequests    int           `json:"max_requests"`
			Window         time.Duration `json:"window"`
			UserMaxRequest int           `json:"user_max_requests"`
			UserWindow     time.Duration `json:"user_window"`
		}{
			Enabled:        true,
			MaxRequests:    cfg.Security.AdminSecurity.MaxRequests,
			Window:         time.Minute,
			UserMaxRequest: cfg.Security.AdminSecurity.UserMaxRequests,
			UserWindow:     time.Minute,
		},
		SecurityHeaders: struct {
			Enabled               bool   `json:"enabled"`
			ContentSecurityPolicy string `json:"content_security_policy"`
			XFrameOptions         string `json:"x_frame_options"`
			XContentTypeOptions   string `json:"x_content_type_options"`
			ReferrerPolicy        string `json:"referrer_policy"`
			StrictTransportSec    string `json:"strict_transport_security"`
		}{
			Enabled:               cfg.Security.SecurityHeaders.Enabled,
			ContentSecurityPolicy: cfg.Security.SecurityHeaders.ContentSecurityPolicy,
			XFrameOptions:         cfg.Security.SecurityHeaders.XFrameOptions,
			XContentTypeOptions:   cfg.Security.SecurityHeaders.XContentTypeOptions,
			ReferrerPolicy:        cfg.Security.SecurityHeaders.ReferrerPolicy,
			StrictTransportSec:    cfg.Security.SecurityHeaders.StrictTransportSecurity,
		},
		InputValidation: struct {
			Enabled           bool     `json:"enabled"`
			MaxRequestSize    int64    `json:"max_request_size"`
			BlockedPatterns   []string `json:"blocked_patterns"`
			AllowedUserAgents []string `json:"allowed_user_agents"`
			BlockedUserAgents []string `json:"blocked_user_agents"`
		}{
			Enabled:           cfg.Security.InputValidation.Enabled,
			MaxRequestSize:    5 * 1024 * 1024, // 管理员接口限制5MB
			BlockedPatterns:   getDefaultBlockedPatterns(),
			AllowedUserAgents: []string{},
			BlockedUserAgents: cfg.Security.InputValidation.BlockedUserAgents,
		},
	}

	middleware := SecurityMiddleware(securityConfig)

	// 如果设置了IP白名单，则添加IP白名单中间件
	if len(cfg.Security.AdminSecurity.IPWhitelist) > 0 {
		return gin.HandlerFunc(func(c *gin.Context) {
			IPWhitelistMiddleware(cfg.Security.AdminSecurity.IPWhitelist)(c)
			if c.IsAborted() {
				return
			}
			middleware(c)
		})
	}

	return middleware
}

// getDefaultBlockedPatterns 获取默认的阻止模式，全部模式经词首边界或取值上下文锚定，
// 避免宽匹配误伤博客正文中 "content ="、"for i = 1" 等正常写法（体检项 BE-07）。
func getDefaultBlockedPatterns() []string {
	return []string{
		`(?i)<script[^>]*>.*?</script>`,           // XSS：script 标签对
		`(?i)javascript:`,                         // XSS：JavaScript 伪协议 URL
		`(?i)\bon\w+\s*=\s*\\?["']`,               // XSS：HTML 事件属性取引号值，词首边界避免命中 content 等单词内部
		`(?i)\bon\w+\s*=\s*\w+\s*\(`,              // XSS：事件属性的无引号函数调用形式，如 onerror=alert(1)
		`(?i)\bunion\s+select`,                    // SQL 注入：联合查询
		`(?i)\binsert\s+into\b`,                   // SQL 注入：插入语句，词边界避免命中英文正文 inserted into
		`(?i)\bdelete\s+from\b`,                   // SQL 注入：删除语句，词边界避免命中英文正文 deleted from
		`(?i)\bdrop\s+table\b`,                    // SQL 注入：删表语句
		`(?i)\b(or|and)\s+\d+\s*=\s*\d+`,          // SQL 注入：恒真式 OR 1=1
		`(?i)\b(or|and)\s+['"]\w+['"]\s*=\s*['"]`, // SQL 注入：恒真式 OR '1'='1
		`(?i)\bexec\s*\(`,                         // 命令执行
		`(?i)\bsystem\s*\(`,                       // 命令执行
		`(?i)\.\.\/`,                              // 路径穿越
		`(?i)\.\.\\`,                              // 路径穿越（Windows）
	}
}
