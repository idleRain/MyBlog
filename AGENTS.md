# MyBlog 项目 Agent 指南

> 本文件为 AI Agent 在本仓库工作时提供项目级上下文与约定。
> 用户全局工作规范见 `~/.dsh/AGENTS.md`，两者冲突时以更具体的规则为准。

## 1. 项目概览

MyBlog 是一个 Monorepo 全栈个人博客应用，采用 Go + SvelteKit 技术栈构建。

```
MyBlog/
├── apps/
│   ├── web/                  # 前台应用（Svelte 5 + TS + TailwindCSS v4，公开博客 + demo + i18n）
│   └── admin/                # 后台应用（SvelteKit，管理控制台 + 登录页，无 i18n）
├── packages/
│   ├── shared/               # 公共纯工具与通用类型
│   ├── http/                 # HTTP 请求器（ky 封装，认证回调注入）
│   ├── api/                  # 后端接口模块与响应类型
│   └── ui/                   # shadcn-svelte 基础组件（stock，主题注入）
├── server/                   # Go 后端服务（Gin + GORM + MySQL）
├── scripts/                  # 跨项目构建/开发脚本（Node.js + tsx）
├── docs/                     # 数据库架构、开发文档、UI 设计系统、schema.sql
├── .husky/                   # Git hooks（commitlint + lint-staged）
└── package.json              # monorepo 根，pnpm workspaces = ["apps/*", "packages/*"]
```

- 包管理：**pnpm**；脚本运行时：**Node.js + tsx**；后端：Go 1.23。
- 配置承载于 `apps/web/.env`、`apps/admin/.env` 与 `server/configs/config.yaml`，默认端口：前台 8899、后台 9988、后端 3000。

## 2. 常用命令

均在仓库根目录通过 `pnpm run <script>` 执行（根 `package.json` 的 scripts）。

```bash
pnpm run setup        # 一键环境设置
pnpm run dev          # 智能启动（含环境检查、端口检查、健康监控）
pnpm run dev:server   # 仅 Go 后端（dev.ts --server）
pnpm run dev:web      # 仅前台 SvelteKit 应用（dev.ts --web）
pnpm run dev:admin    # 仅后台 SvelteKit 应用（dev.ts --admin）

pnpm run build        # 生产构建（可加 --clean / --production / --server-only / --web-only / --skip-tests --skip-lint）
pnpm run build:server # 仅构建 Go 后端二进制
pnpm run build:web    # 仅构建前端静态文件
pnpm run build:clean  # 清理构建产物后构建
pnpm run build:fast   # 跳过测试与 lint 的快速构建
pnpm run test         # test:server + test:web
pnpm run lint         # lint:web + lint:server（go vet + golangci-lint）
pnpm run format       # format:web + format:server
pnpm run quality      # format + lint + test
pnpm run check        # 前后台 SvelteKit 类型检查（svelte-check）
pnpm run clean        # 清理前后端构建产物
pnpm run seed:admin   # 初始化或提升超级管理员账户，命令幂等

# Go 专项
pnpm run go:lint-install / go:quality
# 数据库迁移
pnpm run migrate [create|up|down|version|help]
```

> `scripts/` 下各脚本的用途、参数与 pnpm 入口对照见 `scripts/README.md`。

## 3. 代码质量与格式化约定

- 根 `prettier.config.js`：`semi: false`、`singleQuote: true`、`arrowParens: 'avoid'`、`printWidth: 100`、`tabWidth: 2`、`trailingComma: 'none'`。
- 根 `eslint.config.js` 导出 `baseConfig` 供子项目继承；`apps/*/eslint.config.js` 在其上叠加 Svelte/TS 规则。
- Git hooks：`commitlint`（conventional commits）与 `lint-staged`（对 `apps/**/src/**` 运行 prettier，对 `server/**/*.go` 运行 `gofmt`/`goimports`）。提交信息需符合 conventional commits 规范。
- 更改文件后应运行 `pnpm run lint` 与 `pnpm run format` 保持静态零告警。
- 每完成一个对应功能变更后，使用**简体中文**编写符合 conventional commits 规范的提交信息，并保证提交颗粒度，若单次提交跨度较大需补充 message 描述正文，type 枚举以 `commitlint.config.js` 为准。
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
- **配置**：`internal/config` 包通过 Viper 读取 `configs/config.yaml`，所有配置项均以 YAML 为唯一来源。

## 5. 前端约定（apps/）

- **框架**：SvelteKit + Svelte 5 + TypeScript，`<script setup>` 等价于 Composition API 风格（Svelte 5 runes），**不使用 Options API**。应用分 `apps/web`（前台 toC）与 `apps/admin`（后台 toB）。
- **组件库**：shadcn-svelte 基础组件统一位于 `packages/ui`（保持 stock，主题经各应用 `app.css` token 注入），以 `$ui` 别名引入，由根级 eslint/prettier 排除。别名见各应用 `svelte.config.js`：`$lib`、`$ui`、`$i18n`（仅前台，paraglide messages）、`@/*`（`src/*`）、`#/*`、`~/*`。
- **样式**：TailwindCSS v4（`@tailwindcss/vite`）；前台 `apps/web/src/app.css`（规格书主题，含 `--signal`）、后台 `apps/admin/src/app.css`（原始主题，无 `--signal`）；`packages/ui` 不携带全局样式。
- **API 层**：`packages/http` 提供 `createHttpClient` 工厂，`packages/api` 提供 `createUserAPI` 工厂；应用侧 `src/lib/service` 注入认证与提示回调，`src/lib/api` 实例化接口，一律使用 `POST` 调用后端接口，与后端 POST-Only 规范呼应。
- **状态**：`src/lib/stores/auth.ts` 认证 store 两应用各保留一份；认证域其余代码（`constants`/`guards`/`utils/auth*`/`jwt`/`logout`/`permissions`）仅内置 `apps/admin`。
- **路由**：前台 `src/routes` 使用分组路由 `(app)`、`demo`；后台使用 `(admin)`、`(auth)`（登录页归属后台）。服务端逻辑用 `+page.server.ts`，客户端逻辑用 `+page.ts` / `+page.svelte`。
- **i18n**：仅前台 `apps/web` 使用 `@inlang/paraglide-js`，`project.inlang`/`messages/` 目录，别名 `$i18n`；后台不引入 i18n。
- **类型**：`src/types`、`src/lib/types`（`.d.ts`）；用户/接口类型统一来自 `@myblog/api`。
- 前端代码改动需运行对应应用 `cd apps/web && pnpm run check` 或 `cd apps/admin && pnpm run check`（svelte-check + svelte-kit sync）。

## 6. 注释与代码规范硬约束

遵循全局 `~/.dsh/AGENTS.md`，在此强调项目内高频要求：

- 开发过程中中遇到可维护性差、建议重构或优化的代码，不要忽略，而是寻求是否确认是否进行顺带优化。
- 非异步场景调用 async 函数需用 `void` 显式忽略返回的 Promise。

## 7. 环境配置

- 后端：`server/configs/config.yaml`（数据库、服务器、日志、API、JWT、安全配置）。默认 MySQL `blog` 库，`root/123456`，含校验与沙箱占位。
- 前端：`apps/web/.env`（`VITE_SERVER_PORT=8899`）与 `apps/admin/.env`（`VITE_SERVER_PORT=9988`），均含 `VITE_PROXY_URL=http://localhost:3000`、`VITE_BASE_URL=/api`、`VITE_REQUEST_TIMEOUT=15000`。

## 8. 开发进度概览

已完成：Monorepo 架构、环境与工具链、Git hooks 与质量保证、智能开发脚本与监控、用户管理系统、用户认证与授权、RBAC 权限体系、博客文章 CRUD 与状态及互动管理。其中认证授权包含 JWT 双 token 刷新、auth 中间件与登录、刷新、登出接口；文章模块包含发布、归档、私有化、点赞、收藏与浏览统计。
待办：Markdown 编辑器、评论、全文搜索、文件上传与图片管理、响应式前端完善、部署与 Docker。
