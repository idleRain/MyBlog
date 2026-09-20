package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"

	"github.com/gin-gonic/gin"
)

// newLanguageProbeRouter 构造携带语言上下文并挂载探针处理器的测试路由。
func newLanguageProbeRouter(language domain.Language) (*gin.Engine, *gin.Context) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(domain.LanguageContextKey, language)
		c.Next()
	})

	var captured *gin.Context
	router.POST("/probe", func(c *gin.Context) {
		captured = c
		respondLocalizedArticle(c, newProbeArticle())
	})
	return router, captured
}

// newProbeArticle 构造带英文翻译行的探针文章。
func newProbeArticle() *model.Article {
	return &model.Article{
		ID:    1,
		Title: "你好世界",
		Translations: []model.ArticleTranslation{
			{Locale: "en", Title: "Hello World"},
		},
	}
}

// parseProbeResponse 解析探针响应体中的文章标题与业务码。
func parseProbeResponse(t *testing.T, recorder *httptest.ResponseRecorder) (string, float64) {
	t.Helper()

	var body struct {
		Code float64 `json:"code"`
		Data struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return body.Data.Title, body.Code
}

// TestRespondLocalizedArticleByLanguage 验证文章响应按请求语言本地化并标注实际语言。
func TestRespondLocalizedArticleByLanguage(t *testing.T) {
	cases := []struct {
		name           string
		language       domain.Language
		expectedTitle  string
		expectedHeader string
	}{
		{name: "英文请求输出翻译内容", language: domain.LanguageEnglish, expectedTitle: "Hello World", expectedHeader: "en"},
		{name: "中文请求保留主列内容", language: domain.DefaultLanguage, expectedTitle: "你好世界", expectedHeader: "zh"},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			router, _ := newLanguageProbeRouter(item.language)

			request := httptest.NewRequest(http.MethodPost, "/probe", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			title, code := parseProbeResponse(t, recorder)
			if code != 200 {
				t.Errorf("业务码 = %v, 期望 200", code)
			}
			if title != item.expectedTitle {
				t.Errorf("Title = %q, 期望 %q", title, item.expectedTitle)
			}
			if header := recorder.Header().Get(ContentLanguageHeader); header != item.expectedHeader {
				t.Errorf("Content-Language = %q, 期望 %q", header, item.expectedHeader)
			}
		})
	}
}

// TestRespondLocalizedArticleAllModeKeepsTranslations 验证全量包请求保留翻译行并输出翻译状态。
func TestRespondLocalizedArticleAllModeKeepsTranslations(t *testing.T) {
	router, _ := newLanguageProbeRouter(domain.LanguageAll)

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var body struct {
		Data struct {
			Title              string   `json:"title"`
			TranslationLocales []string `json:"translationLocales"`
			Translations       []struct {
				Locale string `json:"locale"`
				Title  string `json:"title"`
			} `json:"translations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}

	if body.Data.Title != "你好世界" {
		t.Errorf("Title = %q, 期望保留主列内容", body.Data.Title)
	}
	if len(body.Data.Translations) != 1 || body.Data.Translations[0].Title != "Hello World" {
		t.Error("全量包请求应保留翻译行")
	}
	if len(body.Data.TranslationLocales) != 1 || body.Data.TranslationLocales[0] != "en" {
		t.Errorf("TranslationLocales = %v, 期望 [en]", body.Data.TranslationLocales)
	}
	if header := recorder.Header().Get(ContentLanguageHeader); header != "*" {
		t.Errorf("Content-Language = %q, 期望 *", header)
	}
}
