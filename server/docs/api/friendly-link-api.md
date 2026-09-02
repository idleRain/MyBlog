# 友情链接 API 文档

## 概述

友情链接模块提供站点互链申请、审核与展示的完整生命周期管理，支持 `pending/active/hidden/rejected` 状态机。

## 链接状态说明

| 状态 | 说明 |
|------|------|
| pending | 待审核 |
| active | 展示中，对外可见 |
| hidden | 已隐藏 |
| rejected | 已拒绝 |

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 查看展示中的友情链接 | 无 | 无 |
| 创建 / 更新 / 删除 / 审核 | `system:config` | admin及以上 |
| 友情链接列表 | `system:config` | admin及以上 |

## 公开接口（无需认证）

### 1. 获取展示中的友情链接

返回全部 `active` 状态的友情链接，按排序权重排列。

#### 请求信息

- **接口地址**: `/api/friendly-links/list`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/friendly-links/list \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.links | array | 是 | 展示中的友情链接列表 |
| data.links[].id | integer | 是 | 链接ID |
| data.links[].name | string | 是 | 站点名称 |
| data.links[].url | string | 是 | 站点URL |
| data.links[].logo | string | 否 | 站点图标URL |
| data.links[].description | string | 否 | 站点简介 |
| data.links[].sortOrder | integer | 是 | 展示排序权重 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "links": [
      {
        "id": 1,
        "name": "示例站点",
        "url": "https://example.com",
        "sortOrder": 0
      }
    ]
  }
}
```

## 管理接口（需要 system:config 权限）

管理接口需在请求头携带 `Authorization: Bearer {accessToken}`，操作者角色为 admin 及以上。

### 2. 创建友情链接

新链接默认进入 `pending` 待审核状态。

#### 请求信息

- **接口地址**: `/api/admin/friendly-links/create`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| name | string | 是 | 站点名称 | 1-50字符 |
| url | string | 是 | 站点URL，全局唯一 | 最大255字符 |
| logo | string | 否 | 站点图标URL | 最大500字符 |
| description | string | 否 | 站点简介 | 最大255字符 |
| contactEmail | string | 否 | 站长联系邮箱 | 邮箱格式，最大100字符 |
| sortOrder | integer | 否 | 排序权重 | 整数，默认0 |
| isReciprocal | boolean | 否 | 是否已确认回链 | 默认false |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/friendly-links/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "name": "示例站点",
    "url": "https://example.com"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "友情链接创建成功",
  "data": {
    "id": 1,
    "name": "示例站点",
    "url": "https://example.com",
    "status": "pending",
    "sortOrder": 0,
    "isReciprocal": false
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 该站点 URL 已存在 |

### 3. 更新友情链接

#### 请求信息

- **接口地址**: `/api/admin/friendly-links/update`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 链接ID | 大于0的整数 |
| name | string | 否 | 站点名称，省略保留原值 | 1-50字符 |
| url | string | 否 | 站点URL，省略保留原值 | 最大255字符 |
| logo | string | 否 | 图标URL，省略保留原值 | 最大500字符 |
| description | string | 否 | 简介，省略保留原值 | 最大255字符 |
| contactEmail | string | 否 | 邮箱，省略保留原值 | 邮箱格式 |
| sortOrder | integer | 否 | 排序权重，省略保留原值 | 整数 |
| isReciprocal | boolean | 否 | 回链状态，省略保留原值 | 布尔值 |

> 更新接口遵循「省略字段保留原值」约定。

### 4. 删除友情链接

#### 请求信息

- **接口地址**: `/api/admin/friendly-links/delete`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 链接ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/friendly-links/delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 3
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "友情链接删除成功"
}
```

### 5. 状态流转操作

以下接口的请求参数与响应结构一致，仅操作不同：

| 接口地址 | 说明 |
|----------|------|
| `/api/admin/friendly-links/approve` | 审核通过，状态置为 active |
| `/api/admin/friendly-links/hide` | 下架，状态置为 hidden |
| `/api/admin/friendly-links/reject` | 拒绝，状态置为 rejected |

#### 请求参数（通用）

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 链接ID | 大于0的整数 |

#### 请求示例（审核通过）

```bash
curl -X POST http://localhost:3000/api/admin/friendly-links/approve \
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
  "message": "审核通过"
}
```

### 6. 友情链接列表

#### 请求信息

- **接口地址**: `/api/admin/friendly-links/list`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| status | string | 否 | 按状态过滤 | pending/active/hidden/rejected |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/friendly-links/list \
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
| data.links | array | 是 | 友情链接列表（含全量状态） |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
