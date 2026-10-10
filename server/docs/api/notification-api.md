# 通知 API 文档

## 概述

通知模块提供站内消息中心功能，支持评论回复、文章点赞、用户关注、系统通知等类型的消息查看与已读管理。所有接口仅能操作当前登录用户本人的通知。

通知由业务链路自动产生：回复注册用户的评论产生 `comment_reply`，首次点赞文章或评论产生 `article_like` 或 `comment_like`，关注用户产生 `follow`。系统自动跳过自触达场景，即作者点赞自己的文章、用户回复自己的评论不产生通知；游客行为无法定位账号时不产生通知。

## 通知类型说明

| 类型 | 说明 |
|------|------|
| comment_reply | 评论回复 |
| article_like | 文章点赞 |
| comment_like | 评论点赞 |
| system | 系统通知 |
| follow | 用户关注 |
| article_new | 新文章发布 |

> 上表与列表接口 `type` 过滤白名单一致。`article_new` 已完成类型常量与过滤白名单登记，当前业务链路尚无写入点。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 通知列表 / 未读数 | 登录 | user及以上 |
| 标记已读 | 登录 | user及以上 |

> 所有通知接口需要已登录身份，`Authorization: Bearer {accessToken}` 头与会话 Cookie `mb_access_token` 双轨任一即可（Header 优先），且仅能操作本人通知。

## 接口列表（均需登录）

### 1. 通知列表

分页查询当前用户的通知，附带未读数。

#### 请求信息

- **接口地址**: `/api/notifications/list`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |
| type | string | 否 | 按通知类型过滤 | comment_reply/article_like/comment_like/system/follow/article_new |
| isRead | boolean | 否 | 按已读状态过滤 | 布尔值 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/list \
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
| data.notifications | array | 是 | 通知列表，按创建时间倒序 |
| data.notifications[].id | integer | 是 | 通知ID |
| data.notifications[].userId | integer | 是 | 接收用户ID，恒为当前登录用户 |
| data.notifications[].senderId | integer | 是 | 触发用户ID，系统通知为 `null` |
| data.notifications[].type | string | 是 | 通知类型 |
| data.notifications[].title | string | 是 | 通知标题 |
| data.notifications[].content | string | 是 | 通知内容，无内容时为 `null` |
| data.notifications[].actionUrl | string | 是 | 点击跳转地址，无跳转时为空字符串 |
| data.notifications[].relatedType | string | 是 | 关联资源类型（`article`/`comment`/`user`），无关联时为 `null` |
| data.notifications[].relatedId | integer | 是 | 关联资源ID，无关联时为 `null` |
| data.notifications[].isRead | boolean | 是 | 是否已读 |
| data.notifications[].readAt | string | 是 | 已读时间，未已读时为 `null` |
| data.notifications[].createdAt | string | 是 | 创建时间，RFC3339 格式 |
| data.notifications[].updatedAt | string | 是 | 更新时间，RFC3339 格式 |
| data.notifications[].sender | object | 否 | 触发用户实体快照，`senderId` 为空时不返回该字段 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
| data.unreadCount | integer | 是 | 未读通知总数，不受 `type` 与 `isRead` 过滤条件影响 |

> `sender` 为按 `preload` 关联出的用户实体序列化结果，字段集与用户实体一致（含 `email`、`role`、`status` 等），仅 `password` 等 `json:"-"` 字段不输出。
> `readAt` 为模型字段，标记已读的实现只更新 `isRead`，当前不会写入 `read_at`，该字段始终为 `null`。

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | `pageSize` 超过100，或 `type` 不在白名单内 |
| 401 | 未提供或无效的认证令牌 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "notifications": [
      {
        "id": 1,
        "userId": 1,
        "senderId": 2,
        "type": "comment_reply",
        "title": "张三回复了你的评论",
        "content": null,
        "actionUrl": "/blog/hello-world#comment-8",
        "relatedType": "comment",
        "relatedId": 8,
        "isRead": false,
        "readAt": null,
        "createdAt": "2026-09-02T10:00:00Z",
        "updatedAt": "2026-09-02T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10,
    "unreadCount": 1
  }
}
```

### 2. 获取未读数

获取当前用户的未读通知数量。

#### 请求信息

- **接口地址**: `/api/notifications/unreadCount`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/unreadCount \
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
    "unreadCount": 3
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 401 | 未提供或无效的认证令牌 |

### 3. 标记单条通知已读

将指定通知标记为已读，仅本人通知可操作。

#### 请求信息

- **接口地址**: `/api/notifications/read`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| id | integer | 是 | 通知ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/read \
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
  "message": "通知已标记为已读"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 缺少 `id` 或值为0 |
| 401 | 未提供或无效的认证令牌 |
| 404 | 通知不存在或不属于当前用户 |

### 4. 标记全部通知已读

将当前用户的全部未读通知标记为已读，仅更新 `isRead`，不写入 `readAt`。

#### 请求信息

- **接口地址**: `/api/notifications/readAll`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/readAll \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{}'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "全部通知已标记为已读"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 401 | 未提供或无效的认证令牌 |

该接口幂等：无未读通知时同样返回成功。单条已读与全部已读接口均不返回 `data` 字段。
