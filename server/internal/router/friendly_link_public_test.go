package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/handler"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// visibleLinkRepo 友链公开端点所需的仓储替身。
// 内嵌接口继承未触达的方法，被公开列表与访客申请路径调用到的方法显式覆写，
// 避免接口演进时以内嵌空接口的 nil panic 暴露。
type visibleLinkRepo struct {
	repository.FriendlyLinkRepositoryInterface
	links []*model.FriendlyLink
}

// ListVisible 返回全部替身记录，状态过滤由真实仓储承担，此处只验证输出窄化。
func (r *visibleLinkRepo) ListVisible() ([]*model.FriendlyLink, error) {
	return r.links, nil
}

// GetByURL 单站点 URL 未命中，使访客申请路径可直接落库。
func (r *visibleLinkRepo) GetByURL(string) (*model.FriendlyLink, error) {
	return nil, domain.ErrFriendlyLinkNotFound
}

// Create 记录访客申请落库结果。
func (r *visibleLinkRepo) Create(link *model.FriendlyLink) error {
	link.ID = uint(len(r.links) + 1)
	r.links = append(r.links, link)
	return nil
}

// newFriendlyLinkRoutesUnderTest 组装仅包含友链路由模块的被测路由。
func newFriendlyLinkRoutesUnderTest(repo repository.FriendlyLinkRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	routes := NewFriendlyLinkRoutes(
		handler.NewFriendlyLinkHandler(service.NewFriendlyLinkService(repo)),
		&stubTokenService{},
		&stubIdentity{},
		service.NewRBACService(),
	)

	router := gin.New()
	api := router.Group("/api")
	routes.RegisterRoutes(api, api.Group("/admin"))
	return router
}

// performFriendlyLinkPost 以指定请求体调用公开友链端点并返回响应记录器。
func performFriendlyLinkPost(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// TestFriendlyLinkPublicListHidesContactFields 验证公开友链列表的 HTTP 响应不含联系邮箱与备注。
func TestFriendlyLinkPublicListHidesContactFields(t *testing.T) {
	repo := &visibleLinkRepo{
		links: []*model.FriendlyLink{
			{
				ID:           1,
				Name:         "示例站点",
				URL:          "https://example.com",
				Description:  "示例简介",
				ContactEmail: "owner@example.com",
				Note:         "内部备注：2026 年收录",
				Status:       model.LinkStatusActive,
				IsReciprocal: true,
			},
		},
	}
	recorder := performFriendlyLinkPost(newFriendlyLinkRoutesUnderTest(repo), "/api/friendlyLinks/list", `{}`)

	if recorder.Code != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d, 期望统一信封固定 200", recorder.Code)
	}
	if code := parseResponseCode(t, recorder); code != response.CodeSuccess {
		t.Fatalf("公开友链列表业务码 = %d, 期望 %d，body=%s", code, response.CodeSuccess, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, forbidden := range []string{"contactEmail", "owner@example.com", "note", "内部备注"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("公开友链列表响应不应包含 %s，实际为 %s", forbidden, body)
		}
	}

	var envelope struct {
		Data struct {
			Links []map[string]any `json:"links"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("解析公开友链列表响应失败: %v，body=%s", err, body)
	}
	if len(envelope.Data.Links) != 1 {
		t.Fatalf("公开友链条目数 = %d, 期望 1", len(envelope.Data.Links))
	}

	// 输出键集须与窄化视图白名单一致，防止个人信息字段被悄然加回。
	allowedKeys := map[string]bool{
		"id": true, "name": true, "url": true, "logo": true,
		"description": true, "sortOrder": true, "isReciprocal": true,
	}
	for key := range envelope.Data.Links[0] {
		if !allowedKeys[key] {
			t.Errorf("公开友链视图不应输出未知字段 %q，实际响应为 %s", key, body)
		}
	}
}

// TestFriendlyLinkApplyDoesNotEchoContactEmail 验证访客申请响应不回显提交的联系邮箱。
func TestFriendlyLinkApplyDoesNotEchoContactEmail(t *testing.T) {
	repo := &visibleLinkRepo{}
	router := newFriendlyLinkRoutesUnderTest(repo)

	recorder := performFriendlyLinkPost(router, "/api/friendlyLinks/apply",
		`{"name":"朋友的博客","url":"https://friend.example.com","contactEmail":"friend@example.com"}`)

	if code := parseResponseCode(t, recorder); code != response.CodeSuccess {
		t.Fatalf("友链申请业务码 = %d, 期望 %d，body=%s", code, response.CodeSuccess, recorder.Body.String())
	}

	body := recorder.Body.String()
	if strings.Contains(body, "contactEmail") || strings.Contains(body, "friend@example.com") {
		t.Errorf("友链申请响应不应回显联系邮箱，实际为 %s", body)
	}
	// 申请仍须落库保留联系邮箱，供管理员审核时联系站长。
	if len(repo.links) != 1 || repo.links[0].ContactEmail != "friend@example.com" {
		t.Error("友链申请的联系邮箱应保留在落库记录中")
	}
}
