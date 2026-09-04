# MyBlog 项目 Agent 指南

> 本文件为 AI Agent 在本仓库工作时提供项目级上下文与约束。
> 用户全局工作规范见 `~/.dsh/AGENTS.md`；裁决顺序：全局规范 < 本文档 < `docs/architecture-rules.md` 细则，冲突时以更具体者为准。
>
> **文档分级**：本文含【铁律】（违反必须返工，每条附验证方式）与【约定】（应当遵循）两级。
> 铁律的完整裁决标准、历史案例与债务登记见 `docs/architecture-rules.md`。

## 0. 架构铁律（最高优先级，先读这里）

以下规则由 2026-09 全栈架构诊断提炼。诊断结论：**本项目的抽象设计基本正确，但被系统性绕行**（类型寄生、复制粘贴、影子层、私自实例化），导致每次演进成本远超预期。铁律的目的就是终结"绕行"。

### A1 依赖方向，只准向下

- 后端：`handler → service → repository → domain` 单向依赖。禁止：service import handler；repository import service/handler；middleware/router 直接 import repository（D1 已清偿：router 归零、middleware 仅 `identity.go` 实现 1 处，只减不增）。领域类型统一来自 `internal/domain`。
- 前端：`apps → @myblog/api → @myblog/http → @myblog/shared` 单向依赖。禁止：packages 反向 import apps；页面/组件直接 import `ky`（唯一豁免：各应用 `src/lib/service/index.ts` 的令牌刷新直连，为避免循环依赖）。
- 验证（在 `server/` 目录下执行）：

```bash
git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router
# 输出文件数不得高于债务基线 D1（service 11 + middleware 1 + router 0）
```

### A2 类型唯一真相源

- 后端 `model` 实体的 `json` tag 即 API 契约，修改输出字段前必须评估前端影响面（`@myblog/api` 类型 + 页面消费点）。
- 前端接口类型唯一来源为 `@myblog/api/modules/*/types`；**禁止在 apps 内定义与后端请求/响应同构的 interface/type**（影子类型，存量违例见 D7，禁止扩大）。
- 通用响应结构 `{code, message, data}` 唯一来源为 `@myblog/shared` 的 `ApiResponse`。
- 验证：`git grep -n "BaseApiResponse" -- apps` 合法命中仅限 `apps/admin/src/lib/types/`（存量债务，只减不增）。

### A3 契约先行，模块对齐

- 修改 wire 格式（请求/响应结构、错误码语义、认证协议）前，必须先明确双端影响面并在同一变更窗口内同步两端；禁止"先改后端、前端以后再说"。
- 新增业务模块必须双端对齐：后端 `internal/{handler,service,repository}/<module>.go` + 前端 `packages/api/src/modules/<module>/` + 两应用 `lib/api/index.ts` 注册。现存反例：`user_follow` 仅后端存在（D13）。
- 后端能力缺口（如列表接口缺关键词参数）必须推回后端修复；**禁止前端补偿**（全量跨页拉取后客户端过滤，存量见 D8 注记）。

### A4 单一权威，禁止私自实例化

- 业务规则（状态流转、slug 生成、密码强度、权限判定）唯一权威在后端；前端只能做展示层优化（如隐藏按钮），不得复刻规则逻辑。
- RBAC 权限表唯一权威为后端（生产环境 `configs/config.yaml` 的 `rbac` 节），登录响应下发 `permissions[]`；前端 `apps/admin/src/lib/constants/auth.ts` 的映射仅作未下发时的降级兜底，**禁止新增第三份权限定义**。
- 依赖一律经构造函数注入，**禁止在 service/router/middleware 内部私自 `NewXxxService()`**（D4 已收敛：生产实例化仅 `cmd/myblog/main.go` 组合根 1 处，禁止新增实例化点）。

### A5 禁止复制粘贴式共享

- 两 app 之间禁止新增逐字/近似重复的文件；公共逻辑必须下沉到 packages。存量违例 D5（auth store 等 5 文件约 420 行）。
- 修改任一已知重复文件时，必须同步检查另一份并在提交信息中注明同步情况。

### A6 数据加载归位

- `apps/web` 新页面禁止 `onMount` 取数，必须使用 `+page.server.ts` / `+page.ts` load（SSR 应用）。
- `apps/admin` 存量 onMount 模式不强制迁移（SPA 定位），但新页面优先采用 load + 页面状态模块（`.svelte.ts`）模式，禁止复制 400 行级胖组件（UI+数据+状态+权限焊死一个文件）。
- API 调用不得深入叶子组件；数据获取入口限定为 load 函数或页面顶层组件。

## 1. 任务完成自检（Agent 声明任务完成前必跑）

所有指标**只准变好**（棘轮原则），任何一项差于改动前即不得声明完成：

```bash
# 1) 后端：编译 + 静态检查 + 测试
cd server && go build ./... && go vet ./... && go test ./...

# 2) 依赖方向基线不恶化（在 server/ 目录下）
git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router

# 3) 前端：对应应用类型检查
cd apps/web && pnpm run check    # 或 cd apps/admin && pnpm run check

# 4) 格式与 lint 零告警
pnpm run lint && pnpm run format
```

触碰以下区域时追加专项自检：

- 改认证/权限：确认未新增权限定义副本（A4）、未新增认证工具文件（D6）。
- 改 `model` 实体 json tag：确认已评估前端 `@myblog/api` 类型与页面影响（A2）。
- 改两 app 重复文件：确认另一份已同步（A5）。
- 新增模块：确认双端对齐（A3）。

## 2. 项目概览

MyBlog 是一个 Monorepo 全栈个人博客应用，采用 Go + SvelteKit 技术栈构建。

```
MyBlog/
├── apps/
│   ├── web/                  # 前台应用（Svelte 5 + TS + TailwindCSS v4，公开博客 + demo + i18n）
│   └── admin/                # 后台应用（SvelteKit SPA，管理控制台 + 登录页，无 i18n）
├── packages/
│   ├── shared/               # 公共纯工具与通用类型（ApiResponse 等）
│   ├── http/                 # HTTP 请求器（ky 封装，认证回调注入）
│   ├── api/                  # 后端接口模块与响应类型（11 个模块工厂）
│   ├── auth/                 # 认证会话组装（createAuthStore 工厂，注入式）
│   └── ui/                   # shadcn-svelte 基础组件（stock，主题注入）
├── server/                   # Go 后端服务（Gin + GORM + MySQL）
├── scripts/                  # 跨项目构建/开发脚本（Node.js + tsx）
├── docs/                     # 架构铁律、数据库架构、开发指南、UI 设计系统
├── .husky/                   # Git hooks（commitlint + lint-staged）
└── package.json              # monorepo 根，pnpm workspaces = ["apps/*", "packages/*"]
```

- 包管理：**pnpm**（catalog 协议统一版本）；脚本运行时：**Node.js + tsx**；后端：Go 1.23。
- 配置承载于 `apps/web/.env`、`apps/admin/.env` 与 `server/configs/config.yaml`，默认端口：前台 8899、后台 9988、后端 3000。
- 两应用认证基础设施重复为已知债务（D5），新公共代码一律进 packages。

## 3. 常用命令

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
pnpm run test         # test:server（go test）+ typecheck:web（前后台 svelte-check 类型检查）
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

## 4. 代码质量与格式化约定

- 根 `prettier.config.js`：`semi: false`、`singleQuote: true`、`arrowParens: 'avoid'`、`printWidth: 100`、`tabWidth: 2`、`trailingComma: 'none'`。
- 根 `eslint.config.js` 导出 `baseConfig` 供子项目继承；`apps/*/eslint.config.js` 在其上叠加 Svelte/TS 规则。
- Git hooks：`commitlint`（conventional commits）与 `lint-staged`（对 `apps/**/src/**` 运行 prettier，对 `server/**/*.go` 运行 `gofmt`/`goimports`）。提交信息需符合 conventional commits 规范。
- 更改文件后应运行 `pnpm run lint` 与 `pnpm run format` 保持静态零告警。
- 每完成一个对应功能变更后，使用**简体中文**编写符合 conventional commits 规范的提交信息，并保证提交颗粒度，若单次提交跨度较大需补充 message 描述正文，type 枚举以 `commitlint.config.js` 为准。
- **测试现状**：后端已有单测（`service`/`repository`/`handler` 层 `*_test.go`）；前端尚无测试。新增后端关键逻辑必须配套 `*_test.go`（与被测文件同目录）；新增 packages 公共逻辑与页面状态模块（`.svelte.ts`）时应配套 `*.test.ts`。注意 `pnpm run typecheck:web` 当前仅为类型检查，不是单元测试。

## 5. 后端约定（server/）

- **POST-Only 规范**：后端业务接口一律通过 `POST` 方法注册，查询类接口同样使用 `POST`，不使用 `GET`、`PUT`、`DELETE` 等方法。请求参数统一承载于 JSON 请求体，纯 id 类短参数可放在 path 中。
- **接口健壮性**：涉及外部输入的接口不得省略必要校验。请求参数在 handler 层通过 `ShouldBindJSON` 与 `binding` tag 完成必填、长度、格式、枚举等校验，业务规则在 service 层校验，校验全部通过后才进入数据访问层。
- **公共校验方法**：跨接口复用的校验逻辑必须提取为公共方法或工具函数，不得仅内联在单个 handler 或 service 局部。存量私有校验（如 `service/user.go` 的密码强度校验）触碰时应迁移为公共校验工具，禁止其他模块复制其逻辑。
- **分层架构**：`handler`（HTTP 层，参数校验与 DTO 映射）→ `service`（业务逻辑）→ `repository`（数据访问）。路由注册见 `internal/router/router.go`。依赖方向铁律见 A1。
- **错误分档**：service 返回的哨兵错误（`ErrArticleNotFound`、`ErrUserNotFound` 等）必须在 handler 映射为对应语义（404/403/400），**禁止一律 500**；`binding` 校验失败映射 400；权限不足映射 403。
- **依赖注入**：集中在 `cmd/myblog/main.go`（组合根）与 `router.Dependencies`；handler 通过接口注入。禁止包内私自实例化依赖服务（A4）。
- **统一响应**：使用 `pkg/response` 包，响应结构为 `{ code, message, data? }`；预定义响应码 `CodeSuccess=200`、`CodeError=500`、`CodeInvalid=400`、`CodeAuth=401`、`CodeForbid=403`、`CodeNotFound=404`。统一通过 `response.Success` / `response.Error` 等函数返回。
- **模型与类型归属**：领域实体与请求/响应 DTO 统一位于 `internal/domain`（`domain.User`、`domain/dto.go`），为全系统唯一类型语言；`internal/model` 保留各业务 GORM 实体；`repository` 只承载持久化实现。**禁止在 repository 包新增实体或 DTO 定义**（D2 已清偿，禁止回潮）。
- **入口与工具包**：服务入口在 `cmd/myblog/main.go`，种子脚本入口在 `cmd/seed/main.go`，种子逻辑见 `internal/database/seed.go`；通用包 `pkg/response`、`pkg/datetime`、`pkg/slug`。
- **数据库**：MySQL 单库（GORM）；迁移由 `internal/database/migrate.go` 与根 `migrate` 脚本管理（开发模式 AutoMigrate / 生产 golang-migrate 双轨）。模型见 `internal/model/*.go`，架构细节见 `docs/database-architecture.md`。
- **数据库设计可持续性**：表结构以可长期健康演进为目标设计。每张业务表须具备完整的生命周期字段、状态字段、业务字段与必要的预留扩展字段，且每个字段均带 `comment` 说明业务含义，字段类型与长度须贴合真实数据需求。
- **规范化与索引**：遵循第三范式，多对多关系使用独立关联表，树形结构使用 `parent_id`、`root_id`、`level` 字段；唯一性字段加唯一索引，外键与高频查询字段加普通索引，禁止无索引的大表查询。
- **数据一致性**：显式声明 GORM 关联关系与外键删除策略，如 `OnDelete:CASCADE` 与 `OnDelete:SET NULL`；状态类字段使用命名常量枚举，时间字段统一为 `datetime(3)` 精度。多步骤写库操作必须用事务包裹（存量违例：`ArticleService.CreateArticle` 三段写库无统一事务，触碰时必须修复）。
- **密码**：bcrypt（成本常量 `BcryptCost = 12`），密码强度校验在 `service/user.go`（待迁移公共工具）。
- **命名**：Go 结构体字段与 JSON tag 使用小驼峰；接口与实现同包定义，命名统一为 `XxxInterface` 后缀，如 `ArticleHandlerInterface`、`ArticleServiceInterface`、`ArticleRepositoryInterface`。存量不一致：user 模块的 `UserService`、`JWTService`、`RBACService`、`UserRepository` 未带后缀（触碰时统一，不强制专项重构）。各层只依赖接口而非具体实现。
- **接口定义位置**：接口统一声明在各层实现所在包，即 `handler`、`service`、`repository` 内；`router` 只引用各层接口完成依赖注入，**不得在 `router` 包重复定义接口**（存量违例：router 包重复定义 11 个 handler 接口 + `Dependencies` 字段为 `interface{}` 运行时断言，见 D3，禁止扩大）。
- **配置**：`internal/config` 包通过 Viper 读取 `configs/config.yaml`，所有配置项均以 YAML 为唯一来源。
- **中间件**：`middleware` 包含 logger、request ID、CORS、汇总安全、auth、rbac、ratelimit。权限与认证中间件只依赖 `IdentityProvider` 抽象（`middleware/identity.go`），唯一实现 `jwtIdentityProvider` 经组合根注入 jwtService 与 userRepo。

## 6. 前端约定（apps/ + packages/）

- **框架**：SvelteKit + Svelte 5 + TypeScript，Svelte 5 runes 风格，**不使用 Options API**。应用分 `apps/web`（前台 toC，SSR 路线）与 `apps/admin`（后台 toB，SPA 路线 `ssr=false`）。
- **组件库**：shadcn-svelte 基础组件统一位于 `packages/ui`（保持 stock，主题经各应用 `app.css` token 注入），以 `$ui` 别名引入，由根级 eslint/prettier 排除。**禁止在应用内重写 `$ui` 已有组件**（D14 已清偿：admin 分页已回归 `$ui`）。别名见各应用 `svelte.config.js`：`$lib`、`$ui`、`$i18n`（仅前台）、`@/*`、`#/*`、`~/*`。
- **样式**：TailwindCSS v4（`@tailwindcss/vite`）；前台 `apps/web/src/app.css`（规格书主题，含 `--signal`）、后台 `apps/admin/src/app.css`（原始主题，无 `--signal`）；`packages/ui` 不携带全局样式。
- **API 层**：`packages/http` 提供 `createHttpClient` 工厂，`packages/api` 提供 11 个模块工厂（user/article/category/tag/comment/media/setting/friendlyLink/stats/notification/follow）；认证会话由 `@myblog/auth` 的 `createAuthStore` 组装；应用侧 `src/lib/service` 注入认证与提示回调，`src/lib/api` 实例化接口，一律使用 `POST` 调用后端接口，与后端 POST-Only 规范呼应。**新增接口必须先加在 `packages/api`，禁止页面直连 ky**。
- **状态**：认证 store 逻辑已下沉 `@myblog/auth`（两应用薄封装各持一份）；admin 认证域工具 D6 已收敛（`utils/jwt.ts`、`utils/auth.ts` 已删，刷新/登出单轨）。新公共状态逻辑必须下沉 packages，禁止第三处复制。
- **路由**：前台 `src/routes` 使用分组路由 `(app)`、`demo`（i18n 演示沙盒）；后台使用 `(admin)`、`(auth)`（登录页归属后台）。数据加载纪律见 A6：web 用 load，admin 新页面优先 load。
- **i18n**：仅前台 `apps/web` 使用 `@inlang/paraglide-js`，`project.inlang`/`messages/` 目录，别名 `$i18n`；后台不引入 i18n。
- **类型**：接口类型唯一来源 `@myblog/api`（铁律 A2）。D7 已清偿：`types/api.d.ts` 影子层已删并加 eslint 守门，禁止回潮。
- 前端代码改动需运行对应应用 `cd apps/web && pnpm run check` 或 `cd apps/admin && pnpm run check`（svelte-check + svelte-kit sync）。

## 7. 注释与代码规范硬约束

遵循全局 `~/.dsh/AGENTS.md`，在此强调项目内高频要求：

- 开发过程中遇到可维护性差、建议重构或优化的代码，不要忽略，而是寻求是否确认进行顺带优化。
- 非异步场景调用 async 函数需用 `void` 显式忽略返回的 Promise。
- 遇到与铁律冲突的存量代码：**新改动不得加重违规**（棘轮原则），并在提交信息中注明触碰的债务编号。

## 8. 环境配置

- 后端：`server/configs/config.yaml`（数据库、服务器、日志、API、JWT、安全配置）。默认 MySQL `blog` 库，`root/123456`，含校验与沙箱占位。
- 前端：`apps/web/.env`（`VITE_SERVER_PORT=8899`）与 `apps/admin/.env`（`VITE_SERVER_PORT=9988`），均含 `VITE_PROXY_URL=http://localhost:3000`、`VITE_BASE_URL=/api`、`VITE_REQUEST_TIMEOUT=15000`。

## 9. 已知架构债务登记（棘轮：只减不增）

完整债务说明、基线数值与验证命令见 `docs/architecture-rules.md` 第 7 节。触碰相关区域时必须遵守对应红线：

| 编号 | 债务 | 红线 |
|---|---|---|
| D1 | service/middleware/router 依赖 repository 包（**已收敛**：service 11 + middleware 1 + router 0） | 只减不增 |
| D2 | 双 User 模型（**已清偿**：合并为唯一 `domain.User`） | 新字段只加 `domain.User` |
| D3 | router 重复定义 handler 接口 + `interface{}` 断言（**已清偿**） | 禁止回潮 |
| D4 | `RBACService` 生产实例化（**已收敛**：仅 main 组合根 1 处） | 禁止新增实例化点 |
| D5 | 两 app 基础设施逐字重复（auth store 已下沉 `@myblog/auth`，service/theme/layout 仍重复） | 修改任一必须同步另一份 |
| D6 | admin 认证工具三轨并行（**已收敛**：`jwt.ts`/`auth.ts` 已删，刷新/登出单轨） | 禁止新增认证工具文件 |
| D7 | 影子类型 `types/api.d.ts`（**已清偿**，eslint 守门已加） | 禁止回潮；类型一律来自 `@myblog/api` |
| D8 | admin 12 个页面胖组件 + onMount 取数（users 跨页补偿**已清偿**，users/list 支持 keyword） | 新页面禁用；后端缺口推回后端 |
| D9 | web 首页 load 死代码（**已清偿**） | 新页面禁用 load 调认证接口 |
| D10 | 401 文案匹配（**已清偿**：`code === 401` 判定） | 禁止回退文案匹配 |
| D11 | JWT 撤销内存 map（**已加锁**；deprecated `ValidateToken` 已删） | 单实例部署前提；持久化前保持锁 |
| D12 | 文章响应泄漏作者审计字段（**已清偿**：审计字段 `json:"-"`） | 新增审计字段默认 `json:"-"` |
| D13 | `user_follow` 前端零消费（**API 模块已补齐**：`@myblog/api/modules/follow` + 两应用注册；页面待 web 业务接入） | 页面消费前视为功能未完成 |
| D14 | admin 本地 `pagination.svelte` 重写 `$ui` 已有组件（**已清偿**：7 页回归 `$ui`） | 禁止仿效；新分页一律 `$ui` |

## 10. 开发进度概览

已完成：
- 基础设施：Monorepo 架构、环境与工具链、Git hooks、智能开发脚本与监控、pnpm catalog 版本治理。
- 后端：11 个业务模块（用户/认证/JWT 双 token/RBAC、文章 CRUD 与状态及互动、分类、标签、评论、媒体、设置、友链、统计、通知、关注）均已完成三层实现与路由注册。
- 前端 admin：14 个页面（仪表盘、文章管理、分类、标签、评论、媒体、用户、设置、友链、统计、通知、登录）+ markdown 编辑器组件。
- 前端 web：仅应用壳（首页占位数据 + i18n demo 页），业务页面未接入（含 D9 地雷待清理）。

待办：
- web 前台业务接入（文章列表/详情/归档，强制 load 模式 + SSR）。
- `user_follow` 页面级消费（`@myblog/api` 模块已就绪，见 D13）。
- 全文搜索、响应式完善、部署与 Docker。
- 可选细化（非验收口径）：handler 层全面 DTO 分离（D12 已用 `json:"-"` 兜底）、组合根按域装配。

架构大清洗已完成，详见 `docs/architecture-rules.md` §8 分期路线状态：R0/R1/R2 全部完成，R3 的 IdentityProvider 横切归位、RBAC 迁 config 并下发、users/keyword、分页回归 `$ui`、follow 模块、认证工具收敛均已完成；债务 D1-D14 只减不增。
