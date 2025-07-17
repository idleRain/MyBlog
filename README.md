# MyBlog 个人博客应用

🚧 **项目正在建设中** 🚧

::: tip
小孩子不懂事写着玩的
:::

> 一个 Monorepo 全栈个人博客应用，采用 Go + SvelteKit 技术栈构建

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![SvelteKit](https://img.shields.io/badge/SvelteKit-Latest-orange.svg)](https://kit.svelte.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-blue.svg)](https://www.typescriptlang.org)

## ✨ 特性

- 🏗️ **Monorepo 架构** - 统一管理前后端代码和依赖
- 🎨 **现代化设计** - 基于 SvelteKit 的响应式前端界面
- ⚡ **高性能后端** - Go 语言构建的 RESTful API 服务
- 🔧 **智能开发工具** - 自动化环境检查、热更新和健康监控
- 📏 **代码质量保证** - 集成 ESLint、Prettier、golangci-lint
- 🐙 **Git Hooks** - 自动代码检查和提交规范验证

## 🏗️ 项目架构

```
MyBlog/                   # Monorepo 根目录
├── server/               # Go 后端服务
│   ├── cmd/myblog/       # 应用程序入口
│   ├── internal/         # 内部业务逻辑
│   │   ├── handler/      # HTTP 请求处理层
│   │   ├── service/      # 业务逻辑层
│   │   ├── repository/   # 数据访问层
│   │   ├── config/       # 配置管理
│   │   └── database/     # 数据库连接
│   ├── pkg/              # 公共包和工具
│   ├── configs/          # 配置文件
│   └── scripts/          # 后端专用脚本
├── web/                  # SvelteKit 前端应用
│   ├── src/              # 源代码
│   │   ├── routes/       # 页面路由
│   │   ├── lib/          # 可复用组件
│   │   ├── service/      # API 调用服务
│   │   └── utils/        # 工具函数
│   └── static/           # 静态资源
├── scripts/              # 跨项目构建脚本
├── .husky/               # Git hooks 配置
└── docs/                 # 项目文档
```

## 🛠️ 技术栈

### 后端 (server/)

- **Go 1.23+** - 高性能后端语言
- **Gin** - 轻量级 Web 框架
- **GORM** - ORM 数据库操作
- **MySQL** - 关系型数据库
- **Viper** - 配置管理

### 前端 (web/)

- **SvelteKit** - 现代化前端框架
- **TypeScript** - 类型安全的 JavaScript
- **TailwindCSS** - 实用优先的 CSS 框架
- **Vite** - 快速构建工具

### 开发工具

- **Bun** - 快速的 JavaScript 运行时和包管理器
- **Husky** - Git hooks 管理
- **lint-staged** - 暂存文件代码检查
- **commitlint** - 提交信息规范验证
- **Concurrently** - 并行运行多个命令

## 🚀 快速开始

### 环境要求
- **Go 1.23+** - [下载安装](https://golang.org/)
- **Bun 1.1.24+** - [下载安装](https://bun.sh/)
- **MySQL 8.0+** - [下载安装](https://dev.mysql.com/)

### 一键环境设置

```bash
# 克隆项目
git clone https://github.com/idleRain/MyBlog.git
cd MyBlog

# 自动环境设置 (推荐)
bun run setup
```

### 开发环境

```bash
# 智能启动 (推荐) - 包含环境检查、端口检查、健康监控
bun run dev

# 备选启动方式
bun run dev:simple    # 使用 concurrently 简单启动

# 分别启动服务
bun run dev:server    # Go 后端热更新
bun run dev:web       # SvelteKit 前端开发服务器
```

### 访问应用

- **前端应用**: http://localhost:8899 (可配置)
- **后端 API**: http://localhost:3000 (可配置)
- **API 健康检查**: http://localhost:3000/api/health

## 🔧 开发命令

### 基础命令

```bash
# 环境和依赖管理
bun run setup           # 初始化开发环境
bun run deps            # 安装所有依赖

# 开发和构建
bun run dev             # 启动开发环境 (智能模式)
bun run build           # 构建生产版本
bun run test            # 运行所有测试

# 代码质量
bun run lint            # 代码检查
bun run format          # 代码格式化
bun run quality         # 完整质量检查 (格式化 + 检查 + 测试)
```

### 专项命令

```bash
# 前端专用
bun run test:web        # 前端测试
cd web && bun run check # SvelteKit 类型检查

# 后端专用  
bun run test:server     # 后端测试
cd server && go test -v ./...  # 详细测试输出

# Go 工具链
bun run go:lint-install # 安装 Go 代码检查工具
bun run go:quality      # Go 完整质量检查
```

## ⚙️ 配置管理

### 环境配置

- **后端配置**: `server/configs/config.yaml`
  - 数据库连接、服务器端口、日志级别等
- **前端环境**: `web/.env`
  - API 地址、前端端口等

### 默认配置

```yaml
# server/configs/config.yaml
server:
  host: "localhost"
  port: 3000
  mode: "debug"

database:
  host: "host"
  port: 3306
  username: "username"
  password: "password"
  dbname: "blog"
```

```bash
# web/.env
VITE_SERVER_PORT=8899
VITE_PROXY_URL=http://localhost:3000
VITE_BASE_URL=/api
```

## 🔍日志查看

```bash
# 查看开发日志
bun run dev  # 实时日志输出

# Go 后端详细日志
cd server && go run scripts/watcher.go

# 前端开发日志  
cd web && bun run dev
```

## 🏗️ 架构设计

### 后端分层架构

```
Handler Layer    → HTTP 请求处理、参数验证
   ↓
Service Layer    → 业务逻辑实现、数据处理  
   ↓
Repository Layer → 数据库操作、数据访问
```

### API 设计规范

- **统一 POST 方法** - 所有接口使用 POST 请求
- **小驼峰命名** - JSON 字段使用 camelCase
- **统一响应格式** - 标准的 code、message、data 结构
- **错误处理** - 使用 `pkg/response` 统一错误响应

## 📋 开发进度

- [x] Monorepo 架构搭建
- [x] 开发环境配置和工具链
- [x] Git hooks 和代码质量保证
- [x] 智能开发脚本和监控
- [x] 用户管理系统基础功能
- [ ] 用户认证和授权系统
- [ ] 博客文章 CRUD 功能
- [ ] Markdown 编辑器集成
- [ ] 评论系统
- [ ] 搜索功能
- [ ] 文件上传和图片管理
- [ ] 响应式前端界面
- [ ] 部署配置和 Docker 支持
