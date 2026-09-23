package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// TestHandleServiceError 验证不同哨兵错误与未知错误的响应码语义。
func TestHandleServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	notFoundErrors := map[string]error{
		"文章不存在":   domain.ErrArticleNotFound,
		"分类不存在":   domain.ErrCategoryNotFound,
		"标签不存在":   domain.ErrTagNotFound,
		"评论不存在":   domain.ErrCommentNotFound,
		"媒体文件不存在": domain.ErrMediaNotFound,
		"通知不存在":   domain.ErrNotificationNotFound,
		"友情链接不存在": domain.ErrFriendlyLinkNotFound,
		"设置项不存在":  domain.ErrSettingNotFound,
		"用户不存在":   domain.ErrUserNotFound,
		"关注关系不存在": domain.ErrFollowNotFound,
	}

	for name, sentinel := range notFoundErrors {
		t.Run("不存在_"+name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			HandleServiceError(ctx, sentinel)

			assertResponseCode(t, recorder, 404)
		})
	}

	t.Run("权限不足返回403", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		HandleServiceError(ctx, service.ErrPermissionDenied)

		assertResponseCode(t, recorder, 403)
	})

	t.Run("包装后的权限错误经errors.Is识别返回403", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		HandleServiceError(ctx, fmt.Errorf("%w：编辑此文章", service.ErrPermissionDenied))

		assertResponseCode(t, recorder, 403)
	})

	t.Run("未知错误返回500", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		HandleServiceError(ctx, errors.New("数据库连接中断"))

		assertResponseCode(t, recorder, 500)
	})
}

// assertResponseCode 断言响应体中的业务码符合预期。
func assertResponseCode(t *testing.T, recorder *httptest.ResponseRecorder, wantCode int) {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Code != wantCode {
		t.Errorf("业务码 = %d, 期望 %d", body.Code, wantCode)
	}
}
