package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// fakeUserQueryService 用户查询测试替身，模拟 service 层返回的哨兵错误。
type fakeUserQueryService struct {
	service.UserService
	getByIDErr error
}

func (f *fakeUserQueryService) GetUserByID(id uint) (*domain.User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	return &domain.User{}, nil
}

// TestGetUserByIDErrorMapping 验证用户模块错误映射收敛到全局单一映射点：
// ErrUserNotFound 统一返回业务码 404，不再存在 user 模块映射 400 的第二套权威。
func TestGetUserByIDErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "用户不存在统一返回业务码404",
			err:      repository.ErrUserNotFound,
			wantCode: 404,
		},
		{
			name:     "包装后的用户不存在错误经errors.Is识别返回业务码404",
			err:      fmt.Errorf("查询用户失败: %w", repository.ErrUserNotFound),
			wantCode: 404,
		},
		{
			name:     "系统错误返回业务码500",
			err:      errors.New("数据库连接中断"),
			wantCode: 500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userHandler := NewUserHandler(&fakeUserQueryService{getByIDErr: tc.err})
			router := gin.New()
			router.POST("/users/get", userHandler.GetUserByID)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/users/get", strings.NewReader(`{"id":1}`))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			// 统一响应信封固定 HTTP 200，语义以响应体业务码表达。
			if recorder.Code != http.StatusOK {
				t.Fatalf("HTTP状态码 = %d, 期望 %d", recorder.Code, http.StatusOK)
			}
			assertResponseCode(t, recorder, tc.wantCode)
		})
	}
}
