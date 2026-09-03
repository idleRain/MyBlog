# 标签管理 API 文档

## 概述

标签管理模块提供文章标签的创建、编辑、删除与查询功能，支持热门标签与使用次数统计。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 获取标签详情 / 热门标签 | 无 | 无 |
| 全部标签列表（文章编辑选择） | `article:read` | 登录，editor及以上 |
| 创建 / 更新 / 删除标签 | `tag:manage` | admin及以上 |
| 标签列表 | `tag:manage` | admin及以上 |

## 公开接口（无需认证）

### 1. 获取标签详情

根据标签ID获取标签信息。

#### 请求信息

- **接口地址**: `/api/tags/get`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 标签ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/tags/get \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 标签信息 |
| data.id | integer | 是 | 标签ID |
| data.name | string | 是 | 标签名称 |
| data.slug | string | 是 | URL友好标识 |
| data.color | string | 是 | 标签颜色，HEX格式 |
| data.description | string | 否 | 标签描述 |
| data.status | integer | 是 | 状态，1启用 0隐藏 |
| data.usageCount | integer | 是 | 使用次数 |
| data.isHot | boolean | 是 | 是否热门标签 |
| data.createdAt | string | 是 | 创建时间 |
| data.updatedAt | string | 是 | 更新时间 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "name": "Go语言",
    "slug": "go",
    "color": "#808080",
    "status": 1,
    "usageCount": 3,
    "isHot": false,
    "createdAt": "2026-01-01T10:00:00Z",
    "updatedAt": "2026-01-01T10:00:00Z"
  }
}
```

### 2. 获取热门标签

获取使用次数最多的标签，按使用次数倒序排列。

#### 请求信息

- **接口地址**: `/api/tags/popular`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| limit | integer | 否 | 返回数量 | 1-50，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/tags/popular \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 10
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "tags": [
      {
        "id": 1,
        "name": "Go语言",
        "usageCount": 3
      }
    ]
  }
}
```

## 登录可读接口（需要 article:read 权限）

> 供文章编辑时选择标签使用，返回全部标签（含启用与隐藏），按使用次数倒序排列。该接口是文章编辑选择的唯一标签列表来源，权限由服务端按登录者判定，前端不做角色分派。

### 3. 全部标签列表

#### 请求信息

- **接口地址**: `/api/tags/list`
- **请求方式**: `POST`
- **权限要求**: `article:read`（需登录）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

无。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/tags/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{}'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "tags": [
      {
        "id": 1,
        "name": "Go语言",
        "usageCount": 3
      },
      {
        "id": 2,
        "name": "Gin框架",
        "usageCount": 1
      }
    ]
  }
}
```

## 管理接口（需要 tag:manage 权限）

管理接口需在请求头携带 `Authorization: Bearer {accessToken}`，且操作者角色为 admin 及以上。

### 4. 创建标签

#### 请求信息

- **接口地址**: `/api/admin/tags/create`
- **请求方式**: `POST`
- **权限要求**: `tag:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| name | string | 是 | 标签名称，全局唯一 | 1-30字符 |
| slug | string | 否 | URL友好标识，省略时按名称生成 | 最大30字符 |
| color | string | 否 | 标签颜色，HEX格式 | 4-7字符，默认#808080 |
| description | string | 否 | 标签描述 | 最大200字符 |
| status | integer | 否 | 标签状态 | 0或1，默认1启用 |
| isHot | boolean | 否 | 是否热门标签 | 默认false |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/tags/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "name": "Gin框架",
    "color": "#00bcd4"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "标签创建成功",
  "data": {
    "id": 2,
    "name": "Gin框架",
    "slug": "gin",
    "color": "#00bcd4",
    "status": 1,
    "usageCount": 0
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 标签名称已存在 |

### 5. 更新标签

#### 请求信息

- **接口地址**: `/api/admin/tags/update`
- **请求方式**: `POST`
- **权限要求**: `tag:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 标签ID | 大于0的整数 |
| name | string | 否 | 标签名称，省略保留原值 | 1-30字符 |
| slug | string | 否 | URL友好标识，省略保留原值 | 最大30字符 |
| color | string | 否 | 标签颜色，省略保留原值 | 4-7字符 |
| description | string | 否 | 标签描述，省略保留原值 | 最大200字符 |
| status | integer | 否 | 标签状态，省略保留原值 | 0或1 |
| isHot | boolean | 否 | 是否热门，省略保留原值 | 布尔值 |

> 更新接口遵循「省略字段保留原值」约定，仅更新显式传入的字段。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/tags/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 2,
    "isHot": true
  }'
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 标签名称已存在 |
| 404 | 标签不存在 |

### 6. 删除标签

#### 请求信息

- **接口地址**: `/api/admin/tags/delete`
- **请求方式**: `POST`
- **权限要求**: `tag:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 标签ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/tags/delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 5
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "标签删除成功"
}
```

### 7. 标签列表

#### 请求信息

- **接口地址**: `/api/admin/tags/list`
- **请求方式**: `POST`
- **权限要求**: `tag:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| status | integer | 否 | 按状态过滤 | 0或1 |
| isHot | boolean | 否 | 是否仅返回热门标签 | 布尔值 |
| search | string | 否 | 名称或描述模糊搜索 | 字符串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/tags/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.tags | array | 是 | 标签列表，按使用次数倒序 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "tags": [
      {
        "id": 1,
        "name": "Go语言",
        "usageCount": 3
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```
