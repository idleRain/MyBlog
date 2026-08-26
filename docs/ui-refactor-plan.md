# MyBlog 全站 UI 重构开发计划

> 本文档是 MyBlog 前端（`web/`）全站视觉与交互重构的**唯一执行依据**。
> 执行 Agent 应严格按本文档的设计系统规范与阶段顺序改造，逐阶段完成后对照「验收标准」自查。

---

## 0. 目标与原则

将当前「蓝→紫→粉渐变 + 粒子背景 + 渐变文字 + 发光描边」的模板化、AI 味浓厚的外观，统一替换为一套**「工程规格书 / 技术编辑」**风格：

- **与众不同**：以「工程图纸 / 规格书」为视觉隐喻，用网格、发丝线、等宽标注、定位标记构成记忆点。
- **简洁**：克制的层级，单一强调色，大量留白，不做无意义的装饰。
- **设计前卫**：不对称排版、锐利直角、编辑式大字标题、幽灵序号等设计语言的介入。
- **线条分明**：以 1px 发丝线、直角、清晰边界替代柔和的阴影与渐变。

**硬性原则（贯穿所有改造）：**

1. 只允许**一种强调色**（signal 国际橙），其余为中性纸墨色。
2. 全局**禁止 CSS 渐变**用于装饰（背景、文字、描边、按钮、分割线）。
3. **禁止 emoji 作为图标**，一律使用 Lucide SVG 图标或文字序号。
4. **禁止装饰性动画**（`animate-pulse` / `animate-bounce` / 粒子网络 / 发光文字）；功能性的 loading spinner 与 skeleton 除外。
5. 卡片/按钮一律**直角或极小圆角**，禁止 `rounded-full`（头像、开关、单选等功能性圆形除外）。

---

## 1. 设计系统规范（The Spec Sheet）

### 1.1 设计概念

- **隐喻**：工程规格书 / 技术蓝图（blueprint / spec sheet）。
- **特征元素**：坐标网格背景、页边出血线、四角裁切标记、大号幽灵序号、等宽「代码注释式」标注（`// ...`）、发丝线分割。
- **气质**：克制、精确、编辑感，而非「炫技」。

### 1.2 色彩系统

颜色已统一收敛为 `app.css` 中的 CSS 变量，**新代码必须通过语义 token（`bg-signal`、`text-muted-foreground` 等）取色，禁止硬编码色值**。

**亮色模式：**

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(0.982 0.004 95)` | 暖纸底色 |
| `--foreground` | `oklch(0.18 0.012 85)` | 主文字（墨色） |
| `--card` | `oklch(0.992 0.002 95)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.945 0.006 95)` | 弱背景 |
| `--muted-foreground` | `oklch(0.5 0.016 85)` | 次级文字 |
| `--border` | `oklch(0.882 0.007 95)` | 发丝线 / 边框 |
| `--primary` | `oklch(0.18 0.012 85)` | 主按钮（墨色） |
| `--ring` | `oklch(0.63 0.23 30)` | 焦点环（橙） |
| `--signal` | `oklch(0.63 0.23 30)` | **强调色（国际橙）** |
| `--signal-foreground` | `oklch(0.17 0.02 30)` | 橙底上的深墨文字 |

**暗色模式：**

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(0.16 0.007 95)` | 墨底 |
| `--foreground` | `oklch(0.93 0.005 95)` | 主文字 |
| `--card` | `oklch(0.19 0.008 95)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.24 0.009 95)` | 弱背景 |
| `--muted-foreground` | `oklch(0.68 0.012 95)` | 次级文字 |
| `--border` | `oklch(0.27 0.009 95)` | 发丝线 |
| `--primary` | `oklch(0.93 0.005 95)` | 主按钮（纸白） |
| `--signal` | `oklch(0.68 0.21 32)` | **强调色（亮橙）** |
| `--signal-foreground` | `oklch(0.15 0.02 32)` | 橙底上的深墨文字 |

**取色约定：**

- 强调色统一用 `signal`（`text-signal`、`bg-signal`、`border-signal`），透明度用 `/10`、`/30`、`/90`。
- `signal` 已通过 `@theme inline` 暴露为 `--color-signal`，可直接使用 Tailwind 工具类；若某文件里 `bg-signal` 不生效，检查是否在 `@theme inline` 中声明了 `--color-signal`。

### 1.3 字体系统

- **等宽（技术标注专用）**：Fira Mono（400/500/700），已通过 `@fontsource/fira-mono` 自托管，无需联网、兼容大陆网络。
  - 通过 `font-mono` 工具类使用（`@theme` 中已覆盖 `--font-mono`）。
- **正文与标题**：系统 CJK 栈 `-apple-system, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif`。
- **禁止**新增 Google Fonts CDN 引用（大陆被墙）；如需新增字体，必须走 `@fontsource` 自托管并 `bun add`。

**用法约定：**

- 标签、序号、元信息、按钮文字、技术名词：`font-mono text-xs tracking-[0.18em] uppercase text-muted-foreground`。
- 大标题：`text-5xl sm:text-7xl lg:text-8xl font-bold tracking-tight text-balance`。
- 强调词用 `text-signal`，不做渐变文字。

### 1.4 圆角与间距

- 全局 `--radius: 0.25rem`，使 shadcn 的 `rounded-md`（映射到 `--radius-md`）锐利化。
- 卡片/按钮/输入框：`rounded-none` 或继承 `rounded-md`，**禁止** `rounded-full` 与 `rounded-xl`（装饰性）。
- 容器宽度统一 `max-w-6xl` 或 `max-w-7xl`，页边距统一 `px-6 sm:px-10 lg:px-16/20`。

### 1.5 排版层级

| 层级 | 规格 | 用途 |
| --- | --- | --- |
| Display | `text-5xl→8xl / bold / tracking-tight` | 页面主标题 |
| Section 标题 | `text-3xl sm:text-4xl / bold` + 发丝线 | 板块标题 |
| Body | `text-base sm:text-lg / leading-relaxed / text-muted-foreground` | 正文 |
| Meta | `font-mono text-xs / uppercase / tracking-[0.18em] / text-muted-foreground` | 标注、序号、元信息 |

### 1.6 动效规范

遵守 Emil 设计工程规范，只做「必要且有目的」的动效：

- **缓动**：进入用 `cubic-bezier(0.23, 1, 0.32, 1)`（ease-out）；屏内位移用 `cubic-bezier(0.77, 0, 0.175, 1)`（ease-in-out）。
- **时长**：UI 微交互 150–300ms；按钮按压反馈 100–160ms；首屏营销性入场可放宽。
- **入场**：首屏采用阶梯式浮现（30–80ms 间隔），`translateY(16px) + opacity`，**不要**从 `scale(0)` 出现。
- **按钮按压**：`:active` 时 `scale(0.97)`。
- **过渡**：只过渡 `transform`/`opacity`/`color`/`border-color`/`background-color`，**禁止 `transition-all`**。
- **减少动态**：所有位移动画必须在 `@media (prefers-reduced-motion: reduce)` 下关闭位移与渐隐。
- **性能**：只动画 `transform` 与 `opacity`，不动画 `width/height/margin/padding`。

### 1.7 组件风格约定

| 组件 | 规范 |
| --- | --- |
| 主按钮 | `rounded-none bg-signal text-signal-foreground hover:bg-signal/90`，可带 `font-mono` |
| 次按钮 | `rounded-none border border-border hover:border-signal hover:text-signal` |
| 卡片 | `rounded-none border border-border bg-card`，悬停用 `hover:border-signal` 替代位移与重阴影 |
| Badge | `rounded-none`，分类色用 `bg-signal/10 text-signal` 或 `bg-muted text-muted-foreground` |
| 分割线 | `border-t border-border` 或 `h-px bg-border`，**禁止**渐变分割线 |
| Logo 字母标 | 直角描边方块 + `font-mono text-signal` 字母 |

### 1.8 反模式清单（全局禁止）

改造后全站 grep 以下关键词应返回 0（功能性 loading 除外）：

- `bg-gradient-*` / `from-blue-*` / `via-purple-*` / `to-pink-*`（渐变）
- `bg-clip-text text-transparent`（渐变文字）
- `animate-pulse` / `animate-bounce` / `animate-blob`（装饰性动画）
- emoji（`☕ 📚 🎵 🌱 🎨 🔧 💡 🏃 🎯 🚀 ✨` 等作为 UI 图标）
- `particles` / canvas 粒子背景
- `rounded-full`（卡片/按钮/徽章；头像/开关/单选/进度条等功能性圆形除外）
- `transition-all`（应为具体属性）
- `glow-*` / `text-blue-400` 等发光与冷蓝残留

---

## 2. 当前基线（已完成）

以下文件已完成改造，作为后续风格参照与一致性基准：

| 文件 | 已完成的改造 |
| --- | --- |
| `web/src/app.css` | 暖色纸墨中性色 + 国际橙 `--signal`；Fira Mono 引入；`--radius` 收紧；滚动条直角化 |
| `web/src/lib/components/layout/HeroSection.svelte` | 完全重写为「规格书」风格：网格背景、四角标记、等宽标注、幽灵序号、阶梯入场、无渐变无粒子 |
| `web/src/lib/components/layout/Header.svelte` | Logo 去渐变改直角 mono 字母标；导航/按钮强调色统一为 signal |
| `web/src/lib/components/layout/Footer.svelte` | Logo 去渐变；顶部装饰条改实色 signal 发丝线；链接 hover 改 signal |

**基线校验结果：**

- `svelte-check`：全项目存在 **33 个历史遗留错误 + 7 个警告**（分布于 `chart-tooltip`、`utils/*`、`+error.svelte`、`(admin)/*`、`(app)/+page.svelte` 等），**与本次 UI 改造无关**。改造过程**不得新增**错误。
- 已改的 4 个文件通过 svelte-check 与 prettier 校验，无新增告警。

---

## 3. 重构范围清单（Inventory）

### 3.1 页面（`web/src/routes/`）

| 文件 | 现状问题 | 目标改造 |
| --- | --- | --- |
| `(auth)/login/+page.svelte` | 3 个 `animate-blob` 蓝紫粉渐变色块 + `from-slate-50 to-slate-100` 背景 | 移除 blob；背景改纯色或 hero 同款网格；表单卡片按 1.7 规范 |
| `(auth)/register/+page.svelte` | 同上 | 同上 |
| `+error.svelte` | slate 渐变背景、`glow-text animate-pulse text-blue-400`、emoji 🚀 | 改规格书风格：大号 mono 状态码 + signal 强调 + 网格/发丝线 |
| `(admin)/+layout.svelte` | `bg-blue-500` 状态点等冷蓝残留 | 强调色改 signal |
| `(admin)/manage/+page.svelte` | `bg-blue-500` 圆点 | 改 `bg-signal`，直角 |
| `(admin)/manage/users/+page.svelte` | 状态/角色 badge `rounded-full`、角色变体硬编码 | badge 直角化；变体统一走 token |
| `demo/+page.svelte` | `bg-blue-500` / `bg-green-500` 按钮 | 改 signal / muted |
| `(app)/+page.svelte` | 首屏整屏滚动手势含魔法数字（待重构时顺带清理） | 保持功能，抽常量、注释合规 |

### 3.2 布局与内容组件（`web/src/lib/components/layout/`）

| 文件 | 现状问题 | 目标改造 |
| --- | --- | --- |
| `ContentSection.svelte` | 9 个 emoji（☕📚🎵…）作图标；灰色渐变分割线；卡片悬停 `-translate-y-2 shadow-xl`；蓝色 hover | emoji 改「序号 + Lucide 图标」；渐变分割线改实色发丝线；悬停改 `hover:border-signal`；hover 强调改 signal |
| `Footer.svelte` | 残留 `Heart animate-pulse text-red-500` + `and lots of ☕` | 心跳脉动移除；emoji 改 Lucide `Coffee` 或纯文字 |

### 3.3 认证组件（`web/src/lib/components/auth/`）

| 文件 | 现状问题 | 目标改造 |
| --- | --- | --- |
| `LogoutButton.svelte` | spinner 圆形边框（功能性，保留） | 核对按钮样式走 token，无渐变 |
| `TokenMonitor.svelte` | 状态点 `bg-green-500` / `bg-red-500` | 颜色语义化（成功/危险用 token），保持功能 |

### 3.4 基础组件（`web/src/lib/components/ui/`）

- 绝大多数 shadcn 组件**无需改动**，会自动继承新的 `--radius`、`--border`、`--primary` 等 token。
- 保留的 `rounded-full`（功能性）：`avatar`、`switch`、`radio-group`、`progress`、`scroll-area`、`carousel` 翻页按钮、`drawer` 拖拽把手。
- `skeleton.svelte` 的 `animate-pulse` 保留（功能性加载骨架）。
- `button.svelte` 的 `transition-all` 建议改为具体属性（`transition-colors` + `transition-[transform]`），并补 `:active { scale(0.97) }`。

---

## 4. 有序改造计划（Phases）

> 严格按阶段顺序执行，每阶段完成后运行 `cd web && bun run check` 与 prettier format，确认不新增错误后再进入下一阶段。

### Phase 0 — 基线确认

- 拉取最新代码，确认 `app.css` / `HeroSection` / `Header` / `Footer` 已按第 2 节完成。
- 记录基线 `svelte-check` 输出（33 error / 7 warning），作为「不新增错误」的对照。
- 确认 `@fontsource/fira-mono` 已存在于依赖并可正常解析。

### Phase 1 — 全局 token 与基础组件微调

- 核对 `app.css` 的 `--signal`、`--color-signal`、`--font-mono`、`--radius` 是否生效。
- `button.svelte`：`transition-all` → 具体属性；补 `:active` 按压反馈。
- `skeleton.svelte`、`avatar`、`switch` 等确认保留功能性圆形。

### Phase 2 — 布局与内容组件

1. `ContentSection.svelte`（emoji、渐变分割线、卡片悬停、hover 强调色）。
2. `Footer.svelte`（Heart 脉动 + ☕）。
3. `+error.svelte`（规格书风格错误页）。
4. `theme-toggle.svelte` 若含冷蓝则统一。

### Phase 3 — 认证页与认证组件

1. `(auth)/login/+page.svelte`。
2. `(auth)/register/+page.svelte`。
3. `auth/LogoutButton.svelte`、`auth/TokenMonitor.svelte` 颜色语义化。

### Phase 4 — 管理后台

1. `(admin)/+layout.svelte`。
2. `(admin)/manage/+page.svelte`。
3. `(admin)/manage/users/+page.svelte`（badge 直角化 + 变体 token 化 + 冷蓝清理）。

### Phase 5 — 演示页与全站收尾

1. `demo/+page.svelte`。
2. 全站反模式 grep 清零（对照 1.8）。
3. 暗色/亮色对比度抽查（正文 4.5:1，大字 3:1）。
4. 响应式抽查（375 / 768 / 1024 / 1440，无横向滚动）。
5. `prefers-reduced-motion` 验证。
6. i18n 文案核对（中英文 key 齐全）。

---

## 5. 分文件改造要点（速查）

### 5.1 `ContentSection.svelte`

- `dailyShares` 数据：`emoji` 字段改为 `index`（01–09）或 Lucide 图标名；渲染处用 `font-mono text-signal` 序号或 `<Icon />`。
- 分割线：`bg-gradient-to-r from-transparent via-gray-300 to-transparent` → `h-px bg-border`。
- 卡片：`hover:-translate-y-2 hover:shadow-xl` → `hover:border-signal`（去位移与重阴影）。
- `group-hover:text-blue-600` → `group-hover:text-signal`。
- 分类 Badge 色：`bg-blue-100 text-blue-800` 等 → `bg-signal/10 text-signal`（统一或按分类映射到 token）。

### 5.2 `(auth)/login|register/+page.svelte`

- 删除 3 个 `animate-blob` 色块与 `from-slate-* to-slate-*` 容器类。
- 背景可复用 hero 的网格：抽一个可复用类（建议在 `app.css` 增加 `.spec-grid`），避免每页重复。
- 表单卡片：`border border-border rounded-none bg-card`，去掉大阴影。
- 主按钮改 `bg-signal text-signal-foreground hover:bg-signal/90`。

### 5.3 `+error.svelte`

- 移除 slate 渐变背景与 `glow-text animate-pulse text-blue-400`、emoji。
- 视觉：居中大号 `font-mono` 状态码 + 发丝线 + signal 强调字 + 网格背景。

### 5.4 `button.svelte`（基础组件）

- `transition-all` → `transition-colors` + `transition-[transform]`（或 `transition-[color,background-color,border-color,transform,box-shadow]`）。
- 增加 `:active { transform: scale(0.97); }` 与对应 transition。

---

## 6. 验收标准（Definition of Done）

每个阶段、以及最终交付均需满足：

- [ ] 1.8 反模式关键词全局 grep 返回 0（功能性 loading 除外）。
- [ ] 全站仅一种强调色 signal，中性色为暖纸墨，无冷蓝/紫/粉残留。
- [ ] 无 emoji 图标；图标均来自 Lucide，尺寸统一（`h-4 w-4` 起步）。
- [ ] 卡片/按钮直角或极小圆角；功能性圆形保留。
- [ ] 所有交互元素有 `cursor-pointer` 与可见 hover/focus 状态。
- [ ] 亮/暗两模式对比度达标、边框可见。
- [ ] `prefers-reduced-motion` 下位移与渐隐关闭。
- [ ] 375 / 768 / 1024 / 1440 响应式正常，无横向滚动。
- [ ] `cd web && bun run check` 不新增错误（基线 33 error / 7 warning 可保持不变）。
- [ ] 改动文件通过 prettier format 与 eslint。
- [ ] 中文注释为完整陈述句并以句号结尾，无括号补充、无口语词、无魔法数字。

---

## 7. 风险与注意事项

1. **中国大陆字体加载**：禁止 Google Fonts CDN；新增字体必须 `@fontsource` 自托管。正文 CJK 用系统栈，避免引入超大中文字体文件。
2. **SSR 挂起问题**：`(app)/+page.ts` 的 `load` 调用 `API.user.getUser()`，当后端未启动时会导致首页 SSR 长时间挂起（`data` 在页面中并未使用）。建议在重构时将该 load 改为**可降级**（try/catch 返回空数据）或直接移除，这是独立于视觉的健壮性修复。
3. **自定义 token 生效条件**：`bg-signal` / `text-signal` 依赖 `@theme inline` 中的 `--color-signal`；若改动了 token 名称需同步更新映射。
4. **圆角映射**：`--radius: 0.25rem` 使 `rounded-md` 锐利化，需全局回归验证表单、下拉、弹窗观感一致。
5. **历史错误边界**：33 个 svelte-check 错误为既有遗留，改造中不得新增；如需修复应单独立项，避免与视觉重构混在一起。
6. **最小改动**：仅重构视觉与动效，不改动业务逻辑、API 契约、路由结构与数据结构，除非文档明确要求（如 5.1 的 `emoji` 字段）。
7. **i18n**：新增/修改文案需同步 `messages/` 中的中英文 key，不要硬编码中文/英文字符串进组件（英雄屏等营销文案除外，需与产品确认）。

---

## 8. 给执行 Agent 的编码规范约束

1. 输出语言为简体中文，代码注释同样为简体中文。
2. 注释用**完整技术陈述句并以句号结尾**，禁止括号补充、口语化/情绪化用词。
3. **禁止魔法数字与魔法字符串**：颜色、时长、延迟、阈值等必须提取为命名常量或 token。
4. 函数单一职责，函数体 ≤ 30 行；嵌套 ≤ 3 层。
5. 最小改动原则，不影响无关功能。
6. 改动后运行 `cd web && bun run check` 与 `bunx prettier --write <改动文件>`，保持静态分析不新增告警。
7. 非异步场景调用异步函数用 `void` 显式忽略 Promise。
8. 涉及外部 I/O、网络、用户输入处必须有显式错误处理，禁止裸 `try { } catch { /* pass */ }`。

---

*本文档基于 2025 年重构基线（已完成 HeroSection / Header / Footer / app.css）编写，后续若设计系统有变更，需同步更新本文档并在变更处标注修订说明。*
