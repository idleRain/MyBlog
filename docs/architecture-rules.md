# MyBlog 架构铁律与纪律手册

> 本手册是 AGENTS.md 第 0 节铁律的完整裁决细则，效力高于普通文档。
> 产生背景：2026-09 全栈架构诊断发现，本项目的抽象设计基本正确（分层骨架、packages 单向依赖、工厂注入），
> 但被长期系统性绕行，形成七个结构性顽疾。本手册的使命是让"绕行"在代码评审与自检命令层面被拦截。
>
> **裁决顺序**：`~/.dsh/AGENTS.md`（全局）< `AGENTS.md`（项目）< 本手册（细则）。冲突时以更具体者为准。

---

## 1. 规则分级与核心机制

| 级别 | 含义 | 违反后果 |
|---|---|---|
| 【铁律】(A*) | 依赖方向、真相源、契约、权威、复用、数据加载六类结构性约束 | 必须返工 |
| 【约定】 | 命名、格式、注释等风格约束 | 应当修正 |
| 【债务】(D*) | 已存在的违例，登记编号与基线 | **棘轮机制：只减不增** |

**棘轮机制是本手册的核心执行机制**：存量违规不要求立即修复（那是重构计划的事），但任何改动使债务指标恶化（新增违例文件、新增实例化点、新增重复代码）即判定违规。每条债务附带验证命令，任务完成前 Agent 必须自检。

---

## 2. 铁律 A1：依赖方向

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
| middleware → repository | HTTP 横切层不得直捣存储（终态：依赖 IdentityProvider 抽象） | 已收敛：仅 `identity.go` 实现 1 处 |
| router → repository（仅为拼装中间件） | 路由层不应持有数据访问句柄 | 已清偿：router 依赖 repository 归零 |
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
# 在 server/ 目录下（基线：service 12 + middleware 2 + router 10 = 24 文件）
git grep -ln "MyBlog/internal/repository" -- internal/service internal/middleware internal/router

# 前端：packages 反向依赖应用（应为空）
git grep -ln "apps/web" -- packages
git grep -ln "apps/admin" -- packages
```

---

## 3. 铁律 A2：类型唯一真相源

### 3.1 三层类型镜像关系与权威

```
后端 model 实体 json tag（权威）
    ↓ 手工镜像（目标态：OpenAPI 生成，见 §8 R2）
@myblog/api/modules/*/types（唯一合法镜像）
    ↓ 禁止再造
apps 内任何同构 interface/type（影子类型 = 债务 D7）
```

### 3.2 规则

1. 后端修改实体输出字段（json tag）属于 **wire 格式变更**，必须走 A3 契约流程。
2. 前端需要后端没有的类型（如页面视图模型）：允许定义，但**不得与后端请求/响应结构同构**（同构 = 字段集合基本一致）。视图模型应组合/派生自 `@myblog/api` 类型，而非平行重写。
3. `AuthState` 类认证域类型现状有 3 份定义（两 app store + admin `types/auth.d.ts`），属债务，新增认证类型必须收敛到一处。

### 3.3 验证命令

```bash
# 影子类型合法命中仅限债务文件 apps/admin/src/lib/types/（只减不增）
git grep -n "interface BaseApiResponse" -- apps
```

### 3.4 类型生成策略决策：增强方案 1（三把锁）【2026-09 定稿】

**决策**：不引入 swaggo → OpenAPI → 前端类型生成的 codegen 链路，改用**双向类型锚定**三把锁，以最低成本实现与 codegen 同级的「物理不可漂移」：

| 锁 | 机制 | 落地状态 |
|---|---|---|
| ① | Go handler 测试断言响应与 `contracts/fixtures/` 语义逐字节一致 | ✅ `handler/login_fixture_test.go`（login.wrong-password） |
| ② | TS 侧 vitest + `expectTypeOf(fixture).toEqualTypeOf<手写类型>` 双向精确相等 | ✅ `packages/api/src/contracts/fixtures.test.ts`（首批 3 个） |
| ③ | eslint `no-restricted-imports` 禁止应用层定义同构类型 | ✅ `apps/admin/eslint.config.js` |

漂移必炸链：后端改 → Go 测试红 → 改 fixture → 类型测试红 → 改类型 → check 红。

**门禁**：`pnpm run contract:check`（Go fixture 测试 + vitest 类型锚定）。

**升级触发器**（命中任一则改推 codegen 方案 2）：
1. 接口总数 > 80–100；
2. 出现第二客户端（移动端 / 第三方）；
3. 需对外发布 API 文档；
4. 团队扩张超单人。

**备注**：C1（handler DTO 分离）是单行道决策，与本次选择正交；`article.detail` 金样本的类型锚定待 Author 窄化 DTO 后补充。未来升级 codegen 时 DTO 层零返工。

---

## 4. 铁律 A3：契约先行与模块对齐

### 4.1 wire 格式变更流程（必须按序执行）

1. **影响面声明**：列出后端 DTO、`@myblog/api` types、两 app 消费页面三方清单。
2. **双端同窗口变更**：后端与前端类型/页面改动进同一变更集（或紧邻提交），提交信息注明契约面。
3. **错误语义**：后端错误 message 措辞变更必须同步检查前端 `TOKEN_ERROR_MESSAGES`（债务 D10，前端以文案匹配识别 401）；改错误码语义同理。
4. **认证协议**：token 形状、刷新端点、撤销语义的变更属于最高风险契约变更，须先在本文档 §6.3 登记再动代码。

### 4.2 模块对齐表

新增业务模块必须双端同时登记以下位置，缺一即未完成：

| 端 | 必须创建/修改 |
|---|---|
| 后端 | `internal/model/<entity>.go`（如需）、`internal/repository/<module>.go`、`internal/service/<module>.go`、`internal/handler/<module>.go`、`internal/router/<module>.go`、`main.go` 装配 |
| 前端 | `packages/api/src/modules/<module>/{types.ts, index.ts}`、`packages/api/src/index.ts` 导出、两应用 `src/lib/api/index.ts` 注册 |

现存反例：`user_follow` 模块后端完整、前端零消费（债务 D13）。**此为"两端各自演进"的实证，模块对齐表的目的就是让这类偏差在发生当天可见。**

### 4.3 能力缺口推回原则

前端发现后端接口能力不足（如列表缺关键词过滤）时，**唯一合法动作是给后端加参数**，禁止前端补偿（全量拉取 + 客户端过滤）。现存反例：admin users 页跨页抓取补偿（债务 D8 注记），清偿方式为后端 `users/list` 增加 keyword 参数。

---

## 5. 铁律 A4：单一权威

### 5.1 权威归属表

| 事实 | 唯一权威 | 非法副本（存量债务） |
|---|---|---|
| 角色枚举与权限映射 | `server/internal/service/rbac.go` | 前端 `constants/auth.ts`、`utils/permissions.ts`（D6，终态后端下发） |
| "是否管理员"判定 | 后端 RBAC | middleware 内硬编码字符串、前端复刻 |
| 文章状态流转 | `model.Article` 行为方法 + service | 前端硬编码状态字符串比较 |
| slug 生成 | `pkg/slug` + service 调用 | 前端自写 slugify |
| 密码强度 | 后端 `service/user.go` | 前端正则复刻（仅可做输入提示，不得作为校验依据） |
| token 刷新流程 | 各应用 `src/lib/service/index.ts`（裸 ky 直连版） | admin `utils/jwt.ts` 的 `manualRefreshToken`（D6 双轨，触碰时收敛） |

### 5.2 实例化纪律

- 依赖对象只能由组合根（`main.go`）构造并逐层注入。
- **禁止**在 service 构造函数内部 `New` 另一个 service（存量违例：`NewUserService` 内部 `NewRBACService()`，债务 D4）。
- **禁止**中间件、路由注册函数内部实例化服务（存量违例：`router/user.go`、`middleware/rbac.go` 每请求实例化，债务 D4）。

### 5.3 验证命令

```bash
# RBACService 生产实例化点基线为 4 处：main.go / service/user.go / router/user.go / middleware/rbac.go
# （rbac.go 本体的命中为定义，不计入；测试文件不计入）
git grep -n "NewRBACService()" -- server
```

---

## 6. 铁律 A5/A6：复用纪律与数据加载

### 6.1 复用决策树（新增公共逻辑时）

```
该逻辑是否被两个 app 使用？
├─ 是 → 必须进 packages（shared/http/api，认证域终态为独立 auth 包）
└─ 否 → 留在应用内
     该逻辑与另一 app 现有代码是否同构？
     └─ 是 → 违反 A5，停下提取；否 → 正常开发
```

已知跨 app 重复文件（债务 D5，修改任一必须同步另一份）：`stores/auth.ts`、`service/index.ts`、`components/theme-toggle.svelte`、`routes/+layout.svelte`、`routes/+error.svelte`（同构）。

### 6.2 数据加载

| 应用 | 新页面强制模式 | 存量处理 |
|---|---|---|
| web（SSR 路线） | `+page.server.ts`（需 SEO/会话）或 `+page.ts` load | 清理 D9 死 load 后接入业务 |
| admin（SPA 路线） | 优先 load + `.svelte.ts` 页面状态模块 | 12 个胖组件不强制迁移，触碰时拆分 |

胖组件判定标准（任一命中即应拆分）：单文件 > 300 行；`$state` 声明 > 8 个；组件内直接调用 API 且含增删改多个 handler；组件内含权限判定逻辑。

### 6.3 认证协议事实登记（当前线格式）

改动以下任何一项均属最高风险契约变更，须先更新本节再写代码：

- token pair 形状：`{accessToken, refreshToken, expiresIn}`。
- 线格式：**payload-only JWT**（前端存储与传输的是无点号 Base64 payload，后端 `ReconstructFullToken` 重构完整签名后验证）。
- 刷新：`POST /api/auth/refresh`，body `{refreshToken}`；刷新即旋转（旧 refresh token 撤销）。
- 撤销：内存 map（单实例前提，债务 D11）。
- 前端 401 识别：以响应体业务码 `code === 401` 判定（D10 已清偿，`TOKEN_ERROR_MESSAGES` 文案匹配已移除，禁止回退）。

---

## 7. 债务登记表（棘轮基线）

> 触碰相关区域前先读对应条目的红线；任务自检时核对基线不恶化。基线数值以 2026-09 诊断为准。

| 编号 | 债务描述 | 基线 | 验证命令 | 红线 |
|---|---|---|---|---|
| D1 | service/middleware/router import repository | service **11**+middleware **1**+router **0**（R3 后 router 归零，middleware 仅 identity.go） | 见 §2.3 | 只减不增 |
| D2 | 双 User 模型同写 users 表 | **已清偿（R1）**：合并为唯一 `domain.User` 实体 | `git grep -n "type User struct" -- server/internal --include="*.go"`（仅 domain） | 新字段只加 `domain.User` |
| D3 | router 重复定义 handler 接口 + `interface{}` 断言 | **已清偿（R0）**：router 重复接口 0、断言 0 | `git grep -c "HandlerInterface interface" -- server/internal/router`（应为空） | 禁止回潮 |
| D4 | `RBACService` 生产实例化 | **已收敛（R0）**：仅 main.go 组合根 1 处 | 见 §5.3 | 禁止新增实例化点 |
| D5 | 两 app 基础设施逐字重复 | **大幅清偿（R1）**：auth store 下沉 `@myblog/auth`（202 行×2 → 16 行×2）；service/index.ts、theme-toggle、layout、error 仍重复 | `git diff --no-index apps/web/src/lib/stores/auth.ts apps/admin/src/lib/stores/auth.ts`（应近零差异） | 修改任一必须同步另一份 |
| D6 | admin 认证工具三轨并行 | 5 个 utils 约 969 行；`requireAuth`×2、`isAuthenticated`×3、`performLogout`×2、token 刷新×2 | `git grep -ln "requireAuth\|performLogout\|manualRefreshToken" -- apps/admin/src/lib` | 禁止新增认证工具文件；触碰时收敛双轨 |
| D7 | 影子类型层 | **已清偿（R1）**：`types/api.d.ts` 已删，eslint 守门已加 | `git grep -n "interface BaseApiResponse" -- apps`（应为空） | 禁止回潮；类型一律来自 `@myblog/api` |
| D8 | admin 胖组件 + onMount 取数 | **users 跨页补偿已清偿（R3）**：users/list 增加 keyword 参数；12 个胖组件存量保留（新页面禁用） | `git grep -ln "onMount" -- "apps/admin/src/routes/(admin)"` | 新页面禁用；后端缺口推回后端 |
| D9 | web 首页 load 死代码 | **已清偿（R0）**：`(app)/+page.ts` 死 load 已移除 | 读文件确认 | 新页面禁用 load 调认证接口 |
| D10 | 401 文案匹配 | **已清偿（R0）**：`client.ts` 改为响应体 `code === 401` 判定 | `git grep -n "TOKEN_ERROR_MESSAGES" -- packages`（应为空） | 禁止回退文案匹配 |
| D11 | JWT 撤销无锁内存 map | **已加锁（R0）**：`sync.RWMutex` 保护；deprecated `ValidateToken` 已删 | `git grep -n "revokedTokens" -- server` | 单实例部署前提；持久化前保持锁 |
| D12 | 文章响应泄漏作者审计字段 | **已清偿（R2）**：`lastLoginIP` 等审计字段改为 `json:"-"` | 读 `domain/user.go` json tag | 新增审计字段默认 `json:"-"` |
| D13 | follow 模块仅后端 | **API 模块已补齐（R3）**：`@myblog/api/modules/follow` + 两应用注册；页面消费待 web 业务接入 | `git grep -ln "createFollowAPI" -- packages/api/src`（非空即已补齐） | 页面消费前视为功能未完成 |
| D14 | admin 重写 `$ui` 已有组件 | **已清偿（R3）**：本地 `pagination.svelte` 已删，7 页回归 `$ui` | 目录比对 | 禁止仿效；新分页一律 `$ui` |

---

## 8. 债务清偿分期路线（供重构排期参考）

> 详细重构方案见仓库诊断报告（2026-09）；此处只登记执行顺序与验收口径，供逐步推进时对照。
> 分期推进时，每清偿一项债务必须同步更新第 7 节基线数值与本节状态，保持登记表与代码一致。

| 阶段 | 内容 | 清偿债务 | 验收口径 |
|---|---|---|---|
| R0 止血 | 删 router 重复接口与幽灵代码；错误分档（哨兵错误→404/403/400）；JWT 撤销表加锁；web 死 load 清理；`contracts/` 目录 | **D3、D9 已清偿；D11 已加锁；D4 已收敛；D10 已清偿**；not-found 哨兵→404 已落地，403/400 随错误码契约落地 | ✅ 第 1 节自检命令全绿 |
| R1 类型归位 | 建 `internal/domain`，合并双 User，service/middleware/router 签名切 domain 类型；前端 auth 下沉共享包、影子类型清剿 | **D2、D7 已清偿；D5 大幅清偿（auth store 下沉）；D1 service 12→11** | ✅ D1 下降；auth store diff 为零 |
| R2 契约切换 | handler DTO 分离；`contracts/` + 三把锁双向锚定（替代 codegen）；401 改错误码判定 | **D10、D12 已清偿；三把锁已落地**（`pnpm run contract:check`） | ✅ 影子类型归零；漂移必当天变红 |
| R3 深水区 | 中间件坍缩为 IdentityProvider 策略；组合根按域装配；RBAC 权限表迁数据源并下发；admin 胖组件拆分、users 搜索推回后端 | **D4 已收敛；D8 users 补偿已清偿；D14 已清偿；RBAC 迁 config.yaml 完成；permissions 下发完成；D13 API 模块已补齐；IdentityProvider 中间件坍缩完成**（D1 router/middleware 归零）；D6 认证工具收敛待办 | 权限定义全栈唯一；认证工具单轨 |
| R4 扩展点 | 后端 ContentRenderer 内容策略接口；web SSR 业务接入（token 迁 cookie） | — | 新文章类型 = 插入实现，非逐层打洞 |

---

## 9. 历史教训（诊断提炼，供评审对照）

以下模式在本仓库真实发生过并造成结构性损伤，评审见到同类手法应直接引用编号驳回：

1. **类型寄生**：把领域对象/DTO 定义进 `repository` 包（→ 全系统类型倒挂，D1/D2 的起点）。
2. **实体直出**：GORM 实体经 `response.Success` 直接序列化为 API 响应（→ 隐私泄漏 D12、契约锁死）。
3. **复制而非提取**：跨 app 逐字拷贝基础设施（→ D5），"两处保持一致"最终总是失守。
4. **自造影子**：不修镜像源而在应用层再造一套"想象中的 API"类型（→ D7）。
5. **前端补偿后端缺口**：跨页全量拉取 + 客户端过滤（→ D8，把后端债务放大为前端复杂度）。
6. **私自实例化**：绕过组合根在包内 `New` 依赖（→ D4，依赖注入形同虚设）。
7. **双轨并存**：同一职责写两遍而不是收敛（→ token 刷新双轨、认证工具三轨，D6）。
8. **两端各自演进**：模块只在一端落地（→ D13）。
