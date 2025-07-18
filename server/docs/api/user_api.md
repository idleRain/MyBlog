# 用户管理 API 文档

基础路径: `/api`

## 概述

用户管理模块提供用户的增删改查功能，包括用户注册、信息获取、列表查询和删除操作。

## 通用说明

### 请求方式

所有API都使用 `POST` 方法。

### 请求头

```
Content-Type: application/json
```

### 响应格式

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {...},
    "error": ""
}
```

### 状态码说明

- `200`: 操作成功
- `400`: 请求参数错误
- `401`: 认证失败
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

## 接口列表

### 1. 健康检查

检查服务是否正常运行。

**请求**

- URL: `POST /api/health`
- Body: 无

**响应**

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

### 2. 创建用户

创建新用户账户。

**请求**

- URL: `POST /api/users/create`
- Body:

```json
{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "123456",
    "nickname": "John",
    "birthday": "1990-01-01"
}
```

**参数说明**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，1-50字符，支持中文 |
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码，6-100字符 |
| nickname | string | 否 | 昵称，最多50字符 |
| birthday | string | 否 | 生日，格式 YYYY-MM-DD |

**响应**

```json
{
    "code": 200,
    "message": "用户创建成功",
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "nickname": "John",
        "avatar": "",
        "birthday": "1990-01-01",
        "status": 1,
        "createdAt": "2024-01-01",
        "updatedAt": "2024-01-01"
    }
}
```

**错误响应**

```json
{
    "code": 500,
    "message": "服务器内部错误",
    "error": "用户名已存在"
}
```

### 3. 获取用户信息

根据用户ID获取用户详细信息。

**请求**

- URL: `POST /api/users/get`
- Body:

```json
{
    "id": 1
}
```

**参数说明**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | integer | 是 | 用户ID，大于0 |

**响应**

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
        "status": 1,
        "createdAt": "2024-01-01",
        "updatedAt": "2024-01-01"
    }
}
```

**错误响应**

```json
{
    "code": 404,
    "message": "用户不存在",
    "error": "资源不存在"
}
```

### 4. 获取用户列表

分页获取用户列表。

**请求**

- URL: `POST /api/users/list`
- Body:

```json
{
    "page": 1,
    "pageSize": 10
}
```

**参数说明**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | integer | 否 | 页码，默认1 |
| pageSize | integer | 否 | 每页数量，默认10，最大100 |

**响应**

```json
{
    "code": 200,
    "message": "操作成功",
    "data": {
        "users": [
            {
                "id": 1,
                "username": "john_doe",
                "email": "john@example.com",
                "nickname": "John",
                "avatar": "",
                "birthday": "1990-01-01",
                "status": 1,
                "createdAt": "2024-01-01",
                "updatedAt": "2024-01-01"
            }
        ],
        "total": 1,
        "page": 1,
        "pageSize": 10,
        "pages": 1
    }
}
```

### 5. 删除用户

删除指定用户（软删除）。

**请求**

- URL: `POST /api/users/delete`
- Body:

```json
{
    "id": 1
}
```

**参数说明**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | integer | 是 | 用户ID，大于0 |

**响应**

```json
{
    "code": 200,
    "message": "用户删除成功",
    "data": null
}
```

**错误响应**

```json
{
    "code": 404,
    "message": "用户不存在",
    "error": "资源不存在"
}
```

## 调用示例

### curl 示例

```bash
# 健康检查
curl -X POST http://localhost:3000/api/health

# 创建用户
curl -X POST http://localhost:3000/api/users/create \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "123456",
    "nickname": "John",
    "birthday": "1990-01-01"
  }'

# 获取用户信息
curl -X POST http://localhost:3000/api/users/get \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'

# 获取用户列表
curl -X POST http://localhost:3000/api/users/list \
  -H "Content-Type: application/json" \
  -d '{"page": 1, "pageSize": 10}'

# 删除用户
curl -X POST http://localhost:3000/api/users/delete \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### JavaScript 示例

```javascript
// 创建用户
const createUser = async () => {
    const response = await fetch('http://localhost:3000/api/users/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            username: 'john_doe',
            email: 'john@example.com',
            password: '123456',
            nickname: 'John',
            birthday: '1990-01-01'
        })
    });
    
    const result = await response.json();
    console.log(result);
};

// 获取用户列表
const getUserList = async () => {
    const response = await fetch('http://localhost:3000/api/users/list', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            page: 1,
            pageSize: 10
        })
    });
    
    const result = await response.json();
    console.log(result);
};
```

## 错误码参考

| 业务错误信息   | 说明          |
|----------|-------------|
| "用户名已存在" | 注册时用户名重复    |
| "邮箱已存在"  | 注册时邮箱重复     |
| "用户不存在"  | 查询或操作的用户不存在 |
| "请求参数错误" | 请求参数格式不正确   |
| "创建用户失败" | 数据库操作失败     |
| "删除用户失败" | 删除操作失败      |

## 注意事项

1. 密码字段在响应中会被自动过滤，不会返回给客户端
2. 用户删除采用软删除方式，数据不会真正删除
3. 用户名和邮箱具有唯一性约束
4. 分页查询最大每页100条记录
5. 所有时间字段使用ISO 8601格式
