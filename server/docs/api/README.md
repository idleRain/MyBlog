# API 文档总览

## 概述

MyBlog 后端 API 提供完整的博客系统功能，覆盖用户管理、文章管理、分类标签、评论互动、媒体文件、系统设置、友情链接、站点统计、通知与关注等模块。所有接口遵循统一的设计规范，使用 POST 方法和 JSON 数据格式。

## 文档结构

### 核心规范
- [API 规范文档](./api-specification.md) - 统一的 API 设计规范和命名约定

### 功能模块

#### 系统监控
- [健康检查 API](./health-api.md) - 服务器状态检查接口

#### 用户与社交
- [用户管理 API](./user-api.md) - 用户登录、CRUD操作和权限管理
- [用户关注 API](./user-follow-api.md) - 关注、取消关注、粉丝与关注列表
- [通知 API](./notification-api.md) - 站内消息列表与已读管理

#### 内容管理
- [文章管理 API](./article-api.md) - 文章CRUD、搜索、分类标签关联等完整功能
- [分类管理 API](./category-api.md) - 分类树形管理与展示
- [标签管理 API](./tag-api.md) - 标签管理与热门标签
- [评论管理 API](./comment-api.md) - 评论展示、发表、点赞与审核

#### 资源与运营
- [媒体文件 API](./media-api.md) - 文件上传、查看、管理与删除
- [系统设置 API](./setting-api.md) - 站点配置的读取与更新
- [友情链接 API](./friendly-link-api.md) - 友链申请、审核与展示
- [站点统计 API](./stats-api.md) - 站点概览与浏览量趋势

## API 统计

| 模块 | 接口数量 | 说明 |
|------|----------|------|
| 健康检查 | 1 | 系统状态监控 |
| 用户管理 | 8 | 用户认证和管理 |
| 用户关注 | 4 | 关注关系管理 |
| 通知 | 4 | 站内消息中心 |
| 文章管理 | 29 | 文章内容管理 |
| 分类管理 | 6 | 分类树形管理 |
| 标签管理 | 6 | 标签与热门标签 |
| 评论管理 | 10 | 评论与审核 |
| 媒体文件 | 4 | 文件上传与管理 |
| 系统设置 | 3 | 站点配置管理 |
| 友情链接 | 8 | 友链申请与审核 |
| 站点统计 | 2 | 运营数据分析 |
| **总计** | **85** | **完整的博客系统 API** |

## 接口概览

### 认证相关
- `POST /api/users/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌
- `POST /api/auth/logout` - 用户登出

### 用户管理 (需要权限)
- `POST /api/users/create` - 创建用户
- `POST /api/users/get` - 获取用户信息
- `POST /api/users/update` - 更新用户信息
- `POST /api/users/delete` - 删除用户
- `POST /api/users/list` - 获取用户列表

### 用户关注
- `POST /api/users/follow` - 关注用户（登录）
- `POST /api/users/unfollow` - 取消关注（登录）
- `POST /api/users/followers` - 粉丝列表（公开）
- `POST /api/users/following` - 关注列表（公开）

### 通知 (需要登录)
- `POST /api/notifications/list` - 通知列表
- `POST /api/notifications/unread-count` - 未读数
- `POST /api/notifications/read` - 标记单条已读
- `POST /api/notifications/read-all` - 标记全部已读

### 文章管理
#### 公开接口
- `POST /api/articles/get` - 根据ID获取文章
- `POST /api/articles/getBySlug` - 根据Slug获取文章
- `POST /api/articles/list` - 获取文章列表
- `POST /api/articles/byAuthor` - 获取作者文章
- `POST /api/articles/byCategory` - 获取分类文章
- `POST /api/articles/byTag` - 获取标签文章
- `POST /api/articles/search` - 搜索文章
- `POST /api/articles/popular` - 获取热门文章
- `POST /api/articles/recent` - 获取最新文章
- `POST /api/articles/related` - 获取相关文章
- `POST /api/articles/view` - 记录浏览量

#### 认证接口
- `POST /api/articles/like` - 点赞文章
- `POST /api/articles/unlike` - 取消点赞
- `POST /api/articles/bookmark` - 收藏文章
- `POST /api/articles/unbookmark` - 取消收藏

#### 编辑权限接口
- `POST /api/articles/create` - 创建文章
- `POST /api/articles/update` - 更新文章
- `POST /api/articles/delete` - 删除文章
- `POST /api/articles/publish` - 发布文章
- `POST /api/articles/unpublish` - 取消发布
- `POST /api/articles/archive` - 归档文章
- `POST /api/articles/private` - 设为私有

#### 管理权限接口
- `POST /api/admin/articles/list` - 获取所有状态文章列表
- `POST /api/admin/articles/update` - 更新任意文章
- `POST /api/admin/articles/delete` - 删除任意文章
- `POST /api/admin/articles/publish` - 发布任意文章
- `POST /api/admin/articles/unpublish` - 取消发布任意文章
- `POST /api/admin/articles/archive` - 归档任意文章
- `POST /api/admin/articles/private` - 将任意文章设为私有

### 分类管理
- `POST /api/categories/get` - 获取分类详情（公开）
- `POST /api/categories/tree` - 获取分类树（公开）
- `POST /api/admin/categories/create` - 创建分类
- `POST /api/admin/categories/update` - 更新分类
- `POST /api/admin/categories/delete` - 删除分类
- `POST /api/admin/categories/list` - 分类列表

### 标签管理
- `POST /api/tags/get` - 获取标签详情（公开）
- `POST /api/tags/popular` - 获取热门标签（公开）
- `POST /api/admin/tags/create` - 创建标签
- `POST /api/admin/tags/update` - 更新标签
- `POST /api/admin/tags/delete` - 删除标签
- `POST /api/admin/tags/list` - 标签列表

### 评论管理
- `POST /api/comments/list` - 文章评论列表（公开）
- `POST /api/comments/create` - 发表评论（游客/登录）
- `POST /api/comments/like` - 点赞评论（登录）
- `POST /api/comments/unlike` - 取消点赞评论（登录）
- `POST /api/admin/comments/approve` - 审核通过
- `POST /api/admin/comments/reject` - 拒绝评论
- `POST /api/admin/comments/spam` - 标记垃圾
- `POST /api/admin/comments/trash` - 移入回收站
- `POST /api/admin/comments/delete` - 删除评论
- `POST /api/admin/comments/list` - 管理端评论列表

### 媒体文件 (需要权限)
- `POST /api/media/upload` - 上传文件（multipart）
- `POST /api/media/get` - 获取文件详情
- `POST /api/media/list` - 文件列表
- `POST /api/media/delete` - 删除文件

### 系统设置
- `POST /api/settings/public` - 公开设置（免鉴权）
- `POST /api/admin/settings/list` - 设置项列表
- `POST /api/admin/settings/update` - 批量更新设置

### 友情链接
- `POST /api/friendly-links/list` - 展示中的链接（公开）
- `POST /api/admin/friendly-links/create` - 创建链接
- `POST /api/admin/friendly-links/update` - 更新链接
- `POST /api/admin/friendly-links/delete` - 删除链接
- `POST /api/admin/friendly-links/approve` - 审核通过
- `POST /api/admin/friendly-links/hide` - 下架
- `POST /api/admin/friendly-links/reject` - 拒绝
- `POST /api/admin/friendly-links/list` - 链接列表

### 站点统计 (需要权限)
- `POST /api/admin/stats/overview` - 站点概览
- `POST /api/admin/stats/articles` - 浏览量趋势

### 系统监控
- `POST /api/health` - 健康检查

## 权限系统

### 角色层级
| 角色 | 权限级别 | 主要权限 |
|------|----------|----------|
| superadmin | 4 | 系统全部权限 |
| admin | 3 | 用户和内容管理 |
| editor | 2 | 文章创建和编辑 |
| user | 1 | 基础用户权限 |

### 权限模块
- **用户管理**: `user:create`, `user:read`, `user:update`, `user:delete`, `user:list`
- **文章管理**: `article:create`, `article:read`, `article:update`, `article:delete`, `article:list`, `article:publish`, `article:manage`
- **分类标签管理**: `category:manage`, `tag:manage`
- **评论管理**: `comment:create`, `comment:read`, `comment:update`, `comment:delete`, `comment:moderate`
- **文件管理**: `file:upload`, `file:read`, `file:delete`
- **系统管理**: `system:config`, `system:logs`, `system:stats`

## 请求规范

### 统一请求格式
- **请求方式**: POST
- **Content-Type**: application/json（媒体上传为 multipart/form-data）
- **认证头**: Authorization: Bearer {accessToken}

### 统一响应格式
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {}
}
```

## 开发指南

### 本地测试
```bash
# 启动开发服务器
pnpm run dev

# 健康检查
curl -X POST http://localhost:3000/api/health -H "Content-Type: application/json" -d '{}'
```

### 认证流程
1. 使用 `/api/users/login` 登录获取令牌
2. 在请求头中添加 `Authorization: Bearer {accessToken}`
3. 令牌过期时使用 `/api/auth/refresh` 刷新
4. 使用 `/api/auth/logout` 安全登出

### 错误处理
- `400` - 请求参数错误
- `401` - 认证失败或令牌过期
- `403` - 权限不足
- `404` - 资源不存在
- `500` - 服务器内部错误

## 更新日志

### v1.1.0 (当前版本)
- ✅ 分类管理与分类树
- ✅ 标签管理与热门标签
- ✅ 评论系统与审核状态机
- ✅ 媒体文件上传与管理
- ✅ 系统设置管理
- ✅ 友情链接申请与审核
- ✅ 站点统计概览与趋势
- ✅ 站内通知中心
- ✅ 用户关注关系

### v1.0.0
- ✅ 完整的用户认证和授权系统
- ✅ 基于RBAC的权限控制
- ✅ 完整的文章管理系统
- ✅ 文章分类与标签关联
- ✅ 文章关键词搜索
- ✅ 文章统计和分析
- ✅ 健康检查和监控

## 技术架构

### 后端技术栈
- **语言**: Go 1.23+
- **框架**: Gin
- **数据库**: MySQL 8.0 + GORM
- **认证**: JWT
- **权限**: RBAC
