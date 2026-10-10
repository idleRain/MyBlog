# 评论管理 API 文档

## 概述

评论管理模块提供文章的评论展示、发表、点赞与审核管理功能。评论采用 `parent_id`、`root_id`、`level` 描述两级评论树，支持注册用户与游客双通道发表。

## 评论状态说明

| 状态 | 说明 |
|------|------|
| pending | 待审核 |
| approved | 已通过，对外可见 |
| rejected | 已拒绝 |
| spam | 垃圾评论 |
| trash | 回收站 |

> 新发表的评论默认进入 `pending` 待审核状态，仅 `approved` 状态评论对外展示；
> 站点设置 `comment_auto_approve` 为 `true`（或 `1`）时，新评论直接写入 `approved`。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 查看评论列表 | 无 | 无需认证 |
| 发表评论 | 无（游客，受 `allow_guest_comment` 开关控制）/ 登录 | 游客或任意登录角色 |
| 点赞 / 取消点赞评论 | 登录 | 仅校验令牌，不校验角色 |
| 评论审核（approve/reject/spam/trash/delete） | `comment:moderate` | admin及以上（`config.yaml` 中 superadmin 与 admin 持有） |
| 管理端评论列表 | `comment:moderate` | 同上 |

## 公开接口

### 1. 获取文章评论列表

#### 请求信息

- **接口地址**: `/api/comments/list`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| articleId | integer | 是 | 文章ID | 必填且非零 |
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/comments/list \
  -H "Content-Type: application/json" \
  -d '{
    "articleId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.comments | array | 是 | 已审核通过的评论列表，按 `isPinned DESC, createdAt ASC` 排序 |
| data.comments[].id | integer | 是 | 评论ID |
| data.comments[].articleId | integer | 是 | 文章ID |
| data.comments[].userId | integer | 否 | 评论用户ID，游客评论为空 |
| data.comments[].parentId | integer | 否 | 父评论ID，根评论为空 |
| data.comments[].rootId | integer | 否 | 根评论ID |
| data.comments[].level | integer | 是 | 评论层级，根评论为 1，回复为 2 |
| data.comments[].authorName | string | 是 | 游客姓名；登录评论落库为空串，展示名取自 `user` |
| data.comments[].authorWebsite | string | 是 | 游客网站 |
| data.comments[].content | string | 是 | 评论内容，Markdown 格式 |
| data.comments[].contentHtml | string | 是 | 渲染后的 HTML 缓存 |
| data.comments[].status | string | 是 | 评论状态，公开列表恒为 approved |
| data.comments[].likeCount | integer | 是 | 点赞数 |
| data.comments[].replyCount | integer | 是 | 回复数量 |
| data.comments[].reportedCount | integer | 是 | 被举报次数 |
| data.comments[].isAuthor | boolean | 是 | 是否为文章作者回复 |
| data.comments[].isPinned | boolean | 是 | 是否置顶 |
| data.comments[].editedAt | string | 否 | 内容最后编辑时间 |
| data.comments[].createdAt | string | 是 | 创建时间 |
| data.comments[].updatedAt | string | 是 | 更新时间 |
| data.comments[].user | object | 否 | 评论者窄化视图，仅游客评论省略 |
| data.comments[].article | object | 是 | 文章关联对象；公开列表未预加载该关联，实际为零值对象 |
| data.comments[].parent / .root / .children | object | 否 | 树形关联，未预加载时省略 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码，默认1 |
| data.pageSize | integer | 是 | 每页数量，默认10 |

> `data.comments[].user` 为 `domain.AuthorPublic` 窄化视图，仅含 `id`、`username`、`nickname`、`avatar`、`bio`、`website`。
> 评论实体的游客邮箱 `authorEmail`、评论者 IP `authorIP`、`userAgent` 与 `deletedAt` 均为 `json:"-"`，任何响应都不输出。

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "comments": [
      {
        "id": 1,
        "articleId": 1,
        "level": 1,
        "authorName": "游客甲",
        "content": "写得很好，学习了！",
        "likeCount": 0,
        "replyCount": 0,
        "isPinned": false,
        "createdAt": "2026-01-01T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 2. 发表评论

支持注册用户与游客双通道。携带有效令牌时服务层绑定登录身份，并把 `authorName`、`authorEmail`、`authorWebsite` 三个游客字段清空后再落库；游客通道需填写姓名。

受站点设置控制：`allow_guest_comment` 关闭时游客通道被拒绝，仅登录用户可发言；`comment_auto_approve` 开启时新评论跳过待审核直接进入已审核状态。两项设置缺失时保持默认行为，即允许游客评论且需人工审核。

#### 请求信息

- **接口地址**: `/api/comments/create`
- **请求方式**: `POST`
- **权限要求**: 无需认证（游客）/ 登录（可选，携带则绑定身份）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| articleId | integer | 是 | 文章ID | 必填且非零 |
| parentId | integer | 否 | 父评论ID，回复评论时填写 | 无 binding 校验，取值需存在 |
| content | string | 是 | 评论内容 | 1-2000字符 |
| authorName | string | 游客必填 | 游客姓名，登录提交时被服务层清空 | 最大50字符，游客必填由服务层校验 |
| authorEmail | string | 否 | 游客邮箱，登录提交时被服务层清空 | 邮箱格式，最大100字符 |
| authorWebsite | string | 否 | 游客网站，登录提交时被服务层清空 | 最大255字符 |

> 创建成功返回的评论实体经 `GetByID` 重新查询输出，字段集合与公开列表一致；`article` 关联未预加载，为零值对象。

#### 请求示例（游客）

```bash
curl -X POST http://localhost:3000/api/comments/create \
  -H "Content-Type: application/json" \
  -d '{
    "articleId": 1,
    "content": "这是一条游客评论",
    "authorName": "游客甲"
  }'
```

#### 请求示例（登录用户，携带认证头）

```bash
curl -X POST http://localhost:3000/api/comments/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "articleId": 1,
    "content": "这是一条登录用户评论"
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "评论提交成功",
  "data": {
    "id": 2,
    "articleId": 1,
    "status": "pending",
    "level": 1,
    "authorName": "游客甲",
    "content": "这是一条游客评论",
    "createdAt": "2026-01-01T10:00:00Z"
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数校验失败、文章未开启评论或未发布、游客未填写姓名、父评论不属于该文章 |
| 403 | 站点已关闭游客评论，游客通道被拒绝 |
| 404 | 文章不存在、父评论不存在 |

## 认证接口（需要登录）

点赞接口需携带有效访问令牌：`Authorization: Bearer {accessToken}` 头优先，其次读取会话 Cookie `mb_access_token`。

### 3. 点赞评论

#### 请求信息

- **接口地址**: `/api/comments/like`
- **请求方式**: `POST`
- **权限要求**: 登录（仅校验令牌，不校验角色）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 必填且非零 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/comments/like \
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
  "message": "点赞成功"
}
```

> 点赞操作依赖 `(comment_id, user_id)` 唯一索引防重复，重复点赞保持幂等；
> 评论不存在时返回 404，首次点赞才产生点赞通知。

### 4. 取消点赞评论

#### 请求信息

- **接口地址**: `/api/comments/unlike`
- **请求方式**: `POST`
- **权限要求**: 登录（仅校验令牌，不校验角色）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 必填且非零 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/comments/unlike \
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
  "message": "取消点赞成功"
}
```

> 取消点赞直接删除点赞记录，评论不存在或本就未点赞时同样返回成功，与点赞接口的错误分档不对称。

## 管理接口（需要 comment:moderate 权限）

审核接口需携带有效访问令牌：`Authorization: Bearer {accessToken}` 头优先，其次读取会话 Cookie `mb_access_token`；`comment:moderate` 在 `config.yaml` 中由 superadmin 与 admin 持有。

### 5. 评论审核操作

以下接口的请求参数与响应结构一致，仅操作与成功消息不同：

| 接口地址 | 说明 | 成功响应 message |
|----------|------|------------------|
| `/api/admin/comments/approve` | 审核通过，状态置为 approved | 审核通过 |
| `/api/admin/comments/reject` | 拒绝评论，状态置为 rejected | 拒绝评论 |
| `/api/admin/comments/spam` | 标记垃圾，状态置为 spam | 标记垃圾评论 |
| `/api/admin/comments/trash` | 移入回收站，状态置为 trash | 移入回收站 |
| `/api/admin/comments/delete` | 删除评论（软删除，同时重算文章评论数并回退父评论回复数） | 评论删除成功 |

#### 请求参数（通用）

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 必填且非零 |

#### 请求示例（审核通过）

```bash
curl -X POST http://localhost:3000/api/admin/comments/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "id": 2
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "审核通过"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 404 | 评论不存在 |

### 6. 管理端评论列表

#### 请求信息

- **接口地址**: `/api/admin/comments/list`
- **请求方式**: `POST`
- **权限要求**: `comment:moderate`（superadmin 与 admin）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| status | string | 否 | 按状态过滤 | pending/approved/rejected/spam/trash |
| keyword | string | 否 | 模糊搜索 | 无校验，仅匹配 content 与 author_name 两列 |

> `keyword` 匹配 `author_name` 列，登录评论该列为空串，故无法按注册用户昵称检索。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/comments/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "page": 1,
    "pageSize": 10,
    "status": "pending"
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.comments | array | 是 | 评论列表（含全量状态），按 `createdAt DESC` 排序 |
| data.comments[].status | string | 是 | 评论状态 |
| data.comments[].article | object | 是 | 关联文章对象，管理端已预加载真实数据 |
| data.comments[].user | object | 否 | 评论者窄化视图，游客评论省略 |
| data.comments[].authorEmail | string | 是 | 游客邮箱，仅管理端恢复输出 |
| data.comments[].authorIP | string | 是 | 评论者 IP，仅管理端恢复输出 |
| data.comments[].userAgent | string | 是 | 请求 User-Agent，仅管理端恢复输出 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

> 管理端列表经 `service.AdminCommentView` 输出，字段集合为评论实体全部字段加 `authorEmail`、`authorIP`、`userAgent` 三个审计字段；
> 评论者仍只经 `user` 键以 `domain.AuthorPublic` 白名单输出；内嵌 `article` 未预加载作者关联，其 `author` 键省略。
