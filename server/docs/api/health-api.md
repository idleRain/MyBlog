# 健康检查 API 文档

## 概述

健康检查接口用于检测服务器运行状态，区分存活（liveness）与就绪（readiness）两个探针维度：
存活探针仅表明进程在运行，就绪探针额外探测数据库连通性与本地上传目录可写性。
健康探针属基础设施端点，按编排器事实标准使用 `GET` 方法。

就绪探针失败时只返回稳定原因码，故障细节写入服务日志。

## 接口列表

### 1. 存活探针（liveness）

检查服务器进程是否在运行，不探测任何下游依赖。

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

在存活基础上依次探测数据库连通性与本地上传目录可写性，任一项失败即返回 503，供编排器摘除流量。
上传目录不存在时按上传路径的既有语义创建，创建成功后以真实的文件写入与删除判定可写性。

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
| data.reason | string | 否 | 未就绪原因码，仅在 503 时返回 |

#### 原因码取值

| 原因码 | 含义 |
|--------|------|
| database_unavailable | 数据库连通性探测失败 |
| uploads_unwritable | 上传目录不可写，或目录创建失败 |

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

```json
{
  "code": 503,
  "message": "服务未就绪",
  "data": {
    "status": "unready",
    "service": "IdleRain Blogs API",
    "reason": "uploads_unwritable"
  }
}
```

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 503 | 服务未就绪 | 数据库连通性或上传目录可写性探测失败，data.reason 携带稳定原因码，详细原因见服务日志 |
