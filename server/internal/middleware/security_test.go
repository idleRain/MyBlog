package middleware

import (
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
