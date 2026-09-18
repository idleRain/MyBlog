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
├── apps/
│   ├── web/                  # @myblog/web —— 前台 toC（公开博客 + demo + i18n）
│   └── admin/                # @myblog/admin —— 后台 toB（管理控制台 + 登录，无 i18n）
├── packages/
│   ├── shared/               # @myblog/shared —— 框架无关工具、类型、常量
│   ├── http/                 # @myblog/http —— 纯 HTTP 请求器（ky 封装）
│   ├── api/                  # @myblog/api —— 后端接口模块与响应类型
│   └── ui/                   # @myblog/ui —— shadcn-svelte 基础组件（stock，主题注入）
├── server/                   # Go 后端服务
├── scripts/                  # 跨项目脚本（Node.js + tsx）
├── docs/                     # 项目文档
├── .husky/                   # Git hooks
├── pnpm-workspace.yaml       # pnpm 工作区配置
└── package.json              # 根配置文件
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
│                 │  （持久化实现，实体与 DTO 见 internal/domain）
├─────────────────┤
│  Domain Layer   │  internal/domain/
│  (领域类型)      │  - 领域实体 + 请求/响应 DTO（全系统唯一类型语言）
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
- **Slug** (`pkg/slug/`) - URL 友好标识生成工具，供文章、分类、标签复用
- **Markdown** (`pkg/markdown/`) - Markdown 转 HTML 渲染，供文章内容缓存使用

#### 业务模块

后端按领域划分为以下业务模块，每个模块遵循 handler → service → repository 三层结构，接口统一定义在各层实现所在包：

| 模块 | handler | service | repository | 说明 |
|------|---------|---------|------------|------|
| 用户管理 | `handler/user.go` | `service/user.go` | `repository/user.go` | 登录、CRUD、JWT 双 token |
| 文章管理 | `handler/article.go` | `service/article.go` | `repository/article.go` | 文章 CRUD、发布归档、互动 |
| 分类管理 | `handler/category.go` | `service/category.go` | `repository/category.go` | 分类树形管理 |
| 标签管理 | `handler/tag.go` | `service/tag.go` | `repository/tag.go` | 标签与热门标签 |
| 评论管理 | `handler/comment.go` | `service/comment.go` | `repository/comment.go` | 评论与审核状态机 |
| 媒体文件 | `handler/media.go` | `service/media.go` | `repository/media.go` | 上传与管理 |
| 系统设置 | `handler/setting.go` | `service/setting.go` | `repository/setting.go` | 配置与脱敏 |
| 友情链接 | `handler/friendly_link.go` | `service/friendly_link.go` | `repository/friendly_link.go` | 友链申请与审核 |
| 站点统计 | `handler/stats.go` | `service/stats.go` | `repository/stats.go` | 运营数据分析 |
| 站内通知 | `handler/notification.go` | `service/notification.go` | `repository/notification.go` | 消息中心 |
| 用户关注 | `handler/user_follow.go` | `service/user_follow.go` | `repository/user_follow.go` | 关注关系 |

#### 依赖注入

项目使用构造函数注入模式，**组合根唯一**：依赖对象只能在 `cmd/myblog/main.go` 中构造并逐层注入，禁止在 service / router / middleware 内部私自 `New` 依赖服务（历史违例见 [`architecture-rules.md`](./architecture-rules.md) 债务 D4，只减不增）：

```go
// 在 main.go 中（组合根）
userRepo := repository.NewUserRepository(db)
jwtService := service.NewJWTService(cfg)
rbacService := service.NewRBACService()
userSvc := service.NewUserService(userRepo, jwtService, rbacService,
    service.WithLoginLockoutPolicy(loginLockoutPolicyFromConfig(cfg)))
userHandler := handler.NewUserHandler(userSvc)
```

各模块在 `cmd/myblog/main.go` 中完成依赖装配，并通过 `router.Dependencies` 注入路由层。

### 前端架构（apps/ + packages/）

前端在 web-split 重构后拆分为前台与后台两个独立 SvelteKit 应用，公共代码抽取为五个 package（shared/http/api/auth/ui）。

#### 应用拆分

| 应用 | 定位 | 特性 | 开发端口 |
| --- | --- | --- | --- |
| `apps/web`（@myblog/web） | 前台 toC | 博客业务页面（首页/目录/详情/分类/归档/作者页/登录/收藏）+ demo 页 + i18n（paraglide） | 8899 |
| `apps/admin`（@myblog/admin） | 后台 toB | 管理控制台 + 登录页，基准路径 `/admin`，无 i18n | 9988 |

#### 公共包

| 包 | 职责 | 关键依赖 |
| --- | --- | --- |
| `@myblog/shared` | 纯工具（`cn`、深拷贝、防抖节流等）+ 通用类型 + 常量（站点常量与后端响应码 `RESPONSE_CODE_*`），不依赖 SvelteKit/Svelte | clsx、tailwind-merge、mitt |
| `@myblog/http` | `createHttpClient` 工厂（ky 封装，认证回调注入，401 刷新钩子） | ky、@myblog/shared |
| `@myblog/api` | 后端接口模块与响应类型（user/article/category/tag/comment/media/setting/friendlyLink/stats/notification/follow 共 11 个模块工厂） | @myblog/http、@myblog/shared |
| `@myblog/auth` | 认证会话组装：`createAuthStore` 工厂（注入式，逻辑两应用共享） | @myblog/api、@myblog/shared、svelte |
| `@myblog/ui` | shadcn-svelte 基础组件，保持 stock，主题经各应用 `app.css` token 注入 | bits-ui、vaul-svelte 等 |

依赖方向单向无环：应用 → `ui` → `shared`；应用 → `auth`/`http` → `shared`；应用 → `api` → `http` + `shared`。

#### 应用目录结构

```
apps/web/src/                 # 前台
├── routes/                   # 分组路由：(app) 业务页面、(demo) i18n 演示
├── lib/
│   ├── api/                  # 接口实例化（8 个模块工厂实例聚合导出）
│   ├── service/              # http 客户端创建（注入认证与提示回调）
│   ├── components/           # 前台组件（home/ 首页版面、article/ 文章域、comment/ 评论、user/ 用户域、layout/ 全局框架）
│   ├── motion/               # GSAP 动效基建（插件注册、缓动常量、滚动进场 actions）
│   ├── stores/               # 认证状态 store
│   ├── utils/                # 页面工具（日期格式化等）
│   └── paraglide/            # i18n 生成产物
└── app.css                   # 编辑杂志主题（暖纸墨色系，含 --signal 与 --color-line 别名）

apps/admin/src/               # 后台
├── routes/                   # 分组路由：(admin) 后台页面、(auth) 登录页
├── lib/
│   ├── api/                  # 接口实例化
│   ├── service/              # http 客户端创建
│   ├── components/           # 后台组件
│   ├── stores/               # 认证 store 薄封装（逻辑在 @myblog/auth）
│   ├── constants/            # 权限与角色常量（降级兜底，权限判定读登录下发值）
│   ├── types/                # 后台专用类型（接口类型一律来自 @myblog/api）
│   └── utils/                # 认证域工具（logout / permissions / auth-guard / request / navigation）
└── app.css                   # 原始后台主题（无 --signal）
```

#### 关键技术配置

- 路径别名见各应用 `svelte.config.js`：`$lib`、`$ui`（指向 `packages/ui/src` 源码）、`@/*`、`#/*`、`~/*`；前台另有 `$i18n`。
- `packages/ui` 以**源码直连**方式被引用，各应用在 `vite.config.ts` 配置 `ssr.noExternal: ['@myblog/ui']` 参与 SSR 编译；Tailwind v4 在各应用 `app.css` 用 `@source` 显式纳入 `packages/ui/src` 扫描。
- 后台在 `svelte.config.js` 配置 `paths.base = '/admin'` 与 `relative: false`，与前台同源区分部署。
- 前台经 `@inlang/paraglide-js` 做 i18n，产物位于 `src/lib/paraglide`；后台不引入 i18n。
- 两应用均使用 `unplugin-auto-import` 自动导入 SvelteKit / Svelte / toast 常用函数。

## 开发环境设置

### 环境要求

| 工具    | 版本要求 | 说明                            |
| ------- | -------- | ------------------------------- |
| Go      | 1.23+    | 后端开发语言                    |
| Node.js | 20+      | JavaScript 运行时与脚本执行器   |
| MySQL   | 8.0+     | 数据库服务                      |

### 快速启动

> 根 `package.json` 的全部 scripts 均为 `scripts/` 目录下 TypeScript 脚本的薄封装，
> 各脚本的用途、参数与 pnpm 入口对照见 [`scripts/README.md`](../scripts/README.md)。

```bash
# 1. 克隆项目
git clone <repository-url>
cd MyBlog

# 2. 自动环境设置
pnpm run setup

# 3. 启动开发环境
pnpm run dev
```

### 手动设置步骤

如果自动设置失败，可以手动执行以下步骤：

```bash
# 1. 安装根目录依赖（pnpm workspace 会一并安装 apps/ 与 packages/ 的全部依赖）
pnpm install

# 2. 安装后端依赖
cd server && go mod tidy

# 3. 安装 Go 代码检查工具
cd .. && pnpm run go:lint-install

# 4. 配置数据库
# 确保 MySQL 服务运行
# 检查 server/configs/config.yaml 中的数据库配置

# 5. 启动开发服务
pnpm run dev
```

### 初始化管理员账户

数据库初始化不会自动创建管理员账户，需要通过种子命令创建超级管理员：

```bash
# 初始化默认超级管理员
pnpm run seed:admin

# 自定义用户名、密码、邮箱
pnpm run seed:admin --username root --password Root@2025 --email root@myblog.local
```

- **默认账户**：用户名 `admin`，密码 `Admin@123456`，邮箱 `admin@myblog.local`
- 密码规则：至少 6 位，且必须同时包含字母和数字
- 命令幂等：账户已存在且为超级管理员时不重复创建；已存在但角色非超级管理员时自动提升为超级管理员
- 相关实现：`server/cmd/seed/main.go`（命令行入口）、`server/internal/database/seed.go`（核心逻辑）
- 登录接口：`POST /api/users/login`

## 开发工作流

### 日常开发流程

1. **开始开发**

```bash
# 启动所有服务
pnpm run dev
```

2. **代码开发**

- 后端开发：编辑 `server/` 下的文件，自动热重载
- 前端开发：编辑 `apps/web/src/`（前台）或 `apps/admin/src/`（后台）下的文件，自动热重载

3. **代码提交前**

```bash
# 自动代码检查 (通过 git hooks)
git add .
git commit -m "feat: 添加新功能"
```

4. **测试和质量检查**

```bash
# 完整质量检查
pnpm run quality

# 分别运行
pnpm run check     # 前后台 SvelteKit 类型检查
pnpm run format    # 代码格式化
pnpm run lint      # 代码检查
pnpm run test      # 运行测试
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

> 本节仅覆盖风格约定。**架构层面的硬性规则（依赖方向、类型唯一真相源、契约先行、单一权威、禁止复制、数据加载）见 [`architecture-rules.md`](./architecture-rules.md) 与根目录 `AGENTS.md` 第 0 节，违反必须返工。**

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
    userRepo: userRepo,
  }
}
```

#### 3. 错误处理

```go
// 错误映射唯一权威：统一经 handler/error.go 的 HandleServiceError 映射
// （"资源不存在"哨兵 → 404、ErrPermissionDenied → 403、其余 → 500），
// 禁止在模块内新增私有映射函数产生第二套权威（QA-03 收口口径）
func (h *UserHandler) GetUserByID(c *gin.Context) {
  var req GetUserRequest

  if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, "请求参数错误: "+err.Error())
    return
  }

  user, err := h.userService.GetUserByID(req.ID)
  if err != nil {
    HandleServiceError(c, err)
    return
  }

  response.Success(c, user)
}
```

handler 层出现新的"资源不存在"哨兵错误时，扩展 `HandleServiceError` 的 not-found 集合并同步 `contracts/errors.yaml` 的 404 口径；`binding` 校验失败映射 400 的口径不变。

#### 4. 结构体标签（实体与请求 DTO 必须分离）

实体只携带 `json` + `gorm` tag；`binding`（HTTP 校验）只出现在请求 DTO 上。**禁止把三类 tag 写进同一个结构体**——那会让存储结构、API 契约与请求校验互相锁死（历史教训见 [`architecture-rules.md`](./architecture-rules.md) 第 9 节）：

```go
// 实体（internal/domain）：领域实体 + GORM tag，禁止 binding tag
type User struct {
  ID        uint      `json:"id" gorm:"primaryKey"`
  Username  string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
  Email     string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
  CreatedAt time.Time `json:"createdAt"`
  UpdatedAt time.Time `json:"updatedAt"`
}

// 请求 DTO（internal/domain/dto.go）：承载校验，禁止 gorm tag
type CreateUserRequest struct {
  Username string `json:"username" binding:"required,min=3,max=20"`
  Email    string `json:"email" binding:"required,email"`
  Password string `json:"password" binding:"required,min=6,max=100"`
}
```

### TypeScript/Svelte 代码规范

#### 1. 组件结构

```svelte
<script lang="ts">
  // 导入
  import type { User } from '@myblog/api'
  import { UserAPI } from '$lib/api'

  // 属性（Svelte 5 runes）
  let { user }: { user: User } = $props()

  // 响应式状态
  let loading = $state(false)

  // 函数
  async function handleUpdate() {
    loading = true
    try {
      await UserAPI.updateUser(user.id, { username: user.username })
    } catch (error) {
      console.error('更新失败:', error)
    } finally {
      loading = false
    }
  }
</script>

<!-- HTML -->
<div class="user-card">
  <h2>{user.username}</h2>
  <button onclick={handleUpdate} disabled={loading}>
    {loading ? '更新中...' : '更新'}
  </button>
</div>

<!-- CSS：优先使用 Tailwind 工具类，复杂样式才放 <style> -->
<style>
  .user-card {
    @apply p-4 border rounded-lg shadow-sm;
  }
</style>
```

#### 2. API 服务

```typescript
// packages/http 提供 createHttpClient 工厂（ky 封装，认证回调注入）
// packages/api 提供 createUserAPI 接口模块工厂

// 应用侧 src/lib/service/index.ts：创建客户端实例，注入认证与提示回调
import { createHttpClient } from '@myblog/http'

const request = createHttpClient({
  prefixUrl: import.meta.env.VITE_BASE_URL,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  auth: {
    getAccessToken: () => authStore.getCurrentState()?.accessToken ?? null,
    refreshToken: async () => { /* 令牌接近过期时执行刷新 */ },
    onAuthFailure: async message => { /* 清除认证状态并跳转登录页 */ }
  },
  onError: message => { toast.error(message) }
})

// 应用侧 src/lib/api/index.ts：实例化接口模块
import { createUserAPI } from '@myblog/api'
const UserAPI = createUserAPI(request)

// 一律使用 POST 调用后端（与后端 POST-Only 规范呼应）
const list = await UserAPI.getUserList(1, 10)
const updated = await UserAPI.updateUser({ id: user.id, username: 'new-name' })
```

#### 3. 可访问性

- 图标按钮必须携带 `aria-label`（UI-09/21 清偿口径），纯装饰图标加 `aria-hidden="true"`

#### 4. 用户域类型双形状

`@myblog/api` 的 user 模块存在两个合法镜像形状，各自对应不同的后端输出，禁止混用：

| 类型 | 镜像源 | 对应端点 |
| --- | --- | --- |
| `User` | `domain.User.ToResponse()` 裁剪形状 | 登录、`users/list`、`users/get` |
| `ProfileUser` | `domain.User` 自助全量形状（含 bio/website/coverImage 等） | `users/profile`、`users/profile/update` |

修改后端任一形状的输出字段集时，必须同步 `@myblog/api/modules/user/types.ts` 对应类型并运行 `contract:check`（铁律 A2 镜像义务）。

> 已知漂移注记：`User.birthday` 后端经 `datetime.JSONDate` 输出，零值序列化为 `null` 而非常见空字符串；当前前端类型声明为 `string`，触碰该类型时评估改为 `string | null`（历史遗留，金样本以有值样例规避）。

#### 5. 界面多语言（仅 apps/web）

- 用户可见文案经 `$i18n` 的 `m['ui:...']()` 取词；接入现状与红线见 [`architecture-rules.md`](./architecture-rules.md) 债务 D18（Header/Footer/错误页已接入，其余页面待页面大变动后分批，已接入文件禁止回退硬编码）
- `messages/*.json5` 必须保持 **JSON 兼容写法**（双引号、无尾逗号）：paraglide 编译器按严格 JSON 解析，无引号键/尾逗号/单引号会直接编译失败
- 修改词表后需重新编译生成 `src/lib/paraglide` 产物：`pnpm --filter @myblog/web exec paraglide-js compile --project ./project.inlang --outdir ./src/lib/paraglide`（`vite dev`/`build` 亦会自动触发）

## 部署指南

### 开发环境部署

```bash
# 启动开发环境
pnpm run dev

# 访问地址
# 前台: http://localhost:8899
# 后台: http://localhost:9988/admin
# 后端: http://localhost:3000
```

### 生产环境构建

```bash
# 构建所有服务
pnpm run build

# 分别构建
pnpm run build:server  # 构建 Go 二进制文件
pnpm run build:web     # 构建前端静态文件
```

## 最佳实践

### 1. 代码组织

- **单一职责原则**: 每个函数、类只做一件事
- **依赖注入**: 通过构造函数注入依赖，便于测试；依赖只在组合根（main.go）构造
- **接口抽象**: 定义明确的接口，降低耦合
- **架构铁律**: 依赖方向、类型唯一真相源等结构性约束见 [`architecture-rules.md`](./architecture-rules.md)
- **事务化模式**: 多段写库必须包裹单事务；repository 提取 `xxxTx(db *gorm.DB, ...)` 私有方法供事务与独立公开方法复用（`createTx`/`syncTagsTx`/`syncCategoriesTx`/`updateTx`/`recordArticleView` 先例，事务内所有查询与写入均走事务句柄，禁止混用 `r.db`），跨仓储共享的写逻辑下沉为包内共享 helper（`upsertContentStat` 先例），不引入仓储间依赖
- **安全配置默认值双源对齐**: `security.login_lockout` 的 viper 默认值与 service 层 `defaultLoginLockoutPolicy` 必须一致（5 次 / 15 分钟），修改任一处必须同步另一处

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
- **跨域配置**: 跨域来源一律改 `server/configs/config.yaml` 的 `cors` 节白名单（精确匹配），禁止代码内硬编码 Origin 或恢复全放行；生产同源网关形态白名单保持为空
- **WAF 阻止模式**: 新增或修改 `getDefaultBlockedPatterns()` 模式必须先红后绿配"攻击拦截 + 误伤回归"两组用例，模式经词首边界或取值上下文锚定，禁止回退宽匹配（债务 D16）
- **限流键格式**: 用户级限流键统一 `user:<十进制ID>`（经 `getUserKey` 构造），禁止 rune 转换或字符串直拼产生非法键
- **认证协议变更**: token 形状、刷新、撤销语义变更属最高风险契约变更，须先更新 `contracts/auth-protocol.md` 再动代码（流程见 `architecture-rules.md` §6.3）
- **敏感信息**: 不在代码中硬编码密钥；JWT 双密钥经 `MYBLOG_*` 环境变量注入，缺失即启动失败
- **隐私告知同步**: 后端新增或启用的数据收集场景（如评论 IP/UA 写入、新增埋点）必须先同步更新 web 端 `/privacy` 隐私政策页告知再上线（PM-22 承诺口径，隐私页已写明"新增收集场景前先更新本页"）
- **设置键有效登记**: 后端业务新消费某个设置键时，必须同步补录 admin `lib/constants/setting.ts` 的 `EFFECTIVE_SETTING_KEYS`（未登记项在设置中心显示"未生效"角标）；安全分组的真实生效渠道为 `config.yaml` 与安全中间件（PM-11 口径）

### 5. 测试纪律

- **先红后绿**: 修复类任务先写失败测试再改实现，确保用例真实锚定缺陷而非实现细节
- **替身覆写**: service 层 fake 以内嵌空接口继承全部方法，接口新增方法被既有测试路径调用时必须为受影响 fake 显式覆写，否则以 nil panic 暴露（债务 D17）
- **sqlmock 维护**: repository 层 sqlmock 断言与 GORM 生成的 SQL 文本强耦合，GORM 升级或查询改写时须同步维护期望；匹配 SQL 用 `regexp.QuoteMeta` 前缀正则，期望序列须与事务内实际步骤（含 DELETE、计数 UPDATE）完全一致，否则以"剩余期望未匹配"失败
- **契约金样本维护**: 金样本须与 Go 响应逐字节一致（`json.Compact` 豁免空白），修改实现输出字段时必须同步修订金样本并跑 `contract:check`。易错点：`datetime.JSONDate` 序列化为零值 `null`、午夜日期-only、非午夜 `"2006-01-02 15:04:05"`（非 RFC3339）；`gin.H` map 序列化按字典序排键；Go `encoding/json` 默认 HTML 转义（`<` `>` 输出 `\u003c`），金样本取值应避开尖括号；`omitempty` 对 nil 指针与空切片省略键。锁②为 `WidenLiteral`+`toExtend` 单向合法实例校验（JSON 单值无法与 nullable 联合精确相等，见 architecture-rules.md §3.4）
- **gin 中间件测试**: 中间件函数内 `return` 不会中断 gin 的 handler 链，中断必须调用 `c.Abort()`；`IdentityProvider.Resolve` 失败时由实现写响应并 `Abort`，消费方 `if err != nil { return }` 依赖此语义，测试替身须模拟"写响应 + Abort"而非裸返错误。统一响应信封固定 HTTP 200，权限/认证失败以响应体业务码（`code` 字段）断言，不比较 HTTP 状态码
- **依赖基线检查**: `git grep -ln "MyBlog/internal/repository" -- server/internal/...` 须在**仓库根**执行，在 server/ 内跑 pathspec 不匹配会得到假绿

### 5.1 质量门禁

- **契约门禁**: `pnpm run contract:check` = Go handler fixture 测试（锁①）+ `@myblog/api` 的 `tsc --noEmit`（锁② typecheck）+ vitest 类型锚定；契约相关改动必须通过
- **CI**: `.github/workflows/ci.yml` 为五步质量门禁（go 三连、依赖基线 grep、双应用 svelte-check、contract:check、lint），当前**手动触发**（`workflow_dispatch`），push/PR 自动门禁待服务器就绪后改回 `on` 触发；工作流不含自动部署步骤

### 6. 可维护性

- **代码注释**: 为复杂逻辑添加注释
- **文档更新**: 保持文档与代码同步（架构规则变更须同步 `docs/architecture-rules.md` 债务登记表）
- **版本控制**: 使用语义化版本号
- **自动化测试**: 后端关键逻辑配套 `*_test.go`（service/repository 层已有存量）；前端 packages 公共逻辑与页面状态模块应补 `*.test.ts`

## 更多资源

- [Go 官方文档](https://golang.org/doc/)
- [SvelteKit 文档](https://kit.svelte.dev/docs)
- [GORM 文档](https://gorm.io/docs/)
- [TailwindCSS 文档](https://tailwindcss.com/docs)
