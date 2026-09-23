# 认证协议契约（C3）

> 本文档是前后端认证协议的**唯一权威描述**。任何实现变更必须先改本文档，再双端同步切换。
> 生效范围：`server/internal/service/token.go`、`packages/http/src/client.ts`、两应用 `src/lib/service/index.ts`。
> 相关债务登记见 `docs/architecture-rules.md` §6.3 与 D10/D11。

## 1. 令牌对形状

```json
{
  "accessToken": "<opaque token>",
  "refreshToken": "<opaque token>",
  "expiresIn": 1800
}
```

- `expiresIn` 为访问令牌有效期，单位秒，来源于 `config.yaml` 的 `jwt.access_expire`（分钟）换算。
- 刷新令牌有效期以 `jwt.refresh_expire`（小时）为准，不随响应下发。

## 2. 线格式：不透明令牌

**前端存储与传输的是服务端签发的随机串，不含任何可解码的身份信息。**

- 服务端 `GenerateTokenPair` 为每次签发生成 16 字节密码学随机量，十六进制编码为 32 个字符作为令牌串。
- 令牌自身不承载用户标识与有效期，身份与生命周期登记在服务端令牌表；`ValidateAccessToken` 与 `ValidateRefreshToken` 以查表结果为唯一依据。
- 校验同时核对令牌类型，访问令牌与刷新令牌不可互换使用。
- 令牌不可由调用方构造：任何未经服务端签发的串都会因查表失败被拒绝，不存在依据请求内容现算凭证的路径。
- 服务重启会清空令牌表，所有用户需重新登录。这是有状态方案的固有代价，多实例部署前必须将令牌表迁至共享存储，见 D11。

## 3. 刷新协议

- 端点：`POST /api/auth/refresh`，请求体 `{ "refreshToken": "<refresh token>" }`。
- 刷新即旋转：成功后旧 refresh token 被撤销，响应返回全新令牌对。
- 旋转前服务端查库校验令牌归属用户存在且状态正常，被禁用或已删除用户的刷新请求返回业务码 401。
- 前端 `service/index.ts` 的 `refreshAccessToken` 以**裸 ky 直连**该端点（豁免 A1 页面直连 ky 规则，避免循环依赖）。

## 4. 401 语义

- 后端认证失败统一返回 **HTTP 200 + 业务码 `code: 401`**（`pkg/response` 信封），不以 HTTP 状态码标识。
- 前端 `packages/http/src/client.ts` 以响应体 `code === 401` 判定认证失效并尝试刷新（D10 已收敛，禁止回退文案匹配）。
- 刷新成功返回原响应，**调用方不自动重试**——此为已知半成品限制（诊断记录），由上层业务处理。
- 刷新失败或刷新请求自身 401，触发 `onAuthFailure` 回调（应用层清除状态并跳转登录页）。

## 5. 登出语义

- 端点：`POST /api/auth/logout`，请求头 `Authorization: Bearer <access token>`。
- 请求体可选提交 `{ "refreshToken": "<refresh token>" }`，提交后访问与刷新令牌一并撤销；请求体可省略，此时仅撤销访问令牌。
- 修改密码成功后服务端撤销该用户当前全部既有令牌，客户端须清除本地会话并引导重新登录。
- 撤销即从服务端令牌表移除记录，过期记录随签发惰性清理（内存实现，仅单实例生效，见 D11）。

## 6. 已知限制（登记）

| 限制 | 影响 | 计划 |
|---|---|---|
| 令牌表为内存 map（已加互斥锁，过期记录随签发惰性清理） | 服务重启导致全部会话失效；多实例部署即失效 | D11：换持久化存储前保持单实例前提 |
| 刷新成功后不自动重试原请求 | 极端竞态下用户需手动重试 | C3 后续：调用方按新令牌重试 |
| 登录失败（密码错误）同样返回 code 401 | 前端已排除登录端点不触发刷新 | 已在 client.ts 固化 |
