# 媒体文件 API 文档

## 概述

媒体文件模块提供图片、视频、文档等资源的上传、查看、管理与删除功能。文件默认存储于本地 `uploads/` 目录，支持 SHA256 哈希秒传去重。

## 存储说明

- **存储目录**: `server/configs/config.yaml` 中 `media.upload_dir` 配置，默认 `uploads/`
- **访问前缀**: `media.base_url` 配置，默认 `/uploads`
- **单文件上限**: `media.max_size_mb` 配置，默认 10MB
- **文件命名**: 上传后以 UUID 重命名，按年月分目录归档

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 上传文件 | `file:upload` | editor及以上 |
| 查看文件列表/详情 | `file:read` | editor及以上 |
| 删除文件 | `file:delete` | admin及以上，或文件上传者本人 |

> 非管理员仅能查看与删除自己上传的文件。

## 接口列表（均需登录 + 相应权限）

### 1. 上传文件

#### 请求信息

- **接口地址**: `/api/media/upload`
- **请求方式**: `POST`
- **权限要求**: `file:upload`（editor及以上）
- **Content-Type**: `multipart/form-data`

#### 请求参数（表单字段）

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | file | 是 | 待上传的文件，单文件 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/media/upload \
  -H "Authorization: Bearer {accessToken}" \
  -F "file=@/path/to/image.png"
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data | object | 是 | 媒体文件信息 |
| data.id | integer | 是 | 文件ID |
| data.filename | string | 是 | 原始文件名 |
| data.storedName | string | 是 | 存储文件名（UUID） |
| data.filePath | string | 是 | 文件存储路径 |
| data.fileUrl | string | 是 | 文件访问URL |
| data.mimeType | string | 是 | MIME类型 |
| data.fileSize | integer | 是 | 文件大小（字节） |
| data.fileHash | string | 是 | SHA256哈希值 |
| data.status | string | 是 | 文件状态，active表示可用 |
| data.storageType | string | 是 | 存储类型，local表示本地 |
| data.isPublic | boolean | 是 | 是否公开访问 |
| data.createdAt | string | 是 | 上传时间 |

#### 响应示例

```json
{
  "code": 200,
  "message": "文件上传成功",
  "data": {
    "id": 1,
    "filename": "image.png",
    "storedName": "550e8400-e29b-41d4-a716-446655440000.png",
    "filePath": "uploads/2026/01/550e8400-e29b-41d4-a716-446655440000.png",
    "fileUrl": "/uploads/2026/01/550e8400-e29b-41d4-a716-446655440000.png",
    "mimeType": "image/png",
    "fileSize": 10240,
    "fileHash": "a5b8c9d0e1f2...",
    "status": "active",
    "storageType": "local",
    "isPublic": true,
    "createdAt": "2026-01-01T10:00:00Z"
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 文件大小超过限制、读取上传文件失败 |

> 相同内容的文件会通过 SHA256 哈希命中已有记录，返回已存在的媒体文件（秒传去重）。

### 2. 获取文件详情

#### 请求信息

- **接口地址**: `/api/media/get`
- **请求方式**: `POST`
- **权限要求**: `file:read`（editor及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文件ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/media/get \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 1
  }'
```

### 3. 文件列表

#### 请求信息

- **接口地址**: `/api/media/list`
- **请求方式**: `POST`
- **权限要求**: `file:read`（editor及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| folder | string | 否 | 按文件夹过滤 | 字符串 |
| mimeType | string | 否 | 按MIME类型前缀过滤 | 字符串，如 image |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/media/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "page": 1,
    "pageSize": 10,
    "mimeType": "image"
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.media | array | 是 | 媒体文件列表，按上传时间倒序 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

### 4. 删除文件

#### 请求信息

- **接口地址**: `/api/media/delete`
- **请求方式**: `POST`
- **权限要求**: `file:delete`（admin及以上，或文件上传者本人）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文件ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/media/delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "文件删除成功"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 403 | 没有删除此文件的权限 |
| 404 | 文件不存在 |
