package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MyBlog/internal/config"

	"github.com/gin-gonic/gin"
)

// clientIPProbe 以指标中间件位置作为探针捕获引擎解析出的客户端 IP。
// 引擎已挂载静态服务的通配路由，追加探针路由会与之冲突，故复用中间件注入点。
func clientIPProbe(t *testing.T, trustedProxies []string) (*gin.Engine, *string) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	captured := new(string)
	cfg := &config.Config{
		Security: config.SecurityConfig{TrustedProxies: trustedProxies},
		Media:    config.MediaConfig{BaseURL: "/uploads", UploadDir: t.TempDir()},
	}

	manager, err := NewRouter(cfg, func(c *gin.Context) {
		*captured = c.ClientIP()
		c.Next()
	})
	if err != nil {
		t.Fatalf("构建路由失败: %v", err)
	}

	return manager.GetEngine(), captured
}

// performProbeRequest 发送一次指定来源地址与转发头的探针请求。
func performProbeRequest(t *testing.T, engine *gin.Engine, remoteAddr, forwardedFor string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/probe-client-ip", nil)
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}

	engine.ServeHTTP(httptest.NewRecorder(), req)
}

// TestClientIPIgnoresForgedHeaderFromUntrustedPeer 验证非受信来源携带的转发头不被采信，
// 客户端 IP 取连接来源地址。
func TestClientIPIgnoresForgedHeaderFromUntrustedPeer(t *testing.T) {
	engine, captured := clientIPProbe(t, []string{"127.0.0.1/32"})

	performProbeRequest(t, engine, "203.0.113.9:4321", "1.2.3.4")

	if *captured != "203.0.113.9" {
		t.Errorf("ClientIP() = %q, 期望取连接来源地址 203.0.113.9", *captured)
	}
}

// TestClientIPUsesRightmostUntrustedFromTrustedProxy 验证经受信代理转发时取最右侧非受信地址，
// 客户端自带的伪造值位于最左，不参与结果。
func TestClientIPUsesRightmostUntrustedFromTrustedProxy(t *testing.T) {
	engine, captured := clientIPProbe(t, []string{"127.0.0.1/32"})

	performProbeRequest(t, engine, "127.0.0.1:5555", "1.2.3.4, 203.0.113.9")

	if *captured != "203.0.113.9" {
		t.Errorf("ClientIP() = %q, 期望取最右侧非受信地址 203.0.113.9", *captured)
	}
}

// TestClientIPFollowsTrustedProxySingleValue 验证受信代理转发的单值转发头被采信。
func TestClientIPFollowsTrustedProxySingleValue(t *testing.T) {
	engine, captured := clientIPProbe(t, []string{"127.0.0.1/32"})

	performProbeRequest(t, engine, "127.0.0.1:5555", "203.0.113.9")

	if *captured != "203.0.113.9" {
		t.Errorf("ClientIP() = %q, 期望采信受信代理转发的 203.0.113.9", *captured)
	}
}

// TestClientIPIgnoresForwardedHeaderWhenNoProxyTrusted 验证信任列表为空时一律取连接来源地址，
// 即使转发头伪造成受信网段内的地址也不生效。
func TestClientIPIgnoresForwardedHeaderWhenNoProxyTrusted(t *testing.T) {
	engine, captured := clientIPProbe(t, nil)

	performProbeRequest(t, engine, "203.0.113.9:4321", "127.0.0.1, 10.0.0.1")

	if *captured != "203.0.113.9" {
		t.Errorf("ClientIP() = %q, 期望取连接来源地址 203.0.113.9", *captured)
	}
}

// TestNewRouterRejectsInvalidTrustedProxy 验证非法信任项在引擎构造阶段即报错，
// 该分支是配置校验之外的第二道拦截。
func TestNewRouterRejectsInvalidTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Security: config.SecurityConfig{TrustedProxies: []string{"not-an-address"}},
		Media:    config.MediaConfig{BaseURL: "/uploads", UploadDir: t.TempDir()},
	}

	if _, err := NewRouter(cfg, func(c *gin.Context) { c.Next() }); err == nil {
		t.Error("非法代理信任项应返回错误")
	}
}
