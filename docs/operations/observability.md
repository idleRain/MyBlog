# 可观测性操作手册

> 本手册给出指标抓取、告警规则与结构化日志的落地方式与处置流程。
> 指标定义见 `server/internal/metrics/README.md`，日志实现见 `server/internal/logger/README.md`。
> 部署形态与可用性约束（单实例、无自动故障转移、故障即停站）以 `docs/architecture-rules.md` 第 7 节为唯一口径，本手册不重复。

## 1. 交付形态与边界

| 项 | 交付内容 | 不交付内容 |
|---|---|---|
| 指标 | 应用侧 `/metrics` 暴露与记录中间件 | 抓取端与展示端（Prometheus、Grafana） |
| 告警 | 抓取配置与告警规则文件（本目录） | 告警路由与通知渠道（Alertmanager 等） |
| 日志 | JSON 行输出、请求标识关联、级别与输出目标配置 | 日志采集、集中存储与检索（Loki、ELK 等） |

**有意不把抓取端与展示端写入编排**：它们是观测基础设施而非站点的一部分，加入编排会引入两个常驻服务，
与单实例部署的资源上限和形态声明冲突；且其运行验证依赖 Docker 环境，当前环境无法实机确认。
需要抓取时按下文方式在编排网络内单独运行，或接入既有的观测体系。

## 2. 指标暴露方式

指标监听**独立于业务端口**，不进入 `gin.Engine` 路由，也不经网关转发，因此公网不可达。

| 配置项 | 默认值 | 说明 |
|---|---|---|
| `metrics.enabled` | `false` | 是否启动指标监听 |
| `metrics.addr` | `127.0.0.1:9090` | 监听地址，启用时地址须含端口，否则启动期报错 |

容器内由 `docker-compose.yml` 注入 `MYBLOG_METRICS_ENABLED=true` 与 `MYBLOG_METRICS_ADDR=0.0.0.0:9090`，
抓取端必须加入同一 compose 网络（默认 `myblog_default`）才能访问；该端口未发布到宿主机。

`GET /metrics` 属基础设施端点，**不计入 `server/docs/api/README.md` 的 104 条接口口径**。

## 3. 抓取配置

抓取配置见 [`deploy/observability/prometheus.yml`](../../deploy/observability/prometheus.yml)，要点：

- 目标为 `server:9090`（compose 服务名 + 容器内指标端口）；
- `rule_files` 指向同目录的 `alerts.yml`；
- 抓取间隔 15s，与容器健康检查间隔一致。

在编排网络内单独运行抓取端的示例：

```bash
# 镜像版本需按部署环境固定：prom/prometheus:latest 未必是 3.x 线，历史上长期停留在 2.53.x，
# 固定前先确认实际标签。仓库不代管该镜像的版本。
docker run --rm -p 9090:9090 \
  --network myblog_default \
  -v "$PWD/deploy/observability/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  -v "$PWD/deploy/observability/alerts.yml:/etc/prometheus/alerts.yml:ro" \
  prom/prometheus:latest
```

## 4. 指标清单

| 指标 | 类型 | 标签 | 用途 |
|---|---|---|---|
| `myblog_http_requests_total` | Counter | `method`、`route`、`status` | 请求量、错误率、限流与拦截统计 |
| `myblog_http_request_duration_seconds` | Histogram | `method`、`route` | 延迟分位 |
| `myblog_http_requests_in_flight` | Gauge | `method` | 在途请求数 |
| `go_*` | — | — | 运行时指标（协程数、内存、GC） |
| `process_*` | — | — | 进程指标（常驻内存、启动时间、CPU） |

`route` 标签取**路由注册模式**而非原始路径，未匹配路由统一归并为 `unmatched`，标签基数只随路由数量增长。
新增接口无需改动指标代码。

## 5. 告警规则与处置

规则文件见 [`deploy/observability/alerts.yml`](../../deploy/observability/alerts.yml)，共 8 条。阈值为经验初值，
无压测数据支撑，需在真实负载下观测后调整。

| 规则 | 级别 | 触发条件 | 处置动作 |
|---|---|---|---|
| `MyBlogTargetDown` | critical | 抓取失败持续 2 分钟 | 检查 server 容器状态与 `MYBLOG_METRICS_ENABLED`；单实例下抓取失败通常即停站 |
| `MyBlogReadinessFailing` | critical | 就绪探针 503 持续 5 分钟 | 按原因码区分数据库不可连与上传目录不可写，细节在 `/api/health/ready` 的 ERROR 日志中 |
| `MyBlogProcessRestarted` | warning | 进程启动时间在 1 小时内被重置 | 确认是否为计划内发布；非计划重启会使全体用户登出 |
| `MyBlogHighServerErrorRate` | critical | 5xx 占比 > 5% 持续 10 分钟 | 按 `route` 标签定位集中出错的路由，再按访问日志的 `requestId` 检索未分类服务错误 |
| `MyBlogServerErrorBurst` | warning | 10 分钟内 5xx > 20 次 | 低流量下的绝对量补充，检查最近一次发布与数据库连接 |
| `MyBlogRateLimitTriggered` | warning | 429 速率持续 5 分钟 | 按 `clientIp` 与 `route` 判定来源；限流为进程内计数，扩容前必须外置 |
| `MyBlogSlowResponses` | warning | 某路由 P95 > 1s 持续 10 分钟 | 结合 `latencyMs` 与数据库慢查询；前台 SSR 路由优先查 markdown 渲染与文章查询 |
| `MyBlogMemoryNearContainerLimit` | warning | 常驻内存 > 512 MiB 的 90% 持续 10 分钟 | 触顶会被 OOM 终止；确认是否存在大请求体或媒体上传堆积，必要时调整内存上限 |

`for` 时长按告警级别区分：critical 取较短窗口以求快速发现，warning 取较长窗口以抑制抖动。
`MyBlogServerErrorBurst`、`MyBlogProcessRestarted` 与 `MyBlogMemoryNearContainerLimit` 不设 `for` 或设短窗口，
因为其触发条件本身已具备足够的信噪比。

## 6. 结构化日志

日志为 JSON 行，字段键取自 `server/internal/logger` 的 `Field*` 常量，是抓取端查询表达式的契约。

| 字段 | 出现位置 | 说明 |
|---|---|---|
| `time` | 全部记录 | 时间戳 |
| `level` | 全部记录 | `DEBUG` / `INFO` / `WARN` / `ERROR` |
| `msg` | 全部记录 | 事件名或原始文案 |
| `requestId` | 请求级记录 | 请求标识，串联同一次请求的全部日志 |
| `userId` | 访问日志与错误日志 | 操作者标识，未登录时缺省 |
| `clientIp`、`method`、`route`、`path`、`status`、`latencyMs`、`userAgent` | 访问日志 | 请求元信息；`route` 为路由模式，未匹配时为空串 |
| `error` | 错误与 panic 记录 | 原始错误，仅入日志不进响应体 |
| `stack` | panic 记录 | 调用栈 |

级别与输出目标取自 `logger` 配置节：`level` 取 `debug` / `info` / `warn` / `error`，`output` 取 `stdout` / `file`。
非法取值在启动期报错，不会静默回退。`output` 为 `file` 时以追加方式写入 `file_path`，
文件句柄在进程退出时释放；**未内置日志轮转**，需由外部（logrotate 或采集端）负责，否则文件会持续增长。

按请求标识串联一次请求的全部记录：

```bash
# 从访问日志中取出一条 5xx 的 requestId，再检索该请求的全部记录
grep '"status":500' /var/log/myblog/app.log | tail -1
grep '"requestId":"<上一步取到的标识>"' /var/log/myblog/app.log
```

访问日志跳过 `/api/health`（存活探针），`/api/health/ready` 仍会记录：就绪探针的失败原因正是需要被检索的内容。

## 7. 未验证事项（如实记录）

| 项 | 状态 | 原因 |
|---|---|---|
| 抓取端实机运行 | **未执行** | 本机未安装 Docker，无法启动抓取端与告警评估 |
| 告警真实触发 | **未执行** | 依赖抓取端运行；规则表达式仅做静态核对 |
| 规则表达式语法校验 | **未执行 `promtool check rules`** | 无该工具；已核对全部引用的指标名在真实暴露文本中存在 |
| 告警路由与通知渠道 | **未配置** | 属观测基础设施选型，需部署环境决定 |

已完成的核对：`prometheus.yml` 与 `alerts.yml` 均可被 YAML 解析器正确解析，8 条规则的分组、名称与级别符合预期；
规则引用的 `myblog_http_requests_total`、`myblog_http_request_duration_seconds_bucket`、`process_start_time_seconds`、
`process_resident_memory_bytes` 均在真实 `/metrics` 响应中存在。

## 8. 后续待办

1. 在具备 Docker 的环境执行第 3 节的抓取端启动命令，确认目标为 `UP` 且规则加载无报错（`promtool check rules` 一并执行）。
2. 观测一周真实负载后校准阈值：错误率 5%、P95 1 秒、内存 90% 均为经验初值。
3. 决定日志采集与集中检索方案，并在采集端配置轮转；应用侧不内置轮转。
4. 若启用审计需求，`operation_logs` 表已建但业务零写入，可作为关键写操作的审计落点另行立项。
