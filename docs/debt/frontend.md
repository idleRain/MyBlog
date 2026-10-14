# 债务登记：前端与复用

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「前端与复用」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. 两 app 基础设施逐字重复

- **状态**：已接受。取舍已定，不构成债务：接缝必须留在应用侧，因为 `packages/auth` 内无法解析 `$lib/api`；上限 10 行，超出即按复制粘贴式共享返工。
- **上限**：逐字重复仅 `src/lib/stores/auth.ts`（10 行 ×2，SHA256 一致）。它是纯注入接缝：`packages/auth` 的 `createAuthStore` 需要 `logoutApi`，而 `$lib/api` 是应用别名、包内无法解析，故接缝必须留在应用侧，**上限 10 行**，超出即按复制粘贴式共享返工。
  - 本轮收敛：`theme-toggle.svelte`（45 行 ×2）下沉至 `packages/components/src/theme-toggle`，两应用副本与 `lib/components/index.ts` 桶一并删除，消费点改为按包名 `@myblog/components/theme-toggle` 导入；应用侧 `isBrowser` 注入移除，环境判定改由包内复用 `@myblog/shared` 的 `isBrowser`。
  - 另有 3 文件职责同构但内容不同，不计入逐字重复：`service/index.ts`（69 / 67 行）、`routes/+layout.svelte`（20 / 12 行）、`routes/+error.svelte`（95 / 38 行，两 app 版式已各自独立设计）。
- **验证命令**：`git diff --no-index apps/web/src/lib/stores/auth.ts apps/admin/src/lib/stores/auth.ts`（应为零差异）；`git ls-files "*theme-toggle.svelte"` 应仅命中 `packages/components` 一处；形态解析用例经 `pnpm run test:packages` 覆盖。
- **红线**：修改任一逐字重复文件必须同步另一份，且该接缝不得超过 10 行；禁止新增逐字或近似重复文件，公共逻辑下沉 packages，跨应用组合组件下沉 `packages/components`。

## 2. admin 认证工具三轨并行

- **状态**：已统一。
- **基线**：`utils/jwt.ts`、`utils/auth.ts` 已删（约 488 行）；`lib/utils` 现为 6 个文件：`auth-guard.ts`（路由守卫 `requireAuth`）、`logout.ts`（`performLogout`）、`navigation.ts`、`permissions.ts`、`tag-display.ts`（标签状态展示解析，列表与表格组件共用同一降级口径）与 `__tests__/tag-display.test.ts`；刷新单轨在 `service/index.ts`。
- **验证命令**：`git ls-files apps/admin/src/lib/utils`（应无 `jwt.ts`、`auth.ts`）；`git grep -n "manualRefreshToken\|getAuthStatus" -- apps/admin/src`（应为空）。
- **红线**：禁止新增认证工具文件；禁止双轨回退。

## 3. admin 胖组件 + onMount 取数

- **状态**：已修复，行数口径与取数口径均已清零。
- **基线**：
  - 行数口径：`apps/admin/src/routes` 下单文件 > 300 行的页面为 **0** 个；修复前为 7 个，降序为 tags 454 / links 405 / users 404 / dicts 344 / comments 332 / media 323 / categories 322。路由 svelte 文件仍为 19 个，新增的 7 个展示组件下沉在 `lib/components/admin/<domain>/`，不进入该口径。
  - 取数口径：`onMount` 命中 **2** 个路由文件，均属鉴权重定向而非取数：`(admin)/+layout.svelte`、`(auth)/login/+page.svelte`。取数命中已归零；修复前为 12 个页面，本轮前余 6 个（仪表盘、notifications、posts、`posts/[id]`、settings、stats）。
  - 已按「`+page.ts` 的 load 取首屏 → `<domain>-page-state.svelte.ts` 持有列表与全部读写操作 → `+page.svelte` 只做渲染」重构 13 个页面：7 个原胖页面（tags / links / users / dicts / comments / media / categories），加本轮 6 个（仪表盘、notifications、posts、`posts/[id]`、settings、stats）。页面状态模块 **11** 个与配套单测 **11** 个，模块用例 **196** 个；`pnpm run test:apps` 下 admin 全量 13 个测试文件、203 个用例通过。
  - 本轮 6 页合计由 996 行降至 596 行。其中仪表盘与 `posts/[id]` 为只读页，无客户端读写操作，由 load 直出 page data，不建空转状态模块；`stats` 页趋势天数与选项新增 `lib/constants/stats.ts` 承载，不再散落字面量。
  - 权限裁剪依赖的登录态在 `@myblog/auth` 的 store 创建时同步水合，故 `+page.ts` 可读 `authStore.getCurrentState()` 裁决区块可见性，load 内不调用认证接口。
- **验证命令**：`git grep -ln "onMount" -- "apps/admin/src/routes"`（应为 2 个文件，均属鉴权）；行数按「`apps/admin/src/routes` 下单文件 > 300 行」统计；状态模块与单测按 `*-page-state.svelte.ts` / `*-page-state.test.ts` 计数（应各 11 个）；在 `apps/admin` 下执行 `pnpm run check && pnpm run test`。
- **红线**：新页面禁用 onMount 取数，一律 load 取首屏；含客户端读写操作的页面配 `<domain>-page-state.svelte.ts` 与配套单测，只读页可 load 直出 page data；load 内禁止调用认证接口，权限裁剪只读本地认证状态；禁止复制 300 行以上胖组件；后端能力缺口推回后端修复；`lib/constants/*.ts` 不得引入 `@lucide/svelte`，图标映射随其唯一消费者组件声明，否则页面状态模块与单测的编译图会带入图标桶（comments 单测实测由 126 秒降至 5 秒）。

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

- **状态**：admin 侧已修复；web 侧分页为已接受取舍。
- **基线**：admin 本地 `pagination.svelte` 已删，7 页回归 `$ui`。web 侧保留 `apps/web/src/lib/components/article/PaginationNav.svelte` 的锚点版式，以产出可被爬虫与无 JavaScript 环境跟随的真实 `<a href>`。原登记的技术阻塞「`$ui` 包装层未透传 bits-ui 的 `child` 元素替换通道」经复核**不成立**：`pagination-link.svelte` 未解构 `child`，该属性随 `...restProps` 透传至 `PaginationPrimitive.Page`，且 bits-ui 的 Page/PrevButton/NextButton 类型均带 `WithChild`；若要统一，先实测确认后再迁移。
- **验证命令**：读 `packages/ui/src/pagination/pagination-link.svelte` 的 `...restProps` 透传与 bits-ui 类型的 `WithChild`；`git grep -n "child" -- packages/ui/src/pagination` 的命中均为 `children` 通道，不能作为透传缺口的证据。
- **红线**：admin 侧禁止仿效；新分页一律 `$ui`；web 侧在该项统一前不得新增第三处分页实现。

## 7. web 界面多语言局部接入

- **状态**：推进中，分批接入。
- **基线**：已接入面 **12** 个文件——Header、Footer、NotificationBell、FriendlyLinkDialog、`+error`、demo 页，以及**首页整面**：`(app)/+page.svelte`（SEO 与年轮预览）、`(app)/+page.ts`（加载失败文案）与 `lib/components/home/` 下 HeroSection、PostIndexSection、FeaturedStory、BackToTop 四个组件。Header 的站名与作者身份已并入 `ui` 共享词条（`ui:site.name`、`ui:author.*`、`ui:brand.github`、`ui:meta.*`），页面标题不再引用中文站名常量。其余页面文案仍硬编码中文，en 模式下界面为混合语言。
- **词表约定**：按面拆分为独立 namespace 并在 `project.inlang/settings.json` 的 `pathPattern` 注册，`ui` 只承载跨面 chrome 与共享词；插值必须用 i18next 双花括号 `{{param}}`，单花括号会被当作字面量原样输出；计数词条写 `_one`/`_other` 两个变体，参数名必须为 `count` 才参与 `Intl.PluralRules` 选择，变体缺失时取词回退为原始 key 字符串；词表必须保持 JSON 兼容写法（paraglide 编译器按严格 JSON 解析，json5 特性直接编译失败）。
- **验证命令**：`git grep -ln "\$i18n" -- apps/web/src`（当前 12 个文件，逐批递增）；`pnpm --filter @myblog/web run generate:i18n` 后核对 `messages/{zh,en}` 各 namespace 的键集合与条数一致（当前 ui 74 / home 40 / auth 7 / common 4）。
- **红线**：已接入文件禁止回退硬编码，不得在已接入文件里新增中文字面量；新增用户可见文案优先经 `m.*` 取词，页面文案进本面 namespace 不塞回 `ui`；其余页面待页面大变动后分批推进。

## 8. 字典页窄屏交互与工具栏布局遗留

- **状态**：已修复。
- **基线**：`apps/admin/src/routes/(admin)/dicts/+page.svelte` 由 344 行收敛至 174 行，类型列表与字典项表格分别下沉 `lib/components/admin/dict/dict-type-list.svelte` 与 `dict-item-table.svelte`，状态展示配置下沉 `lib/constants/dict.ts`。窄屏交互修复项：类型列表行原先把编辑与删除按钮嵌在选择行按钮内部，构成非法的按钮嵌套并导致窄屏点击区域争抢，现改为选择区按钮与行内操作按钮并列；操作按钮统一 `size="icon"` 并带含字典类型名的 `aria-label`；选择区补 `focus-visible` 焦点环与 `aria-current`。工具栏与表格修复项：字典项卡片头部改 `flex-wrap` 避免窄屏挤压，描述列在窄屏并入显示名列副行并按 `md` 断点恢复独立列，排序列在 `sm` 以下折叠；加载态改用 `$ui/spinner`，空态改用 `$ui/empty`，替换原先手写的旋转圆环与空态容器。
- **验证命令**：读 `lib/components/admin/dict/dict-type-list.svelte` 确认选择区与操作区为并列元素；`get-childitem "apps/admin/src/routes/(admin)/dicts/+page.svelte"` 行数应低于 300；在 `apps/admin` 下执行 `pnpm run check` 与 `pnpm run test`。
- **红线**：触碰该页时必须顺带处理，不得再次遗留；列表行的选择与操作不得回退为嵌套交互元素。
