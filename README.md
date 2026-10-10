# MyBlog 个人博客应用

> [!TIP] 
> 🚧 **项目正在建设中** 🚧
> 小孩子不懂事写着玩的

> 一个 Monorepo 全栈个人博客应用，采用 Go + SvelteKit 技术栈构建

[![Go Version](https://img.shields.io/badge/Go-1.26.9-blue.svg)](https://golang.org)
[![SvelteKit](https://img.shields.io/badge/SvelteKit-Latest-orange.svg)](https://svelte.dev/docs/kit)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-blue.svg)](https://www.typescriptlang.org)

## ✨ 特性

- 🏗️ **Monorepo 架构** - 统一管理前后端代码和依赖
- 🎨 **现代化设计** - 基于 SvelteKit 的响应式前端界面（前台 SSR 编辑杂志主题，后台 SPA 控制台）
- ⚡ **高性能后端** - Go 语言构建的 POST-only API 服务（12 个业务模块，104 条路由）
- 🔐 **用户认证与授权** - 不透明双令牌 + HttpOnly Cookie 会话、RBAC 权限下发、登录失败锁定
- 🌏 **内容多语言** - 文章/分类/标签翻译表，按 `Accept-Language` 协商输出；前台界面经 paraglide 取词
- 📝 **完整内容管理** - 文章 CRUD、发布归档、分类标签、评论互动
- 🖼️ **媒体资源管理** - 文件上传、哈希秒传去重、本地存储
- ⚙️ **站点运营** - 系统设置、友情链接、站点统计、站内通知
- 🔧 **智能开发工具** - 自动化环境检查、热更新和健康监控
- 📏 **代码质量保证** - 集成 ESLint、Prettier、golangci-lint、契约金样本门禁与依赖漏洞扫描
- 🐙 **Git Hooks** - 自动代码检查和提交规范验证

## 🏗️ 项目架构

```
MyBlog/                   # Monorepo 根目录
├── apps/                 # 前端应用
│   ├── web/              # 前台应用（SvelteKit SSR，公开博客 + demo + i18n）
│   └── admin/            # 后台应用（SvelteKit SPA，管理控制台 + 登录页，基准路径 /admin）
├── packages/             # 公共包
│   ├── shared/           # 纯工具与通用类型（ApiResponse、响应码常量）
│   ├── http/             # HTTP 请求器（ky 封装）
│   ├── api/              # 后端接口模块（12 个模块工厂）
│   ├── auth/             # 认证会话组装（createAuthStore 工厂）
│   └── ui/               # shadcn-svelte 基础组件（stock）
├── server/               # Go 后端服务
│   ├── cmd/myblog/       # 应用程序入口与组合根（main.go 生命周期、deps.go 依赖装配）
│   ├── cmd/seed/         # 超级管理员种子命令
│   ├── internal/         # 内部业务逻辑
│   │   ├── handler/      # HTTP 请求处理层
│   │   ├── service/      # 业务逻辑层
│   │   ├── repository/   # 数据访问层
│   │   ├── domain/       # 领域实体与请求/响应 DTO（唯一类型语言）
│   │   ├── model/        # GORM 实体
│   │   ├── router/       # 路由注册
│   │   ├── middleware/   # 中间件（认证、权限、限流、CORS、安全、语言等）
│   │   ├── config/       # 配置管理
│   │   └── database/     # 数据库连接与迁移
│   ├── pkg/              # 通用包（response、datetime、slug、markdown、storage）
│   ├── configs/          # 配置文件
│   ├── migrations/       # golang-migrate 迁移文件
│   └── docs/api/         # 接口清单与逐模块说明
├── contracts/            # 认证与 i18n 协议、错误码口径、契约金样本
├── scripts/              # 跨项目构建脚本（Node.js + tsx）
├── deploy/               # 容器化部署与运维脚本
├── .husky/               # Git hooks 配置
└── docs/                 # 项目文档
```

## 🛠️ 技术栈

### 后端 (server/)

- **Go 1.26.9** - 高性能后端语言（版本唯一来源为 `server/go.mod` 的 `go` 指令）
- **Gin** - 轻量级 Web 框架
- **GORM** - ORM 数据库操作
- **MySQL 8.0** - 关系型数据库
- **Viper** - 配置管理（YAML 为唯一来源，标量项可经环境变量覆盖）

### 前端 (apps/)

- **SvelteKit + Svelte 5** - 现代化前端框架（runes 风格）
- **TypeScript** - 类型安全的 JavaScript
- **TailwindCSS v4** - 实用优先的 CSS 框架
- **Vite** - 快速构建工具

### 开发工具

- **Node.js + tsx** - JavaScript 运行时与 TypeScript 执行器
- **Husky** - Git hooks 管理
- **lint-staged** - 暂存文件代码检查
- **commitlint** - 提交信息规范验证

## 🚀 快速开始

### 环境要求
- **Go 1.26.9** - [下载安装](https://golang.org/)（以 `server/go.mod` 的 `go` 指令为准）
- **Node.js 22+** - [下载安装](https://nodejs.org/)（以仓库根 `.nvmrc` 为准，`engines.node` 为 `>=22`）
- **MySQL 8.0+** - [下载安装](https://dev.mysql.com/)

### 一键环境设置

```bash
# 克隆项目
git clone https://github.com/idleRain/MyBlog.git
cd MyBlog

# 自动环境设置 (推荐)
pnpm run setup
```

### 开发环境

```bash
# 智能启动 (推荐) - 包含环境检查、端口检查、健康监控
pnpm run dev

# 分别启动服务（同样经过环境与端口检查）
pnpm run dev:server    # 仅 Go 后端
pnpm run dev:web       # 仅前台应用
pnpm run dev:admin     # 仅后台应用
```

### 初始化管理员账户

首次部署需要创建超级管理员，未初始化时无法登录后台：

```bash
pnpm run seed:admin
# 自定义用户名、密码、邮箱
pnpm run seed:admin --username root --password Root@2025 --email root@myblog.local
```

默认账户为 `admin` / `Admin@123456` / `admin@myblog.local`；命令幂等，账户已存在且为超级管理员时不重复创建，存在但角色非超级管理员时自动提升。

### 访问应用

- **前台应用**: http://localhost:8899 (可配置)
- **后台应用**: http://localhost:9988/admin (基准路径 /admin，可配置)
- **后端 API**: http://localhost:3000 (可配置)
- **API 存活探针 (liveness)**: http://localhost:3000/api/health
- **API 就绪探针 (readiness)**: http://localhost:3000/api/health/ready（探测数据库连通性与上传目录可写性，未就绪返回 503 与原因码）

## 🔧 开发命令

### 基础命令

```bash
# 环境和依赖管理
pnpm run setup           # 初始化开发环境

# 开发和构建
pnpm run dev             # 启动开发环境 (智能模式)
pnpm run build           # 构建生产版本
pnpm run test            # 运行所有测试（后端 + packages + 两应用 + 契约门禁）

# 代码质量
pnpm run check           # 前后台 SvelteKit 类型检查
pnpm run lint            # 代码检查
pnpm run format          # 代码格式化
pnpm run quality         # 完整质量检查 (格式化 + 检查 + 测试)
```

### 专项命令

```bash
# 前端专用
pnpm run check          # 前后台 SvelteKit 类型检查（svelte-check）
pnpm run typecheck:web  # 与 check 等价的旧入口，两应用并行类型检查
pnpm run test:packages  # packages 单测与类型检查：@myblog/api + @myblog/auth + @myblog/http
pnpm run test:apps      # 两应用单测：@myblog/web + @myblog/admin（vitest）
cd apps/web && pnpm run check   # 前台 SvelteKit 类型检查
cd apps/admin && pnpm run check # 后台 SvelteKit 类型检查

# 后端专用
pnpm run test:server     # 后端测试（go test -count=1）
cd server && go test -v ./...  # 详细测试输出

# Go 工具链
pnpm run go:lint-install # 安装 Go 代码检查工具
pnpm run go:quality      # Go 完整质量检查
pnpm run go:vulncheck    # Go 依赖漏洞扫描
pnpm run audit:deps      # 前端与工具链依赖漏洞扫描

# 契约与数据库
pnpm run contract:check  # 契约三把锁（Go fixture + tsc --noEmit + vitest 类型锚定）
pnpm run migrate [create|up|down|goto|force|version]   # golang-migrate 迁移管理
```

> 前端类型检查与单元测试是两条独立入口：`check` 只做类型检查，单测经 `test:packages` 与 `test:apps` 执行，两者都在 `pnpm run test` 与 CI 中串联。
> 各脚本的用途、参数与 pnpm 入口对照见 `scripts/README.md`。

## ⚙️ 配置管理

### 环境配置

- **后端配置**: `server/configs/config.yaml`
  - 数据库、服务器、日志、API、令牌、安全、CORS、媒体、RBAC 权限表与多语言白名单
  - 标量配置项可经 `MYBLOG_` 前缀的环境变量覆盖，映射规则为配置键的大写下划线形式，例如 `database.password` 对应 `MYBLOG_DATABASE_PASSWORD`；列表与映射结构只能改 YAML
- **前端环境**: `apps/web/.env`（前台，端口 8899，另含 `VITE_ADMIN_URL`）与 `apps/admin/.env`（后台，端口 9988）
  - 生产构建经 `apps/web/.env.prod` 覆盖 `VITE_ADMIN_URL`，未配置时前台回退为同源 `/admin`

### 默认配置

```yaml
# server/configs/config.yaml（开发默认值，生产环境必须覆盖口令与 CORS 白名单）
server:
  host: "localhost"
  port: 3000
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "123456"   # 生产经 MYBLOG_DATABASE_PASSWORD 注入
  dbname: "blog"

token:
  access_expire: 15     # 分钟
  refresh_expire: 168   # 小时（7 天）
  cookie_secure: false  # 会话 Cookie 是否仅经 HTTPS 传输，生产必须置 true

media:
  upload_dir: "uploads"   # 本地媒体存储目录
  base_url: "/uploads"    # 媒体访问 URL 前缀
  max_size_mb: 10         # 单文件大小上限（MB）
```

```bash
# apps/web/.env（前台）与 apps/admin/.env（后台）端口配置示例
VITE_SERVER_PORT=8899   # 前台；后台为 9988
VITE_PROXY_URL=http://localhost:3000
VITE_BASE_URL=/api
VITE_REQUEST_TIMEOUT=15000
```

## 🏗️ 架构设计

### 后端分层架构

```
Handler Layer    → HTTP 请求处理、参数校验、DTO 映射
   ↓
Service Layer    → 业务逻辑、业务规则校验、事务编排
   ↓
Repository Layer → 数据库操作、数据访问
   ↓
Domain Layer     → 领域实体与请求/响应 DTO（全系统唯一类型语言）
```

## 📋 开发进度

- [x] Monorepo 架构搭建
- [x] 开发环境配置和工具链
- [x] Git hooks 和代码质量保证
- [x] 智能开发脚本和监控
- [x] 用户管理系统基础功能
- [x] 用户认证和授权系统（不透明双令牌、HttpOnly Cookie 会话、RBAC 权限下发、登录失败锁定、令牌撤销）
- [x] 博客文章 CRUD 功能（草稿到发布的状态流转、归档、修订记录）
- [x] 文章分类与标签管理（树形分类、热门标签）
- [x] 评论系统（审核状态机与评论开关设置）
- [x] 点赞收藏与用户关注
- [x] 搜索功能（MySQL ngram 全文索引，合并翻译表匹配）
- [x] 文件上传和图片管理（哈希秒传去重、本地存储）
- [x] 系统设置、友情链接、站点统计、站内通知
- [x] Markdown 编辑器集成
- [x] 内容多语言（文章/分类/标签翻译表，按 `Accept-Language` 协商输出）
- [x] 前台界面多语言（Header、Footer、通知铃铛、友链弹窗、错误页与 demo 页已接入 paraglide，其余页面待分批接入）
- [x] 契约金样本与类型锚定门禁、CI 质量流水线与依赖漏洞扫描
- [x] 部署配置和 Docker 支持（MySQL + 后端 + 前台 + Nginx + 备份单实例编排）
- [ ] 前端窄屏与响应式细节打磨（后台字典页工具栏窄屏遗留未完成，见 `docs/debt/frontend.md` 第 8 条）
- [ ] 扩容前置改造三项（令牌表持久化、限流器外置、补 `SetTrustedProxies`，见 `docs/architecture-rules.md` 第 7 节）
