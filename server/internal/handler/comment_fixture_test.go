package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// fakeCommentFixtureService 评论契约场景测试替身，仅覆盖金样本断言涉及的方法。
type fakeCommentFixtureService struct {
	service.CommentServiceInterface
	created      *model.Comment
	listResult   *service.CommentListResponse
	adminResult  *service.AdminCommentListResponse
	moderateCall string
}

func (f *fakeCommentFixtureService) CreateComment(req *service.CreateCommentRequest, userID *uint) (*model.Comment, error) {
	return f.created, nil
}

func (f *fakeCommentFixtureService) GetCommentsByArticle(articleID uint, req *service.ListCommentsRequest) (*service.CommentListResponse, error) {
	return f.listResult, nil
}

func (f *fakeCommentFixtureService) ListComments(req *service.AdminListCommentsRequest) (*service.AdminCommentListResponse, error) {
	return f.adminResult, nil
}

// ApproveComment 记录状态流转入口被调用，金样本只断言响应，故不返回错误。
func (f *fakeCommentFixtureService) ApproveComment(id uint, operatorID uint) error {
	f.moderateCall = "approve"
	return nil
}

// newFixtureComment 构造与金样本一致的游客待审核评论，文章与回复关联均未预加载。
func newFixtureComment() *model.Comment {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	return &model.Comment{
		ID:         1,
		ArticleID:  1,
		Level:      1,
		AuthorName: "访客",
		Content:    "示例评论内容",
		Status:     model.CommentStatusPending,
		CreatedAt:  created,
		UpdatedAt:  created,
	}
}

// newFixtureApprovedComment 构造已审核通过的游客评论，公开列表只返回该状态。
func newFixtureApprovedComment() *model.Comment {
	comment := newFixtureComment()
	comment.Status = model.CommentStatusApproved
	return comment
}

// newFixtureAdminComment 构造管理端评论视图，携带审计字段与已预加载的文章归属。
// 文章归属按真实预加载结果装配：作者与分类标签关联未随评论加载，故不随响应输出。
func newFixtureAdminComment() *service.AdminCommentView {
	comment := newFixtureApprovedComment()
	userID := uint(1)
	comment.UserID = &userID
	comment.AuthorName = ""
	comment.AuthorEmail = "guest@example.com"
	comment.AuthorIP = "203.0.113.7"
	comment.UserAgent = "curl/8"
	comment.AuthorPublic = domain.NewAuthorPublic(&domain.User{ID: 1, Username: "admin", Nickname: "admin"})

	article := newFixtureArticle()
	article.AuthorPublic = nil
	article.Category = nil
	article.Categories = nil
	article.Tags = nil

	return &service.AdminCommentView{
		Comment:     *comment,
		Article:     article,
		AuthorEmail: comment.AuthorEmail,
		AuthorIP:    comment.AuthorIP,
		UserAgent:   comment.UserAgent,
	}
}

// newCommentFixtureContext 构造携带 JSON 请求体的 gin 上下文，供评论契约断言复用。
func newCommentFixtureContext(path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, recorder
}

// TestCommentCreatedMatchesFixture 验证游客评论提交响应与契约金样本一致，构成契约锁 ①。
// 未登录提交走游客通道，评论者公开信息整体缺省，文章归属不出现在公开契约。
func TestCommentCreatedMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCommentHandler(&fakeCommentFixtureService{created: newFixtureComment()})

	ctx, recorder := newCommentFixtureContext("/api/comments/create",
		`{"articleId":1,"content":"示例评论内容","authorName":"访客"}`)
	handler.CreateComment(ctx)

	assertResponseMatchesFixture(t, recorder, "comment.created.json")
}

// TestCommentListMatchesFixture 验证文章评论列表响应与契约金样本一致，构成契约锁 ①。
func TestCommentListMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCommentHandler(&fakeCommentFixtureService{
		listResult: &service.CommentListResponse{
			Comments: []*model.Comment{newFixtureApprovedComment()},
			Total:    1,
			Page:     1,
			PageSize: 10,
		},
	})

	ctx, recorder := newCommentFixtureContext("/api/comments/list", `{"articleId":1,"page":1,"pageSize":10}`)
	handler.GetCommentsByArticle(ctx)

	assertResponseMatchesFixture(t, recorder, "comment.list.json")
}

// TestCommentApprovedMatchesFixture 验证审核通过响应与契约金样本一致，构成契约锁 ①。
// 状态流转接口只回传结果消息，评论状态本身由管理端列表的 status 字段承载。
func TestCommentApprovedMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeCommentFixtureService{}
	handler := NewCommentHandler(fake)

	ctx, recorder := newCommentFixtureContext("/api/admin/comments/approve", `{"id":1}`)
	ctx.Set("userID", uint(1))
	handler.ApproveComment(ctx)

	if fake.moderateCall != "approve" {
		t.Fatalf("审核通过路径未调用 ApproveComment，实际调用 %q", fake.moderateCall)
	}
	assertResponseMatchesFixture(t, recorder, "comment.approved.json")
}

// TestAdminCommentListMatchesFixture 验证管理端评论列表与契约金样本一致，构成契约锁 ①。
// 管理端在公开契约之上恢复审计字段与文章归属，公开链路不含这些键。
func TestAdminCommentListMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCommentHandler(&fakeCommentFixtureService{
		adminResult: &service.AdminCommentListResponse{
			Comments: []*service.AdminCommentView{newFixtureAdminComment()},
			Total:    1,
			Page:     1,
			PageSize: 10,
		},
	})

	ctx, recorder := newCommentFixtureContext("/api/admin/comments/list", `{"page":1,"pageSize":10}`)
	handler.ListComments(ctx)

	assertResponseMatchesFixture(t, recorder, "comment.admin-list.json")
}
