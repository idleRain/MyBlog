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

用户账号密码登录，获取访问令牌。

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
| data.user.status | integer | 是 | 用户状态，1启用0禁用 |
| data.user.createdAt | string | 是 | 创建时间 |
| data.user.updatedAt | string | 是 | 更新时间 |
| data.accessToken | string | 是 | 访问令牌（payload-only，见下方注记） |
| data.refreshToken | string | 是 | 刷新令牌（payload-only） |
| data.expiresIn | integer | 是 | 访问令牌有效期，单位秒 |
| data.permissions | string[] | 是 | 当前角色权限列表（后端唯一权威） |

> **线格式注记**：令牌为 **payload-only JWT**（无点号 Base64 payload，非标准三段式）。完整协议见 `contracts/auth-protocol.md`。

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
    "accessToken": "<payload-only access token>",
    "refreshToken": "<payload-only refresh token>",
    "expiresIn": 3600,
    "permissions": ["article:read", "comment:create", "comment:read", "comment:update"]
  }
}
```

---

### 2. 获取用户信息

根据用户ID获取用户详细信息。

#### 请求信息

- **接口地址**: `/api/users/get`
- **请求方式**: `POST`
- **权限要求**: 需要登录
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
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
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
| password | string | 是 | 密码 | 长度6-100字符，须包含字母和数字 |
| nickname | string | 否 | 昵称 | 长度0-50字符，留空时使用用户名 |
| role | string | 否 | 用户角色 | user/editor/admin/superadmin，默认user |
| birthday | string | 否 | 生日 | 格式 YYYY-MM-DD |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
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
| password | string | 否 | 新密码 | 长度6-100字符，留空则不修改 |
| nickname | string | 否 | 昵称 | 长度0-50字符，留空时使用用户名 |
| role | string | 否 | 用户角色 | user/editor/admin/superadmin |
| birthday | string | 否 | 生日 | 格式 YYYY-MM-DD |
| status | integer | 否 | 用户状态 | 1启用0禁用 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 2,
    "username": "newuser",
    "email": "newuser@example.com",
    "nickname": "更新的昵称",
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
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
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

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
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

使用刷新令牌获取新的访问令牌。

#### 请求信息

- **接口地址**: `/api/auth/refresh`
- **请求方式**: `POST`
- **权限要求**: 无需认证（但需要有效的刷新令牌）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| refreshToken | string | 是 | 刷新令牌 | JWT格式 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "令牌刷新成功",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 3600
  }
}
```

---

### 8. 用户登出

登出用户账号，使令牌失效。

#### 请求信息

- **接口地址**: `/api/auth/logout`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'
```

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
| 400 | 请求参数错误 | 参数格式或内容不正确 |
| 401 | 未提供认证令牌 / 无效的认证令牌 | 需要登录或令牌已过期 |
| 403 | 权限不足 | 当前角色没有操作权限 |
| 404 | 用户不存在 | 指定的用户ID不存在 |
| 500 | 用户名已存在 / 邮箱已存在 | 创建或更新时违反唯一性约束 |
| 500 | 服务器内部错误 | 服务器异常 |

### 错误响应示例

```json
{
  "code": 400,
  "message": "请求参数错误: 用户名长度必须在1到50之间"
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
  "message": "用户名已存在"
}
```