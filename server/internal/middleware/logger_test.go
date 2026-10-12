package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// newLoggedEngine 构造带请求标识与结构化访问日志的最小路由，返回引擎与日志缓冲区。
func newLoggedEngine() (*gin.Engine, *bytes.Buffer) {
	gin.SetMode(gin.TestMode)

	buffer := &bytes.Buffer{}
	if err := logger.Init(logger.Options{Level: logger.LevelDebug, Writer: buffer}); err != nil {
		panic(err)
	}

	engine := gin.New()
	engine.Use(RequestID())
	engine.Use(Logger())
	return engine, buffer
}

// lastRecord 解析缓冲区最后一行为日志记录。
func lastRecord(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	var record map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &record); err != nil {
		t.Fatalf("访问日志不是合法 JSON: %v，原文 %q", err, buffer.String())
	}
	return record
}

// TestLoggerWritesStructuredAccessLog 验证访问日志含请求标识、用户标识与响应元信息。
func TestLoggerWritesStructuredAccessLog(t *testing.T) {
	engine, buffer := newLoggedEngine()
	engine.POST("/api/articles/create", func(c *gin.Context) {
		c.Set("userID", uint(7))
		c.Status(http.StatusCreated)
	})

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/articles/create", nil))

	record := lastRecord(t, buffer)
	wantFields := map[string]any{
		logger.FieldUserID:    float64(7),
		logger.FieldMethod:    "POST",
		logger.FieldRoute:     "/api/articles/create",
		logger.FieldPath:      "/api/articles/create",
		logger.FieldStatus:    float64(http.StatusCreated),
		logger.FieldUserAgent: "",
	}
	for key, want := range wantFields {
		if record[key] != want {
			t.Errorf("%s = %v, 期望 %v", key, record[key], want)
		}
	}
	if requestID, ok := record[logger.FieldRequestID].(string); !ok || requestID == "" {
		t.Errorf("%s 应为非空字符串: %v", logger.FieldRequestID, record[logger.FieldRequestID])
	}
	if _, ok := record[logger.FieldLatencyMs]; !ok {
		t.Errorf("访问日志缺少 %s 字段", logger.FieldLatencyMs)
	}
}

// TestLoggerLevelFollowsStatus 验证访问日志级别随响应状态码变化。
func TestLoggerLevelFollowsStatus(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		wantLevel string
	}{
		{name: "成功", status: http.StatusOK, wantLevel: "INFO"},
		{name: "客户端错误", status: http.StatusBadRequest, wantLevel: "WARN"},
		{name: "服务端错误", status: http.StatusInternalServerError, wantLevel: "ERROR"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			engine, buffer := newLoggedEngine()
			engine.POST("/api/articles/create", func(c *gin.Context) {
				c.Status(testCase.status)
			})

			engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/articles/create", nil))

			if got := lastRecord(t, buffer)["level"]; got != testCase.wantLevel {
				t.Errorf("level = %v, 期望 %s", got, testCase.wantLevel)
			}
		})
	}
}

// TestLoggerSkipsHealthProbe 验证健康探针路径不写入访问日志。
func TestLoggerSkipsHealthProbe(t *testing.T) {
	engine, buffer := newLoggedEngine()
	engine.GET("/api/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if buffer.Len() != 0 {
		t.Errorf("健康探针不应写入访问日志: %q", buffer.String())
	}
}

// TestRequestIDGeneratesUUID 验证缺失调用方标识时生成 UUID。
func TestRequestIDGeneratesUUID(t *testing.T) {
	engine, _ := newLoggedEngine()
	engine.POST("/api/articles/create", func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/articles/create", nil))

	requestID := recorder.Header().Get(requestIDHeader)
	if _, err := uuid.Parse(requestID); err != nil {
		t.Errorf("响应头 %s 应为 UUID，实际 %q", requestIDHeader, requestID)
	}
}

// TestRequestIDReusesValidInboundValue 验证形态合法的调用方标识被原样复用。
func TestRequestIDReusesValidInboundValue(t *testing.T) {
	engine, buffer := newLoggedEngine()
	engine.POST("/api/articles/create", func(c *gin.Context) { c.Status(http.StatusOK) })

	const inboundRequestID = "trace-42.example_1"
	request := httptest.NewRequest(http.MethodPost, "/api/articles/create", nil)
	request.Header.Set(requestIDHeader, inboundRequestID)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(requestIDHeader); got != inboundRequestID {
		t.Errorf("响应头 %s = %q, 期望复用 %q", requestIDHeader, got, inboundRequestID)
	}
	if got := lastRecord(t, buffer)[logger.FieldRequestID]; got != inboundRequestID {
		t.Errorf("访问日志 %s = %v, 期望 %q", logger.FieldRequestID, got, inboundRequestID)
	}
}

// TestRequestIDRejectsUnsafeInboundValue 验证超长或含异常字符的调用方标识被替换为服务端生成的 UUID，
// 避免请求头注入污染日志并放大下游聚合的标签基数。
func TestRequestIDRejectsUnsafeInboundValue(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "超出长度上限", value: strings.Repeat("a", 65)},
		{name: "含换行注入", value: "trace\n{\"level\":\"INFO\"}"},
		{name: "含空格", value: "trace 42"},
		{name: "含非 ASCII 字符", value: "追踪标识"},
		{name: "含双引号", value: `trace"42`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			engine, _ := newLoggedEngine()
			engine.POST("/api/articles/create", func(c *gin.Context) { c.Status(http.StatusOK) })

			request := httptest.NewRequest(http.MethodPost, "/api/articles/create", nil)
			request.Header.Set(requestIDHeader, testCase.value)

			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)

			requestID := recorder.Header().Get(requestIDHeader)
			if requestID == testCase.value {
				t.Fatalf("不合形态的请求标识不应被复用: %q", requestID)
			}
			if _, err := uuid.Parse(requestID); err != nil {
				t.Errorf("替换后的请求标识应为 UUID，实际 %q", requestID)
			}
		})
	}
}

// TestRecoveryLogsStructuredPanic 验证 panic 被恢复为 500 且记录为结构化日志。
func TestRecoveryLogsStructuredPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buffer := &bytes.Buffer{}
	if err := logger.Init(logger.Options{Level: logger.LevelDebug, Writer: buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	engine := gin.New()
	engine.Use(RequestID())
	engine.Use(Logger())
	engine.Use(Recovery())
	engine.POST("/api/articles/create", func(c *gin.Context) {
		panic("示例 panic")
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/articles/create", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("panic 应返回 500，实际 %d", recorder.Code)
	}

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("期望 panic 与访问日志各一行，实际 %d 行: %q", len(lines), buffer.String())
	}

	var panicRecord, accessRecord map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &panicRecord); err != nil {
		t.Fatalf("panic 记录不是合法 JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &accessRecord); err != nil {
		t.Fatalf("访问日志不是合法 JSON: %v", err)
	}

	if panicRecord["level"] != "ERROR" {
		t.Errorf("panic 记录 level = %v, 期望 ERROR", panicRecord["level"])
	}
	if panicRequestID, _ := panicRecord[logger.FieldRequestID].(string); panicRequestID == "" {
		t.Errorf("panic 记录缺少 %s", logger.FieldRequestID)
	} else if panicRequestID != accessRecord[logger.FieldRequestID] {
		t.Errorf("panic 与访问日志的请求标识不一致: %v vs %v",
			panicRequestID, accessRecord[logger.FieldRequestID])
	}
}
