package middleware

import (
	"MyBlog/internal/config"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// serveRequestWithMiddleware 将中间件挂到文章创建路由上执行请求，返回响应记录。
// target 允许携带查询串，headers 用于覆盖需要检查的请求头。
func serveRequestWithMiddleware(
	handler gin.HandlerFunc,
	target string,
	body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(handler)
	router.POST("/api/articles/create", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// serveWithMiddleware 以默认安全配置对文章创建请求执行一次中间件链。
func serveWithMiddleware(handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	return serveRequestWithMiddleware(handler, "/api/articles/create", body, nil)
}

// TestSecurityMiddlewareAllowsNormalArticleBody 通过完整中间件链验证
// 中文正文保存请求不再被 WAF 拦截，覆盖此前被误判为攻击的文章保存 403 场景。
func TestSecurityMiddlewareAllowsNormalArticleBody(t *testing.T) {
	body := `{"title":"正则表达式入门","content":"# 第一节\n\ncontent = '匹配任意字符' 是常见写法，" +
		"例如 for i = 1 到 10 的循环。修饰词 comparison 与 session 的说明如下……"}`

	recorder := serveWithMiddleware(SecurityMiddleware(DefaultSecurityConfig()), body)

	if recorder.Code == http.StatusBadRequest {
		t.Fatalf("正常文章正文被 WAF 误伤，响应: %s", recorder.Body.String())
	}
}

// TestSecurityMiddlewareBlocksXSSBody 通过完整中间件链验证参数位置的攻击载荷仍被拦截。
// 正文以外的字段不参与字段级豁免，攻击载荷放在标题时整条请求必须被拒绝。
func TestSecurityMiddlewareBlocksXSSBody(t *testing.T) {
	body := `{"title":"<img src=x onerror=\"alert(1)\">","content":"正常正文"}`

	recorder := serveWithMiddleware(SecurityMiddleware(DefaultSecurityConfig()), body)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("参数位置的 XSS 载荷应被拦截，实际状态码 %d", recorder.Code)
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

// TestSecurityMiddlewareAllowsMarkdownBody 验证技术文章正文中的 SQL 与 HTML 示例不再被 WAF 拦截。
// 正文限定为富文本字段，其最终输出必经消毒渲染：服务端 pkg/markdown 关闭 goldmark 的 Unsafe，
// 前台 lib/markdown/render 丢弃原始 HTML，两处均有独立用例锚定。参数位置不参与豁免。
func TestSecurityMiddlewareAllowsMarkdownBody(t *testing.T) {
	handler := SecurityMiddleware(DefaultSecurityConfig())

	cases := []struct {
		name string
		body string
	}{
		{
			name: "正文含 SQL 与相对路径示例",
			body: "{\"title\":\"用 SQL 做数据迁移\",\"content\":\"迁移脚本会执行 INSERT INTO users 语句，" +
				"清理时使用 DELETE FROM sessions，必要时才 DROP TABLE tmp_cache。" +
				"```sql\\nSELECT 1;\\n```\\n\\n引用图片可写作 ../../assets/logo.png。\"}",
		},
		{
			name: "正文含脚本标签示例",
			body: "{\"title\":\"脚本加载顺序\",\"content\":\"把 <script src=\\\"app.js\\\"></script> 放在正文末尾之前。\"}",
		},
		{
			name: "正文含事件属性示例",
			body: "{\"title\":\"事件属性写法\",\"content\":\"旧写法 onerror=\\\"handle()\\\" 已不再推荐。\"}",
		},
		{
			name: "翻译正文含脚本标签示例",
			body: "{\"title\":\"Migration\",\"content\":\"正文\",\"i18n\":{\"en\":{\"title\":\"Migration\"," +
				"\"content\":\"示例：<script src=\\\"app.js\\\"></script>\"}}}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := serveWithMiddleware(handler, testCase.body)
			if recorder.Code != http.StatusOK {
				t.Errorf("富文本正文被 WAF 误拦，状态码 %d，响应 %s",
					recorder.Code, recorder.Body.String())
			}
		})
	}
}

// TestSecurityMiddlewareBlocksAttacksOutsideMarkdownBody 验证正文以外的攻击面保持全量扫描。
// 标题、摘要、slug、翻译的非正文字段、URL 参数、请求头与字段名本身都不参与字段级豁免，
// 结构无法解析的请求体退化为整段扫描，避免借畸形结构绕过字段判定。
func TestSecurityMiddlewareBlocksAttacksOutsideMarkdownBody(t *testing.T) {
	handler := SecurityMiddleware(DefaultSecurityConfig())

	cases := []struct {
		name    string
		target  string
		body    string
		headers map[string]string
	}{
		{
			name: "标题含脚本标签",
			body: "{\"title\":\"<script>alert(1)</script>\",\"content\":\"正常正文\"}",
		},
		{
			name: "摘要含事件属性",
			body: "{\"title\":\"正常标题\",\"content\":\"正常正文\",\"summary\":\"<img src=x onerror=alert(1)>\"}",
		},
		{
			name: "slug 含 SQL 恒真式",
			body: "{\"title\":\"正常标题\",\"content\":\"正常正文\",\"slug\":\"' OR 1=1 --\"}",
		},
		{
			name: "翻译 SEO 字段含删表语句",
			body: "{\"title\":\"正常标题\",\"content\":\"正常正文\",\"i18n\":{\"en\":{\"seoTitle\":\"DROP TABLE users\"}}}",
		},
		{
			name:   "URL 查询参数含 SQL 恒真式",
			target: "/api/articles/create?keyword=" + url.QueryEscape("' OR 1=1 --"),
			body:   "{\"title\":\"正常标题\",\"content\":\"正常正文\"}",
		},
		{
			name:    "请求头含脚本标签",
			body:    "{\"title\":\"正常标题\",\"content\":\"正常正文\"}",
			headers: map[string]string{"Referer": "<script>alert(1)</script>"},
		},
		{
			name: "字段名含脚本标签",
			body: "{\"<script>alert(1)</script>\":\"取值\"}",
		},
		{
			name: "正文取值为数组时按类型不符处理",
			body: "{\"title\":\"正常标题\",\"content\":[\"<script>alert(1)</script>\"]}",
		},
		{
			name: "结构无法解析时回退整段扫描",
			body: "{\"title\":\"<script>alert(1)</script>\"",
		},
		{
			// JSON 转义序列在原始字节里以 n 收尾，会遮蔽 \b 词边界；
			// 按解码后的取值扫描后，行首的 SQL 特征重新进入检测范围。
			name: "换行转义不再遮蔽 SQL 特征",
			body: "{\"title\":\"示例\\nINSERT INTO users (name) VALUES ('a');\",\"content\":\"正常正文\"}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := testCase.target
			if target == "" {
				target = "/api/articles/create"
			}

			recorder := serveRequestWithMiddleware(handler, target, testCase.body, testCase.headers)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("参数位置载荷应被拦截，实际状态码 %d，响应 %s",
					recorder.Code, recorder.Body.String())
			}
		})
	}
}
