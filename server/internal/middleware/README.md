# Middleware 中间件模块

HTTP 横切层。认证与权限中间件只依赖 `IdentityProvider` 抽象，不直接持有存储实现。

## 中间件清单

| 文件 | 导出内容 | 职责 |
|---|---|---|
| `logger.go` | `Logger`、`Recovery`、`RequestID` | 结构化访问日志（JSON 行，含 `requestId` 与 `userId`，跳过 `/api/health`）；panic 恢复输出结构化记录并返回 500；请求标识优先复用请求头 `X-Request-ID`（限 64 字符内安全字符集），不合法时生成 UUID v4，写入响应头与 gin 上下文 |
| `cors.go` | `CORSWithConfig` | Origin 白名单精确匹配，白名单外与未携带 Origin 的请求不回显任何 CORS 头，`OPTIONS` 统一 204，所有响应带 `Vary: Origin`；规则来自 `cors` 节 |
| `security.go` | `SecurityMiddleware`、`DefaultSecurityConfig`、`SecurityMiddlewareFromConfig`、`AdminSecurityMiddlewareFromConfig`、`IPWhitelistMiddleware` | IP 与用户两级限流、安全响应头、请求体上限（普通接口 10MB、管理员接口 5MB）、User-Agent 子串黑名单、恶意模式检测（`getDefaultBlockedPatterns`，模式经词首边界锚定；JSON 体按字段路径扫描，正文 `content` 与 `i18n.*.content` 豁免，见 `richContentJSONPaths`） |
| `auth.go` | `Auth`、`OptionalAuth`、`AdminAuth`、`RequireRole` | 访问令牌校验；令牌经 `Authorization: Bearer` 与 `mb_access_token` Cookie 双轨解析，仅向上下文注入 `userID` |
| `identity.go` | `IdentityProvider`、`NewIdentityProvider` | 身份解析抽象与唯一实现，依次校验令牌、用户状态与角色有效性，失败时已写入响应并中止请求链；实现所消费的用户读取能力由本包声明窄接口 `userLookup`，仓储实现经组合根注入，故本包不依赖 `internal/repository` |
| `rbac.go` | `RequirePermission`、`RequireAllPermissions`、`RequireRoleLevel`、`RequireSuperAdmin`、`RequireAdminOrAbove`、`RequireEditorOrAbove`、`RequireOwnershipOrAdmin`、`CanManageUserRole`、`GetCurrentUser`、`GetCurrentUserID`、`GetCurrentUserRole` | RBAC 鉴权与上下文读取，权限判定全部委托注入的 `RBACService` |
| `ratelimit.go` | `RateLimit`、`RateLimitPerUser`、`RateLimiter`、`NewRateLimiter` | 进程内计数限流器，安全中间件复用同一实现；过期记录随新客户端加入惰性清扫 |
| `language.go` | `LanguageMiddlewareFromConfig` | 解析 `Accept-Language` 写入请求上下文，白名单外回退缺省语言；只解析不中断请求 |

## 装配与约定

- 全局中间件顺序见 `internal/router/router.go`：RequestID、指标记录、Logger、Recovery、CORS、Security、Language；管理员分组额外挂 `AdminSecurityMiddlewareFromConfig`。
- 日志输出为 JSON 行，初始化与请求标识读取见 `internal/logger`；请求级记录统一经 `logger.FromGin(c)` 取日志器，字段键取自 `logger.Field*` 常量。
- 权限清单唯一权威为 `configs/config.yaml` 的 `rbac.role_permissions`，本文件不重复罗列。
- 限流计数与令牌表均为进程内存结构，属单实例终态下的已接受取舍，变更前提见 `docs/architecture-rules.md` §7.2。
- `ClientIP()` 的取值受 `security.trusted_proxies` 约束，仅受信来源携带的 `X-Forwarded-For` 参与解析，配置与默认值见 `internal/config/README.md`；漏配容器网段会让 IP 限流、审计落库与访客去重全部归并到网关一个键上。
- 实例化纪律（依赖由组合根注入）见 `docs/architecture-rules.md` §5.2。
