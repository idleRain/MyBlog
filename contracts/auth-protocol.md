# 认证协议契约（C3）

> 本文档是前后端认证协议的**唯一权威描述**。任何实现变更必须先改本文档，再双端同步切换。
> 生效范围：`server/internal/service/token.go`、`server/internal/handler/session.go`、`packages/http/src/client.ts`、两应用 `src/lib/service/index.ts`。
> 相关债务登记见 `docs/architecture-rules.md` §6.3 与 D10/D11。

## 1. 令牌对形状

```json
{
  "accessToken": "<opaque token>",
  "refreshToken": "<opaque token>",
  "expiresIn": 1800
}
```

- `expiresIn` 为访问令牌有效期，单位秒，来源于 `config.yaml` 的 `token.access_expire`（分钟）换算。
- 刷新令牌有效期以 `token.refresh_expire`（小时）为准，不随响应下发。

## 2. 线格式：不透明令牌

**前端存储与传输的是服务端签发的随机串，不含任何可解码的身份信息。**

- 服务端 `GenerateTokenPair` 为每次签发生成 16 字节密码学随机量，十六进制编码为 32 个字符作为令牌串。
- 令牌自身不承载用户标识与有效期，身份与生命周期登记在服务端令牌表；`ValidateAccessToken` 与 `ValidateRefreshToken` 以查表结果为唯一依据。
- 校验同时核对令牌类型，访问令牌与刷新令牌不可互换使用。
- 令牌不可由调用方构造：任何未经服务端签发的串都会因查表失败被拒绝，不存在依据请求内容现算凭证的路径。
- 服务重启会清空令牌表，所有用户需重新登录。这是有状态方案的固有代价，多实例部署前必须将令牌表迁至共享存储，见 D11。

## 3. 刷新协议与 Cookie 会话（R4 起）

### 3.1 会话 Cookie

- 登录后客户端凭刷新令牌调用 `POST /api/auth/session` 建立会话，服务端旋转签发新令牌对并以 `Set-Cookie` 写入两个 HttpOnly Cookie：
  - `mb_access_token`：承载访问令牌，`Max-Age` 与 `token.access_expire` 一致；
  - `mb_refresh_token`：承载刷新令牌，`Max-Age` 与 `token.refresh_expire` 一致。
- 两个 Cookie 的公共属性：`Path=/api`、`HttpOnly`、`SameSite=Lax`、`Secure` 由 `token.cookie_secure` 配置控制（生产环境必须开启）。
- Cookie 为纯传输通道，令牌仍为服务端签发的不透明随机串，身份与生命周期由服务端令牌表唯一权威维护。
- 同站策略采用 Lax 在缓解跨站伪造的同时保留外部链接跳转后的会话识别；业务接口全部为 POST 且同源部署，跨站 POST 不会携带 Cookie。
- `POST /api/auth/session` 同时承担续期：请求体省略时服务端从 Cookie 读取刷新令牌完成旋转并重写 Cookie。
- 刷新即旋转语义下，旋转结果写入共享的浏览器 Cookie 存储，天然规避多标签页并发旋转的令牌竞争。

### 3.2 双轨过渡

- 服务端令牌解析双轨：`Authorization: Bearer` 头优先，其次读取 `mb_access_token` Cookie；RBAC 链路经 `IdentityProvider` 统一获得双轨能力。
- `POST /api/auth/refresh`（第 4 节语义）为 Header 通道的存量客户端过渡保留，新前端不再调用。
- `POST /api/auth/logout` 双轨撤销：Header 令牌与会话 Cookie 任一存在即处理；Cookie 会话存在时随 Cookie 一并撤销并以 `Max-Age=-1` 指示浏览器清除。
- 前端不再持久化任何令牌（localStorage 令牌退役），仅持久化用户信息与权限列表用于界面渲染。

### 3.3 刷新协议（Header 通道，过渡保留）

- 端点：`POST /api/auth/refresh`，请求体 `{ "refreshToken": "<refresh token>" }`。
- 刷新即旋转：成功后旧 refresh token 被撤销，响应返回全新令牌对。
- 旋转前服务端查库校验令牌归属用户存在且状态正常，被禁用或已删除用户的刷新请求返回业务码 401。

## 4. 401 语义

- 后端认证失败统一返回 **HTTP 200 + 业务码 `code: 401`**（`pkg/response` 信封），不以 HTTP 状态码标识。
- 前端 `packages/http/src/client.ts` 以响应体 `code === 401` 判定认证失效并尝试会话续期（D10 已收敛，禁止回退文案匹配）。
- 收到 401 后前端**必须实际执行一次续期**（调用 `refreshSession` 回调）再决定是否重放；不得以本地状态判断短路跳过。
- 重放**以显式标记判定是否已重放过**，标记随请求上下文（`options.context`）传递。续期成功后重放同一请求，会话 Cookie 由浏览器自动携带最新值。
- 重放最多一次：续期成功且该请求尚未重放过时重放原请求；重放后仍返回 401 则触发 `onAuthFailure`。
- 续期失败或续期请求自身 401，触发 `onAuthFailure` 回调（应用层清除状态并跳转登录页）。

## 5. 跨标签页状态一致性

- 会话 Cookie 由同源标签页共享，令牌旋转不产生跨标签页竞争；localStorage 仅持久化用户信息与权限列表。
- `packages/auth` 的 `createAuthStore` 注册 `storage` 事件监听，其他标签页登出或更新用户信息时重新加载本页状态。`storage` 事件只在写入方之外的标签页触发，不会形成回环。

## 6. 登出语义

- 端点：`POST /api/auth/logout`，Authorization 头与会话 Cookie 双轨。
- Header 通道：请求头 `Authorization: Bearer <access token>`，请求体可选提交 `{ "refreshToken": "<refresh token>" }`，提交后访问与刷新令牌一并撤销。
- Cookie 通道：会话 Cookie 存在时访问与刷新令牌一并撤销，响应以 `Max-Age=-1` 指示浏览器清除两个会话 Cookie。
- 修改密码成功后服务端撤销该用户当前全部既有令牌，客户端须清除本地状态并引导重新登录。
- 撤销即从服务端令牌表移除记录，过期记录随签发惰性清理（内存实现，仅单实例生效，见 D11）。

## 7. 已知限制（登记）

| 限制 | 影响 | 计划 |
|---|---|---|
| 令牌表为内存 map（已加互斥锁，过期记录随签发惰性清理） | 服务重启导致全部会话失效；多实例部署即失效 | D11：换持久化存储前保持单实例前提 |
| 登录失败（密码错误）同样返回 code 401 | 前端已排除登录端点不触发刷新 | 已在 client.ts 固化 |
| 跨标签页对齐依赖 `storage` 事件送达时机 | 事件送达前仍可能携带旧刷新令牌发起一次刷新，此时依赖重放退化为登出 | 已在刷新前补同步存储读取，事件延迟窗口内不再误用旧令牌 |
