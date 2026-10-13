# MyBlog 工程规范

> 本文件是 Agent 的默认加载规范，只保留结构性红线、自检口径与文档路由；完整裁决标准与历史案例见 `docs/architecture-rules.md`，债务条目见 `docs/debt/`，各层与各包的实现约束见 `server/internal/*/README.md` 与 `server/pkg/*/README.md`。
> 裁决顺序：全局规范（`~/.dsh/AGENTS.md`）< 本文件 < `docs/architecture-rules.md` 细则，冲突时以更具体者为准。
> 文档引用方向单向：本文件可引用其他文档，其他文档不得引用本文件，因此本文件的结构可自由调整。

## 1. 结构红线

违反必须返工，每条附验证方式。

1. **依赖只准向下**。后端 `handler → service → repository → domain`，禁止 service import handler、repository import service/handler、handler/middleware/router 直接 import repository；前端 `apps → @myblog/api → @myblog/http → @myblog/shared`，禁止 packages 反向 import apps，禁止页面/组件直接 import `ky`（唯一豁免：各应用 `src/lib/service/index.ts` 的令牌刷新直连）。
   验证：在 `server/` 下执行 `git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router ':(exclude)*_test.go'`，基线 13 个文件（service 12 + middleware 仅 `identity.go` + router 0），只减不增。
2. **类型唯一真相源**。后端 `internal/domain` 与 `internal/model` 的 json tag 即 API 契约；前端接口类型唯一来源为 `@myblog/api/modules/*/types`，apps 内禁止定义与后端请求/响应同构的 interface/type；通用响应信封唯一来源为 `@myblog/shared` 的 `ApiResponse`（底层为 `BaseApiResponse`）。
   验证：`git grep -n "interface BaseApiResponse" -- apps` 应为空。
3. **契约先行，双端对齐**。改 wire 格式（请求/响应结构、错误码语义、认证协议）必须同一变更窗口内同步两端，禁止"先改后端、前端以后再说"；改认证协议先更新 `contracts/auth-protocol.md`，改类型镜像跑 `pnpm run contract:check`。新增业务模块必须同时落地后端四层、`packages/api/src/modules/<module>/` 与两应用 `lib/api/index.ts` 注册。
4. **单一权威，禁止私自实例化**。业务规则（状态流转、slug 生成、密码强度、权限判定）权威在后端，前端只做展示层优化；RBAC 权限表权威为生产 `server/configs/config.yaml` 的 `rbac` 节，权限经登录响应 `permissions[]` 下发，`apps/admin/src/lib/constants/auth.ts` 的映射仅作未下发时的降级；依赖一律构造函数注入，生产实例化点仅组合根 `server/cmd/myblog/`（`main.go` 管生命周期、`deps.go` 按域装配），禁止在 service/router/middleware 内 `NewXxxService()`。
   验证：`git grep -n "NewRBACService()" -- server` 的生产命中应仅 `cmd/myblog/deps.go` 一处。
5. **禁止复制粘贴式共享**。两 app 间禁止新增逐字或近似重复文件，公共逻辑下沉 packages；存量逐字重复仅 `src/lib/stores/auth.ts`（15 行 ×2）与 `src/lib/components/theme-toggle.svelte`（45 行 ×2），改任一份必须同步另一份。
   验证：`git diff --no-index` 上述两份文件应零差异。
6. **数据加载归位**。`apps/web` 新页面禁止 `onMount` 取数，用 `+page.server.ts` / `+page.ts` 的 load；`apps/admin` 新页面优先 load + 页面状态模块（`.svelte.ts`）；API 调用不得深入叶子组件。
7. **部署形态为单实例**（有意取舍，非待修缺陷）。应用进程、MySQL、网关各一份，无副本与故障转移；会话令牌表与限流计数均为进程内存结构，服务重启或发布将导致全体用户登出，发布窗口需提前预告；故障即停站，无 RTO 承诺，RPO 为 24 小时。扩容前必须完成令牌表持久化、限流器外置、补 `SetTrustedProxies` 三项改造。
   验证：`git grep -n "SetTrustedProxies" -- server` 与 `git grep -n "replicas" -- docker-compose.yml` 均应为空。
8. **先确认产品定位**。单作者个人博客，不开放注册、无找回密码通道：账号由管理员创建（首次部署 `pnpm run seed:admin`，之后在后台用户管理页新增），忘记密码由超级管理员经 `POST /api/users/update` 重置；游客可浏览并提交评论（受审核设置约束），点赞/收藏/关注需登录。功能提案前先确认是否与之冲突；新增数据收集场景须先同步 `apps/web` 的 `/privacy` 页告知。

## 2. 完成前自检

```bash
cd server && go build ./... && go vet ./... && go test ./...   # 后端编译、静态检查、测试
pnpm run check                                                 # 两应用 SvelteKit 类型检查
pnpm run lint && pnpm run format                               # 静态检查与格式零告警
```

触碰以下区域时追加动作：认证/权限与两 app 重复文件按第 1 节对应条目自检；改 `domain`/`model` 输出字段须评估 `@myblog/api` 类型与页面消费点并跑 `pnpm run contract:check`；新增模块须确认双端对齐；改依赖（`server/go.mod`、`go.sum`、根 lockfile、`pnpm-workspace.yaml` 的 `overrides`）须 `pnpm run go:vulncheck` 与 `pnpm run audit:deps` 零阻塞项。

## 3. 仓库概览

```
MyBlog/
├── apps/
│   ├── web/                  # 前台 toC：SSR 博客 + demo + i18n
│   └── admin/                # 后台 toB：SPA 管理控制台 + 登录页，基准路径 /admin
├── packages/
│   ├── shared/               # 公共纯工具与通用类型（ApiResponse、响应码常量）
│   ├── http/                 # HTTP 请求器（ky 封装，认证回调注入）
│   ├── api/                  # 后端接口模块与响应类型（12 个模块工厂）
│   ├── auth/                 # 认证会话组装（createAuthStore 工厂，注入式）
│   └── ui/                   # shadcn-svelte 基础组件（stock，主题注入）
├── server/                   # Go 后端（Gin + GORM + MySQL）
├── contracts/                # 认证与 i18n 协议、错误码口径、契约金样本
├── scripts/                  # 跨项目构建与开发脚本（Node.js + tsx）
└── docs/                     # 架构细则、开发指南、数据库与 UI 文档
```

- 版本：Go 以 `server/go.mod` 的 `go` 指令为唯一来源（当前 `1.26.9`）；Node 以根 `.nvmrc` 为唯一来源（当前 `22`，`engines.node` 为 `>=22`）。包管理 pnpm（catalog 统一版本），脚本运行时 Node.js + tsx。
- 端口：前台 8899、后台 9988、后端 3000、指标 9090（`metrics.addr`，独立监听，不经网关转发）；健康探针 `GET /api/health`（存活）与 `GET /api/health/ready`（就绪，探测数据库连通性与上传目录可写性，未就绪返回 503 与原因码）。
- 后端业务模块 12 个：user / article / category / tag / comment / media / setting / friendlyLink / stats / notification / follow / dict；路由注册 104 条 = `/api` 业务接口 102 + 健康探针 2，逐条清单见 `server/docs/api/README.md`。

## 4. 常用命令

```bash
pnpm run dev                                     # 智能启动（环境与端口检查、健康监控）
pnpm run dev:server / dev:web / dev:admin        # 分别启动后端 / 前台 / 后台
pnpm run build                                   # 生产构建，可加 --clean --production --skip-tests --skip-lint --server-only --web-only
pnpm run test / check                            # 三类测试串联 / 两应用 svelte-check 类型检查
pnpm run lint / format / quality                 # 静态检查 / 格式化 / 全量质量门禁
pnpm run contract:check                          # 契约三把锁（Go fixture + tsc + vitest 锚定）
pnpm run seed:admin / migrate [create|up|down|goto|force|version]
pnpm run go:quality / go:vulncheck / audit:deps
```

> 环境设置 `pnpm run setup`，各脚本的用途、参数与 pnpm 入口对照见 `scripts/README.md`。

## 5. 后端约定（server/）

- **POST-Only**：业务接口一律 `POST`，参数承载于 JSON 请求体，纯 id 类短参数可放 path；健康探针属基础设施端点，按编排器标准用 `GET`（`internal/router/health.go`）。
- **校验**：handler 用 `ShouldBindJSON` + `binding` tag 完成必填、长度、格式、枚举校验，业务规则在 service 校验且校验通过才进入数据访问层；跨接口复用的校验必须提取公共方法，禁止复制（密码强度校验 `service/password.go` 已按此三处复用）。
- **分层**：`handler`（HTTP 与 DTO 映射）→ `service`（业务）→ `repository`（持久化），路由注册在 `internal/router/`；各层职责与边界见 `docs/development.md`「分层设计」与 `internal/{handler,service,repository,middleware}/README.md`。
- **错误分档**：service 错误唯一经 `internal/handler/error.go` 的 `HandleServiceError` 分档为 401/403/404/400，未分类错误脱敏为 500「服务器内部错误」且只把原文写日志；`binding` 校验失败由 handler 直接 `response.BadRequest` 返回 400；禁止模块内新增私有映射，新增哨兵须同步扩展映射集合与 `contracts/errors.yaml`。完整口径见 `internal/handler/README.md`。
- **统一响应**：统一经 `pkg/response`，业务结果一律 HTTP 200 并以 `code` 区分；信封结构、响应码常量与 404/503 例外见 `server/pkg/response/README.md`。
- **类型归属**：领域实体在 `internal/domain`，GORM 实体在 `internal/model`，`repository` 不得新增实体或 DTO，对外响应禁止直出 GORM 实体（存量违例见 `docs/debt/security-runtime.md` 第 1 条）；请求/响应 DTO 以所属业务域的 service 包内定义为常态（如 `service.CreateArticleRequest`），用户域的 `CreateUserRequest`/`UpdateUserRequest` 在 `internal/domain/dto.go`，短参数 DTO 可就地声明在 handler。
- **命名与接口位置**：Go 字段与 json tag 用小驼峰；接口与实现同包声明并统一 `XxxInterface` 后缀（存量 `UserService`、`RBACService`、`UserRepository` 未带后缀，触碰时统一）；各层依赖接口而非实现，`router` 不得重复定义接口。
- **数据库**：表结构硬性约定（生命周期与状态字段、字段级 `comment`、唯一索引与普通索引、删除策略、时间精度、多段写库事务化）与迁移双轨（开发 `AutoMigrate`、生产 golang-migrate）见 `docs/database-architecture.md`，迁移入口为 `internal/database/migrate.go` 与 `pnpm run migrate`。
- **认证与密码**：bcrypt 成本 `BcryptCost = 12`；令牌为服务端签发的不透明随机串，内存令牌表是身份唯一权威（单实例前提），刷新即旋转、登出撤销令牌对、改密撤销该用户全部令牌。线格式与撤销语义见 `contracts/auth-protocol.md`，实现要点见 `internal/service/README.md`。
- **配置与中间件**：`internal/config` 经 Viper 读 `configs/config.yaml`（YAML 为唯一来源）；中间件覆盖访问日志与 panic 恢复及 RequestID、CORS、安全、认证、RBAC、限流、语言协商与身份解析，认证/权限中间件只依赖 `IdentityProvider` 抽象。配置项与中间件清单以 `internal/config/README.md`、`internal/middleware/README.md` 为准。
- **入口与工具包**：服务入口 `cmd/myblog/`，种子入口 `cmd/seed/`（逻辑 `internal/database/seed.go`），通用包见 `server/pkg/`。

## 6. 前端约定（apps/ + packages/）

- **框架**：SvelteKit + Svelte 5 runes + TypeScript，不使用 Options API；`apps/web` 为前台 toC、SSR 路线（adapter-node），`apps/admin` 为后台 toB、SPA 路线（`ssr=false`、adapter-static，`paths.base='/admin'`）且无 i18n。
- **组件库**：shadcn-svelte 基础组件统一在 `packages/ui` 并保持 stock，经 `$ui` 与 `$ui/*` 引入；组件归属、禁止在应用内重写与主题注入细则见 `docs/ui-design-system.md`。
- **别名**：各应用 `svelte.config.js` 声明 `$ui`、`$ui/*`、`@/*`（应用自身 `src`），前台另有 `$i18n` 与构建期 `#fonts` 别名，`$lib` 由 SvelteKit 隐式提供；对照见 `docs/development.md`「关键技术配置」。
- **样式**：TailwindCSS v4（`@tailwindcss/vite`）；`app.css` 只做聚合入口（含 `@source '../../../packages/ui/src'`），设计令牌在同应用 `src/styles/tokens.css`，`packages/ui` 不携带全局样式；前台为编辑杂志主题（含 `--signal` 与 `--color-line` 别名），后台为原始主题（无 `--signal`）。
- **API 层**：`packages/http` 提供 `createHttpClient` 工厂，`packages/api` 提供 12 个模块工厂，认证会话由 `@myblog/auth` 的 `createAuthStore` 组装；应用侧 `src/lib/service` 注入认证与提示回调、`src/lib/api` 实例化接口，一律 `POST` 调用后端。新增接口必须先加在 `packages/api`，禁止页面直连 ky。
- **状态**：认证 store 逻辑已下沉 `@myblog/auth`（两应用各持薄封装），新公共状态逻辑必须下沉 packages，禁止第三处复制。
- **i18n（仅前台）**：`@inlang/paraglide-js` + `project.inlang` / `messages/*.json5`（必须保持 JSON 兼容写法，否则编译失败）；已接入文件禁止回退硬编码，新增用户可见文案优先经 `m['ui:...']()` 取词，已接入面与其余页面分批口径见 `docs/debt/frontend.md` 第 7 条。
- **类型**：接口类型唯一来源 `@myblog/api`；影子类型层已清除，两应用 eslint 以 patterns 拦截 `$lib/types` 引入形态，禁止重新引入。镜像关系与三把锁见 `docs/architecture-rules.md` 第 3 节。

## 7. 代码风格、注释与测试

- prettier 配置（根 `prettier.config.js`）：`semi: false`、`singleQuote: true`、`arrowParens: 'avoid'`、`printWidth: 100`、`tabWidth: 2`、`trailingComma: 'none'`。
- **导入排序**：prettier 挂载 `prettier-plugin-sort-imports`，排序依据是单条 import 语句的字符长度，语句越长越靠前，既不看模块名字母序也不区分外部依赖与相对路径。这是既定规范，导入块看似错乱时不要手工调整，手工修改会被下次格式化还原。
- **格式化覆盖范围**：lint-staged 只对 `apps/**/src/**` 的脚本与组件文件、任意目录下的 `*.{json,yaml,yml}` 运行 prettier，对 `server/**/*.go` 运行 gofmt 与 goimports；`packages/**` 的源码不在任何 format 门禁覆盖内，新增或改动 packages 文件须手动执行 `npx prettier --write <file>`。
- **注释**：只解释代码无法直接表达的意图、业务规则、约束、原因、风险与副作用，不复述代码；完整写作要求遵循全局规范 `~/.dsh/AGENTS.md` 第 4 节。
- **测试文件位置**：前端测试文件不与业务文件同级，一律置于被测模块同级的 `__tests__/` 目录；跨用例共享的测试替身与夹具置于包内 `src/test/`，不放入 `__tests__/`。
- **测试纪律**：后端各层与 `pkg/*` 均有 `*_test.go`，新增后端关键逻辑必须配套同目录测试；`packages/api`、`packages/auth`、`packages/http` 与两应用单测分别经 `pnpm run test:packages`、`pnpm run test:apps` 纳入门禁，新增 packages 公共逻辑与页面状态模块（`.svelte.ts`）须配套 `*.test.ts`，并为所在 package 补 `typecheck`、`test` 脚本且登记进根 `test:packages`。`pnpm run check` 与 `typecheck:web` 只做类型检查，不是单元测试。
- **提交**：提交信息用简体中文、符合 commitlint 的 conventional commits（type 枚举见 `commitlint.config.js`）；单次提交跨度过大时补充正文说明。

## 8. 文档地图

- `docs/architecture-rules.md`：第 1 节红线的完整裁决细则、债务索引与验证命令、历史教训与重构分期（与本文冲突时以该文件为准）。
- `docs/debt/`：债务条目分册（结构类型 / 前端复用 / 安全运行时 / 测试），基线只减不增。
- `docs/development.md`：环境设置、开发工作流、分层与模块设计、工具陷阱、代码风格示例。
- `docs/database-architecture.md` 与 `docs/database/schema.sql`：数据库设计说明与参考 DDL（实际结构以 GORM 模型为准，生产经 `server/migrations/` 增量迁移）。
- `docs/ui-design-system.md`：视觉、动效与可访问性准则。
- `docs/operations/`：备份恢复与 RPO 口径、指标抓取与告警规则、结构化日志字段与处置流程。
- `contracts/`：认证协议、i18n 协议、错误码口径与契约金样本。
- `server/internal/*/README.md` 与 `server/pkg/*/README.md`：各层与各包的实现约束、字段口径与已知陷阱，改对应目录前先读。
- `server/docs/api/`：接口总数口径与逐模块接口说明。
- `scripts/README.md`：脚本用途、参数与 pnpm 入口对照。
