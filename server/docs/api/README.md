# API 文档总览

## 概述

MyBlog 后端 API 提供完整的博客系统功能，包括用户管理、文章管理、健康检查等模块。所有接口遵循统一的设计规范，使用 POST 方法和 JSON 数据格式。

## 文档结构

### 核心规范
- [API 规范文档](./api-specification.md) - 统一的API设计规范和命名约定

### 功能模块

#### 系统监控
- [健康检查 API](./health-api.md) - 服务器状态检查接口

#### 用户管理
- [用户管理 API](./user-api.md) - 用户注册、登录、CRUD操作和权限管理

#### 内容管理  
- [文章管理 API](./article-api.md) - 文章CRUD、搜索、分类、标签等完整功能

## API 统计

| 模块 | 接口数量 | 说明 |
|------|----------|------|
| 健康检查 | 1 | 系统状态监控 |
| 用户管理 | 8 | 用户认证和管理 |
| 文章管理 | 29 | 文章内容管理 |
| **总计** | **38** | **完整的博客系统API** |

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

### 文章管理
#### 公开接口
- `POST /api/articles/getBySlug` - 根据Slug获取文章
- `POST /api/articles/list` - 获取文章列表
- `POST /api/articles/byAuthor` - 获取作者文章
- `POST /api/articles/byCategory` - 获取分类文章
- `POST /api/articles/byTag` - 获取标签文章
- `POST /api/articles/search` - 搜索文章
- `POST /api/articles/popular` - 获取热门文章
- `POST /api/articles/recent` - 获取最新文章

#### 认证接口
- `POST /api/articles/get` - 获取文章详情
- `POST /api/articles/like` - 点赞文章
- `POST /api/articles/unlike` - 取消点赞

#### 编辑权限接口
- `POST /api/articles/create` - 创建文章
- `POST /api/articles/update` - 更新文章
- `POST /api/articles/delete` - 删除文章
- `POST /api/articles/publish` - 发布文章
- `POST /api/articles/archive` - 归档文章
- `POST /api/articles/draft` - 转为草稿

#### 管理权限接口
- `POST /api/articles/getAll` - 获取所有文章
- `POST /api/articles/getStats` - 获取文章统计
- `POST /api/articles/getByStatus` - 按状态获取文章

### 分类标签管理
- `POST /api/articles/categories/list` - 获取分类列表
- `POST /api/articles/categories/create` - 创建分类
- `POST /api/articles/categories/update` - 更新分类
- `POST /api/articles/categories/delete` - 删除分类
- `POST /api/articles/tags/list` - 获取标签列表
- `POST /api/articles/tags/create` - 创建标签
- `POST /api/articles/tags/update` - 更新标签
- `POST /api/articles/tags/delete` - 删除标签

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
- **用户管理**: `user:create`, `user:update`, `user:delete`, `user:list`
- **文章管理**: `article:create`, `article:publish`, `article:manage`
- **评论管理**: `comment:moderate`
- **文件管理**: `file:upload`, `file:manage`
- **系统管理**: `system:*`

## 请求规范

### 统一请求格式
- **请求方式**: POST
- **Content-Type**: application/json
- **认证头**: Authorization: Bearer {accessToken}

### 统一响应格式
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {},
  "error": ""
}
```

## 开发指南

### 本地测试
```bash
# 启动开发服务器
bun run dev

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

### v1.0.0 (当前版本)
- ✅ 完整的用户认证和授权系统
- ✅ 基于RBAC的权限控制
- ✅ 完整的文章管理系统
- ✅ 分类和标签管理
- ✅ 全文搜索功能
- ✅ 文章统计和分析
- ✅ 健康检查和监控

## 技术架构

### 后端技术栈
- **语言**: Go 1.23+
- **框架**: Gin
- **数据库**: MySQL 8.0 + GORM
- **认证**: JWT
- **权限**: RBAC
