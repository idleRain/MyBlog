# Response 模块

统一的API响应格式包，提供标准化的HTTP响应处理。

## 设计原则

- 统一格式：所有API都使用相同的响应结构
- 状态码标准化：预定义常用的状态码
- 易于使用：提供便捷的方法创建响应
- 错误友好：详细的错误信息

## 响应结构

```go
type Response struct {
    Code    int         `json:"code"`              // 响应码
    Message string      `json:"message"`           // 响应消息
    Data    interface{} `json:"data,omitempty"`    // 响应数据
    Error   string      `json:"error,omitempty"`   // 错误信息
}
```

## 状态码定义

| 状态码 | 常量 | 说明 |
|--------|------|------|
| 200 | CodeSuccess | 操作成功 |
| 400 | CodeInvalid | 请求参数错误 |
| 401 | CodeAuth | 认证失败 |
| 403 | CodeForbid | 权限不足 |
| 404 | CodeNotFound | 资源不存在 |
| 500 | CodeError | 服务器内部错误 |

## 核心方法

### 成功响应

#### Success
```go
response.Success(c, data)
```
返回成功响应，使用默认成功消息。

#### SuccessWithMessage
```go
response.SuccessWithMessage(c, "自定义成功消息", data)
```
返回成功响应，使用自定义消息。

### 错误响应

#### Error
```go
response.Error(c, code, message, err)
```
通用错误响应方法。

#### BadRequest
```go
response.BadRequest(c, "参数错误信息")
```
请求参数错误 (400)。

#### InternalError
```go
response.InternalError(c, "错误详情")
```
服务器内部错误 (500)。

#### NotFound
```go
response.NotFound(c, "资源不存在")
```
资源不存在 (404)。

#### Unauthorized
```go
response.Unauthorized(c, "认证失败")
```
未授权 (401)。

#### Forbidden
```go
response.Forbidden(c, "权限不足")
```
权限不足 (403)。

## 使用示例

### Handler中的使用

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    // 参数验证失败
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "请求参数错误: "+err.Error())
        return
    }
    
    // 业务逻辑处理
    user, err := h.userService.CreateUser(&req)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    
    // 成功响应
    response.SuccessWithMessage(c, "用户创建成功", user)
}
```

### 响应示例

#### 成功响应
```json
{
    "code": 200,
    "message": "用户创建成功",
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com"
    }
}
```

#### 错误响应
```json
{
    "code": 400,
    "message": "请求参数错误",
    "error": "用户名不能为空"
}
```

#### 服务器错误
```json
{
    "code": 500,
    "message": "服务器内部错误",
    "error": "数据库连接失败"
}
```

## 最佳实践

### 1. 选择合适的响应方法

```go
// 成功场景
response.Success(c, user)
response.SuccessWithMessage(c, "操作成功", result)

// 参数错误
response.BadRequest(c, "用户名格式不正确")

// 业务错误
response.InternalError(c, err.Error())

// 资源不存在
response.NotFound(c, "用户不存在")
```

### 2. 错误信息处理

```go
// 不暴露内部错误详情
if err != nil {
    log.Printf("内部错误: %v", err)
    response.InternalError(c, "操作失败，请稍后重试")
    return
}

// 提供有用的错误信息
if user == nil {
    response.NotFound(c, "用户不存在")
    return
}
```

### 3. 数据格式化

```go
// 构建响应数据
data := gin.H{
    "users":     users,
    "total":     total,
    "page":      page,
    "page_size": pageSize,
}
response.Success(c, data)
```

## 扩展指南

### 添加新的状态码

1. 在常量中定义新的状态码
2. 添加对应的便捷方法
3. 更新文档

```go
const (
    CodeRateLimit = 429  // 请求频率限制
)

func RateLimit(c *gin.Context, message string) {
    Error(c, CodeRateLimit, message, "请求过于频繁")
}
```

### 国际化支持

可以扩展支持多语言响应消息：

```go
func SuccessWithI18n(c *gin.Context, key string, data interface{}) {
    message := i18n.GetMessage(c, key)
    SuccessWithMessage(c, message, data)
}
```

## 注意事项

1. 所有响应都使用HTTP 200状态码，业务状态通过code字段区分
2. 敏感错误信息不应暴露给客户端
3. 保持响应格式的一致性
4. error字段只在出错时返回，成功时应该省略