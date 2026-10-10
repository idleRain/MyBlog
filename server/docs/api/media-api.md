# 媒体文件 API 文档

## 概述

媒体文件模块提供图片、文档等资源的上传、查看、管理与删除功能。文件存储于本地目录，支持 SHA256 哈希秒传去重。

## 存储说明

- **存储目录**: `server/configs/config.yaml` 中 `media.upload_dir` 配置，默认 `uploads`（相对服务运行目录）
- **访问前缀**: `media.base_url` 配置，默认 `/uploads`
- **静态访问**: 服务启动时按 `media.base_url` 挂载静态目录（`internal/router/router.go` 的 `engine.Static`），`fileUrl` 可直接公开访问且目录列表关闭；生产部署经 `deploy/nginx.conf` 的 `location /uploads/` 反代到该静态服务
- **单文件上限**: `media.max_size_mb` 配置，默认 10MB
- **允许类型**: `media.allowed_types` 白名单，默认 `image/jpeg`、`image/png`、`image/gif`、`image/webp`、`application/pdf`；校验对象是内容嗅探结果（`http.DetectContentType`），客户端声明的文件名与 Content-Type 均不参与判定，白名单为空表示不限制
- **存储扩展名**: 命中白名单映射的上述 5 种类型按嗅探结果推导扩展名，其余类型沿用原始扩展名；`svg`、`html` 等可承载脚本的类型不在映射内
- **文件命名**: 上传后以 UUID 重命名，按年月分目录归档
- **目录创建**: 上传目录由上传动作按 `0755` 创建；就绪探针 `pkg/storage.CheckWritable` 以真实写入探测目录可写性

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 上传文件 | `file:upload` | editor及以上（editor/admin/superadmin 持有） |
| 查看文件列表/详情 | `file:read` | editor及以上 |
| 删除文件 | `file:delete` | 仅 admin/superadmin 持有该权限；服务层再校验非管理员只能删除本人上传的文件 |

> 列表与详情的可见范围为本人上传的文件，管理员可见全部；越权访问按 404 返回，与文件不存在同响应。

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
| data.storedName | string | 是 | 存储文件名（UUID + 存储扩展名） |
| data.filePath | string | 是 | 文件存储路径，由 `filepath.Join` 拼接，Windows 下为反斜杠分隔 |
| data.fileUrl | string | 是 | 文件访问URL，恒为正斜杠 |
| data.thumbnailUrl | string | 是 | 缩略图URL，上传时不生成，为空串 |
| data.mimeType | string | 是 | MIME类型，取自内容嗅探结果 |
| data.fileSize | integer | 是 | 文件大小（字节） |
| data.fileHash | string | 是 | SHA256哈希值，64位十六进制 |
| data.width / data.height | integer | 是 | 图片宽高，上传时不解析，输出 null |
| data.durationSeconds | integer | 是 | 音视频时长，非媒体文件为 0 |
| data.altText | string | 是 | 替代文本，上传时为空串 |
| data.status | string | 是 | 文件状态，上传后为 active |
| data.processedAt | string | 是 | 后处理完成时间，上传后输出 null |
| data.uploaderId | integer | 是 | 上传者用户ID |
| data.storageType | string | 是 | 存储类型，local表示本地 |
| data.folder | string | 是 | 文件夹分类，上传时为空串 |
| data.usageCount | integer | 是 | 被正文引用次数 |
| data.downloadCount | integer | 是 | 累计下载次数 |
| data.isPublic | boolean | 是 | 是否公开访问，上传后恒为 true |
| data.createdAt | string | 是 | 上传时间 |
| data.updatedAt | string | 是 | 更新时间 |
| data.uploader | object | 否 | 上传者窄化视图，仅普通上传路径输出 |

> `data.uploader` 为 `domain.UploaderPublic` 窄化视图，仅含 `id`、`username`、`nickname`、`avatar`；
> 秒传去重路径复用 `GetByFileHash` 查询结果，未预加载上传者关联，该键整体省略。
> 上传者 IP `uploadIP` 在实体上为 `json:"-"`，任何响应都不输出。

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
    "fileHash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
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
| 400 | 表单缺少 file 字段、打开或读取上传文件失败、文件大小超过限制、文件类型不在白名单、创建目录或写入文件失败、保存媒体记录失败 |
| 401 | 未登录或令牌无效 |

> 相同内容的文件会通过 SHA256 哈希命中已有记录，返回已存在的媒体文件（秒传去重）；
> 哈希比对先于类型白名单校验，白名单调整后旧文件仍可按哈希命中。

### 2. 获取文件详情

#### 请求信息

- **接口地址**: `/api/media/get`
- **请求方式**: `POST`
- **权限要求**: `file:read`（editor及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文件ID | 必填且非零 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/media/get \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 1
  }'
```

#### 响应参数

响应 `data` 为单个媒体文件对象，字段与上传接口一致；上传者关联已预加载，`data.uploader` 按 `domain.UploaderPublic` 白名单输出。

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 404 | 文件不存在，或非管理员访问他人上传的文件（与不存在同响应，防止经 ID 枚举探测） |

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
| folder | string | 否 | 按文件夹精确匹配过滤 | 无校验 |
| mimeType | string | 否 | 按MIME类型前缀过滤 | 无校验，如 image |

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
| data.page | integer | 是 | 当前页码，默认1 |
| data.pageSize | integer | 是 | 每页数量，默认10 |

> 非管理员仅返回本人上传的文件；列表查询未预加载上传者关联，列表项不含 `uploader` 键。
> 列表项的字段集合与上传响应一致（`uploadIP` 始终不输出）。

### 4. 删除文件

#### 请求信息

- **接口地址**: `/api/media/delete`
- **请求方式**: `POST`
- **权限要求**: `file:delete`（`config.yaml` 中仅 admin/superadmin 持有；非 admin 操作者还须是文件上传者本人）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文件ID | 必填且非零 |

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
| 403 | 非管理员且非文件上传者本人 |
| 404 | 文件不存在 |

> 删除为数据库软删除；物理文件同步移除，移除失败只落告警日志，不阻塞删除结果。
