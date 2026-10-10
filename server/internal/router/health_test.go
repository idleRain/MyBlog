package router

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 探测失败时使用的错误文案，用于断言响应体不回传原始错误。
const (
	databaseFailureText = "dial tcp 127.0.0.1:3306: connect: connection refused"
	uploadsFailureText  = "mkdir /app/uploads: read-only file system"
)

func healthyProbe() func() error {
	return func() error { return nil }
}

func failingProbe(message string) func() error {
	return func() error { return errors.New(message) }
}

// performHealthRequest 构造健康检查请求并解析响应信封。
func performHealthRequest(t *testing.T, handlerFunc gin.HandlerFunc) (int, gin.H, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	handlerFunc(ctx)

	rawBody := recorder.Body.String()
	var body gin.H
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return recorder.Code, body, rawBody
}

// TestLivenessCheckReturnsOK 验证存活探针不探测下游依赖且恒返回 200。
func TestLivenessCheckReturnsOK(t *testing.T) {
	routes := NewHealthRoutes(
		func() error {
			t.Error("存活探针不应探测数据库")
			return nil
		},
		func() error {
			t.Error("存活探针不应探测上传目录")
			return nil
		},
	)

	code, _, _ := performHealthRequest(t, routes.livenessCheck)
	if code != 200 {
		t.Errorf("存活探针状态码 = %d, 期望 200", code)
	}
}

// TestReadinessCheckReflectsDatabase 验证就绪探针随数据库状态返回 200 或 503。
func TestReadinessCheckReflectsDatabase(t *testing.T) {
	routes := NewHealthRoutes(healthyProbe(), healthyProbe())

	code, _, _ := performHealthRequest(t, routes.readinessCheck)
	if code != 200 {
		t.Errorf("数据库正常时就绪探针状态码 = %d, 期望 200", code)
	}

	unready := NewHealthRoutes(failingProbe(databaseFailureText), healthyProbe())

	_, body, rawBody := performHealthRequest(t, unready.readinessCheck)
	if body["code"] != float64(503) {
		t.Errorf("数据库异常时响应码 = %v, 期望 503", body["code"])
	}
	if body["message"] != "服务未就绪" {
		t.Errorf("数据库异常时响应消息 = %v, 期望 服务未就绪", body["message"])
	}
	assertReasonCode(t, body, reasonDatabaseUnavailable)
	assertNoRawError(t, rawBody, databaseFailureText)
}

// TestReadinessCheckReflectsUploads 验证上传目录不可写时就绪探针返回 503 且原因码稳定。
func TestReadinessCheckReflectsUploads(t *testing.T) {
	routes := NewHealthRoutes(healthyProbe(), failingProbe(uploadsFailureText))

	code, body, rawBody := performHealthRequest(t, routes.readinessCheck)
	if code != 503 {
		t.Errorf("上传目录不可写时就绪探针状态码 = %d, 期望 503", code)
	}
	assertReasonCode(t, body, reasonUploadsUnwritable)
	assertNoRawError(t, rawBody, uploadsFailureText)
}

// TestReadinessCheckSkipsUploadsProbeWhenDatabaseFails 验证数据库失败时短路，不再触发磁盘探测。
func TestReadinessCheckSkipsUploadsProbeWhenDatabaseFails(t *testing.T) {
	routes := NewHealthRoutes(failingProbe(databaseFailureText), func() error {
		t.Error("数据库已判定未就绪时不应继续探测上传目录")
		return nil
	})

	code, _, _ := performHealthRequest(t, routes.readinessCheck)
	if code != 503 {
		t.Errorf("数据库异常时状态码 = %d, 期望 503", code)
	}
}

// assertReasonCode 断言未就绪原因码。
func assertReasonCode(t *testing.T, body gin.H, expected string) {
	t.Helper()

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("响应 data 字段类型异常: %v", body["data"])
	}
	if data["status"] != "unready" {
		t.Errorf("data.status = %v, 期望 unready", data["status"])
	}
	if data["reason"] != expected {
		t.Errorf("data.reason = %v, 期望 %s", data["reason"], expected)
	}
}

// assertNoRawError 断言响应体不含原始错误文案，防止故障细节经探针外泄。
func assertNoRawError(t *testing.T, rawBody string, rawErrorText string) {
	t.Helper()

	if strings.Contains(rawBody, rawErrorText) {
		t.Errorf("响应体不应包含原始错误文案，实际响应为: %s", rawBody)
	}
}
