package middleware

import (
	"MyBlog/internal/config"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// compileDefaultPatterns 编译默认阻止模式，供模式级断言使用。
func compileDefaultPatterns(t *testing.T) []*regexp.Regexp {
	t.Helper()

	var compiled []*regexp.Regexp
	for _, pattern := range getDefaultBlockedPatterns() {
		re, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatalf("默认阻止模式编译失败 %q: %v", pattern, err)
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// TestDefaultBlockedPatternsCatchAttacks 默认模式必须拦截的攻击载荷。
func TestDefaultBlockedPatternsCatchAttacks(t *testing.T) {
	patterns := compileDefaultPatterns(t)

	mustBlock := []string{
		`<script>alert(1)</script>`,
		`javascript:void(0)`,
		`<img src=x onerror="alert(1)">`,
		`<img src=x onerror='alert(1)'>`,
		`<body onload=alert(1)>`,
		`' OR 1=1 --`,
		`admin' OR '1'='1`,
		`UNION SELECT password FROM users`,
		`DROP TABLE users`,
		`../../etc/passwd`,
	}

	for _, payload := range mustBlock {
		if !containsMaliciousContent(payload, patterns) {
			t.Errorf("攻击载荷未被拦截: %q", payload)
		}
	}
}

// TestDefaultBlockedPatternsAllowNormalContent 正常内容不得被误伤，
// 场景来自博客正文常见写法，含中文正文与代码片段。
func TestDefaultBlockedPatternsAllowNormalContent(t *testing.T) {
	patterns := compileDefaultPatterns(t)

	mustAllow := []string{
		`session 与 comparison 两个词的词源对比`,
		`配置项 content = '示例文本' 的说明`,           // 旧模式 on\w+\s*= 命中 content =
		`button = primary 时组件呈现高亮态`,          // 旧模式命中 button =
		`监视器 monitor = 1 表示开启`,               // 旧模式命中 monitor =
		`for i = 1; i <= 10; i++ 是常见循环写法`,    // 旧模式 or.*= 命中 for i =
		`前端与后端，brand = design 的说明`,           // 旧模式 and.*= 命中 brand 前的 and
		`这条记录已从缓存移除`,                         // 旧模式 delete.*from 类语义的中文表述
		`React 组件的 onClick={handleClick} 写法`, // JSX 事件绑定属正常代码内容
	}

	for _, content := range mustAllow {
		if containsMaliciousContent(content, patterns) {
			t.Errorf("正常内容被误伤拦截: %q", content)
		}
	}
}

// serveWithMiddleware 将中间件挂到最小路由上执行请求，返回响应记录。
func serveWithMiddleware(handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(handler)
	router.POST("/api/articles/create", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/articles/create", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// TestSecurityMiddlewareAllowsNormalArticleBody 通过完整中间件链验证
// 中文正文保存请求不再被 WAF 拦截，对应体检项 BE-07 的文章保存 403 场景。
func TestSecurityMiddlewareAllowsNormalArticleBody(t *testing.T) {
	body := `{"title":"正则表达式入门","content":"# 第一节\n\ncontent = '匹配任意字符' 是常见写法，" +
		"例如 for i = 1 到 10 的循环。修饰词 comparison 与 session 的说明如下……"}`

	recorder := serveWithMiddleware(SecurityMiddleware(DefaultSecurityConfig()), body)

	if recorder.Code == http.StatusBadRequest {
		t.Fatalf("正常文章正文被 WAF 误伤，响应: %s", recorder.Body.String())
	}
}

// TestSecurityMiddlewareBlocksXSSBody 通过完整中间件链验证攻击载荷仍被拦截。
func TestSecurityMiddlewareBlocksXSSBody(t *testing.T) {
	body := `{"title":"test","content":"<img src=x onerror=\"alert(1)\">"}`

	recorder := serveWithMiddleware(SecurityMiddleware(DefaultSecurityConfig()), body)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("XSS 载荷应被拦截，实际状态码 %d", recorder.Code)
	}
}

// TestSecurityMiddlewareRejectsOversizedChunkedBody 验证分块传输编码下的超大请求体仍被拒绝。
// 分块编码不携带 Content-Length，仅凭该头部无法拦截，体积上限必须由读取层的 MaxBytesReader 生效。
func TestSecurityMiddlewareRejectsOversizedChunkedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityConfig := DefaultSecurityConfig()
	securityConfig.InputValidation.MaxRequestSize = 1024

	router := gin.New()
	router.Use(SecurityMiddleware(securityConfig))
	router.POST("/api/articles/create", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 构造超过上限的请求体，并把 ContentLength 置为 -1 模拟分块传输编码。
	oversizedBody := strings.Repeat("a", 4096)
	request := httptest.NewRequest(http.MethodPost, "/api/articles/create", strings.NewReader(oversizedBody))
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = -1

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("超大请求体应返回 413，实际状态码 %d", recorder.Code)
	}
}

// TestSecurityMiddlewarePanicsOnInvalidPattern 验证阻止模式无法编译时在启动期直接失败，
// 而不是静默跳过该模式使 WAF 在无告警的情况下失效。
func TestSecurityMiddlewarePanicsOnInvalidPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityConfig := DefaultSecurityConfig()
	securityConfig.InputValidation.BlockedPatterns = []string{"("}

	defer func() {
		if recover() == nil {
			t.Error("非法阻止模式应触发 panic，实际未发生")
		}
	}()

	SecurityMiddleware(securityConfig)
}

// TestAdminSecurityDisabledDoesNotTighten 验证管理员安全开关关闭时不叠加任何额外限制。
// 原实现在开关关闭时回退到硬编码的更严格配置，与开关语义相反。
func TestAdminSecurityDisabledDoesNotTighten(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Security.AdminSecurity.Enabled = false

	router := gin.New()
	router.Use(AdminSecurityMiddlewareFromConfig(cfg))
	router.POST("/api/admin/users/list", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 开关关闭后不再叠加 WAF 等内容级检查，含攻击特征的请求也应放行。
	request := httptest.NewRequest(http.MethodPost, "/api/admin/users/list",
		strings.NewReader(`{"keyword":"<script>alert(1)</script>"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("开关关闭时请求应放行，实际状态码 %d", recorder.Code)
	}
}
