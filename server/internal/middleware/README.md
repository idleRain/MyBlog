# Middleware 中间件模块

HTTP 横切层。认证与权限中间件只依赖 `IdentityProvider` 抽象，不直接持有存储实现。

## 中间件清单

| 文件 | 导出内容 | 职责 |
|---|---|---|
| `logger.go` | `Logger`、`CustomLogger`、`RequestID` | 请求日志（gin 文本格式，跳过 `/api/health`）；请求 ID 优先复用请求头 `X-Request-ID`，缺失时以 UnixNano 生成，写入响应头 `X-Request-ID` 与上下文键 `RequestID` |
| `cors.go` | `CORSWithConfig` | Origin 白名单精确匹配，白名单外与未携带 Origin 的请求不回显任何 CORS 头，`OPTIONS` 统一 204，所有响应带 `Vary: Origin`；规则来自 `cors` 节 |
| `security.go` | `SecurityMiddleware`、`DefaultSecurityConfig`、`SecurityMiddlewareFromConfig`、`AdminSecurityMiddlewareFromConfig`、`IPWhitelistMiddleware` | IP 与用户两级限流、安全响应头、请求体上限（普通接口 10MB、管理员接口 5MB）、User-Agent 子串黑名单、恶意模式检测（`getDefaultBlockedPatterns`，模式经词首边界锚定） |
| `auth.go` | `Auth`、`OptionalAuth`、`AdminAuth`、`RequireRole` | 访问令牌校验；令牌经 `Authorization: Bearer` 与 `mb_access_token` Cookie 双轨解析，仅向上下文注入 `userID` |
| `identity.go` | `IdentityProvider`、`NewIdentityProvider` | 身份解析抽象与唯一实现，依次校验令牌、用户状态与角色有效性，失败时已写入响应并中止请求链 |
| `rbac.go` | `RequirePermission`、`RequireAllPermissions`、`RequireRoleLevel`、`RequireSuperAdmin`、`RequireAdminOrAbove`、`RequireEditorOrAbove`、`RequireOwnershipOrAdmin`、`CanManageUserRole`、`GetCurrentUser`、`GetCurrentUserID`、`GetCurrentUserRole` | RBAC 鉴权与上下文读取，权限判定全部委托注入的 `RBACService` |
| `ratelimit.go` | `RateLimit`、`RateLimitPerUser`、`RateLimiter`、`NewRateLimiter` | 进程内计数限流器，安全中间件复用同一实现；过期记录随新客户端加入惰性清扫 |
| `language.go` | `LanguageMiddlewareFromConfig` | 解析 `Accept-Language` 写入请求上下文，白名单外回退缺省语言；只解析不中断请求 |

## 装配与约定

- 全局中间件顺序见 `internal/router/router.go`：Logger、`gin.Recovery`、RequestID、CORS、Security、Language；管理员分组额外挂 `AdminSecurityMiddlewareFromConfig`。
- 权限清单唯一权威为 `configs/config.yaml` 的 `rbac.role_permissions`，本文件不重复罗列。
- 限流计数与令牌表均为进程内存结构，单实例部署前提下的取舍见 `AGENTS.md` 第 1.7 节。
- 实例化纪律（依赖由组合根注入）见 `AGENTS.md` 第 1.4 节。
