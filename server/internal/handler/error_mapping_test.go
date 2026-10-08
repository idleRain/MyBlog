package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// fakeCommentErrorService 评论错误映射的测试替身，注入指定错误以验证业务码分档。
type fakeCommentErrorService struct {
	service.CommentServiceInterface
	err error
}

func (f *fakeCommentErrorService) GetCommentsByArticle(_ uint, _ *service.ListCommentsRequest) (*service.CommentListResponse, error) {
	return nil, f.err
}

func (f *fakeCommentErrorService) CreateComment(_ *service.CreateCommentRequest, _ *uint) (*model.Comment, error) {
	return nil, f.err
}

// dispatchCommentError 构造带 JSON 请求体的评论接口调用，返回响应记录。
func dispatchCommentError(t *testing.T, path string, body string, targetErr error) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	commentHandler := NewCommentHandler(&fakeCommentErrorService{err: targetErr})
	router := gin.New()
	switch {
	case strings.Contains(path, "/create"):
		router.POST(path, commentHandler.CreateComment)
	default:
		router.POST(path, commentHandler.GetCommentsByArticle)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

// parseCommentErrorCode 解析统一响应信封的业务码。
func parseCommentErrorCode(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return body.Code
}

// TestGetCommentsByArticleMapsArticleNotFoundToNotFound 验证评论列表的文章不存在分档为 404 而非 500。
func TestGetCommentsByArticleMapsArticleNotFoundToNotFound(t *testing.T) {
	recorder := dispatchCommentError(t, "/api/comments/list", `{"articleId":1}`,
		domain.ErrArticleNotFound)

	if code := parseCommentErrorCode(t, recorder); code != response.CodeNotFound {
		t.Errorf("业务码 = %d，期望 %d", code, response.CodeNotFound)
	}
}

// TestCreateCommentMapsGuestClosedToForbidden 验证游客评论关闭分档为 403 而非 500。
func TestCreateCommentMapsGuestClosedToForbidden(t *testing.T) {
	guestClosed := fmt.Errorf("%w：站点已关闭游客评论，请登录后发言", service.ErrPermissionDenied)
	recorder := dispatchCommentError(t, "/api/comments/create", `{"articleId":1,"content":"评论"}`, guestClosed)

	if code := parseCommentErrorCode(t, recorder); code != response.CodeForbid {
		t.Errorf("业务码 = %d，期望 %d", code, response.CodeForbid)
	}
}

// TestUnclassifiedErrorRedactsInternalDetails 验证未分类错误的原始细节不入响应体。
func TestUnclassifiedErrorRedactsInternalDetails(t *testing.T) {
	leakingErr := errors.New("Error 1054: Unknown column 'password_hint' in table `users`")
	recorder := dispatchCommentError(t, "/api/comments/list", `{"articleId":1}`, leakingErr)

	if code := parseCommentErrorCode(t, recorder); code != response.CodeError {
		t.Errorf("业务码 = %d，期望 %d", code, response.CodeError)
	}

	body := recorder.Body.String()
	if strings.Contains(body, "Unknown column") || strings.Contains(body, "`users`") {
		t.Errorf("未分类错误不得携带内部细节，实际响应 %q", body)
	}
	if !strings.Contains(body, unclassifiedErrorText) {
		t.Errorf("未分类错误应返回统一脱敏文案，实际响应 %q", body)
	}
}
