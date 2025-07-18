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

- 所有API接口统一使用 `POST` 方法

### 请求头

```
Content-Type: application/json
```

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
    "data": {},
    "error": ""
}
```

### 响应字段说明

- `code`: 业务状态码
- `message`: 响应消息
- `data`: 响应数据 (成功时包含)
- `error`: 错误信息 (失败时包含)

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
        "status": 1,
        "createdAt": "2024-01-01T10:00:00Z",
        "updatedAt": "2024-01-01T10:00:00Z"
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

## 时间格式规范

### 时间字段命名

- `createdAt`: 创建日期
- `updatedAt`: 更新日期
- `birthday`: 生日

### 时间格式

- **统一使用 YYYY-MM-DD 格式**
- 示例: `"2024-01-01"`
- 说明: 只包含日期，不包含具体时间

## 错误处理规范

### 错误响应格式

```json
{
    "code": 400,
    "message": "请求参数错误",
    "error": "用户名不能为空"
}
```

### 常见错误信息

- `"请求参数错误: {具体错误信息}"`
- `"用户名已存在"`
- `"邮箱已存在"`
- `"用户不存在"`
- `"权限不足"`
- `"服务器内部错误"`

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

- 统一使用 `status` 作为状态字段名
- 类型为 `integer`
- 1: 正常/启用
- 0: 禁用/删除

## 接口版本管理

### URL 结构

```
POST /api/{模块}/{操作}
```

### 示例

- `POST /api/users/create` - 创建用户
- `POST /api/users/list` - 获取用户列表
- `POST /api/health` - 健康检查

## 开发约定

### Go 结构体 JSON 标签

```go
type User struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### Gin 请求绑定

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=1,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}
```

### 响应构建

```go
data := gin.H{
    "users":    users,
    "total":    total,
    "page":     page,
    "pageSize": pageSize,
    "pages":    totalPages,
}
response.Success(c, data)
```

## 最佳实践

1. **一致性**: 所有接口遵循相同的命名规范
2. **可读性**: 字段名要清晰易懂
3. **简洁性**: 避免冗余的字段名
4. **扩展性**: 为未来的字段预留命名空间

## 工具推荐

### JSON 格式化工具

- [jsonformatter.org](https://jsonformatter.org/)
- VS Code: JSON Tools 插件

### API 测试工具

- Postman
- curl
- HTTPie

### 文档生成工具

- Swagger/OpenAPI
- Postman Documentation
