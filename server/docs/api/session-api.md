# 会话 API 文档

## 概述

会话端点把令牌对写入 HttpOnly Cookie，浏览器随后自动随同源请求携带，页面脚本无法读取。

令牌为服务端签发的**不透明随机串**（16 字节密码学随机量的十六进制编码，32 个字符），自身不承载任何可解码的身份信息；身份与生命周期由服务端令牌表唯一权威维护，Cookie 只是传输通道。协议全文见 [`contracts/auth-protocol.md`](../../../contracts/auth-protocol.md) 第 3 节。

会话端点与登录的衔接：`POST /api/users/login` 返回令牌对（登录接口见 [user-api.md](./user-api.md)），客户端随即调用本端点以刷新令牌换取 Cookie 会话。登录成功后刷新即旋转，旧刷新令牌在成功响应后失效。

## 接口列表

### 1. 建立或续期会话

登录后首次建立会话时刷新令牌经请求体提交；续期场景由浏览器自动携带 `mb_refresh_token` Cookie，请求体可省略。成功时旋转签发新令牌对并重写两个 Cookie。

#### 请求信息

- **接口地址**: `/api/auth/session`
- **请求方式**: `POST`
- **权限要求**: 无需访问令牌，但需要有效的刷新令牌
- **Content-Type**: `application/json`（请求体可省略）

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| refreshToken | string | 否 | 刷新令牌，省略时从 `mb_refresh_token` Cookie 读取 | 最大 512 字符 |

#### 请求示例

```bash
# 首次建立会话：刷新令牌经请求体提交
curl -X POST http://localhost:3000/api/auth/session \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "9f2a5c1e4b7d8056a1c3e5f70b2d4689"}'

# 续期：浏览器自动携带 Cookie，无需请求体
curl -X POST http://localhost:3000/api/auth/session \
  -H "Cookie: mb_refresh_token=9f2a5c1e4b7d8056a1c3e5f70b2d4689"
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200 表示成功 |
| message | string | 是 | 响应消息，成功为「会话已建立」 |
| data | object | 是 | 响应数据 |
| data.expiresIn | integer | 是 | 访问令牌有效期（秒），由 `token.access_expire` 分钟换算 |

#### 响应头

成功响应写入两个 Cookie，属性如下：

| Cookie | 承载 | Max-Age | 公共属性 |
|--------|------|---------|----------|
| `mb_access_token` | 访问令牌 | `token.access_expire` × 60（默认 900 秒） | `Path=/api`、`HttpOnly`、`SameSite=Lax`、`Secure` 由 `token.cookie_secure` 控制 |
| `mb_refresh_token` | 刷新令牌 | `token.refresh_expire` × 3600（默认 604800 秒） | 同上 |

`Path=/api` 把 Cookie 的发送范围限定在业务端点，缩小暴露面；`Secure` 在生产环境必须为 `true`。

#### 响应示例

```json
{
  "code": 200,
  "message": "会话已建立",
  "data": {
    "expiresIn": 900
  }
}
```

#### 错误响应

业务错误经 `pkg/response` 信封返回，**HTTP 状态码恒为 200**，错误语义只体现在响应体 `code` 字段，前端以 `code === 401` 判定并触发续期。

| 业务码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 缺少刷新令牌 | 请求体与 Cookie 均未提供刷新令牌 |
| 401 | 令牌无效或已失效 / 令牌已过期 / 令牌类型不符 | 令牌未登记、已过期，或提交的是访问令牌 |
| 401 | 用户不存在或已失效 / 用户已被禁用 | 旋转前查库校验失败，令牌不再续期 |

```json
{
  "code": 401,
  "message": "令牌无效或已失效"
}
```

## 相关端点

| 端点 | 说明 | 文档位置 |
|------|------|----------|
| `POST /api/auth/refresh` | Header 通道的刷新，为存量客户端过渡保留，新前端不再调用 | [user-api.md](./user-api.md) 第 7 节 |
| `POST /api/auth/logout` | 双轨撤销：Authorization 头与会话 Cookie 任一存在即处理，Cookie 会话随 `Max-Age=-1` 清除 | [user-api.md](./user-api.md) 第 8 节 |
| `POST /api/users/login` | 登录并返回令牌对与权限列表 | [user-api.md](./user-api.md) |

## 前端接入约定

- **401 判定**：以响应体 `code === 401` 判定认证失效，禁止回退文案匹配。
- **续期**：收到 401 后必须实际执行一次会话续期再决定是否重放，不得以本地状态短路跳过；重放最多一次，重放后仍为 401 则触发登出流程。
- **存储**：前端不持久化任何令牌（localStorage 令牌已退役），仅持久化用户信息与权限列表用于界面渲染。

## 已知限制

- 令牌表为**进程内存**结构，服务重启或发布导致全体用户登出；多副本部署前必须完成令牌表持久化（见 [`docs/architecture-rules.md`](../../../docs/architecture-rules.md) 第 7 节）。
- 会话 Cookie 由同源标签页共享，令牌旋转不产生跨标签页竞争。
