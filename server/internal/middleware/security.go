package middleware

import (
	"MyBlog/internal/config"
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
			BlockedPatterns: []string{
				`(?i)<script[^>]*>.*?</script>`, // XSS
				`(?i)javascript:`,               // JavaScript URLs
				`(?i)on\w+\s*=`,                 // Event handlers
				`(?i)union.*select`,             // SQL Injection
				`(?i)insert.*into`,              // SQL Injection
				`(?i)delete.*from`,              // SQL Injection
				`(?i)drop.*table`,               // SQL Injection
				`(?i)exec\s*\(`,                 // Command execution
				`(?i)system\s*\(`,               // Command execution
				`(?i)\.\.\/`,                    // Path traversal
				`(?i)\.\.\\`,                    // Path traversal (Windows)
			},
			AllowedUserAgents: []string{},
			BlockedUserAgents: []string{
				"curl",
				"wget",
				"python-requests",
				"bot",
				"crawler",
				"spider",
			},
		},
	}
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc {
	// 编译正则表达式
	var blockedPatterns []*regexp.Regexp
	for _, pattern := range config.InputValidation.BlockedPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			blockedPatterns = append(blockedPatterns, re)
		}
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
			// 检查请求大小
			if c.Request.ContentLength > config.InputValidation.MaxRequestSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"code":    413,
					"message": "请求体过大",
					"data":    nil,
				})
				c.Abort()
				return
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

			// 检查请求参数中的恶意模式
			if err := validateRequest(c, blockedPatterns); err != nil {
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
			// 读取请求体
			body, err := c.GetRawData()
			if err == nil && len(body) > 0 {
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

// AdminSecurityMiddleware 管理员接口安全中间件
func AdminSecurityMiddleware() gin.HandlerFunc {
	// 更严格的配置
	config := DefaultSecurityConfig()
	config.RateLimit.MaxRequests = 30                       // 更严格的IP限制
	config.RateLimit.UserMaxRequest = 50                    // 更严格的用户限制
	config.InputValidation.MaxRequestSize = 5 * 1024 * 1024 // 5MB

	return SecurityMiddleware(config)
}

// AdminSecurityMiddlewareFromConfig 从配置文件创建管理员安全中间件
func AdminSecurityMiddlewareFromConfig(cfg *config.Config) gin.HandlerFunc {
	if !cfg.Security.AdminSecurity.Enabled {
		return AdminSecurityMiddleware()
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

// getDefaultBlockedPatterns 获取默认的阻止模式
func getDefaultBlockedPatterns() []string {
	return []string{
		`(?i)<script[^>]*>.*?</script>`, // XSS
		`(?i)javascript:`,               // JavaScript URLs
		`(?i)on\w+\s*=`,                 // Event handlers
		`(?i)union.*select`,             // SQL Injection
		`(?i)insert.*into`,              // SQL Injection
		`(?i)delete.*from`,              // SQL Injection
		`(?i)drop.*table`,               // SQL Injection
		`(?i)or.*=`,                     // SQL Injection OR patterns
		`(?i)and.*=`,                    // SQL Injection AND patterns
		`(?i)exec\s*\(`,                 // Command execution
		`(?i)system\s*\(`,               // Command execution
		`(?i)\.\.\/`,                    // Path traversal
		`(?i)\.\.\\`,                    // Path traversal (Windows)
	}
}
