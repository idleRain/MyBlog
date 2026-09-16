package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// benchArticleService 性能基准用的文章服务替身，返回固定的 canned 响应，
// 度量对象为路由、参数绑定与响应封装的 HTTP 处理路径开销。
type benchArticleService struct {
	service.ArticleServiceInterface
	list    *service.ArticleListResponse
	article *model.Article
}

func (s *benchArticleService) GetArticleList(*service.GetArticleListRequest, *uint) (*service.ArticleListResponse, error) {
	return s.list, nil
}

func (s *benchArticleService) GetArticle(uint, *uint) (*model.Article, error) {
	return s.article, nil
}

// newBenchArticleRouter 构造挂载文章列表与详情路由的基准引擎。
func newBenchArticleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	articles := make([]*model.Article, 0, 20)
	for i := 1; i <= 20; i++ {
		articles = append(articles, &model.Article{
			ID:     uint(i),
			Title:  "基准测试文章",
			Slug:   "bench-article",
			Status: model.ArticleStatusPublished,
		})
	}

	svc := &benchArticleService{
		list: &service.ArticleListResponse{Articles: articles, Total: int64(len(articles)), Page: 1, PageSize: 20},
		article: &model.Article{
			ID:     1,
			Title:  "基准测试文章",
			Slug:   "bench-article",
			Status: model.ArticleStatusPublished,
		},
	}

	router := gin.New()
	articleHandler := NewArticleHandler(svc)
	router.POST("/api/articles", articleHandler.GetArticleList)
	router.POST("/api/articles/get", articleHandler.GetArticle)
	return router
}

// BenchmarkArticleListHandler 文章列表接口的 HTTP 处理路径基准。
func BenchmarkArticleListHandler(b *testing.B) {
	router := newBenchArticleRouter()
	body := `{"page":1,"pageSize":20}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, "/api/articles", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			b.Fatalf("列表接口返回 %d", recorder.Code)
		}
	}
}

// BenchmarkArticleDetailHandler 文章详情接口的 HTTP 处理路径基准。
func BenchmarkArticleDetailHandler(b *testing.B) {
	router := newBenchArticleRouter()
	body := `{"id":1}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, "/api/articles/get", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			b.Fatalf("详情接口返回 %d", recorder.Code)
		}
	}
}
