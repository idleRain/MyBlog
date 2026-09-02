# 通知 API 文档

## 概述

通知模块提供站内消息中心功能，支持评论回复、文章点赞、用户关注、系统通知等类型的消息查看与已读管理。所有接口仅能操作当前登录用户本人的通知。

## 通知类型说明

| 类型 | 说明 |
|------|------|
| comment_reply | 评论回复 |
| article_like | 文章点赞 |
| comment_like | 评论点赞 |
| system | 系统通知 |
| follow | 用户关注 |
| article_new | 新文章发布 |

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 通知列表 / 未读数 | 登录 | user及以上 |
| 标记已读 | 登录 | user及以上 |

> 所有通知接口需在请求头携带 `Authorization: Bearer {accessToken}`，仅能操作本人通知。

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
| data.notifications[].type | string | 是 | 通知类型 |
| data.notifications[].title | string | 是 | 通知标题 |
| data.notifications[].content | string | 否 | 通知内容 |
| data.notifications[].actionUrl | string | 否 | 点击跳转地址 |
| data.notifications[].isRead | boolean | 是 | 是否已读 |
| data.notifications[].readAt | string | 否 | 已读时间 |
| data.notifications[].createdAt | string | 是 | 创建时间 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
| data.unreadCount | integer | 是 | 未读通知总数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "notifications": [
      {
        "id": 1,
        "type": "comment_reply",
        "title": "张三回复了你的评论",
        "isRead": false,
        "createdAt": "2026-09-02T10:00:00Z"
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

- **接口地址**: `/api/notifications/unread-count`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/unread-count \
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
| 404 | 通知不存在或不属于当前用户 |

### 4. 标记全部通知已读

将当前用户的全部通知标记为已读。

#### 请求信息

- **接口地址**: `/api/notifications/read-all`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/notifications/read-all \
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
