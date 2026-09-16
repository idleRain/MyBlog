package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// stubCommentService 评论服务的测试替身，绑定失败路径不应触达服务层，
// 替身不覆写任何方法，一旦被调用将以空指针暴露越界访问。
type stubCommentService struct {
	service.CommentServiceInterface
}

// newCommentBindingRouter 构造挂载评论创建路由的最小引擎。
func newCommentBindingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	commentHandler := NewCommentHandler(&stubCommentService{})
	router.POST("/api/comments/create", commentHandler.CreateComment)
	return router
}

// TestCommentHandlerBindingFailureMapsTo400 参数绑定失败应以业务码 400 承载语义，
// 统一响应信封固定 HTTP 200，断言遵循响应体业务码口径（与 QA-01/02 矩阵一致）。
func TestCommentHandlerBindingFailureMapsTo400(t *testing.T) {
	router := newCommentBindingRouter()

	testCases := []struct {
		name string
		body string
	}{
		{name: "缺失必填的文章 ID", body: `{"content":"正常评论内容"}`},
		{name: "评论内容超过 2000 字符上限", body: `{"articleId":1,"content":"` + strings.Repeat("字", 2001) + `"}`},
		{name: "评论内容为空", body: `{"articleId":1,"content":""}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/comments/create", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("响应体解析失败: %v", err)
			}
			if body.Code != http.StatusBadRequest {
				t.Errorf("绑定失败应以业务码 400 承载, 实际为 %d, 响应 %s", body.Code, recorder.Body.String())
			}
		})
	}
}
