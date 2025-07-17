# 开发指南

本文档提供了 MyBlog Monorepo 项目的详细开发指南，包括架构设计、最佳实践和开发工作流。

## 目录

- [项目架构](#项目架构)
- [开发环境设置](#开发环境设置)
- [开发工作流](#开发工作流)
- [代码规范](#代码规范)
- [部署指南](#部署指南)
- [最佳实践](#最佳实践)

## 项目架构

### Monorepo 结构

MyBlog 采用 Monorepo 架构，将前后端代码统一管理在一个仓库中：

```
MyBlog/
├── server/           # Go 后端服务
├── web/             # SvelteKit 前端应用
├── scripts/         # 跨项目脚本
├── docs/            # 项目文档
├── .husky/          # Git hooks
└── package.json     # 根配置文件
```

### 后端架构 (server/)

采用经典的三层架构模式：

#### 分层设计

```
┌─────────────────┐
│   HTTP Layer    │  cmd/myblog/main.go
│   (Gin Router)  │  
├─────────────────┤
│  Handler Layer  │  internal/handler/
│  (请求处理)      │  - 参数验证
│                 │  - 响应格式化
├─────────────────┤
│  Service Layer  │  internal/service/
│  (业务逻辑)      │  - 业务规则
│                 │  - 数据处理
├─────────────────┤
│Repository Layer │  internal/repository/
│  (数据访问)      │  - 数据库操作
│                 │  - 模型定义
├─────────────────┤
│  Database Layer │  MySQL + GORM
│                 │  
└─────────────────┘
```

#### 核心模块

- **Config** (`internal/config/`) - 配置管理，使用 Viper 加载 YAML 配置
- **Database** (`internal/database/`) - 数据库连接池和迁移管理
- **Response** (`pkg/response/`) - 统一 API 响应格式
- **DateTime** (`pkg/datetime/`) - 自定义时间类型处理

#### 依赖注入

项目使用构造函数注入模式：

```go
// 在 main.go 中
userRepo := repository.NewUserRepository(db)
userSvc := service.NewUserService(userRepo)
userHandler := handler.NewUserHandler(userSvc)
```

### 前端架构 (web/)

基于 SvelteKit 的现代前端应用：

#### 目录结构

```
web/
├── src/
│   ├── routes/        # 页面路由 (基于文件系统)
│   ├── lib/           # 可复用组件
│   ├── service/       # API 调用服务
│   └── utils/         # 工具函数
├── static/            # 静态资源
└── vite.config.ts     # 构建配置
```

#### 技术栈

- **SvelteKit** - 全栈框架，支持 SSR/SPA
- **TypeScript** - 类型安全
- **TailwindCSS** - 原子化 CSS
- **Vite** - 快速构建工具

## 开发环境设置

### 环境要求

| 工具    | 版本要求    | 说明                  |
|-------|---------|---------------------|
| Go    | 1.23+   | 后端开发语言              |
| Bun   | 1.1.24+ | JavaScript 运行时和包管理器 |
| MySQL | 8.0+    | 数据库服务               |

### 快速启动

```bash
# 1. 克隆项目
git clone <repository-url>
cd MyBlog

# 2. 自动环境设置
bun run setup

# 3. 启动开发环境
bun run dev
```

### 手动设置步骤

如果自动设置失败，可以手动执行以下步骤：

```bash
# 1. 安装根目录依赖
bun install

# 2. 安装前端依赖
cd web && bun install

# 3. 安装后端依赖
cd ../server && go mod tidy

# 4. 安装 Go 代码检查工具
cd .. && bun run go:lint-install

# 5. 配置数据库
# 确保 MySQL 服务运行
# 检查 server/configs/config.yaml 中的数据库配置

# 6. 启动开发服务
bun run dev
```

## 开发工作流

### 日常开发流程

1. **开始开发**
```bash
# 启动所有服务
bun run dev
```

2. **代码开发**
- 后端开发：编辑 `server/` 下的文件，自动热重载
- 前端开发：编辑 `web/src/` 下的文件，自动热重载

3. **代码提交前**
```bash
# 自动代码检查 (通过 git hooks)
git add .
git commit -m "feat: 添加新功能"
```

4. **测试和质量检查**
```bash
# 完整质量检查
bun run quality

# 分别运行
bun run format    # 代码格式化
bun run lint      # 代码检查
bun run test      # 运行测试
```

### Git 工作流

项目使用 Conventional Commits 规范：

```bash
# 功能开发
git commit -m "feat(api): 添加用户登录接口"
git commit -m "feat(ui): 添加登录页面"

# 问题修复
git commit -m "fix(db): 修复数据库连接池配置"

# 文档更新
git commit -m "docs: 更新开发指南"

# 重构
git commit -m "refactor(auth): 重构认证中间件"

# 样式调整
git commit -m "style: 调整代码格式"

# 测试
git commit -m "test: 添加用户服务单元测试"

# 构建
git commit -m "build: 更新依赖版本"

# CI/CD
git commit -m "ci: 添加自动部署配置"

# 杂项
git commit -m "chore: 清理无用文件"
```

### 分支策略

建议使用 Git Flow 分支模型：

- `main` - 主分支，生产环境代码
- `develop` - 开发分支，集成分支
- `feature/*` - 功能分支
- `release/*` - 发布分支
- `hotfix/*` - 热修复分支

## 代码规范

### Go 代码规范

#### 1. 包命名

```go
// 好的命名
package user
package config
package response

// 避免的命名
package utils
package common
package base
```

#### 2. 接口设计

```go
// 定义接口
type UserService interface {
  CreateUser(req *CreateUserRequest) (*User, error)
  GetUserByID(id uint) (*User, error)
  UpdateUser(id uint, req *UpdateUserRequest) error
  DeleteUser(id uint) error
}

// 实现接口
type userService struct {
  userRepo UserRepository
}

func NewUserService(userRepo UserRepository) UserService {
  return &userService{
    serRepo: userRepo,
  }
}
```

#### 3. 错误处理

```go
// 统一错误响应
func (h *UserHandler) CreateUser(c *gin.Context) {
  var req CreateUserRequest
  
  if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, "请求参数错误: "+err.Error())
    return
  }
  
  user, err := h.userService.CreateUser(&req)
  if err != nil {
    response.InternalError(c, err.Error())
    return
  }

  response.SuccessWithMessage(c, "用户创建成功", user)
}
```

#### 4. 结构体标签

```go
type User struct {
  ID        uint      `json:"id" gorm:"primaryKey"`
  Username  string    `json:"username" gorm:"unique;not null" binding:"required,min=3,max=20"`
  Email     string    `json:"email" gorm:"unique;not null" binding:"required,email"`
  CreatedAt time.Time `json:"createdAt"`
  UpdatedAt time.Time `json:"updatedAt"`
}
```

### TypeScript/Svelte 代码规范

#### 1. 组件结构

```svelte
<script lang="ts">
  // 导入
  import type { User } from '$lib/types';
  import { userService } from '$lib/services';
  
  // 属性
  export let user: User;
  
  // 响应式变量
  let loading = false;
  
  // 函数
  async function handleUpdate() {
    loading = true;
    try {
      await userService.updateUser(user.id, user);
    } catch (error) {
      console.error('更新失败:', error);
    } finally {
      loading = false;
    }
  }
</script>

<!-- HTML -->
<div class="user-card">
  <h2>{user.username}</h2>
  <button on:click={handleUpdate} disabled={loading}>
    {loading ? '更新中...' : '更新'}
  </button>
</div>

<!-- CSS -->
<style>
  .user-card {
    @apply p-4 border rounded-lg shadow-sm;
  }
</style>
```

#### 2. API 服务

```typescript
// src/service/api.ts
import ky from 'ky';

const api = ky.create({
  prefixUrl: import.meta.env.VITE_API_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

export class ApiError extends Error {
  constructor(
    public code: number,
    message: string,
    public data?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function apiPost<T>(
  endpoint: string,
  data?: any
): Promise<T> {
  try {
    const response = await api.post(endpoint, {json: data}).json<ApiResponse<T>>();

    if (response.code !== 200) {
      throw new ApiError(response.code, response.message, response.data);
    }

    return response.data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(500, '网络请求失败', error);
  }
}
```

## 部署指南

### 开发环境部署

```bash
# 启动开发环境
bun run dev

# 访问地址
# 前端: http://localhost:8899
# 后端: http://localhost:3000
```

### 生产环境构建

```bash
# 构建所有服务
bun run build

# 分别构建
bun run build:server  # 构建 Go 二进制文件
bun run build:web     # 构建前端静态文件
```

## 最佳实践

### 1. 代码组织

- **单一职责原则**: 每个函数、类只做一件事
- **依赖注入**: 通过构造函数注入依赖，便于测试
- **接口抽象**: 定义明确的接口，降低耦合

### 2. 错误处理

- **统一错误格式**: 使用统一的错误响应结构
- **适当的错误级别**: 区分用户错误和系统错误
- **错误日志**: 记录详细的错误信息用于调试

### 3. 性能优化

- **数据库连接池**: 合理配置连接池大小
- **缓存策略**: 对频繁访问的数据使用缓存
- **分页查询**: 避免一次性查询大量数据

### 4. 安全考虑

- **输入验证**: 验证所有用户输入
- **SQL 注入防护**: 使用参数化查询
- **跨域配置**: 正确配置 CORS 策略
- **敏感信息**: 不在代码中硬编码密钥

### 5. 可维护性

- **代码注释**: 为复杂逻辑添加注释
- **文档更新**: 保持文档与代码同步
- **版本控制**: 使用语义化版本号
- **自动化测试**: 保持良好的测试覆盖率

## 更多资源

- [Go 官方文档](https://golang.org/doc/)
- [SvelteKit 文档](https://kit.svelte.dev/docs)
- [GORM 文档](https://gorm.io/docs/)
- [TailwindCSS 文档](https://tailwindcss.com/docs)
