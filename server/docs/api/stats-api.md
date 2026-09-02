# 站点统计 API 文档

## 概述

站点统计模块提供博客整体运营数据的概览与趋势分析，供管理后台展示。数据来源于内容统计表聚合与业务表实时计数。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 站点概览 | `system:stats` | admin及以上 |
| 浏览量趋势 | `system:stats` | admin及以上 |

## 管理接口（需要 system:stats 权限）

管理接口需在请求头携带 `Authorization: Bearer {accessToken}`，操作者角色为 admin 及以上。

### 1. 站点概览

聚合返回站点整体运营数据。

#### 请求信息

- **接口地址**: `/api/admin/stats/overview`
- **请求方式**: `POST`
- **权限要求**: `system:stats`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/stats/overview \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data | object | 是 | 站点概览数据 |
| data.articleCount | integer | 是 | 文章总数 |
| data.publishedCount | integer | 是 | 已发布文章数 |
| data.totalViews | integer | 是 | 文章总浏览量 |
| data.totalLikes | integer | 是 | 文章总点赞数 |
| data.commentCount | integer | 是 | 评论总数 |
| data.userCount | integer | 是 | 用户总数 |
| data.categoryCount | integer | 是 | 分类总数 |
| data.tagCount | integer | 是 | 标签总数 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "articleCount": 8,
    "publishedCount": 5,
    "totalViews": 1234,
    "totalLikes": 56,
    "commentCount": 30,
    "userCount": 3,
    "categoryCount": 4,
    "tagCount": 5
  }
}
```

### 2. 文章浏览量趋势

返回指定天数内的文章浏览量趋势，缺失日期补零。

#### 请求信息

- **接口地址**: `/api/admin/stats/articles`
- **请求方式**: `POST`
- **权限要求**: `system:stats`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| days | integer | 否 | 统计天数 | 1-90，默认7 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/stats/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "days": 7
  }'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data | object | 是 | 趋势数据 |
| data.dates | array | 是 | 日期列表，格式 YYYY-MM-DD |
| data.values | array | 是 | 与日期对应的浏览量数组 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "dates": ["2026-08-27", "2026-08-28", "2026-08-29", "2026-08-30", "2026-08-31", "2026-09-01", "2026-09-02"],
    "values": [0, 0, 0, 0, 12, 45, 30]
  }
}
```
