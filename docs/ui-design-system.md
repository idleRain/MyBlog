# MyBlog UI 设计系统（Design System）

> 本文档是 MyBlog **全站 UI 设计与开发的唯一准则**，取代此前前后台拆分与 UI 重构过程中
> 已归档并删除的历史 plan / handoff 系列文档。前台（`apps/web`）、后台（`apps/admin`）
> 与公共包（`packages/ui`）的所有 UI 改动均以本文档为验收依据。

---

## 1. 设计理念与架构

### 1.1 双主题架构

前端在 web-split 重构后拆分为两个独立 SvelteKit 应用，各自维护**独立的视觉主题**：

| 应用 | 主题 | 视觉隐喻 | 圆角基调 | 强调色 |
| --- | --- | --- | --- | --- |
| `apps/web`（前台 toC） | **编辑杂志主题** | 纸质刊物 / 编辑手记 | 锐利直角 `0.25rem` | 朱红 `signal` |
| `apps/admin`（后台 toB） | **shadcn 原始主题** | shadcn-svelte 默认 | 常规圆角 `0.625rem` | slate 蓝 `primary` |

前台以「编辑杂志风」为视觉方向：像一本纸质刊物，而不是数字产品模板。大面积暖纸色与墨色构成安静基底，朱红作为唯一锐利强调色；标题使用衬线字体承担刊头气质，目录、时间线等版式强调「翻阅索引」的阅读体验。

两套主题互不共享样式文件，分别由各自 `app.css` 的 CSS token 定义。`packages/ui` 不携带任何全局样式。

### 1.2 `packages/ui` 边界（stock 原则）

- `packages/ui` 承载 shadcn-svelte 全部基础组件，**必须保持 stock（原始）样式**。
- 组件源码**禁止硬编码业务色与业务样式**，外观一律通过各应用 `app.css` 的语义 token 与调用处 `className` prop 定制。
- 前后台的外观差异全部落在各自 `app.css`，**不得**通过修改 `packages/ui` 组件源码实现。
- 组件库内仅使用标准 shadcn token（`--background` / `--foreground` / `--card` / `--primary` / `--muted` / `--border` / `--ring` / `--destructive` / `--radius` 等）。

### 1.3 主题注入机制（三件套）

任一应用引用 `packages/ui` 源码直连组件时，必须同时具备以下三项，缺一会导致「类不生成」或「SSR 报错」：

1. **`svelte.config.js` 别名**：`$ui` → `../../packages/ui/src`，`$ui/*` → `../../packages/ui/src/*`。
2. **`vite.config.ts`**：`ssr.noExternal: ['@myblog/ui']`，使 `.svelte` 源码参与 SSR 编译。
3. **`app.css` 的 `@source`**：`@source '../../../packages/ui/src'`，让 Tailwind v4 扫描包内组件类。

Token 经 `@theme inline` 映射为 Tailwind 工具类（`--color-*` → `bg-*` / `text-*` 等），组件才能在应用中直接使用。

---

## 2. 色彩系统

**取色硬约束**：所有颜色必须通过语义 token（`bg-signal`、`text-muted-foreground`、`border-line` 等）取用，**禁止硬编码色值**。唯一例外见 8.1 的「精选封面版画」既定画布色板。

### 2.1 前台编辑杂志主题（`apps/web/src/app.css`）

**基调**：「编辑杂志风」暖纸墨色系，色板源自参考稿（纸 `#f7f3ea`、墨 `#221d16`、朱红 `#c83e1d`）。暗色为配套的「深夜书房」暖黑纸墨。

#### 亮色模式

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `#f7f3ea` | 暖纸底色 |
| `--foreground` | `#221d16` | 主文字（墨色） |
| `--card` | `#fbf7ec` | 卡片面 |
| `--secondary` | `#efe8d9` | 纸深面（侧栏卡片、页脚） |
| `--muted` | `#f0e9db` | 弱背景 |
| `--muted-foreground` | `#6f6555` | 次级文字（墨灰） |
| `--accent` | `#e9e0cd` | 悬停/高亮背景 |
| `--primary` | `#221d16` | 主按钮（墨色） |
| `--destructive` | `oklch(0.577 0.245 27.325)` | 危险操作 |
| `--border` / `--input` | `#ded3bf` | 分隔线 / 边框 |
| `--ring` | `#c83e1d` | 焦点环（朱红） |
| `--signal` | `#c83e1d` | **强调色（朱红）** |
| `--signal-foreground` | `#f7f3ea` | 红底上的纸色文字 |

#### 暗色模式（深夜书房）

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `#171310` | 暖黑墨底 |
| `--foreground` | `#ece3d1` | 主文字（陈纸亮调） |
| `--card` | `#1f1914` | 卡片面 |
| `--secondary` | `#2a2219` | 纸深面 |
| `--muted` | `#241d16` | 弱背景 |
| `--muted-foreground` | `#a3947c` | 次级文字 |
| `--primary` | `#ece3d1` | 主按钮（纸白） |
| `--border` / `--input` | `#372e23` | 发丝线 |
| `--ring` / `--signal` | `#d8572f` | 焦点环 / 强调色（提亮朱红） |
| `--signal-foreground` | `#1a120c` | 红底上的深墨文字 |

图表与侧边栏 token（`--chart-1..5`、`--sidebar-*`）随主题在 `app.css` 中定义，业务代码只经 `@theme inline` 映射后的工具类取用。

### 2.2 后台 shadcn 原始主题（`apps/admin/src/app.css`）

**基调**：shadcn-svelte 默认 slate 蓝主题，纯白底，**不含 `--signal`**。后台不使用前台朱红强调色。

#### 亮色模式

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(1 0 0)` | 纯白底 |
| `--foreground` | `oklch(0.129 0.042 264.695)` | 主文字 |
| `--card` | `oklch(1 0 0)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.968 0.007 247.896)` | 弱背景 |
| `--muted-foreground` | `oklch(0.554 0.046 257.417)` | 次级文字 |
| `--primary` | `oklch(0.208 0.042 265.755)` | 主按钮（slate 蓝） |
| `--destructive` | `oklch(0.577 0.245 27.325)` | 危险操作 |
| `--border` / `--input` | `oklch(0.929 0.013 255.508)` | 边框 |
| `--ring` | `oklch(0.704 0.04 256.788)` | 焦点环 |

#### 暗色模式

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(0.129 0.042 264.695)` | 墨蓝底 |
| `--foreground` | `oklch(0.984 0.003 247.858)` | 主文字 |
| `--card` | `oklch(0.208 0.042 265.755)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.279 0.041 260.031)` | 弱背景 |
| `--muted-foreground` | `oklch(0.704 0.04 256.788)` | 次级文字 |
| `--primary` | `oklch(0.929 0.013 255.508)` | 主按钮（亮蓝白） |
| `--border` | `oklch(1 0 0 / 10%)` | 边框 |
| `--ring` | `oklch(0.551 0.027 264.364)` | 焦点环 |

### 2.3 取色约定

- 前台强调色统一用 `signal`（`text-signal` / `bg-signal` / `border-signal`），透明度用 `/10`、`/30`、`/90`。
- `signal` 经 `@theme inline` 暴露为 `--color-signal`，可直接使用 Tailwind 工具类；若 `bg-signal` 不生效，检查 `@theme inline` 中是否声明 `--color-signal`。
- 前台在 `@theme inline` 中额外暴露 `--color-line: var(--border)` 别名，延续参考稿「线框色 line」的编辑版式词汇（`border-line` / `divide-line` / `bg-line`）。
- 后台不得使用 `signal`（该 token 未定义）；状态语义色走 `primary` / `muted` / `destructive`。
- 语义色：成功/危险等状态用 `--destructive` 与 `--muted-foreground`，**不**引入额外强调色。

---

## 3. 字体系统

### 3.1 前台（`apps/web`）

| 角色 | 字体 | 字重 | 引入方式 |
| --- | --- | --- | --- |
| 标题（`font-display`） | Noto Serif SC（思源宋体） | 500 / 700 / 900 | `@fontsource/noto-serif-sc` 自托管 |
| 正文（`font-sans`） | Noto Sans SC（思源黑体） | 400 / 500 / 700 | `@fontsource/noto-sans-sc` 自托管 |
| 等宽（技术标注，`font-mono`） | Fira Mono | 400 / 500 / 700 | `@fontsource/fira-mono` 自托管 |

- 三者均在 `app.css` 顶部按 `chinese-simplified-*.css` 分包引入，Tailwind v4 经 `@theme` 的 `--font-display` / `--font-sans` / `--font-mono` 映射为工具类。
- 全局 `h1..h6` 默认应用 `--font-display`（衬线），正文与 UI 组件应用 `--font-sans`。
- CJK 字体按 unicode-range 分包加载，仅实际用到的字符子集会产生请求。

### 3.2 后台（`apps/admin`）

- 正文声明 `'Manrope'`、标题声明 `'Inter'`，但**两者尚未经 `@fontsource` 自托管**，当前实际回退系统字体栈。如需正式启用，必须自托管（见 3.3）；否则应从 `font-family` 移除以免误导。

### 3.3 自托管约束

- **禁止**新增 Google Fonts CDN 引用（大陆网络不可达）。
- 新增字体必须走 `@fontsource` 自托管并加入依赖，使用前在 `app.css` 的 `@theme` 中完成映射。

---

## 4. 圆角、纹理与间距

### 4.1 前台（编辑杂志，锐利直角）

- 全局 `--radius: 0.25rem`，使 `rounded-md` 锐利化。
- 卡片 / 按钮 / 输入框：`rounded-none` 或继承 `rounded-md`，**禁止**装饰性 `rounded-full` 与 `rounded-xl`。
- 保留的功能性圆形：头像 `avatar`、开关 `switch`、单选 `radio-group`、进度 `progress`、滚动区 `scroll-area`、轮播翻页按钮、抽屉拖拽把手、回到顶部浮动按钮、时间线节点圆点。
- 纸纹底纹：`.texture-grid`（印刷细网格，基于 `color-mix` 前景色叠加，明暗主题自适应），施加于首页正文容器。

### 4.2 后台（shadcn 默认）

- 全局 `--radius: 0.625rem`，沿用 shadcn-svelte 默认圆角体系，不额外收紧。

### 4.3 间距与容器

- 前台容器宽度统一 `max-w-6xl`（杂志版心）。
- 前台页边距统一 `px-4 sm:px-6`，与参考稿版心节奏一致。
- 后台沿用 shadcn 常规间距体系。

---

## 5. 排版层级

### 5.1 前台（编辑杂志排版）

| 层级 | 规格 | 用途 |
| --- | --- | --- |
| Display | `font-display text-5xl sm:text-6xl lg:text-7xl font-black leading-[1.08] tracking-tight` | 首页主标题（刊头） |
| Section 标题 | `font-display text-3xl font-black` + `border-b border-line` 标题栏 | 板块标题 |
| 条目标题 | `font-display text-xl font-bold leading-snug` | 目录行 / 时间线标题 |
| Body | `text-base sm:text-lg leading-relaxed text-muted-foreground` | 正文与摘要 |
| Meta | `font-mono text-xs text-muted-foreground` | 日期、阅读时长、序号、刊号等标注 |

编辑杂志特征元素：印刷细网格纸纹（`.texture-grid`）、目录行「编号 + 日期 + 标题」版式（编号 `font-display` 淡色，悬停转朱红）、朱红竖线引语块（`border-l-2 border-signal`）、时间线年度轨道、大号衬线幽灵刊号（`text-foreground/[0.05]`）。

### 5.2 后台（shadcn 默认）

沿用 shadcn-svelte 标准排版与组件层级，无杂志式刊头与幽灵刊号要求。

---

## 6. 动效规范

动效遵循 Emil 设计工程规范，只做「必要且有目的」的动效。滚动驱动动画统一使用 GSAP。

### 6.1 缓动与时长基准

- **缓动**：进入用 `cubic-bezier(0.23, 1, 0.32, 1)`（强 ease-out，GSAP 对应 `power4.out`）；屏内位移用 `cubic-bezier(0.77, 0, 0.175, 1)`（强 ease-in-out，GSAP 对应 `power4.inOut`）；滚动驱动一律 `linear`，平滑度交给 scrub。
- **时长**：UI 微交互 150–250ms；常规滚动进场 0.6s；首屏与大型版面的编辑感进场 0.9s；封面 `clip-path` 展开等仪式性动效 1s。
- **错峰**：组内元素依次浮现间隔 30–80ms（`MOTION.stagger = 0.07`）。
- **按钮按压**：`:active` 时 `scale(0.97)`。
- **过渡属性**：只过渡 `transform` / `opacity` / `color` / `border-color` / `background-color`，**禁止 `transition-all`**。
- **性能**：只动画 `transform` 与 `opacity`，不动画 `width` / `height` / `margin` / `padding`；进度条用 `scaleX` 而非 `width`。
- **离场快于进场**：出现与消失使用非对称时长（如回到顶部按钮进 0.3s、出 0.2s）。

### 6.2 GSAP 动效基建（`apps/web/src/lib/motion/`）

- **`gsap-setup.ts`**：全站唯一插件注册点（`ScrollTrigger`），统一导出 `gsap` 与 `MOTION` 常量（时长 / 错峰 / 缓动）。组件**禁止**各自注册插件。
- **`reveal.ts`**：`scrollReveal`（单元素滚动进场，支持 `x` / `y` / `delay`）与 `scrollStagger`（组内错峰）两个 Svelte action。两者内部经 `gsap.matchMedia` 适配 `prefers-reduced-motion`，减少动态偏好下不注册任何动画。
- **组件侧模式**：`$effect` 中先取 `const element = root` 再判空（`$state` 变量在嵌套闭包内不保留类型收窄），以 `gsap.context(fn, element)` 建立作用域，返回 `ctx.revert()` 清理。
- **滚动图元约定**：
  - 阅读进度条：`scaleX` + `scrub: 0.3`，`end: 'max'` 挂整页滚动。
  - 时间线年度轨道：`data-rail` / `data-rail-progress` 数据属性标记，`fromTo` 显式声明起止，`scrub: true`。
  - `clip-path` 动画必须用 `gsap.fromTo`，起止值均显式声明，自然值 `none` 无法参与插值。
  - 首屏视差退场：`trigger` 挂 Hero 区块自身，`start: 'top top'` / `end: 'bottom top'`。
- **禁止整屏滚动劫持**：不得拦截 `wheel` / 键盘事件做页面级吸附，滚动始终交还原生。

### 6.3 后台动效

- 沿用 shadcn 组件自带过渡。
- 后台登录页保留 blob 装饰动画（`animate-blob`），属历史保留的登录页特色，不属于前台杂志主题反模式约束范围；后续如需清理应单独立项。

### 6.4 减少动态

- GSAP 侧统一经 `gsap.matchMedia('(prefers-reduced-motion: no-preference)')` 门控，减少动态偏好下不注册位移与渐隐动画。
- CSS 侧装饰性循环动画（如滚动提示流动线）在 `@media (prefers-reduced-motion: reduce)` 下停用，保留静态形态。
- 减少动态意味着更少与更温和的动效，而非零动效；颜色与透明度的理解性过渡保留。

---

## 7. 组件风格约定

### 7.1 前台（编辑杂志风格）

| 组件 | 规范 |
| --- | --- |
| 主按钮 | `bg-primary text-primary-foreground hover:bg-signal hover:text-signal-foreground active:scale-[0.97]` |
| 次按钮 | `border border-border hover:border-signal hover:text-signal active:scale-[0.97]` |
| 面板卡片 | `border border-line bg-secondary`（纸深面），内边距 `p-6` |
| 目录行 | 「编号 + 日期 + 标题」索引行，编号 `font-display` 淡色（`var(--border)`）悬停转 `var(--signal)`，左侧朱红竖条悬停 `scaleY` 展开 |
| 时间线 | 年度网格 + 粘性年份标记 + 左侧轨道（基线 `bg-line`，进度线 `bg-signal` 随滚动绘制）+ 节点圆点（`border-2 border-signal`） |
| Badge | `bg-signal px-2 py-0.5 text-xs font-bold text-signal-foreground`（精选标）或 `border border-line bg-card`（标签徽章） |
| 订阅面板 | `bg-primary text-primary-foreground` 墨底反白，按钮 `bg-signal` |
| Logo 字母标 | 直角描边方块 + `font-mono text-signal` 字母，站名用 `font-display` |
| 图标 | 一律 Lucide SVG（`@lucide/svelte`），尺寸 `h-4 w-4` 起步；禁止 emoji 作图标 |

### 7.2 后台（shadcn 默认）

- 全部使用 `packages/ui` 的 stock 组件，样式经后台 `app.css` token 生效。
- 状态/角色 Badge 等走 shadcn 变体（`secondary` / `outline` / `destructive` 等），不硬编码颜色类。

### 7.3 `packages/ui` 维护流程

- `components.json` 位于 `packages/ui` 根目录，新增组件在包内执行 `npx shadcn-svelte@latest add <组件>`。
- 各应用不各自维护 `components.json`。
- 新增组件保持 stock；外观定制一律经应用 `app.css` token 与调用处 `className` 完成。

---

## 8. 反模式清单

### 8.1 前台（严格禁止）

改造与后续开发中，`apps/web/src` 下 grep 以下关键词应返回 0（功能性 loading 除外）：

- `bg-clip-text text-transparent`（渐变文字）
- `animate-pulse` / `animate-bounce` / `animate-blob`（装饰性动画；骨架屏 `skeleton` 的 pulse 除外）
- emoji 作为 UI 图标（`☕ 📚 🎵 🌱 🎨 🔧 💡 🏃 🎯 🚀 ✨` 等）
- `particles` / canvas 粒子背景
- `rounded-full`（卡片/按钮/徽章；功能性圆形除外）
- `transition-all`
- `glow-*` / `text-blue-*` / 冷蓝、紫、粉残留
- 硬编码色值（应走 token；唯一例外见下）
- wheel / 键盘事件拦截式整屏滚动劫持

**唯一例外——精选封面版画**：`FeaturedStory` 的封面是杂志版面的「画布」，允许使用内容性渐变（`bg-gradient-to-br from-[#3a2f26] via-[#6b4f3a] to-[#c83e1d]`）。该色板为既定艺术设定，不进入语义 token 体系；除封面外任何位置不得复用此渐变或色值。

### 8.2 后台（相对宽松）

- 后台遵循 shadcn 原始主题，不强制前台杂志主题反模式。
- 但**禁止**引入前台朱红之外的第三种强调色、禁止渐变文字与粒子背景，保持与前台一致的基本克制。

---

## 9. 现状记录与待办

| 项 | 现状 | 建议 |
| --- | --- | --- |
| 前台 `.texture-grid` 纸纹 | 已启用，施加于首页正文容器 | 随页面扩充按需复用 |
| 前台 `.spec-grid` | 无组件引用，为死代码 | 后台登录页如需网格底纹可复用；否则待清理 |
| 前台 `.animate-blob` / 延迟类 | 前台无组件引用，为死代码 | 随清理项一并移除 |
| 首页文章数据 | `lib/data/home-content.ts` 占位内容 | 文章业务接入后由 `@myblog/api` 真实数据替换 |
| 后台 Manrope / Inter 字体 | 已声明于 `font-family` 但未自托管，实际回退系统栈 | 自托管或移除声明 |
| 后台登录页 blob 动画 | 历史保留的登录页特色 | 保持；如需清理单独立项 |
| Header 触发器 button 嵌套 | `Dialog.Trigger`/`Sheet.Trigger` 包裹 `$ui` Button 产生 `hydration_mismatch` 警告 | 触碰 Header 时顺带修复（改用 `Button asChild` 或原生按钮） |

---

## 10. 开发与校验

1. 取色必须走 token；动效时长、延迟等必须使用 `MOTION` 常量或命名常量，禁止魔法数字。
2. 前台杂志版式元素（目录行、朱红竖线引语、幽灵刊号、纸纹）须与既有组件（`lib/components/home/` 各版面、`Header` / `Footer`）保持一致的实现方式。
3. 新增滚动动效时：插件只在 `gsap-setup.ts` 注册；进场优先使用 `scrollReveal` / `scrollStagger` actions；复杂编排参考 `HeroSection.svelte` 的 timeline 模式；清理一律走 `ctx.revert()`。
4. 改动后运行对应校验：
   - 前台：`cd apps/web && pnpm run check`（类型检查）+ `pnpm --filter @myblog/web lint`
   - 后台：`cd apps/admin && pnpm run check`
   - 全量：`pnpm run quality`
5. 动效验收：慢速播放或逐帧检查时序；确认控制台无 GSAP 警告；开启系统减少动态偏好后页面保持静态可读。
6. 反模式自查：按 8.1 在 `apps/web/src` 下 grep 验证。
7. 注释与对话使用简体中文；注释为完整技术陈述句并以句号结尾，无括号补充、无口语词、无魔法数字。
