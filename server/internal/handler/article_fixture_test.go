package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// fakeArticleFixtureService 文章契约场景测试替身，仅覆盖金样本断言涉及的方法。
type fakeArticleFixtureService struct {
	service.ArticleServiceInterface
	article    *model.Article
	articleErr error
	listResult *service.ArticleListResponse
	archives   []service.ArticleArchiveYear
	deleteErr  error
}

func (f *fakeArticleFixtureService) GetArticle(id uint, userID *uint) (*model.Article, error) {
	if f.articleErr != nil {
		return nil, f.articleErr
	}
	return f.article, nil
}

func (f *fakeArticleFixtureService) CreateArticle(req *service.CreateArticleRequest, authorID uint) (*model.Article, error) {
	return f.article, nil
}

func (f *fakeArticleFixtureService) DeleteArticle(id uint, userID uint) error {
	return f.deleteErr
}

func (f *fakeArticleFixtureService) GetArticleList(req *service.GetArticleListRequest, userID *uint) (*service.ArticleListResponse, error) {
	return f.listResult, nil
}

func (f *fakeArticleFixtureService) GetArticleArchives() ([]service.ArticleArchiveYear, error) {
	return f.archives, nil
}

// newArticleFixtureContext 构造携带 JSON 请求体的 gin 上下文，供文章契约断言复用。
func newArticleFixtureContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, recorder
}

// TestArticleListMatchesFixture 验证文章列表响应与契约金样本一致，构成契约锁 ①。
func TestArticleListMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewArticleHandler(&fakeArticleFixtureService{
		listResult: &service.ArticleListResponse{
			Articles: []*model.Article{newFixtureArticle()},
			Total:    1,
			Page:     1,
			PageSize: 10,
		},
	})

	ctx, recorder := newArticleFixtureContext(http.MethodPost, "/api/articles/list",
		`{"page":1,"pageSize":10}`)
	handler.GetArticleList(ctx)

	assertResponseMatchesFixture(t, recorder, "article.list.json")
}

// TestArticleArchivesMatchesFixture 验证归档响应与契约金样本一致，构成契约锁 ①。
func TestArticleArchivesMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewArticleHandler(&fakeArticleFixtureService{
		archives: []service.ArticleArchiveYear{{
			Year:  2026,
			Total: 1,
			Months: []service.ArticleArchiveMonth{{
				Month:    1,
				Articles: []*model.Article{newFixtureArticle()},
			}},
		}},
	})

	ctx, recorder := newArticleFixtureContext(http.MethodPost, "/api/articles/archives", `{}`)
	handler.GetArticleArchives(ctx)

	assertResponseMatchesFixture(t, recorder, "article.archives.json")
}

// TestArticleCreateDraftMatchesFixture 验证草稿创建响应与契约金样本一致，构成契约锁 ①。
// 未指定状态时创建为草稿，publishedAt 为空是该状态的对外契约。
func TestArticleCreateDraftMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	draft := newFixtureArticle()
	draft.Status = model.ArticleStatusDraft
	draft.PublishedAt = nil

	handler := NewArticleHandler(&fakeArticleFixtureService{article: draft})

	ctx, recorder := newArticleFixtureContext(http.MethodPost, "/api/articles/create",
		`{"title":"示例文章标题","content":"Markdown 正文内容"}`)
	ctx.Set("userID", uint(1))
	handler.CreateArticle(ctx)

	assertResponseMatchesFixture(t, recorder, "article.created.json")
}

// TestArticleNotFoundMatchesFixture 验证文章不存在响应与契约金样本一致，构成契约锁 ①。
func TestArticleNotFoundMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewArticleHandler(&fakeArticleFixtureService{articleErr: domain.ErrArticleNotFound})

	ctx, recorder := newArticleFixtureContext(http.MethodPost, "/api/articles/get", `{"id":1}`)
	handler.GetArticle(ctx)

	assertResponseMatchesFixture(t, recorder, "article.not-found.json")
}

// TestArticleDeletedMatchesFixture 验证删除文章响应与契约金样本一致，构成契约锁 ①。
func TestArticleDeletedMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewArticleHandler(&fakeArticleFixtureService{})

	ctx, recorder := newArticleFixtureContext(http.MethodPost, "/api/articles/delete", `{"id":1}`)
	ctx.Set("userID", uint(1))
	handler.DeleteArticle(ctx)

	assertResponseMatchesFixture(t, recorder, "article.deleted.json")
}
