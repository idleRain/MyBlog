# 文章管理 API 文档

## 概述

文章管理模块提供文章的完整生命周期管理，包括创建、发布、编辑、删除等功能，支持分类、标签、搜索、统计等高级特性。

## 内容多语言说明

语言协商、`Content-Language` 标注与 `i18n` 写入的通用规则以 [`contracts/i18n-protocol.md`](../../../contracts/i18n-protocol.md) 为唯一权威，本模块的事实如下：

- 输出语言作用于 `title`、`summary`、`content`、`contentHtml`、`seoTitle`、`seoDescription`、`seoKeywords`，缺失翻译的字段按字段回退中文；
- 具体语言请求会清空文章对象的 `translations` 并装配 `translationLocales`（存在有效翻译内容的语言列表，无翻译内容时该字段省略）；`Accept-Language: *` 时单资源响应保留 `translations` 全量翻译行并装配 `translationLocales`，列表与归档响应保留各项 `translations`，但不装配 `translationLocales`；
- 创建与更新请求体可选 `i18n` 字段按语言提交翻译补丁（`{"en": {"title": "...", "content": "..."}}`），键必须为白名单内的非缺省语言，缺省语言键与白名单外语言键返回 400；
- 翻译正文随保存渲染 `contentHtml` 并统计 `wordCount`，阅读时长按当前语言字数估算；
- 全文搜索同时匹配主表与翻译表，任一语言命中即纳入结果；
- 修订快照暂不包含翻译内容。

## 文章状态说明

| 状态 | 说明 | 权限要求 |
|------|------|----------|
| draft | 草稿 | 作者和管理员可查看 |
| published | 已发布 | 所有人可查看 |
| archived | 已归档 | 作者和管理员可查看 |
| private | 私有 | 作者和管理员可查看 |

状态流转接口写入的时间字段：发布写入 `published_at`（覆盖既有值），归档写入 `archived_at`；取消发布与设为私有只改状态，不清空既有发布时间。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 查看已发布文章 | 无 | 无 |
| 创建文章 | `article:create` | editor及以上 |
| 编辑自己的文章 | `article:create` | editor及以上 |
| 管理所有文章 | `article:manage` | admin及以上 |
| 点赞收藏 | 登录 | user及以上 |

> 上表为阅读用的文章模块口径，权限标识与角色映射的唯一权威是后端 `configs/config.yaml` 的 `rbac` 节：`article:manage` 当前仅 superadmin 与 admin 持有，`article:create` 由 superadmin、admin、editor 持有。

## 公开接口（可选认证）

> 公开接口携带有效令牌时，服务端注入登录者身份并按角色决定可见范围；未携带或令牌无效时按游客处理。令牌为双轨通道：`Authorization: Bearer {accessToken}` 头优先，其次读取会话 Cookie `mb_access_token`。文章读接口与列表的"按角色可见"逻辑由服务端统一判定。

### 1. 获取文章详情

根据文章ID获取文章详细信息。

#### 请求信息

- **接口地址**: `/api/articles/get`
- **请求方式**: `POST`
- **权限要求**: 可选认证，未发布文章仅作者或具备 `article:manage` 的管理员可见
- **错误分档**: 文章不存在返回 404；未发布文章对无权查看者返回 403，不回落为 404
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
| code | integer | 是 | 响应码，200表示成功 |
| message | string | 是 | 响应消息，成功时固定为 `操作成功` |
| data | object | 是 | 文章对象 |
| data.id | integer | 是 | 文章ID |
| data.title | string | 是 | 文章标题 |
| data.slug | string | 是 | 文章别名 |
| data.summary | string | 是 | 文章摘要，请求未提供时由正文自动截取 |
| data.content | string | 是 | 文章内容，Markdown 源文本 |
| data.contentHtml | string | 是 | 渲染后的 HTML 内容缓存，写入时生成，存量数据读取时按需补渲染 |
| data.coverImage | string | 是 | 封面图片URL |
| data.authorId | integer | 是 | 作者ID |
| data.author | object | 是 | 作者公开信息，窄化输出且仅含 `id`、`username`、`nickname`、`avatar`、`bio`、`website` |
| data.categoryId | integer | 否 | 主分类ID，未设置时为 null |
| data.category | object | 否 | 主分类对象，完整字段见[分类管理 API](./category-api.md) |
| data.categories | array | 否 | 全部关联分类，无关联时字段省略 |
| data.tags | array | 否 | 标签列表，无标签时字段省略 |
| data.status | string | 是 | 文章状态，draft/published/archived/private |
| data.originType | string | 是 | 来源类型，original/translation/reprint |
| data.sourceUrl | string | 是 | 原文链接，原创文章为空串 |
| data.sourceAuthor | string | 是 | 原文作者，原创文章为空串 |
| data.isFeatured | boolean | 是 | 是否推荐 |
| data.isTop | boolean | 是 | 是否置顶 |
| data.commentEnabled | boolean | 是 | 是否允许评论 |
| data.viewCount | integer | 是 | 浏览量 |
| data.likeCount | integer | 是 | 点赞数 |
| data.bookmarkCount | integer | 是 | 收藏数 |
| data.commentCount | integer | 是 | 评论数 |
| data.wordCount | integer | 是 | 字数统计，按 Unicode 字符数 |
| data.readingTime | integer | 是 | 预计阅读时间（分钟），最小为 1 |
| data.version | integer | 是 | 内容版本号，当前无递增写入路径，恒为 1 |
| data.seoTitle | string | 是 | SEO标题 |
| data.seoDescription | string | 是 | SEO描述 |
| data.seoKeywords | string | 是 | SEO关键词 |
| data.scheduledAt | string | 否 | 定时发布时间，未设置为 null |
| data.publishedAt | string | 否 | 发布时间，未发布为 null |
| data.editedAt | string | 否 | 正文最后编辑时间，当前无写入路径，恒为 null |
| data.archivedAt | string | 否 | 归档时间，未归档为 null |
| data.lastCommentAt | string | 否 | 最新评论时间，无评论为 null |
| data.createdAt | string | 是 | 创建时间 |
| data.updatedAt | string | 是 | 更新时间 |
| data.translations | array | 否 | 全量翻译行数组，仅 `Accept-Language: *` 返回；行字段为 `id`、`articleId`、`locale`、`title`、`summary`、`content`、`contentHtml`、`wordCount`、`seoTitle`、`seoDescription`、`seoKeywords`、`createdAt`、`updatedAt` |
| data.translationLocales | array | 否 | 存在有效翻译内容的语言列表，无翻译内容时字段省略 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "title": "Hello World",
    "slug": "hello-world",
    "summary": "这是我的第一篇文章",
    "content": "# Hello World\n\n这是文章内容...",
    "contentHtml": "<h1>Hello World</h1>\n<p>这是文章内容...</p>",
    "coverImage": "https://example.com/cover.jpg",
    "authorId": 1,
    "categoryId": 1,
    "status": "published",
    "originType": "original",
    "sourceUrl": "",
    "sourceAuthor": "",
    "isFeatured": true,
    "isTop": false,
    "commentEnabled": true,
    "viewCount": 100,
    "likeCount": 5,
    "bookmarkCount": 2,
    "commentCount": 3,
    "wordCount": 1500,
    "readingTime": 8,
    "version": 1,
    "seoTitle": "Hello World - 我的博客",
    "seoDescription": "这是我的第一篇文章的SEO描述",
    "seoKeywords": "Hello,World,博客",
    "scheduledAt": null,
    "publishedAt": "2024-01-01T10:00:00Z",
    "editedAt": null,
    "archivedAt": null,
    "lastCommentAt": null,
    "createdAt": "2024-01-01T09:30:00Z",
    "updatedAt": "2024-01-01T10:00:00Z",
    "author": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "avatar": "",
      "bio": "",
      "website": ""
    },
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
    ]
  }
}
```

> 该文章无翻译内容，故响应不含 `translationLocales` 与 `translations`；`category`、`categories`、`tags` 为节选字段。

---

### 2. 根据Slug获取文章

使用友好URL获取文章信息。

#### 请求信息

- **接口地址**: `/api/articles/getBySlug`
- **请求方式**: `POST`
- **权限要求**: 可选认证，未发布文章仅作者或具备 `article:manage` 的管理员可见
- **错误分档**: slug 不存在返回 404；未发布文章对无权查看者返回 403，不回落为 404
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
- **权限要求**: 可选认证；具备 `article:manage` 的管理员可按任意状态筛选，其他登录角色可见范围为已发布文章加本人全部状态文章，游客仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |
| status | string | 否 | 状态筛选 | draft/published/archived/private；非管理角色的筛选叠加在本人可见边界上，如 draft 仅命中本人草稿 |
| authorId | integer | 否 | 作者ID筛选 | 大于0的整数 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count，缺省 created_at |
| order | string | 否 | 排序方向 | asc/desc，缺省 desc |
| search | string | 否 | 关键词模糊筛选 | 对标题、正文、摘要做 LIKE 匹配，非全文索引 |

> `page` 与 `pageSize` 由 `binding` 约束，请求体缺省这两个字段时零值不通过校验，直接返回 400。

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
| code | integer | 是 | 响应码，200表示成功 |
| message | string | 是 | 响应消息，成功时固定为 `操作成功` |
| data | object | 是 | 响应数据 |
| data.articles | array | 是 | 文章列表，项字段见"获取文章详情"的响应参数表，但不含 `categories` 关联 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
          "nickname": "管理员",
          "avatar": "",
          "bio": "",
          "website": ""
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

> 示例中的列表项省略了 `content`、`contentHtml` 等长文本与无翻译内容时的 `translationLocales` 字段，字段全集见"获取文章详情"的响应参数表。

---

### 4. 获取作者文章列表

获取指定作者的文章列表，可见性规则与"获取文章列表"接口一致。

#### 请求信息

- **接口地址**: `/api/articles/byAuthor`
- **请求方式**: `POST`
- **权限要求**: 可选认证；具备 `article:manage` 的管理员可按任意状态筛选，作者本人查看自己的文章时可见全部状态，其余查看者仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| authorId | integer | 是 | 作者ID | 大于0的整数 |
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |
| status | string | 否 | 状态筛选 | draft/published/archived/private；管理员与作者本人（查看自己时）生效，其余角色服务端强制 published |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count，缺省 created_at |
| order | string | 否 | 排序方向 | asc/desc，缺省 desc |
| search | string | 否 | 关键词模糊筛选 | 对标题、正文、摘要做 LIKE 匹配 |

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
- **权限要求**: 无需认证，仅返回已发布文章（管理员身份也不放宽）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| categoryId | integer | 是 | 分类ID | 大于0的整数 |
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count，缺省 created_at |
| order | string | 否 | 排序方向 | asc/desc，缺省 desc |
| search | string | 否 | 关键词模糊筛选 | 对标题、正文、摘要做 LIKE 匹配 |

> 命中主分类或任一关联分类的文章均计入该分类。请求体复用列表参数结构，`status` 与 `authorId` 被服务端忽略。

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
- **权限要求**: 无需认证，仅返回已发布文章（管理员身份也不放宽）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| tagId | integer | 是 | 标签ID | 大于0的整数 |
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count，缺省 created_at |
| order | string | 否 | 排序方向 | asc/desc，缺省 desc |
| search | string | 否 | 关键词模糊筛选 | 对标题、正文、摘要做 LIKE 匹配 |

> 请求体复用列表参数结构，`status` 与 `authorId` 被服务端忽略。

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

全文搜索文章，基于 ngram 全文索引匹配标题、内容与摘要。ngram 按双字切分中文，单个汉字的关键词无法命中，建议输入至少两个字符。

#### 请求信息

- **接口地址**: `/api/articles/search`
- **请求方式**: `POST`
- **权限要求**: 无需认证，仅返回已发布文章（管理员身份也不放宽）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| keyword | string | 是 | 搜索关键词 | 非空字符串 |
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |
| sortBy | string | 否 | 排序字段 | created_at/updated_at/published_at/view_count/like_count，缺省 created_at |
| order | string | 否 | 排序方向 | asc/desc，缺省 desc |

> 请求体复用列表参数结构，`status`、`authorId`、`search` 在本接口被服务端忽略。每次搜索按关键词、结果数、耗时与成败写入搜索日志，日志写入失败不影响结果返回。

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
- **权限要求**: 无需认证，仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| limit | integer | 否 | 返回数量 | 缺省或小于等于 0 时取 10，无上限校验 |

> 排序为浏览量降序、点赞数降序。

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
  "message": "操作成功",
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
- **权限要求**: 无需认证，仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| limit | integer | 否 | 返回数量 | 缺省或小于等于 0 时取 10，无上限校验 |

> 按 `published_at` 倒序，未写入发布时间的记录排在末尾。

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
- **权限要求**: 无需认证，仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |
| limit | integer | 否 | 返回数量 | 缺省或小于等于 0 时取 5，无上限校验 |

> 候选按浏览量降序，优先取同主分类文章，不足时以同标签文章补齐，且排除当前文章本身；`id` 不存在时返回 404。

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

服务端在递增文章浏览计数的同时写入浏览明细与日统计：明细按文章、访客与日期三元组去重，同一访客当日重复浏览仅累加当日次数；访客标识缺失时回落到客户端 IP。日统计按 `content_stats` 表的日浏览维度累加，供管理端浏览量趋势图使用。

#### 请求信息

- **接口地址**: `/api/articles/view`
- **请求方式**: `POST`
- **权限要求**: 无需认证，接口不校验文章可见性，未发布文章同样计入浏览
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

> 访客标识经可选请求头 `Visitor-ID` 传入。

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
  "message": "操作成功",
  "data": {
    "message": "浏览记录成功"
  }
}
```

---

## 认证用户接口（需要登录）

> 点赞、取消点赞、收藏与取消收藏均为幂等：重复调用不报错，点赞数与收藏数只在真正新增或删除记录时增减。

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

> 文章不存在返回 404；对当前用户不可见的文章返回 403。首次点赞会向作者写入站内通知，作者本人点赞不通知。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/like \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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

> 文章不存在返回 404；对当前用户不可见的文章返回 403。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/bookmark \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "message": "取消收藏成功"
  }
}
```

---

## 文章操作与查询接口

> 创建、更新、删除与状态流转接口统一挂在 `/api/articles` 下，路由层统一挂 `article:create` 权限中间件，业务授权由服务端判定：作者（具备 `article:create`）可操作自己的文章，具备 `article:manage` 的管理员可操作任意文章。前端不再按角色切换接口。
> 本节内 `articles/archives` 为公开接口；`articles/isLiked`、`articles/isBookmarked`、`articles/bookmarks` 仅需登录；其余接口需要 `article:create` 权限。

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
| slug | string | 否 | 文章别名 | 长度0-200字符，为空时按标题生成；标题过滤后无可用字符时回退为 `article-<时间戳>-<随机数>`，与既有值冲突时自动追加 `-1`、`-2` 后缀 |
| summary | string | 否 | 文章摘要 | 长度0-500字符，为空时由正文自动截取 |
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
| i18n | object | 否 | 按语言提交翻译补丁 | 键为白名单内的非缺省语言（如 `en`），值为对象；可用字段 `title`、`summary`、`content`、`seoTitle`、`seoDescription`、`seoKeywords`，长度校验同主字段；缺省语言键与白名单外语言键返回 400 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
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
    "seoKeywords": "Go,语言,编程",
    "i18n": {
      "en": {
        "title": "My New Article",
        "content": "# My New Article\n\nContent..."
      }
    }
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
| slug | string | 否 | 文章别名 | 长度0-200字符，未传则保留原值，传空串时按标题重新生成 |
| summary | string | 否 | 文章摘要 | 长度0-500字符，未传则保留原值 |
| content | string | 是 | 文章内容 | 非空字符串，每次更新须整体提交 |
| coverImage | string | 否 | 封面图片URL | 长度0-500字符，未传则保留原值 |
| categoryId | integer | 否 | 主分类ID | 大于0的整数；未传时清空主分类，与字符串字段的保留语义不同 |
| categoryIds | array | 否 | 分类ID列表 | 整数数组；未传或传空数组时保持既有分类关联不变 |
| tagIds | array | 否 | 标签ID列表 | 整数数组；未传或传空数组时保持既有标签关联不变 |
| status | string | 否 | 文章状态 | draft/published/archived/private，未传则保留原状态 |
| isFeatured | boolean | 否 | 是否推荐 | 布尔值，未传则保留原值 |
| isTop | boolean | 否 | 是否置顶 | 布尔值，未传则保留原值 |
| commentEnabled | boolean | 否 | 是否允许评论 | 布尔值，未传则保留原值 |
| seoTitle | string | 否 | SEO标题 | 长度0-100字符，未传则保留原值 |
| seoDescription | string | 否 | SEO描述 | 长度0-255字符，未传则保留原值 |
| seoKeywords | string | 否 | SEO关键词 | 长度0-200字符，未传则保留原值 |
| i18n | object | 否 | 按语言提交翻译补丁 | 键为白名单内的非缺省语言（如 `en`）；补丁字段提供即更新，未提供的字段保留既有翻译值；缺省语言键与白名单外语言键返回 400 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1,
    "title": "更新后的文章标题",
    "content": "# 更新后的内容\n\n这是更新后的文章内容...",
    "status": "published",
    "i18n": {
      "en": {
        "title": "Updated Title"
      }
    }
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
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
  -H "Authorization: Bearer 3f2a9c1e4b7d8056a1c3e5f70b2d4689" \
  -d '{
    "id": 1
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "message": "文章设置为私有成功"
  }
}
```

---

### 23. 获取文章归档（公开）

获取全部已发布文章的归档数据，按年份与月份两级分组，供归档页时间线使用。

#### 请求信息

- **接口地址**: `/api/articles/archives`
- **请求方式**: `POST`
- **权限要求**: 无需认证，仅返回已发布文章
- **Content-Type**: `application/json`

#### 请求参数

无请求参数，请求体传空对象。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/articles/archives \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 响应码，200表示成功 |
| message | string | 是 | 响应消息，成功时固定为 `操作成功` |
| data | array | 是 | 年份分组列表，按年份倒序 |
| data[].year | integer | 是 | 年份 |
| data[].total | integer | 是 | 该年文章总数 |
| data[].months | array | 是 | 月份分组列表，按月份倒序 |
| data[].months[].month | integer | 是 | 月份，1-12 |
| data[].months[].articles | array | 是 | 该月已发布文章，按发布时间倒序 |

> 归档为最小列集查询，文章项只有 `id`、`title`、`slug`、`summary`、`publishedAt`、`createdAt` 携带数据，其余字段为零值且不输出 `author`、`category`、`tags` 等关联。

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": [
    {
      "year": 2026,
      "total": 2,
      "months": [
        {
          "month": 2,
          "articles": [
            {
              "id": 1,
              "title": "Hello World",
              "slug": "hello-world",
              "publishedAt": "2026-02-03T10:00:00Z"
            }
          ]
        },
        {
          "month": 1,
          "articles": [
            {
              "id": 2,
              "title": "设计令牌先行",
              "slug": "design-tokens-first",
              "publishedAt": "2026-01-18T10:00:00Z"
            }
          ]
        }
      ]
    }
  ]
}
```

---

### 24. 查询点赞状态

查询当前用户对指定文章的点赞状态。

#### 请求信息

- **接口地址**: `/api/articles/isLiked`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "isLiked": true
  }
}
```

---

### 25. 查询收藏状态

查询当前用户对指定文章的收藏状态。

#### 请求信息

- **接口地址**: `/api/articles/isBookmarked`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 文章ID | 大于0的整数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "isBookmarked": false
  }
}
```

---

### 26. 我的收藏列表

分页查询当前用户收藏的文章，按收藏时间倒序。

#### 请求信息

- **接口地址**: `/api/articles/bookmarks`
- **请求方式**: `POST`
- **权限要求**: 需要登录
- **Content-Type**: `application/json`
- **Authorization**: `Bearer {accessToken}`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 是 | 页码 | `binding:"min=1"`，必须显式传入且不小于 1 |
| pageSize | integer | 是 | 每页数量 | `binding:"min=1,max=100"`，必须显式传入，取值 1-100 |

> 结果按收藏时间倒序，不按文章状态过滤：收藏后文章状态变为草稿、私有或归档，仍会出现在本列表中；列表项包含 `categories` 关联。

#### 响应示例

响应格式同"获取文章列表"接口。

---

## 错误响应

### 常见错误码

业务错误经 `pkg/response` 信封返回，**HTTP 状态码恒为 200**，错误语义只在响应体 `code` 字段中。

| 响应码(code) | 错误信息 | 说明 |
|--------|----------|------|
| 400 | `参数错误: <校验详情>` 或服务层业务校验文案 | `binding` 校验失败或 service 层业务规则不通过（如 `i18n` 语言键非法） |
| 401 | `未提供认证令牌` / `无效的认证令牌` | 认证中间件拦截，令牌缺失或已失效 |
| 403 | `权限不足，无法访问该资源` / `没有执行此操作的权限：<动作>` | 权限中间件拦截，或服务层判定无权操作该文章 |
| 404 | `文章不存在` | 指定的文章ID或 slug 不存在、已软删除 |
| 500 | `服务器内部错误` | 未分类的服务端异常，原始错误只入日志不外泄 |

说明：

- 未发布文章（draft/archived/private）对无权查看者返回 403 而非 404，避免接口语义与权限判定混淆；
- slug 冲突不会报错，系统会自动追加 `-1`、`-2` 等后缀保证唯一。

### 错误响应示例

```json
{
  "code": 404,
  "message": "文章不存在"
}
```

```json
{
  "code": 403,
  "message": "没有执行此操作的权限：编辑此文章"
}
```