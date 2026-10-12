package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/metrics"
)

// TestStartMetricsServerDisabled 验证未启用指标时不创建监听句柄。
func TestStartMetricsServerDisabled(t *testing.T) {
	cfg := &config.Config{}

	server, err := startMetricsServer(cfg, metrics.New())
	if err != nil {
		t.Fatalf("未启用指标时不应报错: %v", err)
	}
	if server != nil {
		t.Fatal("未启用指标时不应启动监听")
	}
}

// TestStartMetricsServerServesMetrics 验证启用后暴露文本可经 HTTP 抓取。
// 监听端口取 0 由内核分配，避免测试占用固定端口。
func TestStartMetricsServerServesMetrics(t *testing.T) {
	cfg := &config.Config{}
	cfg.Metrics.Enabled = true
	cfg.Metrics.Addr = "127.0.0.1:0"

	server, err := startMetricsServer(cfg, metrics.New())
	if err != nil {
		t.Fatalf("启动指标监听失败: %v", err)
	}
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			t.Errorf("关停指标监听失败: %v", err)
		}
	}()

	response, err := http.Get("http://" + server.Addr() + "/metrics")
	if err != nil {
		t.Fatalf("抓取指标失败: %v", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("关闭响应体失败: %v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("抓取指标状态码 = %d, 期望 200", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("读取指标响应失败: %v", err)
	}
	if !strings.Contains(string(body), "go_goroutines") {
		t.Error("指标暴露文本缺少运行时采集器指标")
	}
}
