# 用户关注 API 文档

## 概述

用户关注模块提供用户间的关注关系管理，支持关注、取消关注、粉丝列表与关注列表查询。关注关系通过复合唯一索引防重复，通过检查约束阻止自我关注。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 关注 / 取消关注 | 登录 | user及以上 |
| 粉丝列表 / 关注列表 | 无 | 无 |

> 关注与取消关注接口需在请求头携带 `Authorization: Bearer {accessToken}`；粉丝与关注列表公开可查。

## 认证接口（需要登录）

### 1. 关注用户

#### 请求信息

- **接口地址**: `/api/users/follow`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| followingId | integer | 是 | 被关注用户ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/follow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "followingId": 2
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "关注成功"
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 不能关注自己、目标用户不存在 |

> 关注操作依赖 `(follower_id, following_id)` 唯一索引防重复，重复关注保持幂等。

### 2. 取消关注

#### 请求信息

- **接口地址**: `/api/users/unfollow`
- **请求方式**: `POST`
- **权限要求**: 登录（user及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| followingId | integer | 是 | 被取消关注的用户ID | 大于0的整数 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/unfollow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "followingId": 2
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "取消关注成功"
}
```

## 公开接口（无需认证）

### 3. 粉丝列表

分页查询指定用户的粉丝。

#### 请求信息

- **接口地址**: `/api/users/followers`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| userId | integer | 是 | 目标用户ID | 大于0的整数 |
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/followers \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.follows | array | 是 | 关注关系列表，按关注时间倒序 |
| data.follows[].id | integer | 是 | 关注关系ID |
| data.follows[].followerId | integer | 是 | 粉丝用户ID |
| data.follows[].follower | object | 是 | 粉丝用户信息 |
| data.follows[].follower.id | integer | 是 | 粉丝ID |
| data.follows[].follower.username | string | 是 | 粉丝用户名 |
| data.follows[].follower.nickname | string | 是 | 粉丝昵称 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "follows": [
      {
        "id": 10,
        "followerId": 2,
        "follower": {
          "id": 2,
          "username": "user2",
          "nickname": "用户二"
        }
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 4. 关注列表

分页查询指定用户关注的对象。

#### 请求信息

- **接口地址**: `/api/users/following`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| userId | integer | 是 | 目标用户ID | 大于0的整数 |
| page | integer | 否 | 页码 | 最小1，默认1 |
| pageSize | integer | 否 | 每页数量 | 1-100，默认10 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/users/following \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1,
    "page": 1,
    "pageSize": 10
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.follows | array | 是 | 关注关系列表 |
| data.follows[].followingId | integer | 是 | 被关注用户ID |
| data.follows[].following | object | 是 | 被关注用户信息 |
| data.follows[].following.id | integer | 是 | 用户ID |
| data.follows[].following.username | string | 是 | 用户名 |
| data.follows[].following.nickname | string | 是 | 昵称 |
| data.total | integer | 是 | 总记录数 |
| data.page | integer | 是 | 当前页码 |
| data.pageSize | integer | 是 | 每页数量 |
