# 债务登记：前端与复用

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「前端与复用」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. 两 app 基础设施逐字重复

- **状态**：存量，只减不增。
- **基线**：逐字重复仅 **2 文件 60 行**：`src/lib/stores/auth.ts`（15 行 ×2，SHA256 一致）与 `src/lib/components/theme-toggle.svelte`（45 行 ×2，SHA256 一致）。另有 3 文件职责同构但内容不同，不计入逐字重复：`service/index.ts`（69 / 67 行）、`routes/+layout.svelte`（20 / 12 行）、`routes/+error.svelte`（95 / 38 行，两 app 版式已各自独立设计）。
- **验证命令**：`git diff --no-index apps/web/src/lib/stores/auth.ts apps/admin/src/lib/stores/auth.ts`（应为零差异）；同法核对 `theme-toggle.svelte`。
- **红线**：修改任一逐字重复文件必须同步另一份；禁止新增逐字或近似重复文件，公共逻辑下沉 packages。

## 2. admin 认证工具三轨并行

- **状态**：已统一。
- **基线**：`utils/jwt.ts`、`utils/auth.ts` 已删（约 488 行）；`lib/utils` 现仅 `auth-guard.ts`（路由守卫 `requireAuth`）、`logout.ts`（`performLogout`）、`navigation.ts`、`permissions.ts` 四个文件；刷新单轨在 `service/index.ts`。
- **验证命令**：`git ls-files apps/admin/src/lib/utils`（应无 `jwt.ts`、`auth.ts`）；`git grep -n "manualRefreshToken\|getAuthStatus" -- apps/admin/src`（应为空）。
- **红线**：禁止新增认证工具文件；禁止双轨回退。

## 3. admin 胖组件 + onMount 取数

- **状态**：存量。
- **基线**：口径为「`apps/admin/src/routes` 下单文件 > 300 行」，当前 **7 个**，降序为 tags 454 / links 405 / users 404 / dicts 344 / comments 331 / media 323 / categories 322；路由 svelte 文件共 19 个。`onMount` 取数命中 14 个路由文件。users 跨页补偿已移除，`users/list` 支持 keyword。
- **验证命令**：`git grep -ln "onMount" -- "apps/admin/src/routes"`；行数按上述口径统计。
- **红线**：新页面禁用 onMount 取数，优先 load + 页面状态模块（`.svelte.ts`）；禁止复制 300 行以上胖组件；后端能力缺口推回后端修复。

## 4. web 首页 load 死代码

- **状态**：已修复。
- **基线**：`(app)/+page.ts` 的死 load 已移除。
- **验证命令**：读文件确认，或按 §6.2 的数据加载模式核对新增页面。
- **红线**：新页面禁止用 load 调认证接口。

## 5. 401 文案匹配

- **状态**：已修复。
- **基线**：`packages/http/src/client.ts` 改为响应体 `code === 401` 判定，文案匹配常量已移除。
- **验证命令**：`git grep -n "TOKEN_ERROR_MESSAGES" -- packages`（应为空）。
- **红线**：禁止回退文案匹配。

## 6. admin 重写 `$ui` 已有组件（含 web 侧分页例外）

- **状态**：admin 侧已修复；web 侧分页为受接受存量。
- **基线**：admin 本地 `pagination.svelte` 已删，7 页回归 `$ui`。web 侧保留 `apps/web/src/lib/components/article/PaginationNav.svelte` 的锚点版式，以产出可被爬虫与无 JavaScript 环境跟随的真实 `<a href>`。原登记的技术阻塞「`$ui` 包装层未透传 bits-ui 的 `child` 元素替换通道」经复核**不成立**：`pagination-link.svelte` 未解构 `child`，该属性随 `...restProps` 透传至 `PaginationPrimitive.Page`，且 bits-ui 的 Page/PrevButton/NextButton 类型均带 `WithChild`；若要统一，先实测确认后再迁移。
- **验证命令**：读 `packages/ui/src/pagination/pagination-link.svelte` 的 `...restProps` 透传与 bits-ui 类型的 `WithChild`；`git grep -n "child" -- packages/ui/src/pagination` 的命中均为 `children` 通道，不能作为透传缺口的证据。
- **红线**：admin 侧禁止仿效；新分页一律 `$ui`；web 侧在该项统一前不得新增第三处分页实现。

## 7. web 界面多语言局部接入

- **状态**：存量，分批推进。
- **基线**：语言切换对 Header、Footer、错误页真实生效（含 NotificationBell、FriendlyLinkDialog 两个 Header 子组件与 demo 页），其余页面文案硬编码中文，en 模式下界面为混合语言；词表文件必须保持 JSON 兼容写法（paraglide 编译器按严格 JSON 解析，json5 特性直接编译失败）。
- **验证命令**：`git grep -ln "\$i18n" -- apps/web/src`（已接入面：Header、Footer、NotificationBell、FriendlyLinkDialog、`+error`、demo）。
- **红线**：已接入文件禁止回退硬编码；新增用户可见文案优先经 `m.*` 词表取词；其余页面接入待页面大变动后分批推进。

## 8. 字典页窄屏交互与工具栏布局遗留

- **状态**：存量。
- **基线**：`apps/admin/src/routes/(admin)/dicts/+page.svelte` 的类型列表窄屏交互与工具栏布局打磨未完成。该页同时是 7 个胖组件之一（344 行），两个问题可在同一次触碰中一并处理。
- **验证命令**：读该页的类型列表与工具栏区块在窄屏下的布局。
- **红线**：触碰该页时必须顺带处理，不得再次遗留；不得以「已记录在提交信息」替代债务登记。
