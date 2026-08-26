# MyBlog 项目 Agent 指南

> 本文件为 AI Agent 在本仓库工作时提供项目级上下文与约定。
> 用户全局工作规范见 `~/.dsh/AGENTS.md`，两者冲突时以更具体的规则为准。

## 1. 项目概览

MyBlog 是一个 Monorepo 全栈个人博客应用，采用 Go + SvelteKit 技术栈构建。

```
MyBlog/
├── server/               # Go 后端服务（Gin + GORM + MySQL）
├── web/                  # SvelteKit 前端应用（Svelte 5 + TS + TailwindCSS v4 + shadcn-svelte）
├── scripts/              # 跨项目构建/开发脚本（Bun/TypeScript）
├── docs/                 # 数据库架构、开发文档、schema.sql
├── .husky/               # Git hooks（commitlint + lint-staged）
└── package.json          # monorepo 根，workspaces = ["web"]
```

- 包管理/运行时：**Bun**（根脚本）、Go 1.23（后端）。
- 配置承载于 `web/.env` 与 `server/configs/config.yaml`，默认端口：前端 8899、后端 3000。

## 2. 常用命令

均在仓库根目录通过 `bun run <script>` 执行（根 `package.json` 的 scripts）。

```bash
bun run setup        # 一键环境设置
bun run dev          # 智能启动（含环境检查、端口检查、健康监控）
bun run dev:simple   # concurrently 并行启动 server 与 web
bun run dev:server   # 仅 Go 后端热更新（air watcher）
bun run dev:web      # 仅 SvelteKit 前端

bun run build        # 生产构建（可加 --clean / --production / --server-only / --web-only / --skip-tests --skip-lint）
bun run test         # test:server + test:web
bun run lint         # lint:web + lint:server（go vet + golangci-lint）
bun run format       # format:web + format:server
bun run quality      # format + lint + test
bun run check        # 即 lint

# Go 专项
bun run go:lint-install / go:quality
# 数据库迁移
bun run migrate [create|up|down|version|help]
```

## 3. 代码质量与格式化约定

- 根 `prettier.config.js`：`semi: false`、`singleQuote: true`、`arrowParens: 'avoid'`、`printWidth: 100`、`tabWidth: 2`、`trailingComma: 'none'`。
- 根 `eslint.config.js` 导出 `baseConfig` 供子项目继承；`web/eslint.config.js` 在其上叠加 Svelte/TS 规则。
- Git hooks：`commitlint`（conventional commits）与 `lint-staged`（对 `web/src/**` 运行 prettier，对 `server/**/*.go` 运行 `gofmt`/`goimports`）。提交信息需符合 conventional commits 规范。
- 更改文件后应运行 `bun run lint` 与 `bun run format` 保持静态零告警。项目名非 `lotus` 前缀，必须执行 lint 与 check。

## 4. 后端约定（server/）

- **分层架构**：`handler`（HTTP 层，参数校验）→ `service`（业务逻辑）→ `repository`（数据访问）。路由注册见 `internal/router/router.go`。
- 依赖注入集中在 `router.Dependencies` 结构，handler 通过接口（如 `UserHandlerInterface`、`ArticleHandlerInterface`）注入。
- **统一响应**：使用 `pkg/response` 包，响应结构为 `{ code, message, data? }`；预定义响应码 `CodeSuccess=200`、`CodeError=500`、`CodeInvalid=400`、`CodeAuth=401`、`CodeForbid=403`、`CodeNotFound=404`。统一通过 `response.Success` / `response.Error` 等函数返回。
- **中间件**：`middleware` 包含 logger、request ID、CORS、汇总安全（限流、安全头、输入校验、管理员接口校验）、auth、rbac、ratelimit。
- **模型**：`internal/model`（GORM 实体）、`internal/repository`（含请求/响应结构体`create/update` 请求与 `User` 领域对象）、`internal/service`（业务与 JWT、RBAC）。
- **数据库**：MySQL 单库（GORM）；迁移由 `internal/database/migrate.go` 与根 `migrate` 脚本管理。
- **密码**：bcrypt（成本常量 `BcryptCost = 12`），密码强度校验在 `service/user.go`。
- **命名**：Go 结构体字段与 JSON tag 使用小驼峰；接口以 `XXX` 形式声明并只依赖接口而非具体实现。
- **配置**：Viper 读取 `configs/config.yaml`，支持 `${ENV:default}` 变量替换（如 JWT 密钥）。

## 5. 前端约定（web/）

- **框架**：SvelteKit + Svelte 5 + TypeScript，`<script setup>` 等价于 Composition API 风格（Svelte 5 runes），**不使用 Options API**。
- **组件库**：shadcn-svelte，组件位于 `src/lib/components/ui/*`（已忽略 ESPLint）。别名见 `svelte.config.js`：`$lib`、`$ui`、`$i18n`（paraglide messages）、`@/*`（`src/*`）、`#/*`、`~/*`。
- **样式**：TailwindCSS v4（`@tailwindcss/vite`），`src/app.css`；已引入 `tailwind-merge`、`tailwind-variants`、`tw-animate-css`。
- **API 层**：`src/lib/service` 为核心请求封装（基于 `ky`），`src/lib/utils/request.ts` 提供 `requestWithRetry`（401 自动刷新令牌并重试）、`safeApiCall`（统一错误处理与 toast）、`isApiSuccess`、`extractApiData`。模块化 API 见 `src/lib/api/modules/*`。
- **状态**：`src/lib/stores`（如 `auth`），token 刷新/失效处理见 `src/lib/utils/auth.ts`、`jwt.ts`、`logout.ts`。
- **路由**：`src/routes` 使用分组路由 `(admin)`、`(app)`、`(auth)`、`demo`；服务端逻辑用 `+page.server.ts`，客户端逻辑用 `+page.ts` / `+page.svelte`。
- **i18n**：`@inlang/paraglide-js`，`messages/` 目录，别名 `$i18n`。
- **类型**：`src/types`、`src/lib/types`（`.d.ts`）。
- 前端代码改动需运行 `cd web && bun run check`（svelte-check + svelte-kit sync）。

## 6. 注释与代码规范硬约束

遵循全局 `~/.dsh/AGENTS.md`，在此强调项目内高频要求：

- 与用户对话使用**简体中文**；代码注释同样使用简体中文。
- 注释用完整技术陈述句并以句号结尾，不使用括号补充解释，不使用口语化/情绪化用词（如「兜底」「保证」「弄好」）。
- 禁止魔法数字/字符串，提升为有意义的命名常量。
- 函数遵守单一职责，函数体建议不超过 30 行；嵌套不超过 3 层。
- 命名自解释，禁止无意义缩写。
- 涉及外部 I/O、网络、用户输入处必须显式错误处理，不得使用裸的 `try-except: pass`。
- 非异步场景调用 async 函数需用 `void` 显式忽略返回的 Promise。

## 7. 环境配置

- 后端：`server/configs/config.yaml`（数据库、服务器、日志、API、JWT、安全配置）。默认 MySQL `blog` 库，`root/123456`，含校验与沙箱占位。
- 前端：`web/.env`（`VITE_SERVER_PORT=8899`、`VITE_PROXY_URL=http://localhost:3000`、`VITE_BASE_URL=/api`、`VITE_REQUEST_TIMEOUT=15000`）。
- 后端示例环境变量见 `server/.env.example`。

## 8. 开发进度概览

已完成：Monorepo 架构、环境与工具链、Git hooks 与质量保证、智能开发脚本与监控、用户管理系统基础功能。
待办：用户认证与授权、博客文章 CRUD、Markdown 编辑器、评论、搜索、文件上传与图片管理、响应式前端、部署与 Docker。
