// MyBlog 指标监听的启动与生命周期管理
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"MyBlog/internal/config"
	"MyBlog/internal/logger"
	"MyBlog/internal/metrics"
)

// metricsServer 指标监听的运行句柄，持有监听器以便暴露实际绑定地址。
type metricsServer struct {
	server   *http.Server
	listener net.Listener
}

// startMetricsServer 在独立端口启动指标暴露服务。
// 该监听不进入业务路由也不经网关转发，公网不可达，仅同网络内的抓取端可访问。
// 未启用指标时返回 nil 句柄与空错误，调用方据此跳过关停。
func startMetricsServer(cfg *config.Config, registry *metrics.Registry) (*metricsServer, error) {
	if !cfg.Metrics.Enabled {
		return nil, nil
	}

	// 先完成端口绑定，把监听失败暴露在启动阶段而不是后台 goroutine 内。
	listener, err := net.Listen("tcp", cfg.Metrics.Addr)
	if err != nil {
		return nil, fmt.Errorf("指标监听绑定失败: %w", err)
	}

	// 指标端点只读且响应体小，沿用业务服务的四项超时以防连接资源被长期占用。
	server := &http.Server{
		Handler:           registry.Handler(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		logger.L().Info("指标监听启动", slog.String("addr", listener.Addr().String()))
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.L().Error("指标监听异常退出", slog.Any(logger.FieldError, err))
		}
	}()

	return &metricsServer{server: server, listener: listener}, nil
}

// Addr 返回实际绑定地址，供启动日志与抓取配置核对。
func (m *metricsServer) Addr() string {
	return m.listener.Addr().String()
}

// Shutdown 停止接收新的抓取请求并排空在途请求。
func (m *metricsServer) Shutdown(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}
