# 认证协议契约（C3）

> 本文档是前后端认证协议的**唯一权威描述**。任何实现变更必须先改本文档，再双端同步切换。
> 生效范围：`server/internal/service/jwt.go`、`packages/http/src/client.ts`、两应用 `src/lib/service/index.ts`。
> 相关债务登记见 `docs/architecture-rules.md` §6.3 与 D10/D11。

## 1. 令牌对形状

```json
{
  "accessToken": "<payload-only jwt>",
  "refreshToken": "<payload-only jwt>",
  "expiresIn": 1800
}
```

- `expiresIn` 为访问令牌有效期，单位秒，来源于 `config.yaml` 的 `jwt.access_expire`（分钟）换算。
- 刷新令牌有效期以 `jwt.refresh_expire`（小时）为准，不随响应下发。

## 2. 线格式：payload-only JWT（已知怪癖）

**前端存储与传输的是无点号的 Base64 payload，不是标准三段式 JWT。**

- 服务端 `GenerateTokenPair` 仅序列化 `{u: userID, exp: unix}` 为 JSON 后 Base64 URL 编码，**不含签名**。
- 服务端 `ValidateAccessToken` / `ValidateRefreshToken` 在收到无点号 token 时，经 `ReconstructFullToken` 补回固定 Header 与 HMAC-SHA256 签名后校验。
- 任何新前端实现者必须知晓此怪癖，否则无法通过服务端校验。

> 迁移标记：若未来切换为标准 JWT，按本文档流程「先改文档 → 双端同窗口切换 → 更新 fixtures」。

## 3. 刷新协议

- 端点：`POST /api/auth/refresh`，请求体 `{ "refreshToken": "<refresh token>" }`。
- 刷新即旋转：成功后旧 refresh token 被撤销，响应返回全新令牌对。
- 前端 `service/index.ts` 的 `refreshAccessToken` 以**裸 ky 直连**该端点（豁免 A1 页面直连 ky 规则，避免循环依赖）。

## 4. 401 语义

- 后端认证失败统一返回 **HTTP 200 + 业务码 `code: 401`**（`pkg/response` 信封），不以 HTTP 状态码标识。
- 前端 `packages/http/src/client.ts` 以响应体 `code === 401` 判定认证失效并尝试刷新（D10 已收敛，禁止回退文案匹配）。
- 刷新成功返回原响应，**调用方不自动重试**——此为已知半成品限制（诊断记录），由上层业务处理。
- 刷新失败或刷新请求自身 401，触发 `onAuthFailure` 回调（应用层清除状态并跳转登录页）。

## 5. 登出语义

- 端点：`POST /api/auth/logout`，请求头 `Authorization: Bearer <access token>`。
- 服务端撤销访问令牌（内存实现，仅单实例生效，见 D11）。

## 6. 已知限制（登记）

| 限制 | 影响 | 计划 |
|---|---|---|
| 撤销表为无锁内存 map（A1 已加互斥锁） | 多实例部署即失效 | D11：换持久化存储前保持单实例前提 |
| 刷新成功后不自动重试原请求 | 极端竞态下用户需手动重试 | C3 后续：调用方按新令牌重试 |
| 登录失败（密码错误）同样返回 code 401 | 前端已排除登录端点不触发刷新 | 已在 client.ts 固化 |
