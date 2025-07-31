# 文章管理 API 文档

## 概述

文章管理模块提供文章的完整生命周期管理，包括创建、发布、编辑、删除等功能，支持分类、标签、搜索、统计等高级特性。

## 文章状态说明

| 状态 | 说明 | 权限要求 |
|------|------|----------|
| draft | 草稿 | 作者和管理员可查看 |
| published | 已发布 | 所有人可查看 |
| archived | 已归档 | 作者和管理员可查看 |
| private | 私有 | 仅作者可查看 |

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 查看已发布文章 | 无 | 无 |
| 创建文章 | `article:create` | editor及以上 |
| 编辑自己的文章 | `article:create` | editor及以上 |
| 管理所有文章 | `article:manage` | admin及以上 |
| 点赞收藏 | 登录 | user及以上 |

## 公开接口（无需认证）

### 1. 获取文章详情

根据文章ID获取文章详细信息。

#### 请求信息

- **接口地址**: `/api/articles/get`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/get \
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
| data | object | 是 | 文章信息 |
| data.id | integer | 是 | 文章ID |
| data.title | string | 是 | 文章标题 |
| data.slug | string | 是 | 文章别名 |
| data.summary | string | 是 | 文章摘要 |
| data.content | string | 是 | 文章内容 |
| data.coverImage | string | 是 | 封面图片URL |
| data.authorId | integer | 是 | 作者ID |
| data.author | object | 是 | 作者信息 |
| data.author.id | integer | 是 | 作者ID |
| data.author.username | string | 是 | 作者用户名 |
| data.author.nickname | string | 是 | 作者昵称 |
| data.categoryId | integer | 否 | 主分类ID |
| data.category | object | 否 | 主分类信息 |
| data.categories | array | 是 | 所有分类列表 |
| data.tags | array | 是 | 标签列表 |
| data.status | string | 是 | 文章状态 |
| data.isFeatured | boolean | 是 | 是否推荐 |
| data.isTop | boolean | 是 | 是否置顶 |
| data.commentEnabled | boolean | 是 | 是否允许评论 |
| data.viewCount | integer | 是 | 浏览量 |
| data.likeCount | integer | 是 | 点赞数 |
| data.commentCount | integer | 是 | 评论数 |
| data.wordCount | integer | 是 | 字数统计 |
| data.readingTime | integer | 是 | 预计阅读时间（分钟） |
| data.seoTitle | string | 是 | SEO标题 |
| data.seoDescription | string | 是 | SEO描述 |
| data.seoKeywords | string | 是 | SEO关键词 |
| data.publishedAt | string | 否 | 发布时间 |
| data.createdAt | string | 是 | 创建时间 |
| data.updatedAt | string | 是 | 更新时间 |

#### 响应示例

```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1,
    "title": "Hello World",
    "slug": "hello-world",
    "summary": "这是我的第一篇文章",
    "content": "# Hello World\n\n这是文章内容...",
    "coverImage": "https://example.com/cover.jpg",
    "authorId": 1,
    "author": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员"
    },
    "categoryId": 1,
    "category": {
      "id": 1,
      "name": "技术分享",
      "slug": "tech"
    },
    "categories": [
      {
        "id": 1,
        "name": "技术分享",
        "slug": "tech"
      }
    ],
    "tags": [
      {
        "id": 1,
        "name": "Go语言",
        "slug": "golang",
        "color": "#00ADD8"
      }
    ],
    "status": "published",
    "isFeatured": true,
    "isTop": false,
    "commentEnabled": true,
    "viewCount": 100,
    "likeCount": 5,
    "commentCount": 3,
    "wordCount": 1500,
    "readingTime": 8,
    "seoTitle": "Hello World - 我的博客",
    "seoDescription": "这是我的第一篇文章的SEO描述",
    "seoKeywords": "Hello,World,博客",
    "publishedAt": "2024-01-01T10:00:00Z",
    "createdAt": "2024-01-01T09:30:00Z",
    "updatedAt": "2024-01-01T10:00:00Z"
  }
}
```

---

### 2. 根据Slug获取文章

使用友好URL获取文章信息。

#### 请求信息

- **接口地址**: `/api/articles/getBySlug`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| slug | string | 是 | 文章别名 | 非空字符串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/getBySlug \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "hello-world"
  }'
```

#### 响应示例

响应格式同"获取文章详情"接口。

---

### 3. 获取文章列表

分页获取文章列表，支持多种筛选条件。

#### 请求信息

- **接口地址**: `/api/articles/list`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| status | string | 否 | 状态筛选 | draft/published/archived/private |
| authorId | integer | 否 | 作者ID筛选 | 大于0的整数 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count |
| order | string | 否 | 排序方向 | asc/desc，默认desc |
| search | string | 否 | 搜索关键词 | 搜索标题、内容、摘要 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/list \
  -H "Content-Type: application/json" \
  -d '{
    "page": 1,
    "pageSize": 10,
    "status": "published",
    "sortBy": "created_at",
    "order": "desc"
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.articles | array | 是 | 文章列表 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

#### 响应示例

```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "articles": [
      {
        "id": 1,
        "title": "Hello World",
        "slug": "hello-world",
        "summary": "这是我的第一篇文章",
        "coverImage": "https://example.com/cover.jpg",
        "author": {
          "id": 1,
          "username": "admin",
          "nickname": "管理员"
        },
        "category": {
          "id": 1,
          "name": "技术分享"
        },
        "tags": [
          {
            "id": 1,
            "name": "Go语言",
            "color": "#00ADD8"
          }
        ],
        "status": "published",
        "isFeatured": true,
        "viewCount": 100,
        "likeCount": 5,
        "commentCount": 3,
        "readingTime": 8,
        "publishedAt": "2024-01-01T10:00:00Z",
        "createdAt": "2024-01-01T09:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

---

### 4. 获取作者文章列表

获取指定作者的文章列表。

#### 请求信息

- **接口地址**: `/api/articles/byAuthor`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| authorId | integer | 是 | 作者ID | 大于0的整数 |
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count |
| order | string | 否 | 排序方向 | asc/desc，默认desc |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/byAuthor \
  -H "Content-Type: application/json" \
  -d '{
    "authorId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应示例

响应格式同"获取文章列表"接口。

---

### 5. 获取分类文章列表

获取指定分类的文章列表。

#### 请求信息

- **接口地址**: `/api/articles/byCategory`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| categoryId | integer | 是 | 分类ID | 大于0的整数 |
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count |
| order | string | 否 | 排序方向 | asc/desc，默认desc |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/byCategory \
  -H "Content-Type: application/json" \
  -d '{
    "categoryId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应示例

响应格式同"获取文章列表"接口。

---

### 6. 获取标签文章列表

获取指定标签的文章列表。

#### 请求信息

- **接口地址**: `/api/articles/byTag`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| tagId | integer | 是 | 标签ID | 大于0的整数 |
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count |
| order | string | 否 | 排序方向 | asc/desc，默认desc |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/byTag \
  -H "Content-Type: application/json" \
  -d '{
    "tagId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应示例

响应格式同"获取文章列表"接口。

---

### 7. 搜索文章

全文搜索文章。

#### 请求信息

- **接口地址**: `/api/articles/search`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| keyword | string | 是 | 搜索关键词 | 非空字符串 |
| page | integer | 否 | 页码 | 大于0的整数，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100之间，默认10 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count |
| order | string | 否 | 排序方向 | asc/desc，默认desc |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/search \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "Go语言",
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应示例

响应格式同"获取文章列表"接口。

---

### 8. 获取热门文章

获取浏览量和点赞数最高的文章。

#### 请求信息

- **接口地址**: `/api/articles/popular`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| limit | integer | 否 | 返回数量 | 1-50之间，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/popular \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 5
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "articles": [
      {
        "id": 1,
        "title": "Hello World",
        "slug": "hello-world",
        "summary": "这是我的第一篇文章",
        "viewCount": 100,
        "likeCount": 5,
        "publishedAt": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

---

### 9. 获取最新文章

获取最近发布的文章。

#### 请求信息

- **接口地址**: `/api/articles/recent`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| limit | integer | 否 | 返回数量 | 1-50之间，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/recent \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 5
  }'
```

#### 响应示例

响应格式同"获取热门文章"接口。

---

### 10. 获取相关文章

根据分类和标签获取相关文章。

#### 请求信息

- **接口地址**: `/api/articles/related`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |
| limit | integer | 否 | 返回数量 | 1-20之间，默认5 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/related \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "limit": 5
  }'
```

#### 响应示例

响应格式同"获取热门文章"接口。

---

### 11. 记录文章浏览

记录文章浏览量，支持防重复统计。

#### 请求信息

- **接口地址**: `/api/articles/view`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/view \
  -H "Content-Type: application/json" \
  -H "Visitor-ID: unique-visitor-id" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "浏览记录成功",
  "data": {
    "message": "浏览记录成功"
  }
}
```

---

## 认证用户接口（需要登录）

### 12. 点赞文章

为文章点赞。

#### 请求信息

- **接口地址**: `/api/articles/like`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/like \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "点赞成功",
  "data": {
    "message": "点赞成功"
  }
}
```

---

### 13. 取消点赞文章

取消对文章的点赞。

#### 请求信息

- **接口地址**: `/api/articles/unlike`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/unlike \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "取消点赞成功",
  "data": {
    "message": "取消点赞成功"
  }
}
```

---

### 14. 收藏文章

收藏文章到个人收藏夹。

#### 请求信息

- **接口地址**: `/api/articles/bookmark`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/bookmark \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "收藏成功",
  "data": {
    "message": "收藏成功"
  }
}
```

---

### 15. 取消收藏文章

从个人收藏夹中移除文章。

#### 请求信息

- **接口地址**: `/api/articles/unbookmark`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/unbookmark \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "取消收藏成功",
  "data": {
    "message": "取消收藏成功"
  }
}
```

---

## 编辑者接口（需要编辑权限）

### 16. 创建文章

创建新文章。

#### 请求信息

- **接口地址**: `/api/articles/create`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| title | string | 是 | 文章标题 | 长度1-200字符 |
| slug | string | 否 | 文章别名 | 长度0-200字符，为空时自动生成 |
| summary | string | 否 | 文章摘要 | 长度0-500字符 |
| content | string | 是 | 文章内容 | 非空字符串 |
| coverImage | string | 否 | 封面图片URL | 长度0-500字符 |
| categoryId | integer | 否 | 主分类ID | 大于0的整数 |
| categoryIds | array | 否 | 分类ID列表 | 整数数组 |
| tagIds | array | 否 | 标签ID列表 | 整数数组 |
| status | string | 否 | 文章状态 | draft/published/private，默认draft |
| isFeatured | boolean | 否 | 是否推荐 | 默认false |
| isTop | boolean | 否 | 是否置顶 | 默认false |
| commentEnabled | boolean | 否 | 是否允许评论 | 默认true |
| seoTitle | string | 否 | SEO标题 | 长度0-100字符 |
| seoDescription | string | 否 | SEO描述 | 长度0-255字符 |
| seoKeywords | string | 否 | SEO关键词 | 长度0-200字符 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "title": "我的新文章",
    "slug": "my-new-article",
    "summary": "这是一篇关于Go语言的文章",
    "content": "# 我的新文章\n\n这里是文章内容...",
    "coverImage": "https://example.com/cover.jpg",
    "categoryId": 1,
    "tagIds": [1, 2],
    "status": "draft",
    "isFeatured": false,
    "commentEnabled": true,
    "seoTitle": "我的新文章 - 技术博客",
    "seoDescription": "一篇关于Go语言的深度分析文章",
    "seoKeywords": "Go,语言,编程"
  }'
```

#### 响应示例

响应格式同"获取文章详情"接口，返回新创建的文章信息。

---

### 17. 更新文章

更新文章信息。

#### 请求信息

- **接口地址**: `/api/articles/update`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |
| title | string | 是 | 文章标题 | 长度1-200字符 |
| slug | string | 否 | 文章别名 | 长度0-200字符 |
| summary | string | 否 | 文章摘要 | 长度0-500字符 |
| content | string | 是 | 文章内容 | 非空字符串 |
| coverImage | string | 否 | 封面图片URL | 长度0-500字符 |
| categoryId | integer | 否 | 主分类ID | 大于0的整数 |
| categoryIds | array | 否 | 分类ID列表 | 整数数组 |
| tagIds | array | 否 | 标签ID列表 | 整数数组 |
| status | string | 否 | 文章状态 | draft/published/archived/private |
| isFeatured | boolean | 否 | 是否推荐 | 布尔值 |
| isTop | boolean | 否 | 是否置顶 | 布尔值 |
| commentEnabled | boolean | 否 | 是否允许评论 | 布尔值 |
| seoTitle | string | 否 | SEO标题 | 长度0-100字符 |
| seoDescription | string | 否 | SEO描述 | 长度0-255字符 |
| seoKeywords | string | 否 | SEO关键词 | 长度0-200字符 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1,
    "title": "更新后的文章标题",
    "content": "# 更新后的内容\n\n这是更新后的文章内容...",
    "status": "published"
  }'
```

#### 响应示例

响应格式同"获取文章详情"接口，返回更新后的文章信息。

---

### 18. 删除文章

删除文章（软删除）。

#### 请求信息

- **接口地址**: `/api/articles/delete`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "文章删除成功",
  "data": {
    "message": "文章删除成功"
  }
}
```

---

### 19. 发布文章

将草稿文章发布。

#### 请求信息

- **接口地址**: `/api/articles/publish`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/publish \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "文章发布成功",
  "data": {
    "message": "文章发布成功"
  }
}
```

---

### 20. 取消发布文章

将已发布文章设为草稿。

#### 请求信息

- **接口地址**: `/api/articles/unpublish`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/unpublish \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "取消发布成功",
  "data": {
    "message": "取消发布成功"
  }
}
```

---

### 21. 归档文章

将文章设为归档状态。

#### 请求信息

- **接口地址**: `/api/articles/archive`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/archive \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "文章归档成功",
  "data": {
    "message": "文章归档成功"
  }
}
```

---

### 22. 设置文章为私有

将文章设为私有状态。

#### 请求信息

- **接口地址**: `/api/articles/private`
- **请求方式**: `POST`
- **权限要求**: `article:create` 权限（作者）或 `article:manage` 权限（管理员）
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/private \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "文章设置为私有成功",
  "data": {
    "message": "文章设置为私有成功"
  }
}
```

---

## 管理员接口（需要管理权限）

### 23. 管理员文章列表

获取所有状态的文章列表（管理员专用）。

#### 请求信息

- **接口地址**: `/api/admin/articles/list`
- **请求方式**: `POST`
- **权限要求**: `article:manage` 权限
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

参数同"获取文章列表"接口，但可以查看所有状态的文章。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/articles/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "page": 1,
    "pageSize": 10,
    "status": "draft"
  }'
```

#### 响应示例

响应格式同"获取文章列表"接口。

---

### 24-29. 管理员文章操作

管理员可以对任意文章执行以下操作：

- `POST /api/admin/articles/update` - 更新任意文章
- `POST /api/admin/articles/delete` - 删除任意文章  
- `POST /api/admin/articles/publish` - 发布任意文章
- `POST /api/admin/articles/unpublish` - 取消发布任意文章
- `POST /api/admin/articles/archive` - 归档任意文章
- `POST /api/admin/articles/private` - 设置任意文章为私有

请求参数和响应格式与对应的编辑者接口相同，区别在于管理员可以操作任意文章，不受作者限制。

---

## 错误响应

### 常见错误码

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 请求参数错误 | 参数格式或内容不正确 |
| 401 | 未提供认证令牌 / 无效的认证令牌 | 需要登录或令牌已过期 |
| 403 | 权限不足 | 当前角色没有操作权限 |
| 404 | 文章不存在 | 指定的文章ID不存在或已删除 |
| 409 | 文章别名已存在 | slug重复 |
| 500 | 服务器内部错误 | 服务器异常 |

### 错误响应示例

```json
{
  "code": 404,
  "message": "文章不存在",
  "error": "article not found"
}
```

```json
{
  "code": 403,
  "message": "没有编辑此文章的权限",
  "error": "insufficient permissions"
}
```