# API 规范文档

## 命名规范

### JSON 字段命名

- **统一使用小驼峰命名 (camelCase)**
- 首字母小写，后续单词首字母大写
- 例如：`userId`, `createdAt`, `pageSize`

### 示例对比

#### ❌ 错误的命名方式

```json
{
    "user_id": 1,
    "created_at": "2024-01-01T10:00:00Z",
    "page_size": 10
}
```

#### ✅ 正确的命名方式 (小驼峰)

```json
{
    "userId": 1,
    "createdAt": "2024-01-01T10:00:00Z",
    "pageSize": 10
}
```

## 请求规范

### 请求方式

- 所有业务接口统一使用 `POST` 方法
- 健康探针等基础设施端点按编排器标准使用 `GET`

### 请求头

```
Content-Type: application/json
```

媒体上传接口（`/api/media/upload`）使用 `multipart/form-data`。

### 请求体格式

```json
{
    "字段名": "值"
}
```

### 请求参数命名示例

```json
{
    "username": "john_doe",
    "email": "john@example.com",
    "birthday": "1990-01-01",
    "pageSize": 10
}
```

## 响应规范

### 统一响应格式

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {}
}
```

### 响应字段说明

- `code`: 业务状态码
- `message`: 响应消息，失败时包含具体错误信息
- `data`: 响应数据 (成功时包含，纯操作类接口可能省略)

### 状态码规范

- `200`: 操作成功
- `400`: 请求参数错误
- `401`: 认证失败
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

### 响应数据命名示例

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "nickname": "John",
        "avatar": "",
        "birthday": "1990-01-01",
        "role": "user",
        "status": 1,
        "createdAt": "2024-01-01 10:00:00",
        "updatedAt": "2024-01-01 10:00:00"
    }
}
```

## 分页数据规范

### 分页请求

```json
{
    "page": 1,
    "pageSize": 10
}
```

### 分页响应

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {
        "users": [...],
        "total": 100,
        "page": 1,
        "pageSize": 10,
        "pages": 10
    }
}
```

### 分页字段说明

- `page`: 当前页码
- `pageSize`: 每页数量
- `total`: 总记录数
- `pages`: 总页数

> 分页响应主体位于 `data` 下，列表键与字段集随模块而异（用户列表为 `{ users, total, page, pageSize, pages }`，字典类型列表为 `{ types, total, page, pageSize }`，字典项列表为 `{ items, total, page, pageSize }`）；`page`、`pageSize`、`total` 为各分页接口共有，`pages` 目前仅用户列表返回。各模块的准确形状以对应模块文档为准。

## 时间格式规范

### 时间字段命名

- `createdAt`: 创建日期
- `updatedAt`: 更新日期
- `birthday`: 生日

### 时间格式

时间字段分为两种序列化格式，与后端模型类型对应：

- **JSONDate 类型字段**（如用户信息的 `createdAt`、`updatedAt`、`birthday`）输出 `YYYY-MM-DD HH:mm:ss` 格式
- 示例: `"2024-01-01 10:00:00"`
- JSONDate 字段当天零点时输出纯日期 `YYYY-MM-DD`，空时间输出 `null`
- **time.Time 类型字段**（如文章信息的 `createdAt`、`publishedAt`、`updatedAt`）输出 RFC3339 格式
- 示例: `"2024-01-01T10:00:00Z"`

## 错误处理规范

### 业务响应 HTTP 状态约定

**业务接口（含错误响应）HTTP 状态码恒为 200**，调用方以响应体 `code` 字段判断业务成败，不应依赖 HTTP 状态码。

- `code=200` 表示业务成功，`code=400/401/403/404/500` 表示业务失败
- 例外：仅中间件层与基础设施端点返回真实非 200 的 HTTP 状态 —— 限流 429、UA 拦截与管理员 IP 白名单拦截 403、内容拦截 400、请求体过大 413、未匹配路由 404、CORS 预检 204，以及健康探针未就绪 503
- 原因：业务错误码与 HTTP 状态码语义解耦，便于调用方按业务码统一处理

### 错误响应格式

```json
{
    "code": 400,
    "message": "请求参数错误: 用户名不能为空"
}
```

### 常见错误信息

- `"请求参数错误: {具体错误信息}"`（`binding` 校验失败）
- `"用户名已存在"` / `"邮箱已存在"` / `"邮箱已被其他用户使用"`
- `"用户不存在"`
- `"没有执行此操作的权限"` / `"权限不足，无法访问该资源"`
- `"服务器内部错误"`（未分类错误的脱敏文案，原始错误只写入服务日志）

## 数据类型规范

### 基础类型

- `string`: 字符串
- `integer`: 整数
- `boolean`: 布尔值
- `array`: 数组
- `object`: 对象

### ID 字段

- 统一使用 `id` 作为主键字段名
- 类型为 `integer`
- 大于 0 的正整数

### 状态字段

- 用户状态统一使用 `status` 字段，类型为 `integer`
- 1: 正常/启用
- 0: 禁用
- 用户状态的领域枚举另定义 `2`（锁定），批量状态更新仅开放 `0`/`1`；登录失败锁定经 `lockedUntil` 截止时间判定，不改写 `status`
- 文章状态使用 `status` 字段，类型为 `string`，取值为 `draft`、`published`、`archived`、`private`

## 接口版本管理

### URL 结构

```
POST /api/{模块}/{操作}
```

### 示例

路径由资源复数段与小驼峰操作段组成，查询类接口同样使用 `POST`；健康探针等基础设施端点按编排器标准使用 `GET`。

完整路由清单以 [接口概览](./README.md#接口概览) 为**唯一索引**，本文件不重复罗列具体路由，避免两处清单各自漂移。

## 服务端实现约定

- 请求 DTO 统一定义在 `internal/domain/dto.go`；`binding` tag 只出现在请求 DTO 上，GORM 实体只携带 `json` 与 `gorm` tag。
- 响应统一经 `pkg/response` 构建：成功用 `response.Success`，错误经 `handler.HandleServiceError` 分档，不把内部错误原文写入 `message`。
- Go 结构体命名、注释与代码风格示例见 [`docs/development.md`](../../../docs/development.md) 「代码规范」。

