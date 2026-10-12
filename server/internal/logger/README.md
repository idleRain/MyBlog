# Logger 日志模块

结构化日志的初始化、请求标识传递与标准库日志桥接。日志统一输出为 JSON 行，字段键集中声明在本包。

## 输出与级别

| 配置项 | 取值 | 说明 |
|---|---|---|
| `logger.level` | `debug` / `info` / `warn` / `error` | 最低记录级别，非法取值在启动期报错 |
| `logger.output` | `stdout` / `file` | 输出目标，非法取值在启动期报错 |
| `logger.file_path` | 文件路径 | `output` 为 `file` 时的目标文件，以追加方式打开 |

## 请求标识

- 标识由 `middleware.RequestID` 经 `logger.AttachRequestID` 写入 gin 上下文，并回写响应头 `X-Request-ID`。
- 调用方传入的 `X-Request-ID` 仅在形如长度不超过 64 且只含 `[A-Za-z0-9._-]` 时复用，其余取值一律替换为服务端生成的 UUID v4，避免请求头注入污染日志与放大标签基数。
- 日志侧统一经 `logger.FromGin(c)` 取日志器，记录的 `requestId` 字段即可串联同一次请求的全部日志；上下文缺少标识时退化为默认日志器，不因标识缺失而丢日志。
- 标识经 gin 上下文而非 `context.Context` 传递，与 `domain.LanguageContextKey` 的既有通道一致；`gin.CreateTestContext` 构造的上下文不携带 `*http.Request`，该通道在测试中同样可用。

## 标准库与 gin 兼容通道

`Init` 会把标准库 `log` 的默认输出与 gin 的调试、错误输出一并重定向到同一处理器，使 `internal/database`、`internal/service` 等不经 slog 的既有调用点以及 gin 的启动横幅同样输出 JSON 行。桥接按消息前缀 `[ERROR]`、`[WARN]`、`[DEBUG]` 恢复级别，未带标记的记录按 `info` 处理；标准库自身的日期前缀已关闭，时间戳只由结构化记录提供。

## 约定

- 新增日志一律经 `slog` 写入，禁止新增 `log.Printf` 调用点。
- 日志字段键取自本包的 `Field*` 常量，禁止在调用处另写字面量。
