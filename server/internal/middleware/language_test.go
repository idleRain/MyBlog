package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/domain"

	"github.com/gin-gonic/gin"
)

// languageTestConfig 受支持语言为中文与英文的测试配置。
func languageTestConfig() config.I18NConfig {
	return config.I18NConfig{
		DefaultLanguage:    "zh",
		SupportedLanguages: []string{"zh", "en"},
	}
}

// newLanguageTestRouter 构造挂载语言中间件与探针处理器的测试路由，探针记录解析结果。
func newLanguageTestRouter(cfg config.I18NConfig) (*gin.Engine, *domain.Language, *bool) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LanguageMiddlewareFromConfig(cfg))

	resolved := domain.DefaultLanguage
	nextReached := false
	router.POST("/probe", func(c *gin.Context) {
		value, _ := c.Get(domain.LanguageContextKey)
		language, _ := value.(domain.Language)
		resolved = language
		nextReached = true
		c.Status(http.StatusOK)
	})
	return router, &resolved, &nextReached
}

// TestLanguageMiddleware 覆盖请求头各形态下的语言解析与上下文写入。
func TestLanguageMiddleware(t *testing.T) {
	cases := []struct {
		name     string
		header   string
		expected domain.Language
	}{
		{name: "未携带请求头回退缺省语言", header: "", expected: domain.LanguageChinese},
		{name: "地区变体归一化命中英文", header: "en-US", expected: domain.LanguageEnglish},
		{name: "白名单外语言回退缺省语言", header: "fr-FR", expected: domain.LanguageChinese},
		{name: "多标签按出现顺序命中", header: "fr, en;q=0.8", expected: domain.LanguageEnglish},
		{name: "通配符整头透传全量包标识", header: "*", expected: domain.LanguageAll},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			router, resolved, nextReached := newLanguageTestRouter(languageTestConfig())

			request := httptest.NewRequest(http.MethodPost, "/probe", nil)
			if item.header != "" {
				request.Header.Set(AcceptLanguageHeader, item.header)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if *resolved != item.expected {
				t.Errorf("解析语言 = %q, 期望 %q", *resolved, item.expected)
			}
			if !*nextReached {
				t.Error("语言中间件不应中断请求链")
			}
			if recorder.Code != http.StatusOK {
				t.Errorf("探针响应状态码 = %d, 期望 200", recorder.Code)
			}
		})
	}
}

// TestLanguageMiddlewareUsesConfiguredFallback 验证缺省语言来自配置而非硬编码。
func TestLanguageMiddlewareUsesConfiguredFallback(t *testing.T) {
	cfg := languageTestConfig()
	cfg.DefaultLanguage = "en"

	router, resolved, _ := newLanguageTestRouter(cfg)

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if *resolved != domain.LanguageEnglish {
		t.Errorf("解析语言 = %q, 期望回退到配置缺省语言 en", *resolved)
	}
}
