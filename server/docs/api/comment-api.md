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

> 新发表的评论默认进入 `pending` 待审核状态，仅 `approved` 状态评论对外展示。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 查看评论列表 | 无 | 无 |
| 发表评论 | 无（游客）/ 登录 | 游客或 user及以上 |
| 点赞 / 取消点赞评论 | 登录 | user及以上 |
| 评论审核（approve/reject/spam/trash/delete） | `comment:moderate` | admin及以上 |
| 管理端评论列表 | `comment:moderate` | admin及以上 |

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
| articleId | integer | 是 | 文章ID | 大于0的整数 |
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
| data.comments | array | 是 | 已审核通过的评论列表 |
| data.comments[].id | integer | 是 | 评论ID |
| data.comments[].articleId | integer | 是 | 文章ID |
| data.comments[].userId | integer | 否 | 评论用户ID，游客为空 |
| data.comments[].parentId | integer | 否 | 父评论ID |
| data.comments[].rootId | integer | 否 | 根评论ID |
| data.comments[].level | integer | 是 | 评论层级 |
| data.comments[].authorName | string | 是 | 评论者名称 |
| data.comments[].content | string | 是 | 评论内容 |
| data.comments[].likeCount | integer | 是 | 点赞数 |
| data.comments[].replyCount | integer | 是 | 回复数量 |
| data.comments[].isPinned | boolean | 是 | 是否置顶 |
| data.comments[].createdAt | string | 是 | 创建时间 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

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

支持注册用户与游客双通道。注册用户经认证自动绑定身份；游客需填写姓名。

#### 请求信息

- **接口地址**: `/api/comments/create`
- **请求方式**: `POST`
- **权限要求**: 无需认证（游客）/ 登录（可选，携带则绑定身份）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| articleId | integer | 是 | 文章ID | 大于0的整数 |
| parentId | integer | 否 | 父评论ID，回复评论时填写 | 正整数 |
| content | string | 是 | 评论内容 | 1-2000字符 |
| authorName | string | 游客必填 | 游客姓名 | 最大50字符 |
| authorEmail | string | 否 | 游客邮箱 | 邮箱格式，最大100字符 |
| authorWebsite | string | 否 | 游客网站 | 最大255字符 |

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
| 400 | 文章不允许评论、游客未填写姓名、父评论不存在或不属于该文章 |

## 认证接口（需要登录）

点赞接口需在请求头携带 `Authorization: Bearer {accessToken}`。

### 3. 点赞评论

#### 请求信息

- **接口地址**: `/api/comments/like`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 大于0的整数 |

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

> 点赞操作依赖 `(comment_id, user_id)` 唯一索引防重复，重复点赞保持幂等。

### 4. 取消点赞评论

#### 请求信息

- **接口地址**: `/api/comments/unlike`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 大于0的整数 |

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

## 管理接口（需要 comment:moderate 权限）

审核接口需在请求头携带 `Authorization: Bearer {accessToken}`，操作者角色为 admin 及以上。

### 5. 评论审核操作

以下接口的请求参数与响应结构一致，仅操作不同：

| 接口地址 | 说明 |
|----------|------|
| `/api/admin/comments/approve` | 审核通过，状态置为 approved |
| `/api/admin/comments/reject` | 拒绝评论，状态置为 rejected |
| `/api/admin/comments/spam` | 标记垃圾，状态置为 spam |
| `/api/admin/comments/trash` | 移入回收站，状态置为 trash |
| `/api/admin/comments/delete` | 删除评论（软删除） |

#### 请求参数（通用）

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 评论ID | 大于0的整数 |

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
- **权限要求**: `comment:moderate`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| status | string | 否 | 按状态过滤 | pending/approved/rejected/spam/trash |
| keyword | string | 否 | 内容或评论者名称模糊搜索 | 字符串 |

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
| data.comments | array | 是 | 评论列表（含全量状态） |
| data.comments[].status | string | 是 | 评论状态 |
| data.comments[].article | object | 否 | 关联文章信息 |
| data.comments[].user | object | 否 | 评论用户信息 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
