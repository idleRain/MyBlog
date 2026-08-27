# 健康检查 API 文档

## 概述

健康检查接口用于检测服务器运行状态，通常用于负载均衡器、监控系统等场景。

## 接口列表

### 1. 健康检查

检查服务器运行状态。

#### 请求信息

- **接口地址**: `/api/health`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/health \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.status | string | 是 | 服务状态，"healthy"表示健康 |
| data.service | string | 是 | 服务名称 |

#### 响应示例

```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "healthy",
    "service": "MyBlog API"
  }
}
```

#### 错误响应

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 500 | 服务器内部错误 | 服务器异常 |

#### 错误响应示例

```json
{
  "code": 500,
  "message": "服务器内部错误"
}
```