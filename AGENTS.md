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
bun run build:server # 仅构建 Go 后端二进制
bun run build:web    # 仅构建前端静态文件
bun run build:clean  # 清理构建产物后构建
bun run build:fast   # 跳过测试与 lint 的快速构建
bun run test         # test:server + test:web
bun run lint         # lint:web + lint:server（go vet + golangci-lint）
bun run format       # format:web + format:server
bun run quality      # format + lint + test
bun run check        # 即 lint
bun run clean        # 清理前后端构建产物
bun run deps         # 安装前后端全部依赖
bun run seed:admin   # 初始化或提升超级管理员账户，命令幂等

# Go 专项
bun run go:lint-install / go:quality
# 数据库迁移
bun run migrate [create|up|down|version|help]
```

## 3. 代码质量与格式化约定

- 根 `prettier.config.js`：`semi: false`、`singleQuote: true`、`arrowParens: 'avoid'`、`printWidth: 100`、`tabWidth: 2`、`trailingComma: 'none'`。
- 根 `eslint.config.js` 导出 `baseConfig` 供子项目继承；`web/eslint.config.js` 在其上叠加 Svelte/TS 规则。
- Git hooks：`commitlint`（conventional commits）与 `lint-staged`（对 `web/src/**` 运行 prettier，对 `server/**/*.go` 运行 `gofmt`/`goimports`）。提交信息需符合 conventional commits 规范。
- 更改文件后应运行 `bun run lint` 与 `bun run format` 保持静态零告警。
- 每完成一个对应功能变更后，使用**简体中文**编写符合 conventional commits 规范的提交信息，并保证提交颗粒度，type 枚举以 `commitlint.config.js` 为准。
- 关键逻辑应配套单元测试，测试文件与被测文件同目录命名，Go 为 `*_test.go`，前端为 `*.test.ts`。当前仓库尚未引入测试，新增功能时应一并补齐。

## 4. 后端约定（server/）

- **POST-Only 规范**：后端业务接口遵循 `POST-Only` 规范，一律通过 `POST` 方法注册，查询类接口同样使用 `POST`，不使用 `GET`、`PUT`、`DELETE` 等方法。请求参数统一承载于 JSON 请求体，纯 id 类短参数可放在 path 中。
- **接口健壮性**：后端接口必须保证健壮性，涉及外部输入、用户输入的接口不得省略必要校验。请求参数在 handler 层通过 `ShouldBindJSON` 与 `binding` tag 完成必填、长度、格式、枚举等校验，业务规则在 service 层校验，校验全部通过后才进入数据访问层。
- **公共校验方法**：跨接口复用的校验逻辑必须提取为公共方法或工具函数，不得仅内联在单个 handler 或 service 局部。现有私有校验如密码强度校验，后续改动时应迁移为公共校验工具。
- **分层架构**：`handler`（HTTP 层，参数校验）→ `service`（业务逻辑）→ `repository`（数据访问）。路由注册见 `internal/router/router.go`。
- 依赖注入集中在 `router.Dependencies` 结构，handler 通过接口（如 `UserHandlerInterface`、`ArticleHandlerInterface`）注入。
- **统一响应**：使用 `pkg/response` 包，响应结构为 `{ code, message, data? }`；预定义响应码 `CodeSuccess=200`、`CodeError=500`、`CodeInvalid=400`、`CodeAuth=401`、`CodeForbid=403`、`CodeNotFound=404`。统一通过 `response.Success` / `response.Error` 等函数返回。
- **中间件**：`middleware` 包含 logger、request ID、CORS、汇总安全（限流、安全头、输入校验、管理员接口校验）、auth、rbac、ratelimit。
- **模型**：`internal/model`（GORM 实体）、`internal/repository`（含请求/响应结构体`create/update` 请求与 `User` 领域对象）、`internal/service`（业务与 JWT、RBAC）。
- **入口与工具包**：服务入口在 `cmd/myblog/main.go`，种子脚本入口在 `cmd/seed/main.go`，种子逻辑见 `internal/database/seed.go`；通用包 `pkg/response` 与 `pkg/datetime`。
- **数据库**：MySQL 单库（GORM）；迁移由 `internal/database/migrate.go` 与根 `migrate` 脚本管理。模型见 `internal/model/*.go`，架构细节见 `docs/database-architecture.md`。
- **数据库设计可持续性**：表结构以可长期健康演进为目标设计，不得只象征性定义 ID、名称、时间等少数字段应付了事。每张业务表须具备完整的生命周期字段、状态字段、业务字段与必要的预留扩展字段，且每个字段均带 `comment` 说明业务含义，字段类型与长度须贴合真实数据需求。
- **规范化与索引**：遵循第三范式，多对多关系使用独立关联表，树形结构使用 `parent_id`、`root_id`、`level` 字段；唯一性字段加唯一索引，外键与高频查询字段加普通索引，禁止无索引的大表查询。
- **数据一致性**：显式声明 GORM 关联关系与外键删除策略，如 `OnDelete:CASCADE` 与 `OnDelete:SET NULL`；状态类字段使用命名常量枚举，时间字段统一为 `datetime(3)` 精度。
- **密码**：bcrypt（成本常量 `BcryptCost = 12`），密码强度校验在 `service/user.go`。
- **命名**：Go 结构体字段与 JSON tag 使用小驼峰；接口与实现同包定义，命名统一为 `XxxInterface` 后缀，如 `ArticleHandlerInterface`、`ArticleServiceInterface`、`ArticleRepositoryInterface`。各层只依赖接口而非具体实现。
- **接口定义位置**：接口统一声明在各层实现所在包，即 `handler`、`service`、`repository` 内；`router` 只引用各层接口完成依赖注入与类型断言，不得在 `router` 包重复定义接口。
- **配置**：`internal/config` 包通过 Viper 读取 `configs/config.yaml`，支持 `${ENV:default}` 变量替换，如 JWT 密钥。

## 5. 前端约定（web/）

- **框架**：SvelteKit + Svelte 5 + TypeScript，`<script setup>` 等价于 Composition API 风格（Svelte 5 runes），**不使用 Options API**。
- **组件库**：shadcn-svelte，组件位于 `src/lib/components/ui/*`（已忽略 ESLint）。别名见 `svelte.config.js`：`$lib`、`$ui`、`$i18n`（paraglide messages）、`@/*`（`src/*`）、`#/*`、`~/*`。
- **样式**：TailwindCSS v4（`@tailwindcss/vite`），`src/app.css`；已引入 `tailwind-merge`、`tailwind-variants`、`tw-animate-css`。
- **API 层**：`src/lib/service` 为核心请求封装（基于 `ky`），`src/lib/utils/request.ts` 提供 `requestWithRetry`（401 自动刷新令牌并重试）、`safeApiCall`（统一错误处理与 toast）、`isApiSuccess`、`extractApiData`。模块化 API 见 `src/lib/api/modules/*`，一律使用 `POST` 调用后端接口，与后端 POST-Only 规范呼应。
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

已完成：Monorepo 架构、环境与工具链、Git hooks 与质量保证、智能开发脚本与监控、用户管理系统、用户认证与授权、RBAC 权限体系、博客文章 CRUD 与状态及互动管理。其中认证授权包含 JWT 双 token 刷新、auth 中间件与登录、刷新、登出接口；文章模块包含发布、归档、私有化、点赞、收藏与浏览统计。
待办：Markdown 编辑器、评论、全文搜索、文件上传与图片管理、响应式前端完善、部署与 Docker。
