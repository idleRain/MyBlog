# 健康检查 API 文档

## 概述

健康检查接口用于检测服务器运行状态，区分存活（liveness）与就绪（readiness）两个探针维度：
存活探针仅表明进程在运行，就绪探针额外探测数据库连通性。
健康探针属基础设施端点，按编排器事实标准使用 `GET` 方法。

## 接口列表

### 1. 存活探针（liveness）

检查服务器进程是否在运行，不探测数据库。

#### 请求信息

- **接口地址**: `/api/health`
- **请求方式**: `GET`
- **权限要求**: 无需认证

#### 请求示例

```bash
curl http://localhost:3000/api/health
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
    "service": "IdleRain Blogs API"
  }
}
```

### 2. 就绪探针（readiness）

在存活基础上探测数据库连通性，供编排器在数据库异常时摘除流量。

#### 请求信息

- **接口地址**: `/api/health/ready`
- **请求方式**: `GET`
- **权限要求**: 无需认证

#### 请求示例

```bash
curl http://localhost:3000/api/health/ready
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示就绪，503表示未就绪 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.status | string | 是 | "ready"表示就绪，"unready"表示未就绪 |
| data.service | string | 是 | 服务名称 |
| data.reason | string | 否 | 未就绪原因，仅在 503 时返回 |

#### 响应示例

```json
{
  "code": 200,
  "message": "服务就绪",
  "data": {
    "status": "ready",
    "service": "IdleRain Blogs API"
  }
}
```

#### 错误响应

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 503 | 服务未就绪 | 数据库连通性探测失败，data.reason 携带原因 |
