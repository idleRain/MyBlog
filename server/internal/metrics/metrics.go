// Package metrics 定义并暴露服务端运行指标。
// 指标端点由组合根以独立监听启动，不进入业务路由，也不经网关转发。
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// metricNamespace 自定义指标的统一前缀，避免与采集器指标重名。
	metricNamespace = "myblog"

	// unmatchedRouteLabel 未匹配到路由的请求在路由标签中的取值。
	// 直接取原始路径会让标签基数随扫描流量无界增长。
	unmatchedRouteLabel = "unmatched"

	// httpSubsystem 请求类指标的子系统名。
	httpSubsystem = "http"
)

// durationBuckets 请求耗时直方图的分桶边界，单位秒。
// 上界与 HTTP 服务器的写超时保持同一量级，覆盖前台 SSR 的慢响应区间。
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30}

// Registry 指标注册表，集中登记本站指标与运行时、进程采集器。
type Registry struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight *prometheus.GaugeVec
}

// New 创建指标注册表并注册全部采集器。
func New() *Registry {
	registry := prometheus.NewRegistry()

	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: httpSubsystem,
		Name:      "requests_total",
		Help:      "按方法、路由与状态码统计的 HTTP 请求总数。",
	}, []string{"method", "route", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Subsystem: httpSubsystem,
		Name:      "request_duration_seconds",
		Help:      "HTTP 请求耗时分布，单位秒。",
		Buckets:   durationBuckets,
	}, []string{"method", "route"})

	inFlight := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Subsystem: httpSubsystem,
		Name:      "requests_in_flight",
		Help:      "当前正在处理的 HTTP 请求数。",
	}, []string{"method"})

	registry.MustRegister(
		requests,
		duration,
		inFlight,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Registry{
		registry: registry,
		requests: requests,
		duration: duration,
		inFlight: inFlight,
	}
}

// Handler 返回指标抓取处理器，输出 Prometheus 文本暴露格式。
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// Middleware 记录请求计数、耗时分布与在途数量。
// 路由标签取路由注册模式而非原始路径，标签基数只随路由数量增长。
func (r *Registry) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		r.inFlight.WithLabelValues(method).Inc()
		start := time.Now()

		c.Next()

		r.inFlight.WithLabelValues(method).Dec()
		route := routeLabel(c)
		r.duration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
		r.requests.WithLabelValues(method, route, strconv.Itoa(c.Writer.Status())).Inc()
	}
}

// routeLabel 返回用于指标标签的路由模式，未匹配到路由时返回固定占位。
func routeLabel(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return unmatchedRouteLabel
}
