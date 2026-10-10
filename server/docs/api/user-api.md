# 用户管理 API 文档

## 概述

用户管理模块提供用户登录、信息管理等功能，支持基于角色的权限控制（RBAC）。

## 角色权限说明

| 角色 | 权限级别 | 说明 |
|------|----------|------|
| superadmin | 4 | 超级管理员，拥有所有权限 |
| admin | 3 | 管理员，可管理内容和普通用户 |
| editor | 2 | 编辑者，可发布和管理文章 |
| user | 1 | 普通用户，基础读写权限 |

## 接口列表

### 1. 用户登录

用户账号密码登录，获取访问令牌。连续密码失败达到 `security.login_lockout` 配置阈值后账户将被锁定一段时间，到期自动解除，登录成功后失败计数清零；配置缺省时按内置默认值 5 次 / 15 分钟执行。锁定检查先于密码校验，锁定期间密码正确同样拒绝。

#### 请求信息

- **接口地址**: `/api/users/login`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| username | string | 是 | 用户名或邮箱 | 非空字符串 |
| password | string | 是 | 密码 | 非空字符串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "123456"
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.user | object | 是 | 用户信息 |
| data.user.id | integer | 是 | 用户ID |
| data.user.username | string | 是 | 用户名 |
| data.user.email | string | 是 | 邮箱 |
| data.user.nickname | string | 是 | 昵称 |
| data.user.avatar | string | 是 | 头像URL |
| data.user.birthday | string | 否 | 生日，格式 YYYY-MM-DD |
| data.user.role | string | 是 | 用户角色 |
| data.user.status | integer | 是 | 用户状态，1启用0禁用2锁定 |
| data.user.createdAt | string | 是 | 创建时间 |
| data.user.updatedAt | string | 是 | 更新时间 |
| data.accessToken | string | 是 | 访问令牌（不透明令牌，见下方注记） |
| data.refreshToken | string | 是 | 刷新令牌（不透明令牌） |
| data.expiresIn | integer | 是 | 访问令牌有效期，单位秒 |
| data.permissions | string[] | 是 | 当前角色权限列表（后端唯一权威） |

> **线格式注记**：令牌为**不透明随机串**（32 位十六进制，服务端令牌表是身份唯一权威，不携带可解码信息）。完整协议见 `contracts/auth-protocol.md`。

> **响应形状注记**：本模块存在两个用户响应形状。`users/login`、`users/get`、`users/list`、`users/create`、`users/update` 输出裁剪形状（字段见上表，不含手机号、简介、时区等）；`users/profile` 与 `users/profile/update` 直接输出用户实体的自助全量形状，字段清单与时间格式见第 9 节。

#### 响应示例

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "nickname": "管理员",
      "avatar": "",
      "birthday": "1990-01-01",
      "role": "admin",
      "status": 1,
      "createdAt": "2024-01-01 10:00:00",
      "updatedAt": "2024-01-01 10:00:00"
    },
    "accessToken": "<opaque access token>",
    "refreshToken": "<opaque refresh token>",
    "expiresIn": 900,
    "permissions": ["article:read", "comment:create", "comment:read", "comment:update"]
  }
}
```

#### 错误响应

登录失败全部映射为业务码 401，`message` 区分原因：

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 401 | 用户不存在 | 用户名与邮箱两种查找方式均未命中 |
| 401 | 密码错误 | 密码校验失败，失败计数自增 |
| 401 | 密码错误，失败次数过多，账户已被锁定，请稍后再试 | 本次失败使计数达到阈值并写入锁定 |
| 401 | 账户已被锁定，请稍后再试 | 请求落在锁定窗口内，未做密码校验 |
| 401 | 用户已被禁用 | 密码正确但账号状态非启用 |

---

### 2. 获取用户信息

根据用户ID获取用户详细信息。

#### 请求信息

- **接口地址**: `/api/users/get`
- **请求方式**: `POST`
- **权限要求**: `user:list` 权限（管理员及以上），普通用户查看他人资料走 `/api/users/publicProfile`
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 用户ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/get \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应参数

用户信息字段与登录接口的 `data.user` 一致，直接返回在 `data` 下。

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "nickname": "管理员",
    "avatar": "",
    "birthday": "1990-01-01",
    "role": "admin",
    "status": 1,
    "createdAt": "2024-01-01 10:00:00",
    "updatedAt": "2024-01-01 10:00:00"
  }
}
```

---

### 3. 创建用户

创建新用户账号（需要用户创建权限）。

#### 请求信息

- **接口地址**: `/api/users/create`
- **请求方式**: `POST`
- **权限要求**: `user:create` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| username | string | 是 | 用户名 | 长度1-50字符，唯一 |
| email | string | 是 | 邮箱 | 有效邮箱格式，唯一 |
| password | string | 是 | 密码 | 长度6-100字符，须包含字母和数字，且不在弱口令表内 |
| nickname | string | 否 | 昵称 | 长度0-50字符，留空时使用用户名 |
| role | string | 否 | 用户角色 | user/editor/admin/superadmin，默认user |
| birthday | string | 否 | 生日 | 格式 YYYY-MM-DD |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "Newpass123",
    "nickname": "新用户",
    "role": "user",
    "birthday": "1995-06-15"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "用户创建成功",
  "data": {
    "id": 2,
    "username": "newuser",
    "email": "newuser@example.com",
    "nickname": "新用户",
    "avatar": "",
    "birthday": "1995-06-15",
    "role": "user",
    "status": 1,
    "createdAt": "2024-01-01 11:00:00",
    "updatedAt": "2024-01-01 11:00:00"
  }
}
```

---

### 4. 更新用户信息

更新用户信息（需要用户更新权限）。

#### 请求信息

- **接口地址**: `/api/users/update`
- **请求方式**: `POST`
- **权限要求**: `user:update` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 用户ID | 大于0的整数 |
| username | string | 是 | 用户名 | 长度1-50字符，唯一 |
| email | string | 是 | 邮箱 | 有效邮箱格式，唯一 |
| password | string | 否 | 新密码 | 长度6-100字符，须包含字母和数字，且不在弱口令表内，留空则不修改 |
| nickname | string | 否 | 昵称 | 长度0-50字符，留空时使用用户名 |
| role | string | 否 | 用户角色 | user/editor/admin/superadmin |
| birthday | string | 否 | 生日 | 格式 YYYY-MM-DD |
| status | integer | 否 | 用户状态 | 1启用0禁用 |

> **字段语义注记**：服务层对 `role` 与 `status` 均为无条件赋值。省略 `role` 时角色按默认值 `user` 写回，省略 `status` 时状态按 `0`（禁用）写回，客户端需回填当前值后再提交。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 2,
    "username": "newuser",
    "email": "newuser@example.com",
    "nickname": "更新的昵称",
    "role": "user",
    "status": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "用户更新成功",
  "data": {
    "id": 2,
    "username": "newuser",
    "email": "newuser@example.com",
    "nickname": "更新的昵称",
    "avatar": "",
    "birthday": "1995-06-15",
    "role": "user",
    "status": 1,
    "createdAt": "2024-01-01 11:00:00",
    "updatedAt": "2024-01-01 12:00:00"
  }
}
```

---

### 5. 删除用户

删除用户账号（软删除，需要用户删除权限）。

#### 请求信息

- **接口地址**: `/api/users/delete`
- **请求方式**: `POST`
- **权限要求**: `user:delete` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 用户ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 2
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "用户删除成功"
}
```

说明：删除类接口不返回 `data` 字段。

---

### 6. 获取用户列表

分页获取用户列表（需要用户列表权限）。

#### 请求信息

- **接口地址**: `/api/users/list`
- **请求方式**: `POST`
- **权限要求**: `user:list` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| keyword | string | 否 | 关键词，按用户名、邮箱或昵称模糊匹配 | 无长度约束，空值不参与过滤 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.users | array | 是 | 用户列表 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
| data.pages | integer | 是 | 总页数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "nickname": "管理员",
        "avatar": "",
        "birthday": "1990-01-01",
        "role": "admin",
        "status": 1,
        "createdAt": "2024-01-01 10:00:00",
        "updatedAt": "2024-01-01 10:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10,
    "pages": 1
  }
}
```

---

### 7. 刷新访问令牌

使用刷新令牌获取新的访问令牌。服务端在旋转前查库校验令牌归属用户存在且状态正常，被禁用或已删除用户的刷新令牌无法换取新令牌对。

> **刷新即旋转**：成功响应后旧刷新令牌立即从令牌表撤销，同一刷新令牌不能重复兑换；校验失败时旧令牌保持不变。

> 该端点为 Header 通道的存量客户端保留，浏览器场景请改用 [`POST /api/auth/session`](./session-api.md) 的 Cookie 会话与续期。

#### 请求信息

- **接口地址**: `/api/auth/refresh`
- **请求方式**: `POST`
- **权限要求**: 无需认证（但需要有效的刷新令牌）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| refreshToken | string | 是 | 刷新令牌 | 不透明令牌线格式（见登录接口注记） |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "9f2a5c1e4b7d8056a1c3e5f70b2d4689"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "令牌刷新成功",
  "data": {
    "accessToken": "<opaque access token>",
    "refreshToken": "<opaque refresh token>",
    "expiresIn": 900
  }
}
```

> `expiresIn` 由 `token.access_expire`（分钟）换算为秒，示例值对应默认的 15 分钟。

---

### 8. 用户登出

登出用户账号，撤销访问令牌与刷新令牌构成的对。

> 双轨语义与 Cookie 清除行为见 [`session-api.md`](./session-api.md) 的「相关端点」。

#### 请求信息

- **接口地址**: `/api/auth/logout`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| refreshToken | string | 否 | 刷新令牌，提交后与访问令牌一并撤销 | 最大512字符，不透明令牌线格式 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{"refreshToken": "<opaque refresh token>"}'
```

请求体可省略，此时仅撤销访问令牌。

#### 响应示例

```json
{
  "code": 200,
  "message": "登出成功"
}
```

说明：登出接口不返回 `data` 字段。

---

## 错误响应

### 常见错误码

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 请求参数错误: ... | `binding` 校验失败，参数格式或内容不正确 |
| 400 | 用户名已存在 | 创建用户时用户名冲突（哨兵错误 `ErrUsernameTaken`） |
| 400 | 邮箱已被其他用户使用 | 更新用户时邮箱被占用（哨兵错误 `ErrEmailTaken`） |
| 400 | 不能删除自己的账户 | `users/delete` 目标为当前操作者本人 |
| 400 | 超级管理员角色只能通过系统管理员设置 / 超级管理员角色不能被降级 / 无效的目标角色: x | 更新用户的角色转换校验失败 |
| 400 | 请求参数不合法：批量操作的目标不能为空 / 批量操作的单次人数不能超过 100 / 批量操作不能包含自己 / 用户 N 不存在 | 批量接口的业务规则校验失败 |
| 401 | 未提供认证令牌 / 无效的认证令牌 | 缺少访问令牌，或令牌未登记、已过期、类型不符 |
| 401 | 用户不存在 / 密码错误 / 用户已被禁用 / 账户已被锁定，请稍后再试 | 登录失败，`users/login` 全部错误映射为 401 |
| 401 | 未登录 | 认证中间件通过但上下文缺少用户ID，自助接口直接返回 |
| 403 | 权限不足，无法访问该资源 | 缺少接口所需 RBAC 权限 |
| 403 | 权限不足，无法管理该角色的用户 / 权限不足，无法分配该角色 / 权限不足，无法删除该角色的用户 | 角色管理规则拒绝，操作者级别低于目标用户 |
| 404 | 用户不存在 | `users/get`、`users/delete`、`users/profile` 等目标用户不存在 |
| 500 | 服务器内部错误 | 未归入哨兵错误的业务失败与内部异常，错误原文仅落日志 |

> 现行实现中「邮箱已存在」（创建用户邮箱冲突）、「用户名已被其他用户使用」（更新用户用户名冲突）、「密码长度不能少于6位」等密码强度错误、「旧密码不正确」均未使用哨兵错误，实际统一返回 500 服务器内部错误，见各接口小节。

### 错误响应示例

```json
{
  "code": 400,
  "message": "请求参数错误: Key: 'CreateUserRequest.Username' Error:Field validation for 'Username' failed on the 'max' tag"
}
```

```json
{
  "code": 401,
  "message": "无效的认证令牌"
}
```

```json
{
  "code": 403,
  "message": "权限不足，无法访问该资源"
}
```

```json
{
  "code": 500,
  "message": "服务器内部错误"
}
```

---

### 9. 获取当前用户资料

获取当前登录用户的完整资料。

#### 请求信息

- **接口地址**: `/api/users/profile`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

无请求参数，请求体传空对象。

#### 响应参数

该端点直接输出用户实体的自助全量形状（不含密码、失败计数、锁定时间等 `json:"-"` 字段），与登录等接口的裁剪形状不同：

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 用户ID |
| username | string | 是 | 用户名 |
| email | string | 是 | 邮箱 |
| phone | string | 否 | 手机号，未绑定时不返回该字段 |
| nickname | string | 是 | 昵称 |
| avatar | string | 是 | 头像URL |
| coverImage | string | 是 | 个人主页封面图URL |
| bio | string | 是 | 个人简介 |
| website | string | 是 | 个人网站URL |
| location | string | 是 | 常居地描述 |
| gender | integer | 否 | 性别，未设置时不返回该字段 |
| birthday | string | 是 | 生日，当天零点输出 `YYYY-MM-DD`，未设置为 `null` |
| timezone | string | 是 | 时区标识 |
| locale | string | 是 | 界面语言标识 |
| role | string | 是 | 用户角色 |
| status | integer | 是 | 用户状态，1启用0禁用2锁定 |
| createdAt | string | 是 | 创建时间，RFC3339 格式 |
| updatedAt | string | 是 | 更新时间，RFC3339 格式 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "nickname": "管理员",
    "avatar": "",
    "coverImage": "",
    "bio": "白天写代码，晚上写字。",
    "website": "https://example.com",
    "location": "杭州",
    "birthday": "1990-01-01",
    "timezone": "Asia/Shanghai",
    "locale": "zh-CN",
    "role": "admin",
    "status": 1,
    "createdAt": "2024-01-01T10:00:00+08:00",
    "updatedAt": "2024-01-01T10:00:00+08:00"
  }
}
```

---

### 10. 更新当前用户资料

更新当前登录用户的自助资料，仅昵称、头像、简介与网站四个字段，显式传入才更新。用户名、角色等敏感字段请使用管理端 `users/update`。

#### 请求信息

- **接口地址**: `/api/users/profile/update`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| nickname | string | 否 | 昵称，清空时回退为用户名 | 最大50字符 |
| avatar | string | 否 | 头像URL | 最大255字符 |
| bio | string | 否 | 个人简介 | 最大500字符 |
| website | string | 否 | 个人网站URL | 最大255字符 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/profile/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "nickname": "新昵称",
    "bio": "新的个人简介"
  }'
```

#### 响应示例

响应格式与第 9 节一致，返回更新后的自助全量形状资料。

---

### 11. 修改密码

校验旧密码后更新为新的登录密码，新密码需满足强度要求。修改成功后服务端撤销该用户当前全部既有令牌，本次请求使用的令牌同样失效，客户端应清除本地会话并引导重新登录。

#### 请求信息

- **接口地址**: `/api/users/changePassword`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| oldPassword | string | 是 | 旧密码 | 非空字符串 |
| newPassword | string | 是 | 新密码 | 8-64字符，且须通过密码强度校验（含字母与数字、不少于6位、不在弱口令表内） |

#### 响应示例

```json
{
  "code": 200,
  "message": "密码修改成功"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 401 | 未提供或无效的认证令牌 |
| 500 | 旧密码不正确或新密码强度不足：实现未使用哨兵错误，错误原文不进入响应体，统一返回「服务器内部错误」 |

---

### 12. 批量删除用户

按 ID 集合批量软删除用户，逐项复用单删的业务规则，任一目标违规整批拒绝。

#### 请求信息

- **接口地址**: `/api/users/batchDelete`
- **请求方式**: `POST`
- **权限要求**: `user:delete` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| ids | integer[] | 是 | 目标用户ID集合 | 1-100 项，每项大于0 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/batchDelete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "ids": [2, 3]
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "批量删除用户成功"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 批量操作的目标不能为空、单次人数超过100、集合包含操作者本人、目标用户不存在 |
| 403 | 目标用户角色超出操作者可管理的范围 |

---

### 13. 批量更新用户状态

按 ID 集合批量启用或禁用用户，业务规则与批量删除一致；仅开放启用与禁用两态，锁定状态由登录锁定机制专用。

#### 请求信息

- **接口地址**: `/api/users/batchUpdateStatus`
- **请求方式**: `POST`
- **权限要求**: `user:update` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| ids | integer[] | 是 | 目标用户ID集合 | 1-100 项，每项大于0 |
| status | integer | 是 | 目标状态 | 取值 0 或 1，省略时按 0（禁用）写回 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/batchUpdateStatus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "ids": [2, 3],
    "status": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "批量更新用户状态成功"
}
```

---

### 14. 建立或续期会话

凭刷新令牌换取新令牌对并写入 HttpOnly Cookie，是浏览器场景的默认认证通道。

- **接口地址**: `/api/auth/session`
- **请求方式**: `POST`
- **权限要求**: 无需访问令牌，但需要有效的刷新令牌

请求参数、Cookie 属性、错误分档与前端接入约定见 [`session-api.md`](./session-api.md)。
