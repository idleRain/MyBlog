# MyBlog 个人博客应用

🚧 **项目正在建设中** 🚧

::: tip
小孩子不懂事写着玩的
:::

> 一个个人博客应用，采用 Go + MySQL + SvelteKit 全栈技术栈构建。

## ✨ 特性

- 🎨 **现代化设计** - 基于 SvelteKit 的响应式前端界面
- ⚡ **高性能后端** - Go 语言构建的 RESTFUL API 服务

## 🏗️ 项目架构

```
MyBlog/
├── server/          # Go 后端服务
│   ├── api/         # API 路由和处理器
│   ├── internal/    # 内部业务逻辑
│   ├── pkg/         # 公共包和工具
│   └── configs/     # 配置文件
├── web/             # SvelteKit 前端应用
│   ├── src/         # 源代码
│   ├── static/      # 静态资源
└── scripts/         # 构建和部署脚本
```

## 🛠️ 技术栈

### 后端
- **Go 1.23+** - 高性能后端语言
- **Gin** - 轻量级 Web 框架
- **GORM** - ORM 数据库操作
- **MySQL** - 关系型数据库

### 前端
- **SvelteKit** - 现代化前端框架
- **TypeScript** - 类型安全的 JavaScript
- **Tailwind CSS** - 实用优先的 CSS 框架
- **Vite** - 快速构建工具

### 开发工具
- **Bun** - 快速的 JavaScript 运行时
- **ESLint** - 代码质量检查
- **Prettier** - 代码格式化
- **golangci-lint** - Go 代码检查

## 🚀 快速开始

### 环境要求
- Go 1.23+ 
- Bun 1.1.24+
- MySQL 8.0+

### 安装依赖
```bash
# 克隆项目
git clone <repository-url>
cd MyBlog

# 安装项目依赖
bun run deps
```

### 开发环境
```bash
# 同时启动前后端开发服务
bun run dev

# 或者分别启动
bun run dev:server  # 启动后端服务 (端口: 8080)
bun run dev:web     # 启动前端服务 (端口: 5173)
```

### 构建部署
```bash
# 构建生产版本
bun run build

# 代码质量检查
bun run quality
```

## 📋 开发进度

- [x] 项目架构搭建
- [x] 开发环境配置
- [x] 构建脚本完善
- [ ] 数据库设计
- [ ] 用户认证系统
- [ ] 博客文章 CRUD
- [ ] 前端界面设计
- [ ] Markdown 编辑器
- [ ] 评论系统
- [ ] 搜索功能
- [ ] 部署配置

## 🔧 可用脚本

```bash
# 开发相关
bun run dev          # 启动开发环境
bun run build        # 构建生产版本
bun run test         # 运行测试

# 代码质量
bun run lint         # 代码检查
bun run format       # 代码格式化
bun run quality      # 完整质量检查

# Go 工具
bun run go:lint-install  # 安装 Go 代码检查工具
bun run go:quality       # Go 代码质量检查

# 清理
bun run clean        # 清理构建文件
```

## 📝 配置文件

- **后端配置**: `server/configs/config.yaml`
- **前端环境**: `web/.env`
- **数据库配置**: `server/configs/database.yaml`
