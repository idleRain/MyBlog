# Handler 模块

HTTP请求处理层（Handler Layer），负责处理HTTP请求和响应。

## 设计原则

- 请求解析：解析和验证HTTP请求参数
- 业务调用：调用Service层处理业务逻辑
- 响应格式化：统一响应格式和错误处理
- 轻量级：Handler层只负责HTTP协议相关的处理

## 用户控制器 (UserHandler)

### 核心方法

#### 创建用户

- **路由**: `POST /api/v1/users/create`
- **功能**: 创建新用户
- **请求体**:

```json
{
    "username": "john_doe",
    "email": "john@example.com", 
    "password": "123456",
    "nickname": "John"
}
```

#### 获取用户

- **路由**: `POST /api/v1/users/get`
- **功能**: 根据ID获取用户信息
- **请求体**:

```json
{
    "id": 1
}
```

#### 用户列表

- **路由**: `POST /api/v1/users/list`
- **功能**: 分页获取用户列表
- **请求体**:

```json
{
    "page": 1,
    "page_size": 10
}
```

#### 删除用户

- **路由**: `POST /api/v1/users/delete`
- **功能**: 删除指定用户
- **请求体**:

```json
{
    "id": 1
}
```

#### 健康检查

- **路由**: `POST /api/v1/health`
- **功能**: 服务健康状态检查
- **请求体**: 无

### 响应格式

所有接口都使用统一的响应格式：

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {...},
    "error": ""
}
```

### 状态码定义

- `200`: 操作成功
- `400`: 请求参数错误
- `401`: 认证失败
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

### 请求示例

#### 创建用户

```bash
curl -X POST http://localhost:8080/api/v1/users/create \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "123456",
    "nickname": "John"
  }'
```

响应：

```json
{
    "code": 200,
    "message": "用户创建成功",
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "nickname": "John",
        "status": 1,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
    }
}
```

#### 获取用户列表

```bash
curl -X POST http://localhost:8080/api/v1/users/list \
  -H "Content-Type: application/json" \
  -d '{
    "page": 1,
    "page_size": 10
  }'
```

响应：

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {
        "users": [...],
        "total": 50,
        "page": 1,
        "page_size": 10,
        "pages": 5
    }
}
```

#### 健康检查

```bash
curl -X POST http://localhost:8080/api/v1/health
```

响应：

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {
        "status": "ok",
        "message": "服务运行正常"
    }
}
```

### 参数验证

使用Gin的binding功能进行参数验证：

- `required`: 必填字段
- `min/max`: 字符串长度或数值范围
- `email`: 邮箱格式验证
- `omitempty`: 可选字段

### 错误处理

Handler层会将不同类型的错误映射到相应的HTTP响应：

- 参数错误 → BadRequest (400)
- 业务错误 → InternalError (500)
- 资源不存在 → NotFound (404)

## 扩展指南

### 添加新的Handler方法

1. 定义请求和响应结构体
2. 实现Handler方法
3. 添加参数验证
4. 调用Service层方法
5. 格式化响应
6. 注册路由

### 中间件支持

可以添加各种中间件：

```go
// 跨域中间件
r.Use(cors.Default())

// 日志中间件
r.Use(gin.Logger())

// 恢复中间件
r.Use(gin.Recovery())

// 认证中间件
api.Use(authMiddleware())
```

### 最佳实践

1. **统一响应**: 使用response包统一响应格式
2. **参数验证**: 使用binding进行参数验证
3. **错误处理**: 适当的错误分类和响应
4. **轻量级**: Handler只做协议转换，不处理业务逻辑
5. **POST**: 遵循POST-only API设计规范
