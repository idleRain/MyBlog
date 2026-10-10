# MyBlog 架构铁律与纪律手册

> 本手册是 `AGENTS.md` 第 1 节「结构红线」的完整裁决细则，效力高于普通文档。
> 本项目的抽象设计基本正确（分层骨架、packages 单向依赖、工厂注入），但历史上被长期系统性绕行，
> 形成了若干结构性顽疾。本手册的使命是让"绕行"在代码评审与自检命令层面被拦截。
>
> **裁决顺序**：`~/.dsh/AGENTS.md`（全局）< `AGENTS.md`（项目）< 本手册（细则）。冲突时以更具体者为准。
>
> **分册**：第 1 至 7 节为规则本体，第 8 节为债务索引，第 9 至 10 节为历史登记；债务的完整条目（状态、基线数值、验证命令、完整红线）按关注域分册在 [`debt/`](./debt/README.md)。

---

## 1. 规则分级与核心机制

| 级别 | 含义 | 违反后果 |
|---|---|---|
| 【铁律】 | 依赖方向、真相源、契约、权威、复用、数据加载六类结构性约束 | 必须返工 |
| 【约定】 | 命名、格式、注释等风格约束 | 应当修正 |
| 【债务】 | 已存在的违例，登记基线与验证命令 | **基线只减不增**（条目分册见 [`debt/`](./debt/README.md)） |

**基线只减不增是本手册的核心执行机制**：存量违规不要求立即修复，但任何改动使债务指标恶化，即新增违例文件、新增实例化点或新增重复代码时判定违规。每条债务附带验证命令，任务完成前 Agent 必须自检。

---

## 2. 依赖方向

### 2.1 后端允许的依赖图

```
cmd/myblog/（组合根：main.go 管生命周期、deps.go 按域装配，唯一允许 new 一切的地方）
    ↓
router → handler → service → repository → domain / model
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
| 任何包 → `database.GetDB()` 全局单例 | 破坏可测试性 | 已合规（仅 `cmd/myblog` 与 `cmd/seed` 使用） |

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
# 在 server/ 目录下执行（路径相对当前目录解析，在仓库根执行会得到零命中的假绿）
# 基线：service 12 + middleware 1（仅 identity.go）+ router 0 = 13 文件
git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router ':(exclude)*_test.go'

# 前端：packages 反向依赖应用（应为空；在仓库根执行）
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

**当前状态**：路由注册总数已达 104 条（`/api` 业务接口 102 + 健康探针 2），触发器 1 已被命中，但 codegen 迁移尚未启动，现行机制仍为三把锁。是否迁移属待决事项，决策变化时必须同步本节。

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
| 后端 | `internal/model/<entity>.go`（如需）、领域实体与共享 DTO 进 `internal/domain`、请求 DTO 就近定义在所属 service 包、`internal/repository/<module>.go`、`internal/service/<module>.go`、`internal/handler/<module>.go`、`internal/router/<module>.go`、`cmd/myblog/deps.go` 装配 |
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

- 依赖对象只能由组合根构造并逐层注入。组合根为 `cmd/myblog/`：`main.go` 负责配置加载、数据库初始化与 HTTP 生命周期，`deps.go` 按业务域装配仓储、服务与处理器；二者同属组合根，不构成第二个实例化点。
- **禁止**在 service 构造函数内部 `New` 另一个 service。`NewUserService` 不得内部实例化 `RBACService`。
- **禁止**在中间件、路由注册函数内部实例化服务。`router/user.go`、`middleware/rbac.go` 的私自实例化已被移除。

### 5.3 验证命令

```bash
# RBACService 生产实例化点为 1 处，即组合根 cmd/myblog/deps.go
# rbac.go 本体的命中为定义，不计入；*_test.go 测试文件不计入
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

已知跨 app 重复文件分两档，处理方式不同：

1. **逐字重复，改任一必须同步另一份**（`git diff --no-index` 应零差异）：`src/lib/stores/auth.ts`（15 行 ×2）、`src/lib/components/theme-toggle.svelte`（45 行 ×2）。
2. **职责同构但内容不同，不计入逐字重复**，也不要求逐字同步：`service/index.ts`（69 / 67 行，差异在 `getLanguage` 取值与 `goto` 导入来源）、`routes/+layout.svelte`（20 / 12 行）、`routes/+error.svelte`（95 / 38 行，web 编辑杂志版式与 admin 居中版式已各自独立设计）。

新增公共逻辑仍按 6.1 决策树优先下沉 packages；两档文件均不得再增加第三份同构实现。

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

## 7. 部署形态与可用性约束（单实例）

> 本节的取舍是**有意接受**的架构决策，不是待修复的缺陷：它记录当前形态能做什么、不能做什么，以及扩容前必须先完成的改造。
> 代码事实来源：`docker-compose.yml`（五个服务各一份，无副本与故障转移）、`service/token.go`（进程内令牌表）、`middleware/ratelimit.go`（进程内计数 map）。

### 7.1 当前约束

| 维度 | 当前形态 | 后果 |
|---|---|---|
| 部署形态 | 应用进程、MySQL、网关各**单实例** | 无横向扩容能力，单副本容量即整站容量 |
| 会话令牌 | 存于**进程内存**，不透明随机串，身份唯一权威在服务端令牌表 | **服务重启或发布导致全体用户登出**，发布窗口需提前预告用户 |
| 请求限流 | 进程内计数 map，按客户端 IP 或用户维度计数 | 阈值是**单进程**口径，多副本会使实际阈值成倍放大 |
| 故障处理 | **无自动故障转移** | **故障即停站**，恢复依赖人工介入与重启；无 RTO 承诺，RPO 为 24 小时（见 `operations/backup-restore.md`） |

### 7.2 扩容前置改造三项

三项须**全部完成**才能以增加副本的方式扩容，缺任一项即产生静默缺陷：

1. **令牌表持久化**：把 `service/token.go` 的内存表切换到共享存储，并保持「查表即唯一权威」的既有校验语义。`model.AuthToken` 的 `TokenHash`（SHA256 十六进制）通道已就绪，可沿用该存储口径。
2. **限流器外置**：把 `middleware/ratelimit.go` 的进程内计数迁至共享存储，否则限流阈值随副本数成倍放大，防护形同失效。
3. **反向代理信任链**：为 `gin.Engine` 补 `SetTrustedProxies` 配置（当前零配置）。缺此配置时经网关转发后 `ClientIP()` 恒为网关地址，同时影响限流、登录失败锁定、访问审计与 WAF 四项能力；`ClientIP()` 现有 7 处调用。

### 7.3 验证命令

```bash
# 改造前两处均应为空
git grep -n "SetTrustedProxies" -- server
git grep -n "replicas" -- docker-compose.yml
```

---

## 8. 债务登记表（基线只减不增）

> 完整条目（状态、基线数值、验证命令、完整红线）按关注域分册登记在 `docs/debt/`，机制与维护流程见 [`debt/README.md`](./debt/README.md)。
> 本节只保留索引：触碰相关区域前先按「所属分册」读对应条目，任务自检时核对基线不恶化。

| 债务 | 所属分册 | 红线（一句话） |
|---|---|---|
| service/middleware/router import repository | [`debt/architecture.md`](./debt/architecture.md) 第 1 条 | 只减不增 |
| router 重复定义 handler 接口 + `interface{}` 断言 | [`debt/architecture.md`](./debt/architecture.md) 第 2 条 | 禁止重新引入 |
| `RBACService` 生产实例化 | [`debt/architecture.md`](./debt/architecture.md) 第 3 条 | 禁止新增实例化点 |
| 双 User 模型同写 users 表 | [`debt/architecture.md`](./debt/architecture.md) 第 4 条 | 新字段只加 `domain.User` |
| 应用层影子类型层 | [`debt/architecture.md`](./debt/architecture.md) 第 5 条 | 禁止重新引入，类型一律来自 `@myblog/api` |
| follow 模块仅后端 | [`debt/architecture.md`](./debt/architecture.md) 第 6 条 | 关注数据仅经 service 域端点读写 |
| 两 app 基础设施逐字重复 | [`debt/frontend.md`](./debt/frontend.md) 第 1 条 | 修改任一必须同步另一份 |
| admin 认证工具三轨并行 | [`debt/frontend.md`](./debt/frontend.md) 第 2 条 | 禁止新增认证工具文件与双轨回退 |
| admin 胖组件 + onMount 取数 | [`debt/frontend.md`](./debt/frontend.md) 第 3 条 | 新页面禁用 onMount 取数，缺口推回后端 |
| web 首页 load 死代码 | [`debt/frontend.md`](./debt/frontend.md) 第 4 条 | 新页面禁用 load 调认证接口 |
| 401 文案匹配 | [`debt/frontend.md`](./debt/frontend.md) 第 5 条 | 禁止回退文案匹配 |
| admin 重写 `$ui` 已有组件（含 web 分页例外） | [`debt/frontend.md`](./debt/frontend.md) 第 6 条 | 新分页一律 `$ui`，禁止新增第三处实现 |
| web 界面多语言局部接入 | [`debt/frontend.md`](./debt/frontend.md) 第 7 条 | 已接入文件禁止回退硬编码 |
| 字典页窄屏交互与工具栏布局遗留 | [`debt/frontend.md`](./debt/frontend.md) 第 8 条 | 触碰该页必须顺带处理 |
| 公开端点直出实体泄漏个人信息 | [`debt/security-runtime.md`](./debt/security-runtime.md) 第 1 条 | 个人信息须经窄化 DTO 或字段白名单 |
| 文章响应泄漏作者审计字段 | [`debt/security-runtime.md`](./debt/security-runtime.md) 第 2 条 | 新增审计字段默认 `json:"-"` |
| WAF 内容级黑名单的固有误伤面 | [`debt/security-runtime.md`](./debt/security-runtime.md) 第 3 条 | 阻止模式必须先红后绿并配两组用例 |
| 令牌表为内存 map | [`debt/security-runtime.md`](./debt/security-runtime.md) 第 4 条 | 单实例前提；持久化前保持锁 |
| 测试替身内嵌空接口的运行时脆性 | [`debt/testing.md`](./debt/testing.md) 第 1 条 | 受影响 fake 必须显式覆写 |

---

## 9. 历史重构分期路线

> 此处登记历史重构的分期内容与验收口径，供回顾时对照。
> 每完成一项债务修复必须同步更新 `docs/debt/` 对应条目与第 8 节索引，并更新本节状态，保持登记与代码一致。

| 阶段 | 内容 | 完成的债务 | 验收口径 |
|---|---|---|---|
| 阶段一 结构归位与错误分档 | 删 router 重复接口与不可达代码；错误分档（哨兵错误→404/403/400）；令牌表加锁；web 死 load 清理；建立 `contracts/` 目录 | router 重复接口、web 死 load、401 判定均已修复；令牌表已加锁；`RBACService` 生产实例化点仅 1 处；not-found 哨兵→404 已落地，403/400 随错误码契约落地 | ✅ `AGENTS.md` 第 2 节自检命令全绿 |
| 阶段二 类型归位 | 建立 `internal/domain`，合并双 User，service/middleware/router 签名切换为 domain 类型；前端 auth 下沉共享包、影子类型清除 | 双 User 模型与影子类型层已修复；两 app 重复代码大幅减少（auth store 下沉）；service 层 repository 依赖由 12 降至 11 | ✅ 依赖方向基线下降；auth store diff 为零 |
| 阶段三 契约切换 | handler DTO 分离；`contracts/` + 三把锁双向锚定（替代 codegen）；401 改错误码判定 | 401 文案匹配与审计字段泄漏已修复；三把锁已落地（`pnpm run contract:check`） | ✅ 影子类型归零；漂移必当天变红 |
| 阶段四 横切归位与权限下发 | 中间件坍缩为 IdentityProvider 策略；组合根按域装配；RBAC 权限表迁数据源并下发；admin 胖组件拆分、users 搜索推回后端 | `RBACService` 生产实例化点仅 1 处；users 跨页补偿与 admin 分页均已修复；RBAC 迁 config.yaml 完成；permissions 下发完成；follow API 模块已补齐；IdentityProvider 中间件坍缩完成（router 依赖 repository 归零、middleware 1 处）；认证工具已统一 | 权限定义全栈唯一；认证工具单轨 ✅ |
| 阶段五 扩展点 | 后端 ContentRenderer 内容策略接口（当前 `pkg/markdown` 已直接承接 Markdown 渲染，策略接口化待做）；web 业务页面已按"公开数据 SSR + 个性化数据客户端补拉"现状接入，token 迁 cookie 已完成 | — | 新文章类型 = 插入实现，非逐层打洞 |

---

## 10. 历史教训

以下模式在本仓库真实发生过并造成结构性损伤，评审见到同类手法应直接驳回：

1. **类型寄生**：把领域对象/DTO 定义进 `repository` 包（→ 全系统类型倒挂）。
2. **实体直出**：GORM 实体经 `response.Success` 直接序列化为 API 响应（→ 隐私泄漏、契约锁死）。
3. **复制而非提取**：跨 app 逐字拷贝基础设施，"两处保持一致"最终总是失守。
4. **自造影子**：不修镜像源而在应用层再造一套"想象中的 API"类型。
5. **前端补偿后端缺口**：跨页全量拉取 + 客户端过滤（→ 把后端债务放大为前端复杂度）。
6. **私自实例化**：绕过组合根在包内 `New` 依赖（→ 依赖注入形同虚设）。
7. **双轨并存**：同一职责写两遍而不是合并为一处（→ token 刷新双轨、认证工具三轨）。
8. **两端各自演进**：模块只在一端落地。
