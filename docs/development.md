# 开发指南

本文档提供 MyBlog Monorepo 的环境设置、开发工作流、工具陷阱与代码风格示例；架构层面的结构性规则与裁决细则见根 `AGENTS.md` 与 [`architecture-rules.md`](./architecture-rules.md)。

## 目录

- [项目架构](#项目架构)
- [开发环境设置](#开发环境设置)
- [开发工作流](#开发工作流)
- [代码规范](#代码规范)
- [部署指南](#部署指南)
- [最佳实践](#最佳实践)
- [更多资源](#更多资源)

## 项目架构

### Monorepo 结构

MyBlog 采用 Monorepo 架构，将前后端代码统一管理在一个仓库中。仓库目录树与各包职责见根 `AGENTS.md` 第 3 节「仓库概览」，本节不复述。

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
- **Storage** (`pkg/storage/`) - 存储侧通用逻辑，当前为上传目录可写性校验（`CheckWritable`），供就绪探针复用

#### 业务模块

后端按领域划分为以下业务模块，每个模块遵循 handler → service → repository 三层结构，接口统一定义在各层实现所在包：

| 模块 | handler | service | repository | 说明 |
|------|---------|---------|------------|------|
| 用户管理 | `handler/user.go` | `service/user.go` | `repository/user.go` | 登录、CRUD、不透明双令牌 |
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
| 数据字典 | `handler/dict.go` | `service/dict.go` | `repository/dict.go` | 字典类型与字典项 |

#### 依赖注入

依赖一律由组合根 `cmd/myblog/`（`main.go` 管生命周期、`deps.go` 按域装配）构造，再经 `router.Dependencies` 逐层注入，禁止在 service / router / middleware 内私自 `New`；规则原文与验证命令见根 `AGENTS.md` 第 1 节第 4 条与 [`architecture-rules.md`](./architecture-rules.md) §5.2。

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
| `@myblog/http` | `createHttpClient` 工厂（ky 封装，认证回调注入） | ky、@myblog/shared |
| `@myblog/api` | 后端接口模块与响应类型（article/category/comment/dict/follow/friendlyLink/media/notification/setting/stats/tag/user 共 12 个模块工厂） | @myblog/http、@myblog/shared |
| `@myblog/auth` | 认证会话组装：`createAuthStore` 工厂（注入式，逻辑两应用共享） | @myblog/api、@myblog/shared（svelte 为 peerDependency） |
| `@myblog/ui` | shadcn-svelte 基础组件，保持 stock，主题经各应用设计令牌注入 | bits-ui、vaul-svelte 等 |

依赖方向单向无环：应用 → `ui` → `shared`；应用 → `auth`/`http` → `shared`；应用 → `api` → `http` + `shared`。

#### 应用目录结构

```
apps/web/src/                 # 前台
├── routes/                   # 分组路由：(app) 业务页面、(demo) i18n 演示
├── lib/
│   ├── api/                  # 接口实例化（9 个模块工厂实例聚合导出）
│   ├── service/              # http 客户端创建（注入认证与提示回调）
│   ├── components/           # 前台组件（home/ 首页版面、article/ 文章域、comment/ 评论、user/ 用户域、layout/ 全局框架、icons/）
│   ├── markdown/             # Markdown 渲染与代码块增强（render.ts 等）
│   ├── motion/               # GSAP 动效基建（插件注册、缓动常量、滚动进场 actions）
│   ├── stores/               # 认证状态 store
│   ├── styles/               # 字体来源模块（#fonts 别名的本地与 CDN 两个实现）
│   ├── utils/                # 页面工具（日期格式化等）
│   └── paraglide/            # i18n 生成产物，不入库，由 paraglide 编译生成
├── styles/                   # 设计令牌与基础层（tokens.css 定义 --signal 与 --color-line）
└── app.css                   # 样式聚合入口（仅 @import 与 @source）

apps/admin/src/               # 后台
├── routes/                   # 分组路由：(admin) 后台页面、(auth) 登录页
├── lib/
│   ├── api/                  # 接口实例化（12 个模块工厂实例聚合导出）
│   ├── service/              # http 客户端创建
│   ├── components/           # 后台组件（admin/ 域组件）
│   ├── stores/               # 认证 store 薄封装（逻辑在 @myblog/auth）
│   ├── constants/            # 权限、角色与各业务域常量（降级逻辑，权限判定读登录下发值）
│   └── utils/                # 认证域工具（auth-guard / logout / navigation / permissions）
├── styles/                   # 设计令牌与基础层（无 --signal）
└── app.css                   # 样式聚合入口（仅 @import 与 @source）
```

> 影子类型层已清除：admin `lib/types/` 目录不存在，接口类型一律来自 `@myblog/api`，两应用 eslint 以 patterns 拦截 `$lib/types` 引入形态。

#### 关键技术配置

- 路径别名见各应用 `svelte.config.js`：`$ui` 与 `$ui/*`（指向 `packages/ui/src` 源码）、`@/*`（应用自身 `src`）；`$lib` 由 SvelteKit 隐式提供，未在配置中声明；前台另有 `$i18n`（指向 `src/lib/paraglide/messages`）。`svelte.config.js` 与 `tsconfig.json` 均未声明 `#/*` 或 `~/*` 别名；前台 `vite.config.ts` 的构建期别名只有 `#fonts`（dev 走本地 fontsource、prod 走 CDN 模块），与通配形态的 `#/*` 不是一回事。
- `packages/ui` 以**源码直连**方式被引用，各应用在 `vite.config.ts` 配置 `ssr.noExternal: ['@myblog/ui']` 参与 SSR 编译；Tailwind v4 在各应用 `app.css` 用 `@source` 显式纳入 `packages/ui/src` 扫描。
- 后台在 `svelte.config.js` 配置 `paths.base = '/admin'` 与 `relative: false`，与前台同源区分部署。
- 前台经 `@inlang/paraglide-js` 做 i18n，产物位于 `src/lib/paraglide`；后台不引入 i18n。
- 两应用均使用 `unplugin-auto-import` 自动导入 SvelteKit / Svelte / toast 常用函数。

## 开发环境设置

### 环境要求

| 工具    | 版本要求 | 说明                            |
| ------- | -------- | ------------------------------- |
| Go      | 1.26.9   | 后端开发语言                    |
| Node.js | 22+      | JavaScript 运行时与脚本执行器   |
| MySQL   | 8.0+     | 数据库服务                      |

Go 版本的单一事实来源为 `server/go.mod` 的 `go` 指令（当前 `1.26.9`），CI 经 `actions/setup-go` 的 `go-version-file` 读取该值；本地低于该版本时由 `GOTOOLCHAIN` 自动下载对应工具链。

Node.js 版本的单一事实来源为根 `.nvmrc`（当前 `22`），与 CI 的 `actions/setup-node` 保持一致；根 `package.json` 的 `engines.node` 声明为 `>=22`，版本过低时 pnpm 会给出告警。新机器建议先执行 `nvm use` 再执行 `pnpm install`。

### 快速启动

> 根 `package.json` 的 scripts 分两类：多数是 `scripts/` 目录下 TypeScript 脚本的薄封装，
> 另一部分（`test`、`test:packages`、`test:apps`、`lint:web`、`format:web`、`check`、`typecheck:web`、`contract:check`、`audit:deps`）直接编排 pnpm、go 与 shell 命令。
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
- 密码规则：经服务层创建或修改用户时要求 6 至 100 位且同时包含字母和数字，并拒绝常见弱口令（`service/user.go` 的 `validatePasswordStrength`）；种子命令直连 `database.EnsureSuperAdmin`，不经该校验
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

仓库当前只维护 `master` 一个长期分支，CI 在 push 到 `master`/`main` 与所有 pull request 上触发。

- 修复与功能类工作按需建立短生命周期分支，历史远端分支形如 `fix/health-remediation-ph1-5`、`fix/health-remediation-ph6-9`。
- 未采用 Git Flow 的 `develop`、`release/*`、`hotfix/*` 常驻分支，本节不再保留该模型的建议清单。

### 工具陷阱与生成物维护

以下三项均为实测结论，触碰相关区域前先读本节。

#### 1. lint-staged 的暂存区行为

`lint-staged` 在存在未暂存改动时会先 stash 再执行任务。历史实测（15.2.10）曾出现「恢复阶段把未暂存改动重新加入索引」，使无关文件混入提交；**升级到 17.6.0 后按同一路径复测未再现**（暂存 1 个 `*.json`、未暂存 1 个 Markdown，任务确实改写了暂存文件，索引清单前后一致）。

无论版本如何，「先暂存目标文件并核对清单，再提交」是稳定做法：

```bash
git add <目标文件>
git diff --cached --name-only   # 核对清单后再执行 git commit
```

#### 2. 自动导入生成物的维护方式

两应用各有两个由 `unplugin-auto-import` 写入的文件，仓库根另有一个无消费方的历史残留，**均入库跟踪但不由人工维护**：

| 文件 | 写入者 | 说明 |
|---|---|---|
| `apps/<app>/typings/auto-imports.d.ts` | 该应用的 vite 插件 | 自动导入的全局类型声明 |
| `apps/<app>/.eslintrc-auto-import.js` | 该应用的 vite 插件 | 同名全局变量的 eslint globals，被该应用 `eslint.config.js` 直接 import |
| `.eslintrc-auto-import.js`（仓库根） | 某次以仓库根为工作目录的运行 | **无任何消费方**，改它不影响任何门禁 |

该插件以追加方式写入声明：**误加或误删的声明不会随下次运行自动消失**，改错必须手工修正。调整自动导入白名单后，应显式重跑一次前端构建（vite 插件只在构建与开发服务启动时写入，`svelte-check` 不会触发），并核对上述文件的差异是否符合预期；出现与本轮改动无关的差异时不要顺手提交。

#### 3. 依赖漏洞中「已发布 advisory 但上游无修复版本」的处置顺序

`pnpm audit` 的个别条目在 npm 上**不存在已发布的修复版本**（advisory 标注的版本并未发布），此时按以下顺序处置：

1. 先查该依赖的**父包**能否升级以移除整条依赖链（先例：`lint-staged` 升级后不再包含 `braces` 与 `micromatch`）；
2. 父包无法升级时，再用 `pnpm-workspace.yaml` 的 `overrides` 按版本范围收敛；
3. 两者都不可行时才登记为已知残余风险，并写明退出条件。

对无修复版本的包直接写 `overrides` 只会把问题掩盖成「已处理」，实际并未消除。

## 代码规范

> 本节仅覆盖风格约定。**架构层面的硬性规则（依赖方向、类型唯一真相源、契约先行、单一权威、禁止复制、数据加载）见 [`architecture-rules.md`](./architecture-rules.md) 与根目录 `AGENTS.md` 第 1 节，违反必须返工。**

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
// 定义接口；接口与实现同包声明，接口名统一 XxxInterface 后缀
// （存量的 UserService、RBACService、UserRepository 未带后缀，触碰时统一）
type UserService interface {
  CreateUser(req *domain.CreateUserRequest) (*domain.User, error)
  GetUserByID(id uint) (*domain.User, error)
  DeleteUser(id uint) error
}

// 实现接口；各层依赖接口而非实现
type userService struct {
  userRepo     repository.UserRepository
  tokenService TokenServiceInterface
  rbacService  RBACService
}

// 依赖一律由组合根注入，禁止在包内私自 New
func NewUserService(userRepo repository.UserRepository, tokenService TokenServiceInterface,
  rbacService RBACService, opts ...UserServiceOption) UserService {
  svc := &userService{
    userRepo:     userRepo,
    tokenService: tokenService,
    rbacService:  rbacService,
  }
  for _, opt := range opts {
    opt(svc)
  }
  return svc
}
```

#### 3. 错误处理

```go
// 错误映射唯一权威：统一经 handler/error.go 的 HandleServiceError 映射——
// not-found 哨兵集合 → 404、ErrPermissionDenied → 403、
// service.ErrInvalidRequest / ErrUsernameTaken / ErrEmailTaken → 400、
// 其余 → 500 且响应体脱敏为固定文案（原始错误仅入日志）。
// 禁止在模块内新增私有映射函数产生第二套权威。
func (h *UserHandler) GetUserByID(c *gin.Context) {
  // 请求 DTO 就近声明，由 binding tag 承载必填与取值校验
  type GetUserByIDRequest struct {
    ID uint `json:"id" binding:"required,min=1"`
  }

  var req GetUserByIDRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, "请求参数错误: "+err.Error())
    return
  }

  user, err := h.userService.GetUserByID(req.ID)
  if err != nil {
    HandleServiceError(c, err)
    return
  }

  // 输出经 ToResponse 裁剪，禁止实体直出
  response.Success(c, user.ToResponse())
}
```

handler 层出现新的"资源不存在"哨兵错误时，扩展 `HandleServiceError` 的 not-found 集合并同步 `contracts/errors.yaml` 的 404 口径；`binding` 校验失败由 handler 直接 `response.BadRequest` 映射 400，不走 `HandleServiceError`。

#### 4. 结构体标签（实体与请求 DTO 必须分离）

实体只携带 `json` + `gorm` tag；`binding`（HTTP 校验）只出现在请求 DTO 上。**禁止把三类 tag 写进同一个结构体**——那会让存储结构、API 契约与请求校验互相锁死（历史教训见 [`architecture-rules.md`](./architecture-rules.md) 第 10 节）：

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
  Username string `json:"username" binding:"required,min=1,max=50"`
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
      // 接口方法按 @myblog/api 的实际签名调用，更新用户收单个请求对象
      await UserAPI.updateUser({ id: user.id, username: user.username, email: user.email })
    } catch (error) {
      console.error('更新失败:', error)
    } finally {
      loading = false
    }
  }
</script>

<!-- HTML：样式一律用 Tailwind 工具类；全局样式与设计令牌统一放 src/styles/，组件内不写 <style> -->
<div class="user-card rounded-lg border border-border p-4">
  <h2>{user.username}</h2>
  <button onclick={handleUpdate} disabled={loading}>
    {loading ? '更新中...' : '更新'}
  </button>
</div>
```

#### 2. API 服务

```typescript
// packages/http 提供 createHttpClient 工厂（ky 封装，认证回调注入）
// packages/api 提供 createUserAPI 接口模块工厂

// 应用侧 src/lib/service/index.ts：创建客户端实例，注入会话续期与界面提示回调
import { createHttpClient } from '@myblog/http'

const request = createHttpClient({
  prefixUrl: import.meta.env.VITE_BASE_URL,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  // 内容语言经 Accept-Language 下发，后端按此输出本地化字段
  getLanguage: () => getLocale(),
  auth: {
    // 会话令牌存放于 HttpOnly Cookie，由浏览器自动携带；
    // 回调不读取任何令牌，只负责在服务端拒绝当前会话时发起续期。
    refreshSession: () => authStore.refreshSingleFlight(refreshSession),
    onAuthFailure: async message => { /* 清除本地认证状态并跳转登录页 */ }
  },
  onError: message => { /* 界面统一提示，如 toast.error(message) */ }
})

// 应用侧 src/lib/api/index.ts：实例化接口模块
import { createUserAPI } from '@myblog/api'
const UserAPI = createUserAPI(request)

// 一律使用 POST 调用后端（与后端 POST-Only 规范呼应）
const list = await UserAPI.getUserList(1, 10)
const updated = await UserAPI.updateUser({ id: user.id, username: 'new-name', email: user.email })
```

#### 3. 可访问性

- 图标按钮必须携带 `aria-label`，纯装饰图标加 `aria-hidden="true"`

#### 4. 用户域类型双形状

`@myblog/api` 的 user 模块存在两个合法镜像形状，各自对应不同的后端输出，禁止混用：

| 类型 | 镜像源 | 对应端点 |
| --- | --- | --- |
| `User` | `domain.User.ToResponse()` 裁剪形状 | 登录、`users/list`、`users/get` |
| `ProfileUser` | `domain.User` 自助全量形状（含 bio/website/coverImage 等） | `users/profile`、`users/profile/update` |

修改后端任一形状的输出字段集时，必须同步 `@myblog/api/modules/user/types.ts` 对应类型并运行 `contract:check`（类型镜像义务）。两类形状的 `birthday` 均为 `string | null`：后端经 `datetime.JSONDate` 输出，零值序列化为 `null` 而非空字符串。

#### 5. 界面多语言（仅 apps/web）

- 用户可见文案经 `$i18n` 的 `m['ui:...']()` 取词；接入现状与红线见 [`architecture-rules.md`](./architecture-rules.md)（Header/Footer/错误页已接入，其余页面待页面大变动后分批，已接入文件禁止回退硬编码）
- `messages/*.json5` 必须保持 **JSON 兼容写法**（双引号、无尾逗号）：paraglide 编译器按严格 JSON 解析，无引号键/尾逗号/单引号会直接编译失败
- 修改词表后需重新编译生成 `src/lib/paraglide` 产物：`pnpm --filter @myblog/web exec paraglide-js compile --project ./project.inlang --outdir ./src/lib/paraglide`，等价入口为 `pnpm --filter @myblog/web run generate:i18n`（`vite dev`/`build` 亦会自动触发）

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

- **架构铁律**: 依赖方向、类型唯一真相源等结构性约束见 [`architecture-rules.md`](./architecture-rules.md) 与根 `AGENTS.md` 第 1 节
- **事务化模式**: 多段写库必须包裹单事务；repository 提取 `xxxTx(db *gorm.DB, ...)` 私有方法供事务与独立公开方法复用（`createTx`/`syncTagsTx`/`syncCategoriesTx`/`updateTx`/`recordArticleView` 先例，事务内所有查询与写入均走事务句柄，禁止混用 `r.db`），跨仓储共享的写逻辑下沉为包内共享 helper（`upsertContentStat` 先例），不引入仓储间依赖
- **安全配置默认值双源对齐**: `security.login_lockout` 的 viper 默认值与 service 层 `defaultLoginLockoutPolicy` 必须一致（5 次 / 15 分钟），修改任一处必须同步另一处

### 2. 错误处理

- **错误分档与统一响应**: 哨兵错误到响应码的映射口径与响应信封结构见根 `AGENTS.md` 第 5 节，唯一映射点为 `internal/handler/error.go`，本节不复述。

### 3. 性能优化

- **数据库连接池**: 连接数经 `server/configs/config.yaml` 的 `database.max_idle_conns` / `database.max_open_conns` 配置，默认 10 / 100，同值在 `internal/config/config.go` 经 viper 声明
- **渲染结果缓存**: 文章、评论与多语言正文的渲染后 HTML 以 `contentHtml` 列落库缓存，存量空缓存由 `ensureContentHTML` 按需补渲染，避免请求路径重复渲染 Markdown
- **分页查询**: 列表接口统一经请求体传入 `page` / `pageSize`，禁止一次性全量返回

### 4. 安全考虑

- **输入验证与注入防护**: 入参经 `binding` tag 完成必填、长度、格式与枚举校验（口径见根 `AGENTS.md` 第 5 节「校验」）；数据访问一律经 GORM 参数化查询，禁止字符串拼接 SQL
- **跨域配置**: 跨域来源一律改 `server/configs/config.yaml` 的 `cors` 节白名单（精确匹配），禁止代码内硬编码 Origin 或恢复全放行；生产同源网关形态白名单保持为空
- **WAF 阻止模式**: 新增或修改 `getDefaultBlockedPatterns()` 模式必须先红后绿配"攻击拦截 + 误伤回归"两组用例，模式经词首边界或取值上下文锚定，禁止回退宽匹配
- **限流键格式**: 用户级限流键统一 `user:<十进制ID>`（经 `getUserKey` 构造），禁止 rune 转换或字符串直拼产生非法键
- **认证协议变更**: token 形状、刷新、撤销语义变更属最高风险契约变更，须先更新 `contracts/auth-protocol.md` 再动代码（流程见 `architecture-rules.md` §6.3）
- **敏感信息**: 不在代码中硬编码密钥；数据库口令经 `MYBLOG_DATABASE_PASSWORD` 环境变量注入。令牌为服务端签发的不透明随机串，不存在签名密钥配置项
- **隐私告知同步**: 后端新增或启用的数据收集场景（如评论 IP/UA 写入、新增埋点）必须先同步更新 web 端 `/privacy` 隐私政策页告知再上线，隐私页已写明"新增收集场景前先更新本页"
- **设置键有效登记**: 后端业务新消费某个设置键时，必须同步补录 admin `lib/constants/setting.ts` 的 `EFFECTIVE_SETTING_KEYS`（未登记项在设置中心显示"未生效"角标）；安全分组的真实生效渠道为 `config.yaml` 与安全中间件

### 5. 测试纪律

- **先红后绿**: 修复类任务先写失败测试再改实现，确保用例真实锚定缺陷而非实现细节
- **替身覆写**: service 层 fake 以内嵌空接口继承全部方法，接口新增方法被既有测试路径调用时必须为受影响 fake 显式覆写，否则以 nil panic 暴露
- **sqlmock 维护**: repository 层 sqlmock 断言与 GORM 生成的 SQL 文本强耦合，GORM 升级或查询改写时须同步维护期望；匹配 SQL 用 `regexp.QuoteMeta` 前缀正则，期望序列须与事务内实际步骤（含 DELETE、计数 UPDATE）完全一致，否则以"剩余期望未匹配"失败
- **契约金样本维护**: 金样本须与 Go 响应逐字节一致（`json.Compact` 豁免空白），修改实现输出字段时必须同步修订金样本并跑 `contract:check`。易错点：`datetime.JSONDate` 序列化为零值 `null`、午夜日期-only、非午夜 `"2006-01-02 15:04:05"`（非 RFC3339）；`gin.H` map 序列化按字典序排键；Go `encoding/json` 默认 HTML 转义（`<` `>` 输出 `\u003c`），金样本取值应避开尖括号；`omitempty` 对 nil 指针与空切片省略键。锁②为 `WidenLiteral`+`toExtend` 单向合法实例校验（JSON 单值无法与 nullable 联合精确相等，见 architecture-rules.md §3.4）
- **gin 中间件测试**: 中间件函数内 `return` 不会中断 gin 的 handler 链，中断必须调用 `c.Abort()`；`IdentityProvider.Resolve` 失败时由实现写响应并 `Abort`，消费方 `if err != nil { return }` 依赖此语义，测试替身须模拟"写响应 + Abort"而非裸返错误。统一响应信封固定 HTTP 200，权限/认证失败以响应体业务码（`code` 字段）断言，不比较 HTTP 状态码
- **依赖基线检查**: 相对 pathspec 必须在 `server/` 下执行：`git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router ':(exclude)*_test.go'`。在仓库根用同一段相对 pathspec 会零命中而得到假绿，仓库根执行须改写为 `-- server/internal/service server/internal/middleware server/internal/router ':(exclude)*_test.go'`。基线 **13 个文件**（service 12 + middleware 仅 `identity.go` + router 0），只减不增；`*_test.go` 须排除，否则计数会包含测试替身文件

### 5.1 质量门禁

- **契约门禁**: `pnpm run contract:check` = Go handler fixture 测试（锁①）+ `@myblog/api` 的 `tsc --noEmit`（锁② typecheck）+ vitest 类型锚定；契约相关改动必须通过
- **CI**: `.github/workflows/ci.yml` 为质量门禁流水线，触发方式为 push 到 master/main 以及所有 pull request；步骤依次为 checkout、setup-node 22 + corepack、setup-go（读 `server/go.mod`）、`pnpm install --frozen-lockfile`、后端 go 三连（`go build` / `go vet` / `go test -count=1`）、`go:vulncheck`、`audit:deps`、依赖方向基线 grep（在 `server/` 内执行并断言命中数 ≤ 13）、`test:packages`、生成前台 i18n 产物（`pnpm --filter @myblog/web run generate:i18n`）、`test:apps`、`check`、`contract:check`、`build:apps`、`measure:assets`、`check:fonts`、`format:check`、`lint`；工作流不含自动部署步骤
- **依赖漏洞扫描**: Go 侧经 `pnpm run go:vulncheck` 运行固定版本的 `govulncheck`，判定口径为「存在可达调用链即失败」；前端与工具链侧经 `pnpm run audit:deps` 运行 `pnpm audit`，按 high/critical 阻塞、medium 及以下告警的阈值执行。扫描数据源与已知漏洞的收敛清单见 `scripts/README.md`；依赖升级由 `.github/dependabot.yml` 按周提交 PR

### 6. 可维护性

- **文档更新**: 架构规则变更时按 [`architecture-rules.md`](./architecture-rules.md) 同步规则本体与第 8 节债务索引；债务条目的状态与基线数值按 [`debt/README.md`](./debt/README.md) 的维护纪律更新对应分册，历史分期状态更新第 9 节
- **自动化测试**: 后端关键逻辑配套 `*_test.go`（service/repository 层已有存量）；前端 packages 公共逻辑与页面状态模块应补 `*.test.ts`。应用侧单测经根 `test:apps` 汇聚（`@myblog/web` 覆盖 markdown 渲染守卫、日期格式化与服务端会话 Cookie 转发，`@myblog/admin` 覆盖字典页状态模块），并与 `test:packages` 一样纳入 `pnpm run test` 与 CI。

## 更多资源

- [Go 官方文档](https://golang.org/doc/)
- [SvelteKit 文档](https://svelte.dev/docs/kit)
- [GORM 文档](https://gorm.io/docs/)
- [TailwindCSS 文档](https://tailwindcss.com/docs)
