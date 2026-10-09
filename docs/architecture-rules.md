# MyBlog 架构铁律与纪律手册

> 本手册是 AGENTS.md 第 0 节铁律的完整裁决细则，效力高于普通文档。
> 本项目的抽象设计基本正确（分层骨架、packages 单向依赖、工厂注入），但历史上被长期系统性绕行，
> 形成了若干结构性顽疾。本手册的使命是让"绕行"在代码评审与自检命令层面被拦截。
>
> **裁决顺序**：`~/.dsh/AGENTS.md`（全局）< `AGENTS.md`（项目）< 本手册（细则）。冲突时以更具体者为准。

---

## 1. 规则分级与核心机制

| 级别 | 含义 | 违反后果 |
|---|---|---|
| 【铁律】 | 依赖方向、真相源、契约、权威、复用、数据加载六类结构性约束 | 必须返工 |
| 【约定】 | 命名、格式、注释等风格约束 | 应当修正 |
| 【债务】 | 已存在的违例，登记基线与验证命令 | **基线只减不增** |

**基线只减不增是本手册的核心执行机制**：存量违规不要求立即修复，但任何改动使债务指标恶化，即新增违例文件、新增实例化点或新增重复代码时判定违规。每条债务附带验证命令，任务完成前 Agent 必须自检。

---

## 2. 依赖方向

### 2.1 后端允许的依赖图

```
cmd/myblog/main.go（组合根，唯一允许 new 一切的地方）
    ↓
router → handler → service → repository → model
              ↑ 严禁越过        ↓
           middleware（只依赖抽象接口，不依赖 repository 实现）
```

**禁止的依赖（黑名单）**：

| 禁止 | 理由 | 现状 |
|---|---|---|
| service → handler | 业务逻辑不得感知 HTTP | 已合规 |
| repository → service / handler | 数据层不得承载业务 | 已合规 |
| middleware → repository | HTTP 横切层不得直捣存储（终态：依赖 IdentityProvider 抽象） | 仅 `identity.go` 实现 1 处 |
| router → repository（仅为拼装中间件） | 路由层不应持有数据访问句柄 | router 依赖 repository 归零 |
| 任何包 → `database.GetDB()` 全局单例 | 破坏可测试性 | 已合规（仅 main/seed 使用） |

### 2.2 前端允许的依赖图

```
apps/web, apps/admin（组合根：实例化与回调注入）
    ↓
@myblog/api → @myblog/http → @myblog/shared
@myblog/ui（独立，被应用经 $ui 别名消费）
```

**禁止的依赖（黑名单）**：

| 禁止 | 豁免 |
|---|---|
| packages/* → apps/* | 无（反向依赖一律非法） |
| 页面/组件 → 直接 import ky | 仅各应用 `src/lib/service/index.ts` 的令牌刷新直连（避免循环依赖） |
| 页面/组件 → 自造 HTTP 请求 | 无，新接口一律先进 `packages/api` |

### 2.3 验证命令

```bash
# 在 server/ 目录下（基线：service 12 + middleware 1 + router 0 = 13 文件，`*_test.go` 测试替身不计入）
git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router

# 前端：packages 反向依赖应用（应为空）
git grep -ln "apps/web" -- packages
git grep -ln "apps/admin" -- packages
```

---

## 3. 类型唯一真相源

### 3.1 三层类型镜像关系与权威

```
后端 domain 实体 json tag（权威）
    ↓ 手工镜像（三把锁双向锚定，见 §3.4；不引入 codegen）
@myblog/api/modules/*/types（唯一合法镜像）
    ↓ 禁止再造
apps 内任何同构 interface/type（影子类型，禁止定义）
```

### 3.2 规则

1. 后端修改实体输出字段（json tag）属于 **wire 格式变更**，必须走契约先行流程（见 §4）。
2. 前端需要后端没有的类型（如页面视图模型）：允许定义，但**不得与后端请求/响应结构同构**（同构 = 字段集合基本一致）。视图模型应组合/派生自 `@myblog/api` 类型，而非平行重写。
3. `AuthState` 类认证域类型统一位于 `@myblog/auth`，应用内已无旧副本；新增认证类型一律来自 `@myblog/auth`。

### 3.3 验证命令

```bash
# apps 内不得再出现 BaseApiResponse 定义
git grep -n "interface BaseApiResponse" -- apps
```

### 3.4 类型生成策略：三把锁双向锚定

**决策**：不引入 swaggo → OpenAPI → 前端类型生成的 codegen 链路，改用**双向类型锚定**三把锁，以最低成本实现与 codegen 同级的「物理不可漂移」：

| 锁 | 机制 | 落地状态 |
|---|---|---|
| ① | Go handler 测试断言响应与 `contracts/fixtures/` 语义逐字节一致 | ✅ `handler/fixture_locks_test.go` + `login_fixture_test.go`（**4/4 金样本**：login.success、login.wrong-password、users.list、article.detail） |
| ② | TS 侧 `tsc --noEmit`（挂入 contract:check）+ `expectTypeOf(fixture).toExtend<WidenLiteral<手写类型>>` 单向合法实例校验 | ✅ `packages/api/src/contracts/fixtures.test.ts`（4 用例） |
| ③ | eslint `no-restricted-imports` 禁止应用层定义同构类型 | ✅ `apps/admin` 与 `apps/web` 双侧 eslint.config.js |

**锁②的语义**：JSON 导入值在 TS 中推导为宽类型，与手写联合类型无法精确相等（`toEqualTypeOf` 对 `string|null` 等 nullable 联合字段必然失败）。因此引入 `WidenLiteral`（字面量与联合放宽为基础类型）后，以 `toExtend` 单向断言"金样本为手写契约类型的合法实例"：金样本缺字段、字段类型不符或手写类型收窄即编译失败。**已知弱项**：单向断言不检测"金样本多字段而类型未声明"（多余属性在 assignability 中允许），此类漂移由锁①逐字节校验覆盖。

漂移必炸链：后端改 → Go 测试红 → 改 fixture → 类型测试红 → 改类型 → check 红。

**门禁**：`pnpm run contract:check`（Go fixture 测试 + `@myblog/api` 的 `tsc --noEmit` + vitest 类型锚定）。

**升级触发器**（命中任一则改用 codegen 方案）：
1. 接口总数 > 80–100；
2. 出现第二客户端（移动端 / 第三方）；
3. 需对外发布 API 文档；
4. 团队扩张超单人。

**备注**：handler DTO 分离是既定方向，与类型生成策略的选择正交；`article.detail` 金样本的类型锚定已补齐（Author 窄化后直接锚定 `Article` 类型）。未来升级 codegen 时 DTO 层零返工。

---

## 4. 契约先行与模块对齐

### 4.1 wire 格式变更流程（必须按序执行）

1. **影响面声明**：列出后端 DTO、`@myblog/api` types、两 app 消费页面三方清单。
2. **双端同窗口变更**：后端与前端类型/页面改动进同一变更集（或紧邻提交），提交信息注明契约面。
3. **错误语义**：后端错误 message 措辞变更属文案层面，前端以 `code === 401` 判定，不再匹配文案；改错误码语义仍需双端同步。
4. **认证协议**：token 形状、刷新端点、撤销语义的变更属于最高风险契约变更，须先更新 `contracts/auth-protocol.md` 再动代码。

### 4.2 模块对齐表

新增业务模块必须双端同时登记以下位置，缺一即未完成：

| 端 | 必须创建/修改 |
|---|---|
| 后端 | `internal/model/<entity>.go`（如需）、`internal/repository/<module>.go`、`internal/service/<module>.go`、`internal/handler/<module>.go`、`internal/router/<module>.go`、`main.go` 装配 |
| 前端 | `packages/api/src/modules/<module>/{types.ts, index.ts}`、`packages/api/src/index.ts` 导出、两应用 `src/lib/api/index.ts` 注册 |

**历史反例**：`user_follow` 曾后端完整、前端零消费，后补齐 API 模块（`@myblog/api/modules/follow` + 两应用注册）。**此为"两端各自演进"的实证，模块对齐表的目的就是让这类偏差在发生当天可见。**

### 4.3 能力缺口推回原则

前端发现后端接口能力不足（如列表缺关键词过滤）时，**唯一合法动作是给后端加参数**，禁止前端补偿（全量拉取 + 客户端过滤）。admin users 页的跨页抓取补偿已移除，后端 `users/list` 已增加 keyword 参数。

---

## 5. 单一权威

### 5.1 权威归属表

| 事实 | 唯一权威 | 非法副本 |
|---|---|---|
| 角色枚举与权限映射 | 后端（生产 `configs/config.yaml` `rbac` 节，经 `LoadRBACConfig` 加载） | 前端 `constants/auth.ts` 映射仅作未下发时的降级逻辑，**禁止新增第三份定义** |
| "是否管理员"判定 | 后端 RBAC | 前端复刻（仅可做展示层优化） |
| 文章状态流转 | `domain.Article` 行为方法 + service | 前端硬编码状态字符串比较 |
| slug 生成 | `pkg/slug` + service 调用 | 前端自写 slugify |
| 密码强度 | 后端 `service/user.go` | 前端正则复刻（仅可做输入提示，不得作为校验依据） |
| token 刷新流程 | 各应用 `src/lib/service/index.ts`（裸 ky 直连版，单一实现） | —（`utils/jwt.ts` 的 `manualRefreshToken` 已删） |

### 5.2 实例化纪律

- 依赖对象只能由组合根（`main.go`）构造并逐层注入。
- **禁止**在 service 构造函数内部 `New` 另一个 service。`NewUserService` 不得内部实例化 `RBACService`。
- **禁止**在中间件、路由注册函数内部实例化服务。`router/user.go`、`middleware/rbac.go` 的私自实例化已被移除。

### 5.3 验证命令

```bash
# RBACService 生产实例化点为 1 处，即 cmd/myblog/main.go 组合根
# rbac.go 本体的命中为定义，不计入；测试文件不计入
git grep -n "NewRBACService()" -- server
```

---

## 6. 复用纪律与数据加载

### 6.1 复用决策树（新增公共逻辑时）

```
该逻辑是否被两个 app 使用？
├─ 是 → 必须进 packages（shared/http/api，认证域终态为独立 auth 包）
└─ 否 → 留在应用内
     该逻辑与另一 app 现有代码是否同构？
     └─ 是 → 违反复用纪律，停下提取；否 → 正常开发
```

已知跨 app 重复文件（修改任一必须同步另一份）：`stores/auth.ts`、`service/index.ts`（仅 goto 导入路径一处既定差异，语义逐字一致）、`components/theme-toggle.svelte`（经 `git diff --no-index` 校验逐字一致）、`routes/+layout.svelte`（仅导入排序与类型标注差异）。error 页两 app 已各自独立设计（web 编辑杂志版式、admin 居中后台版式），仅同名职责，已移出同构清单。

### 6.2 数据加载

| 应用 | 新页面强制模式 | 存量处理 |
|---|---|---|
| web（SSR 路线） | `+page.server.ts`（需 SEO/会话）或 `+page.ts` load | 死 load 已清理，按此模式接入业务 |
| admin（SPA 路线） | 优先 load + `.svelte.ts` 页面状态模块 | 7 个胖组件不强制迁移，触碰时拆分 |

胖组件判定标准（任一命中即应拆分）：单文件 > 300 行；`$state` 声明 > 8 个；组件内直接调用 API 且含增删改多个 handler；组件内含权限判定逻辑。

### 6.3 认证协议事实登记（当前线格式）

改动以下任何一项均属最高风险契约变更，须先更新本节再写代码：

- token pair 形状：`{accessToken, refreshToken, expiresIn}`。
- 线格式：**不透明令牌**（服务端签发 32 位十六进制随机串，身份与生命周期登记在服务端令牌表；令牌自身不携带任何可解码信息，校验以令牌表为唯一权威，不重构也不重新签名）。
- 刷新：`POST /api/auth/refresh`，body `{refreshToken}`；刷新即旋转（旧 refresh token 撤销），旋转前查库校验用户存在且状态正常。
- 撤销：令牌表为内存 map（单实例前提）；撤销即从表中移除记录，过期记录随签发惰性清理；登出撤销令牌对（refreshToken 经请求体可选提交），改密成功按用户撤销全部既有令牌。
- access 链路取舍：信任短有效期，不逐请求查库；被禁用用户的存量 access 至多存活 `token.access_expire` 分钟，实时失效诉求由 Cookie 会话方案承接。
- 前端 401 识别：以响应体业务码 `code === 401` 判定（`TOKEN_ERROR_MESSAGES` 文案匹配已移除，禁止回退）。

---

## 7. 债务登记表（基线只减不增）

> 触碰相关区域前先读对应条目的红线；任务自检时核对基线不恶化。

| 债务 | 基线 | 验证命令 | 红线 |
|---|---|---|---|
| service/middleware/router import repository | service **12** + middleware **1** + router **0**。router 已归零，middleware 仅 `identity.go`；`*_test.go` 测试替身文件不计入。dict 模块接入曾使 service 层由 11 增至 12 | 见 §2.3 | 只减不增 |
| 双 User 模型同写 users 表 | 已修复：合并为唯一 `domain.User` 实体 | `git grep -n "type User struct" -- server/internal --include="*.go"`（仅 domain） | 新字段只加 `domain.User` |
| router 重复定义 handler 接口 + `interface{}` 断言 | 已修复：router 重复接口 0、断言 0 | `git grep -c "HandlerInterface interface" -- server/internal/router`（应为空） | 禁止重新引入 |
| `RBACService` 生产实例化 | 生产实例化点 1 处，即 main.go 组合根 | 见 §5.3 | 禁止新增实例化点 |
| 两 app 基础设施逐字重复 | 逐字重复仅 **2 文件 60 行**：`src/lib/stores/auth.ts`（15 行 ×2，SHA256 一致）与 `src/lib/components/theme-toggle.svelte`（45 行 ×2，SHA256 一致）。另有 3 文件职责同构但内容不同，不计入逐字重复：`service/index.ts`（69 / 67 行）、`routes/+layout.svelte`（20 / 12 行）、`routes/+error.svelte`（95 / 38 行，两 app 版式已各自独立设计） | `git diff --no-index apps/web/src/lib/stores/auth.ts apps/admin/src/lib/stores/auth.ts`（应为零差异）；同法核对 `theme-toggle.svelte` | 修改任一必须同步另一份 |
| admin 认证工具三轨并行 | 已统一：`utils/jwt.ts`、`utils/auth.ts` 已删（约 488 行）；`performLogout` 单轨（utils/logout）、刷新单轨（service/index.ts） | `git grep -ln "requireAuth\|performLogout\|manualRefreshToken\|getAuthStatus" -- apps/admin/src/lib` | 禁止新增认证工具文件；禁止双轨回退 |
| 影子类型层 | 已修复：`types/api.d.ts` 与 admin `lib/types`（admin/common/auth/index 共 535 行）已删，两应用 eslint 守门由 paths 改 patterns，拦截 `$lib/types` 全部引入形态 | `git grep -n "interface BaseApiResponse" -- apps`（应为空） | 禁止重新引入；类型一律来自 `@myblog/api` |
| admin 胖组件 + onMount 取数 | 口径为「`apps/admin/src/routes` 下单文件 > 300 行」，当前 **7 个**，降序为 tags 448 / users 401 / links 399 / dicts 341 / comments 324 / categories 321 / media 320；路由 svelte 文件共 19 个。`onMount` 取数命中 13 个路由文件。users 跨页补偿已移除，users/list 支持 keyword | `git grep -ln "onMount" -- "apps/admin/src/routes/(admin)"`；行数按上述口径统计 | 新页面禁用；后端缺口推回后端 |
| web 首页 load 死代码 | 已修复：`(app)/+page.ts` 死 load 已移除 | 读文件确认 | 新页面禁用 load 调认证接口 |
| 401 文案匹配 | 已修复：`client.ts` 改为响应体 `code === 401` 判定 | `git grep -n "TOKEN_ERROR_MESSAGES" -- packages`（应为空） | 禁止回退文案匹配 |
| 令牌表为内存 map | 已加锁并支持过期惰性清理：`sync.RWMutex` 保护；令牌为不透明随机串，身份唯一权威在服务端令牌表，过期记录随签发清理 | `git grep -n "tokensByUser" -- server` | 单实例部署前提；持久化前保持锁；服务重启即全部会话失效 |
| 文章响应泄漏作者审计字段 | 已修复：`lastLoginIP` 等审计字段改为 `json:"-"` | 读 `domain/user.go` json tag | 新增审计字段默认 `json:"-"` |
| follow 模块仅后端 | API 模块已补齐：`@myblog/api/modules/follow` + 两应用注册；页面消费待 web 业务接入 | `git grep -ln "createFollowAPI" -- packages/api/src`（非空即已补齐） | 页面消费前视为功能未完成 |
| admin 重写 `$ui` 已有组件 | **admin 侧已修复**：本地 `pagination.svelte` 已删，7 页回归 `$ui`。**web 侧保留自有实现**：`apps/web/src/lib/components/article/PaginationNav.svelte` 使用锚点版式，因为 `$ui/pagination` 的包装层未透传 bits-ui 原语的 `child` 元素替换通道，前台改用该包装层将失去可爬取的真实 `<a href>`，与 SSR 站点的 SEO 与渐进增强目标冲突。该组件已补齐 `aria-label` 与 `rel` 语义，包装层补齐 `child` 透传后方可统一 | `git diff --no-index <(git show 865b613^:apps/admin/src/lib/components/admin/pagination.svelte) apps/web/src/lib/components/article/PaginationNav.svelte` 仅供对照；`git grep -n "child" -- packages/ui/src/pagination`（当前无命中，即透传缺口） | admin 侧禁止仿效；新分页一律 `$ui`；web 侧在该项统一前不得新增第三处分页实现 |
| 公开端点直出实体泄漏个人信息 | **评论域与作者域已窄化**：`model.Comment` 游客邮箱/IP/UserAgent 改 `json:"-"`，评论 `user` 与文章 `author` 经 `domain.AuthorPublic` 窄化视图输出，管理端审计走 `AdminCommentView`。**`/users/get` 已收紧**：由仅挂基础认证改为挂 `user:list` 权限仅限管理端访问，普通用户查看他人资料走窄化的 `publicProfile` 端点，回归测试在 `internal/router/user_rbac_test.go`。**媒体域已收紧**：媒体接口的上传者经 `domain.UploaderPublic` 窄化视图输出、`uploadIP` 审计字段改 `json:"-"`，媒体详情补齐水平越权校验（与列表归属规则对称），回归测试在 `model/media_public_test.go` 与 `service/media_test.go` | 读 `model/comment.go`、`model/article.go`、`model/media.go` json tag 与 `domain/author.go`；service 层测试断言公开响应无审计字段；`go test ./internal/router/ -run TestUsersGetRequiresUserListPermission -count=1` | 公开端点输出个人信息前必须经窄化 DTO 或字段白名单；实体新增隐私/审计字段默认 `json:"-"` |
| WAF 内容级黑名单的固有误伤面 | 误伤回归用例已就位：默认模式全部经词首边界或取值上下文锚定，攻击拦截与误伤回归两组用例在位；残余风险为讲解 SQL/XSS 的技术文章正文命中关键词模式仍会被拦，根治需内容感知解析或按路由豁免 | `go test ./internal/middleware/ -run "TestDefaultBlockedPatterns\|TestSecurityMiddleware" -count=1` | 新增或修改阻止模式必须先红后绿配"攻击拦截 + 误伤回归"两组用例，禁止回退宽匹配 |
| 测试替身内嵌空接口的运行时脆性 | 既有约定，三处实证：service 层 fake 以内嵌接口继承全部方法，接口新增方法被既有测试路径调用时以 nil panic 暴露而非编译错误（`recordedTokenService.GenerateTokenPair`、`loginUserRepo.Update`、`lockoutUserRepo.GetByUsername` 三例） | `go test ./internal/... -count=1` | 接口新增方法被既有测试路径触达时，必须为受影响 fake 显式覆写；禁止依赖内嵌空接口的静默兼容 |
| web 界面多语言局部接入 | 语言切换对 Header/Footer/错误页真实生效（含 NotificationBell、FriendlyLinkDialog 两个 Header 子组件），其余页面文案硬编码中文，en 模式下界面为混合语言；词表文件必须保持 JSON 兼容写法（paraglide 编译器按严格 JSON 解析，json5 特性直接编译失败） | `git grep -ln "\$i18n" -- apps/web/src`（已接入面：Header、Footer、NotificationBell、FriendlyLinkDialog、+error） | 已接入文件禁止回退硬编码；新增用户可见文案优先经 `m.*` 词表取词；其余页面接入待页面大变动后分批推进 |
| 字典页窄屏交互与工具栏布局遗留 | `apps/admin/src/routes/(admin)/dicts/+page.svelte` 的类型列表窄屏交互与工具栏布局打磨未完成。该页同时是 7 个胖组件之一（341 行），两个问题可在同一次触碰中一并处理 | 读 `apps/admin/src/routes/(admin)/dicts/+page.svelte` 的类型列表与工具栏区块在窄屏下的布局 | 触碰该页时必须顺带处理，不得再次遗留；不得以「已记录在提交信息」替代债务登记 |

---

## 8. 历史重构分期路线

> 此处登记历史重构的分期内容与验收口径，供回顾时对照。
> 每完成一项债务修复必须同步更新第 7 节基线数值与本节状态，保持登记表与代码一致。

| 阶段 | 内容 | 完成的债务 | 验收口径 |
|---|---|---|---|
| 阶段一 结构归位与错误分档 | 删 router 重复接口与不可达代码；错误分档（哨兵错误→404/403/400）；令牌表加锁；web 死 load 清理；建立 `contracts/` 目录 | router 重复接口、web 死 load、401 判定均已修复；令牌表已加锁；`RBACService` 生产实例化点仅 1 处；not-found 哨兵→404 已落地，403/400 随错误码契约落地 | ✅ 第 1 节自检命令全绿 |
| 阶段二 类型归位 | 建立 `internal/domain`，合并双 User，service/middleware/router 签名切换为 domain 类型；前端 auth 下沉共享包、影子类型清除 | 双 User 模型与影子类型层已修复；两 app 重复代码大幅减少（auth store 下沉）；service 层 repository 依赖由 12 降至 11 | ✅ 依赖方向基线下降；auth store diff 为零 |
| 阶段三 契约切换 | handler DTO 分离；`contracts/` + 三把锁双向锚定（替代 codegen）；401 改错误码判定 | 401 文案匹配与审计字段泄漏已修复；三把锁已落地（`pnpm run contract:check`） | ✅ 影子类型归零；漂移必当天变红 |
| 阶段四 横切归位与权限下发 | 中间件坍缩为 IdentityProvider 策略；组合根按域装配；RBAC 权限表迁数据源并下发；admin 胖组件拆分、users 搜索推回后端 | `RBACService` 生产实例化点仅 1 处；users 跨页补偿与 admin 分页均已修复；RBAC 迁 config.yaml 完成；permissions 下发完成；follow API 模块已补齐；IdentityProvider 中间件坍缩完成（router 依赖 repository 归零、middleware 1 处）；认证工具已统一 | 权限定义全栈唯一；认证工具单轨 ✅ |
| 阶段五 扩展点 | 后端 ContentRenderer 内容策略接口（当前 `pkg/markdown` 已直接承接 Markdown 渲染，策略接口化待做）；web 业务页面已按"公开数据 SSR + 个性化数据客户端补拉"现状接入，token 迁 cookie 已完成 | — | 新文章类型 = 插入实现，非逐层打洞 |

---

## 9. 历史教训

以下模式在本仓库真实发生过并造成结构性损伤，评审见到同类手法应直接驳回：

1. **类型寄生**：把领域对象/DTO 定义进 `repository` 包（→ 全系统类型倒挂）。
2. **实体直出**：GORM 实体经 `response.Success` 直接序列化为 API 响应（→ 隐私泄漏、契约锁死）。
3. **复制而非提取**：跨 app 逐字拷贝基础设施，"两处保持一致"最终总是失守。
4. **自造影子**：不修镜像源而在应用层再造一套"想象中的 API"类型。
5. **前端补偿后端缺口**：跨页全量拉取 + 客户端过滤（→ 把后端债务放大为前端复杂度）。
6. **私自实例化**：绕过组合根在包内 `New` 依赖（→ 依赖注入形同虚设）。
7. **双轨并存**：同一职责写两遍而不是合并为一处（→ token 刷新双轨、认证工具三轨）。
8. **两端各自演进**：模块只在一端落地。
