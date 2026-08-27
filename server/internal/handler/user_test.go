package handler

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/repository"

	"github.com/gin-gonic/gin"
)

// TestHandleUserQueryError 验证用户查询错误的响应码语义。
func TestHandleUserQueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		err        error
		wantCode   int
		wantStatus int
	}{
		{
			name:       "用户不存在时返回业务码400",
			err:        repository.ErrUserNotFound,
			wantCode:   400,
			wantStatus: 200,
		},
		{
			name:       "系统错误时返回业务码500",
			err:        errors.New("查询用户失败: 数据库连接中断"),
			wantCode:   500,
			wantStatus: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			handleUserQueryError(ctx, tc.err)

			if recorder.Code != tc.wantStatus {
				t.Errorf("HTTP状态码 = %d, 期望 %d", recorder.Code, tc.wantStatus)
			}

			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("解析响应体失败: %v", err)
			}
			if body.Code != tc.wantCode {
				t.Errorf("业务码 = %d, 期望 %d", body.Code, tc.wantCode)
			}
			if body.Message == "" {
				t.Error("响应消息不能为空")
			}
		})
	}
}
