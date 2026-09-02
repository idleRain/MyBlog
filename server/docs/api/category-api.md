# 分类管理 API 文档

## 概述

分类管理模块提供文章分类的树形管理与展示功能。分类采用 `parent_id`、`root_id`、`level`、`path` 四字段描述层级结构，支持子树查询与排序。

## 分类树结构说明

| 字段 | 类型 | 说明 |
|------|------|------|
| parentId | integer | 父分类 ID，顶级分类为空 |
| rootId | integer | 根分类 ID，用于整棵子树的聚合查询 |
| level | integer | 分类层级，顶级为 1 |
| path | string | 物化路径，形如 `/1/5/12` |

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 获取分类详情 / 分类树 | 无 | 无 |
| 创建 / 更新 / 删除分类 | `category:manage` | admin及以上 |
| 分类列表 | `category:manage` | admin及以上 |

## 公开接口（无需认证）

### 1. 获取分类详情

根据分类ID获取分类信息。

#### 请求信息

- **接口地址**: `/api/categories/get`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 分类ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/categories/get \
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
| data | object | 是 | 分类信息 |
| data.id | integer | 是 | 分类ID |
| data.name | string | 是 | 分类名称 |
| data.slug | string | 是 | URL友好标识 |
| data.description | string | 否 | 分类描述 |
| data.coverImage | string | 否 | 封面图URL |
| data.parentId | integer | 否 | 父分类ID |
| data.rootId | integer | 否 | 根分类ID |
| data.level | integer | 是 | 分类层级 |
| data.path | string | 是 | 物化路径 |
| data.sortOrder | integer | 是 | 排序权重 |
| data.status | integer | 是 | 状态，1显示 0隐藏 |
| data.articleCount | integer | 是 | 文章数量 |
| data.isFeatured | boolean | 是 | 是否精选分类 |
| data.createdAt | string | 是 | 创建时间 |
| data.updatedAt | string | 是 | 更新时间 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "name": "技术分享",
    "slug": "tech",
    "description": "技术文章分享",
    "level": 1,
    "path": "/1",
    "sortOrder": 0,
    "status": 1,
    "articleCount": 3,
    "isFeatured": false,
    "createdAt": "2026-01-01T10:00:00Z",
    "updatedAt": "2026-01-01T10:00:00Z"
  }
}
```

### 2. 获取分类树

获取完整的分类树形结构，按 sortOrder 排序。

#### 请求信息

- **接口地址**: `/api/categories/tree`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/categories/tree \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data | object | 是 | 响应数据 |
| data.tree | array | 是 | 根分类节点数组 |
| data.tree[].id | integer | 是 | 分类ID |
| data.tree[].name | string | 是 | 分类名称 |
| data.tree[].children | array | 否 | 子分类节点数组 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "tree": [
      {
        "id": 1,
        "name": "技术分享",
        "children": [
          {
            "id": 2,
            "name": "后端开发",
            "children": []
          }
        ]
      }
    ]
  }
}
```

## 管理接口（需要 category:manage 权限）

管理接口需在请求头携带 `Authorization: Bearer {accessToken}`，且操作者角色为 admin 及以上。

### 3. 创建分类

#### 请求信息

- **接口地址**: `/api/admin/categories/create`
- **请求方式**: `POST`
- **权限要求**: `category:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| name | string | 是 | 分类名称 | 1-50字符 |
| slug | string | 否 | URL友好标识，省略时按名称自动生成 | 最大50字符 |
| description | string | 否 | 分类描述 | 最大1000字符 |
| coverImage | string | 否 | 封面图URL | 最大255字符 |
| parentId | integer | 否 | 父分类ID，顶级分类不传 | 正整数 |
| sortOrder | integer | 否 | 排序权重，数值小靠前 | 整数，默认0 |
| status | integer | 否 | 分类状态 | 0或1，默认1显示 |
| isFeatured | boolean | 否 | 是否精选分类 | 默认false |
| seoTitle | string | 否 | SEO标题 | 最大100字符 |
| seoDescription | string | 否 | SEO描述 | 最大255字符 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/categories/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "name": "后端开发",
    "parentId": 1
  }'
```

#### 响应参数

成功时返回创建的完整分类对象，字段同「获取分类详情」。

#### 响应示例

```json
{
  "code": 200,
  "message": "分类创建成功",
  "data": {
    "id": 2,
    "name": "后端开发",
    "slug": "backend",
    "parentId": 1,
    "rootId": 1,
    "level": 2,
    "path": "/1/2",
    "sortOrder": 0,
    "status": 1,
    "articleCount": 0
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 父分类不存在、请求参数错误 |

### 4. 更新分类

#### 请求信息

- **接口地址**: `/api/admin/categories/update`
- **请求方式**: `POST`
- **权限要求**: `category:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 分类ID | 大于0的整数 |
| name | string | 否 | 分类名称，省略保留原值 | 1-50字符 |
| slug | string | 否 | URL友好标识，省略保留原值 | 最大50字符 |
| description | string | 否 | 分类描述，省略保留原值 | 最大1000字符 |
| coverImage | string | 否 | 封面图URL，省略保留原值 | 最大255字符 |
| sortOrder | integer | 否 | 排序权重，省略保留原值 | 整数 |
| status | integer | 否 | 分类状态，省略保留原值 | 0或1 |
| isFeatured | boolean | 否 | 是否精选，省略保留原值 | 布尔值 |
| seoTitle | string | 否 | SEO标题，省略保留原值 | 最大100字符 |
| seoDescription | string | 否 | SEO描述，省略保留原值 | 最大255字符 |

> 更新接口遵循「省略字段保留原值」约定，仅更新显式传入的字段。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/categories/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 2,
    "name": "后端开发进阶"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "分类更新成功",
  "data": {
    "id": 2,
    "name": "后端开发进阶",
    "slug": "backend"
  }
}
```

### 5. 删除分类

#### 请求信息

- **接口地址**: `/api/admin/categories/delete`
- **请求方式**: `POST`
- **权限要求**: `category:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 分类ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/categories/delete \
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
  "message": "分类删除成功"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 分类下存在子分类，无法删除 |
| 404 | 分类不存在 |

### 6. 分类列表

#### 请求信息

- **接口地址**: `/api/admin/categories/list`
- **请求方式**: `POST`
- **权限要求**: `category:manage`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| status | integer | 否 | 按状态过滤 | 0或1 |
| search | string | 否 | 名称或描述模糊搜索 | 字符串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/categories/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "page": 1,
    "pageSize": 10,
    "search": "技术"
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.categories | array | 是 | 分类列表 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "categories": [
      {
        "id": 1,
        "name": "技术分享",
        "level": 1,
        "path": "/1",
        "status": 1,
        "articleCount": 3
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```
