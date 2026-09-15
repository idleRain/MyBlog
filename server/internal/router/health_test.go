package router

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// performHealthRequest 以指定探针函数构造请求并解析响应信封。
func performHealthRequest(t *testing.T, handlerFunc gin.HandlerFunc, dbCheck func() error) (int, gin.H) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	handlerFunc(ctx)

	var body gin.H
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return recorder.Code, body
}

// TestLivenessCheckReturnsOK 验证存活探针不探测数据库且恒返回 200。
func TestLivenessCheckReturnsOK(t *testing.T) {
	routes := NewHealthRoutes(func() error {
		t.Error("存活探针不应探测数据库")
		return nil
	})

	code, _ := performHealthRequest(t, routes.livenessCheck, nil)
	if code != 200 {
		t.Errorf("存活探针状态码 = %d, 期望 200", code)
	}
}

// TestReadinessCheckReflectsDatabase 验证就绪探针随数据库状态返回 200 或 503。
func TestReadinessCheckReflectsDatabase(t *testing.T) {
	routes := NewHealthRoutes(func() error {
		return nil
	})

	code, _ := performHealthRequest(t, routes.readinessCheck, routes.dbCheck)
	if code != 200 {
		t.Errorf("数据库正常时就绪探针状态码 = %d, 期望 200", code)
	}

	unready := NewHealthRoutes(func() error {
		return errors.New("数据库连接测试失败")
	})

	code, body := performHealthRequest(t, unready.readinessCheck, unready.dbCheck)
	if code != 503 {
		t.Errorf("数据库异常时就绪探针状态码 = %d, 期望 503", code)
	}
	if body["message"] != "服务未就绪" {
		t.Errorf("数据库异常时响应消息 = %v, 期望 服务未就绪", body["message"])
	}
}
