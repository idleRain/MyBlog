package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newMetricsTestRouter 构造带指标中间件的最小路由，返回注册表与路由引擎。
func newMetricsTestRouter() (*Registry, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	registry := New()
	engine := gin.New()
	engine.Use(registry.Middleware())
	return registry, engine
}

// scrape 抓取一次指标暴露文本，抓取失败或采集报错时判定为失败。
func scrape(t *testing.T, registry *Registry) string {
	t.Helper()

	recorder := httptest.NewRecorder()
	registry.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	// promhttp 默认在采集出错时返回 500，此处同时充当采集器平台兼容性断言。
	if recorder.Code != http.StatusOK {
		t.Fatalf("抓取指标失败，状态码 %d，响应 %s", recorder.Code, recorder.Body.String())
	}
	return recorder.Body.String()
}

// serve 以指定方法与路径发起一次请求。
func serve(engine *gin.Engine, method, target string) {
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, target, nil))
}

// TestMiddlewareRecordsRequestMetrics 验证请求计数、耗时分布与在途数量被记录。
func TestMiddlewareRecordsRequestMetrics(t *testing.T) {
	registry, engine := newMetricsTestRouter()
	engine.POST("/api/articles/create", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	serve(engine, http.MethodPost, "/api/articles/create")

	body := scrape(t, registry)
	wantLines := []string{
		`myblog_http_requests_total{method="POST",route="/api/articles/create",status="201"} 1`,
		`myblog_http_request_duration_seconds_count{method="POST",route="/api/articles/create"} 1`,
		`myblog_http_requests_in_flight{method="POST"} 0`,
	}
	for _, line := range wantLines {
		if !strings.Contains(body, line) {
			t.Errorf("指标暴露文本缺少 %q\n%s", line, body)
		}
	}
}

// TestMiddlewareSeparatesRoutes 验证不同路由各自独立计数。
func TestMiddlewareSeparatesRoutes(t *testing.T) {
	registry, engine := newMetricsTestRouter()
	engine.POST("/api/articles/list", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.POST("/api/users/list", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve(engine, http.MethodPost, "/api/articles/list")
	serve(engine, http.MethodPost, "/api/articles/list")
	serve(engine, http.MethodPost, "/api/users/list")

	body := scrape(t, registry)
	wantLines := []string{
		`myblog_http_requests_total{method="POST",route="/api/articles/list",status="200"} 2`,
		`myblog_http_requests_total{method="POST",route="/api/users/list",status="200"} 1`,
	}
	for _, line := range wantLines {
		if !strings.Contains(body, line) {
			t.Errorf("指标暴露文本缺少 %q\n%s", line, body)
		}
	}
}

// TestMiddlewareKeepsRouteCardinalityBounded 验证未匹配路由归并到固定标签，
// 原始路径不进入指标标签，避免扫描流量把标签基数推高。
func TestMiddlewareKeepsRouteCardinalityBounded(t *testing.T) {
	registry, engine := newMetricsTestRouter()
	engine.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })

	rawPaths := []string{"/api/admin/secret-probe-1", "/api/admin/secret-probe-2"}
	for _, path := range rawPaths {
		serve(engine, http.MethodPost, path)
	}

	body := scrape(t, registry)
	wantLine := `myblog_http_requests_total{method="POST",route="unmatched",status="404"} 2`
	if !strings.Contains(body, wantLine) {
		t.Errorf("未匹配路由应归并到固定标签，期望 %q\n%s", wantLine, body)
	}
	for _, rawPath := range rawPaths {
		if strings.Contains(body, rawPath) {
			t.Errorf("原始路径不应进入指标标签: %q", rawPath)
		}
	}
}

// TestHandlerExposesRuntimeCollectors 验证暴露文本包含 Go 运行时采集器指标。
func TestHandlerExposesRuntimeCollectors(t *testing.T) {
	registry := New()

	body := scrape(t, registry)
	for _, metricName := range []string{"go_goroutines", "go_memstats_alloc_bytes"} {
		if !strings.Contains(body, metricName) {
			t.Errorf("指标暴露文本缺少采集器指标 %q", metricName)
		}
	}
}
