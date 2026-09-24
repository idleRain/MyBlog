package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/service"
	"MyBlog/pkg/datetime"

	"github.com/gin-gonic/gin"
)

// fakeUserListService 用户列表场景测试替身，仅覆盖 GetUserList 方法。
type fakeUserListService struct {
	service.UserService
	users []*domain.User
	total int64
	err   error
}

func (f *fakeUserListService) GetUserList(page, pageSize int, keyword string) ([]*domain.User, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.users, f.total, nil
}

// fakeArticleService 文章详情场景测试替身，仅覆盖 GetArticle 方法。
type fakeArticleService struct {
	service.ArticleServiceInterface
	article *model.Article
	err     error
}

func (f *fakeArticleService) GetArticle(id uint, userID *uint) (*model.Article, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.article, nil
}

// newFixtureUser 构造与金样本一致的测试用户，时间字段落在 JSONDate 序列化格式内。
func newFixtureUser() *domain.User {
	return &domain.User{
		ID:        1,
		Username:  "admin",
		Email:     "admin@myblog.local",
		Nickname:  "admin",
		Role:      "superadmin",
		Status:    1,
		Birthday:  datetime.JSONDate{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)},
		CreatedAt: time.Date(2026, 1, 1, 8, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 8, 30, 0, 0, time.UTC),
	}
}

// testSessionCookieConfig 测试用会话 Cookie 配置。
// 有效期取值与生产配置同数量级，仅承载写入断言，不参与业务断言。
func testSessionCookieConfig() SessionCookieConfig {
	return SessionCookieConfig{
		AccessMaxAge:  15 * 60,
		RefreshMaxAge: 168 * 3600,
		Secure:        false,
	}
}

// TestLoginSuccessMatchesFixture 验证登录成功响应与契约金样本一致，构成契约锁 ①。
func TestLoginSuccessMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserHandler(&fakeLoginUserService{loginResp: &service.LoginResponse{
		User:         newFixtureUser(),
		AccessToken:  "3f2a9c1e4b7d8056a1c3e5f70b2d4689",
		RefreshToken: "8b1d4e7a2c9f0356b8e1a4d7c0f32695",
		ExpiresIn:    1800,
		Permissions:  []string{"article:read", "comment:create", "comment:read", "comment:update"},
	}}, testSessionCookieConfig())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/users/login",
		strings.NewReader(`{"username":"admin","password":"correct-pass1"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Login(ctx)

	assertResponseMatchesFixture(t, recorder, "login.success.json")
}

// TestUserListMatchesFixture 验证用户列表响应与契约金样本一致，构成契约锁 ①。
func TestUserListMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserHandler(&fakeUserListService{users: []*domain.User{newFixtureUser()}, total: 1},
		testSessionCookieConfig())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/users/list",
		strings.NewReader(`{"page":1,"pageSize":10}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.GetUserList(ctx)

	assertResponseMatchesFixture(t, recorder, "users.list.json")
}

// newFixtureArticle 构造与金样本一致的测试文章，作者经公开视图输出并携带分类与标签关联。
func newFixtureArticle() *model.Article {
	loc := time.FixedZone("UTC+8", 8*3600)
	categoryID := uint(1)
	category := &model.Category{
		ID: 1, Name: "示例分类", Slug: "example-category", Level: 1, Path: "/1", Status: 1,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
	}
	tag := model.Tag{
		ID: 1, Name: "示例标签", Slug: "example-tag", Color: "#808080", Status: 1,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
	}
	return &model.Article{
		ID:             1,
		Title:          "示例文章标题",
		Slug:           "example-article",
		Summary:        "示例文章摘要",
		Content:        "Markdown 正文内容",
		ContentHTML:    "",
		CoverImage:     "",
		AuthorID:       1,
		CategoryID:     &categoryID,
		Status:         model.ArticleStatusPublished,
		OriginType:     "original",
		SourceURL:      "",
		SourceAuthor:   "",
		IsFeatured:     false,
		IsTop:          false,
		CommentEnabled: true,
		Version:        1,
		PublishedAt:    timePtr(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
		UpdatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
		AuthorPublic: &domain.AuthorPublic{
			ID: 1, Username: "admin", Nickname: "admin", Avatar: "", Bio: "", Website: "",
		},
		Category:   category,
		Categories: []model.Category{*category},
		Tags:       []model.Tag{tag},
	}
}

// timePtr 返回指向给定时间的指针，用于构造可选时间字段。
func timePtr(t time.Time) *time.Time {
	return &t
}

// TestArticleDetailMatchesFixture 验证文章详情响应与契约金样本一致，构成契约锁 ①。
func TestArticleDetailMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewArticleHandler(&fakeArticleService{article: newFixtureArticle()})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/articles/get",
		strings.NewReader(`{"id":1}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.GetArticle(ctx)

	assertResponseMatchesFixture(t, recorder, "article.detail.json")
}

// TestFixtureFilesParse 验证全部契约金样本为合法 JSON，防止手工修订引入语法错误。
func TestFixtureFilesParse(t *testing.T) {
	fixtureNames := []string{
		"login.success.json",
		"login.wrong-password.json",
		"users.list.json",
		"article.detail.json",
	}
	for _, name := range fixtureNames {
		fixtureBytes, err := os.ReadFile(fixtureContractPath + "/" + name)
		if err != nil {
			t.Fatalf("读取金样本 %s 失败: %v", name, err)
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, fixtureBytes); err != nil {
			t.Errorf("金样本 %s 不是合法 JSON: %v", name, err)
		}
	}
}
