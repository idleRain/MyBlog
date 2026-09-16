package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newCORSTestRouter 构造挂载 CORS 中间件与探针处理器的测试路由。
func newCORSTestRouter(config CORSConfig) (*gin.Engine, *bool) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSWithConfig(config))

	nextReached := false
	router.POST("/probe", func(c *gin.Context) {
		nextReached = true
		c.Status(http.StatusOK)
	})
	return router, &nextReached
}

// corsTestConfig 白名单仅含本地前台开发源的测试配置。
func corsTestConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8899"},
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}
}

// TestCORSAllowedOriginReceivesHeaders 白名单内 Origin 应获得回显的 CORS 头。
func TestCORSAllowedOriginReceivesHeaders(t *testing.T) {
	router, nextReached := newCORSTestRouter(corsTestConfig())

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	request.Header.Set("Origin", "http://localhost:8899")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8899" {
		t.Errorf("Allow-Origin = %q, 期望回显白名单 Origin", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("允许凭证时应返回 Allow-Credentials: true")
	}
	methods := recorder.Header().Get("Access-Control-Allow-Methods")
	if methods != "POST, OPTIONS" {
		t.Errorf("Allow-Methods = %q, 期望收敛为 POST 与 OPTIONS", methods)
	}
	if !*nextReached {
		t.Error("白名单内请求应放行到后续处理器")
	}
}

// TestCORSDisallowedOriginGetsNoHeaders 白名单外 Origin 不返回任何 CORS 放行头。
func TestCORSDisallowedOriginGetsNoHeaders(t *testing.T) {
	router, _ := newCORSTestRouter(corsTestConfig())

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	request.Header.Set("Origin", "http://evil.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("白名单外 Origin 不应获得 Allow-Origin 头")
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Error("白名单外 Origin 不应获得 Allow-Credentials 头")
	}
	if recorder.Header().Get("Vary") != "Origin" {
		t.Error("响应应声明随 Origin 变化，防止缓存串用 CORS 头")
	}
}

// TestCORSPreflightReturns204 白名单内预检请求应直接返回 204 并中止后续链。
func TestCORSPreflightReturns204(t *testing.T) {
	router, nextReached := newCORSTestRouter(corsTestConfig())

	request := httptest.NewRequest(http.MethodOptions, "/probe", nil)
	request.Header.Set("Origin", "http://localhost:8899")
	request.Header.Set("Access-Control-Request-Method", "POST")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("预检响应状态码 = %d, 期望 204", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("白名单内预检应返回 Allow-Origin 头")
	}
	if *nextReached {
		t.Error("预检请求不应到达业务处理器")
	}
}

// TestCORSWithoutOriginSkipsCorsHeaders 同源请求不携带 Origin 头时不应返回 CORS 头。
func TestCORSWithoutOriginSkipsCorsHeaders(t *testing.T) {
	router, nextReached := newCORSTestRouter(corsTestConfig())

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("未携带 Origin 的请求不应获得 Allow-Origin 头")
	}
	if !*nextReached {
		t.Error("同源请求应正常放行到后续处理器")
	}
}

// TestCORSCredentialsWithWildcardRejected 通配符放行与凭证携带互斥，
// 白名单模式下不再提供全放行配置，防止出现规范禁止的组合。
func TestCORSCredentialsWithWildcardRejected(t *testing.T) {
	config := corsTestConfig()
	config.AllowedOrigins = []string{}

	router, _ := newCORSTestRouter(config)

	request := httptest.NewRequest(http.MethodPost, "/probe", nil)
	request.Header.Set("Origin", "http://any-origin.example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("空白名单不应向任何 Origin 放行")
	}
}
