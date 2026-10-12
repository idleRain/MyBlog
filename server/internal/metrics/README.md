# Metrics 指标模块

Prometheus 指标的定义、采集与暴露。指标监听由组合根以独立端口启动，不进入业务路由，也不经网关转发。

## 指标清单

| 指标 | 类型 | 标签 | 含义 |
|---|---|---|---|
| `myblog_http_requests_total` | Counter | `method`、`route`、`status` | 按方法、路由与状态码统计的请求总数 |
| `myblog_http_request_duration_seconds` | Histogram | `method`、`route` | 请求耗时分布（秒），分桶见 `durationBuckets` |
| `myblog_http_requests_in_flight` | Gauge | `method` | 当前正在处理的请求数 |
| `go_*` | — | — | 由 `collectors.NewGoCollector` 提供的运行时指标 |
| `process_*` | — | — | 由 `collectors.NewProcessCollector` 提供的进程指标 |

## 约定

- **路由标签取路由注册模式**（`c.FullPath()`）而非原始路径，标签基数只随路由数量增长；未匹配到路由的请求统一归并为 `unmatched`，避免扫描流量把标签基数推高。
- 指标记录中间件在 `internal/router/router.go` 中置于中间件链第二位，度量范围覆盖恢复、安全与业务处理的完整耗时。
- `metrics.enabled` 只控制是否启动指标监听，记录本身始终进行；该开关关闭时指标无抓取端，记录开销为常数级。
- 指标监听默认绑定回环地址，容器内需经 `MYBLOG_METRICS_ADDR` 覆盖为 `0.0.0.0:9090` 才能被同网络的抓取端访问；抓取配置与告警规则见 `deploy/observability/`，处置流程见 `docs/operations/observability.md`。
