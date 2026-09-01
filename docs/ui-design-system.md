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
| `apps/web`（前台 toC） | **规格书主题** | 工程规格书 / 技术蓝图 | 锐利直角 `0.25rem` | 国际橙 `signal` |
| `apps/admin`（后台 toB） | **shadcn 原始主题** | shadcn-svelte 默认 | 常规圆角 `0.625rem` | slate 蓝 `primary` |

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

**取色硬约束**：所有颜色必须通过语义 token（`bg-signal`、`text-muted-foreground`、`border-border` 等）取用，**禁止硬编码色值**。色值一律以 `oklch()` 定义。

### 2.1 前台规格书主题（`apps/web/src/app.css`）

**基调**：暖色纸墨 + 国际橙强调色。中性色统一采用极低的黄色相（hue 95），避免常见冷灰偏蓝。

#### 亮色模式

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(0.982 0.004 95)` | 暖纸底色 |
| `--foreground` | `oklch(0.18 0.012 85)` | 主文字（墨色） |
| `--card` | `oklch(0.992 0.002 95)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.945 0.006 95)` | 弱背景 |
| `--muted-foreground` | `oklch(0.5 0.016 85)` | 次级文字 |
| `--accent` | `oklch(0.935 0.007 95)` | 悬停/高亮背景 |
| `--primary` | `oklch(0.18 0.012 85)` | 主按钮（墨色） |
| `--destructive` | `oklch(0.577 0.245 27.325)` | 危险操作 |
| `--border` / `--input` | `oklch(0.882 0.007 95)` | 发丝线 / 边框 |
| `--ring` | `oklch(0.63 0.23 30)` | 焦点环（橙） |
| `--signal` | `oklch(0.63 0.23 30)` | **强调色（国际橙）** |
| `--signal-foreground` | `oklch(0.17 0.02 30)` | 橙底上的深墨文字 |

#### 暗色模式

| Token | 值 | 用途 |
| --- | --- | --- |
| `--background` | `oklch(0.16 0.007 95)` | 墨底 |
| `--foreground` | `oklch(0.93 0.005 95)` | 主文字 |
| `--card` | `oklch(0.19 0.008 95)` | 卡片面 |
| `--muted` / `--secondary` | `oklch(0.24 0.009 95)` | 弱背景 |
| `--muted-foreground` | `oklch(0.68 0.012 95)` | 次级文字 |
| `--primary` | `oklch(0.93 0.005 95)` | 主按钮（纸白） |
| `--border` / `--input` | `oklch(0.27 0.009 95)` | 发丝线 |
| `--ring` / `--signal` | `oklch(0.68 0.21 32)` | 焦点环 / 强调色（亮橙） |
| `--signal-foreground` | `oklch(0.15 0.02 32)` | 橙底上的深墨文字 |

图表与侧边栏 token（`--chart-1..5`、`--sidebar-*`）随主题在 `app.css` 中定义，业务代码只经 `@theme inline` 映射后的工具类取用。

### 2.2 后台 shadcn 原始主题（`apps/admin/src/app.css`）

**基调**：shadcn-svelte 默认 slate 蓝主题，纯白底，**不含 `--signal`**。后台不使用规格书强调色。

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
- 后台不得使用 `signal`（该 token 未定义）；状态语义色走 `primary` / `muted` / `destructive`。
- 语义色：成功/危险等状态用 `--destructive` 与 `--muted-foreground`，**不**引入额外强调色。

---

## 3. 字体系统

### 3.1 前台（`apps/web`）

- **等宽（技术标注专用）**：Fira Mono（400/500/700），经 `@fontsource/fira-mono` 自托管，`font-mono` 工具类即可用。
- **正文与标题**：系统 CJK 栈 `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif`。标题依靠字号、字重与字距营造层级，不另设标题字体。

### 3.2 后台（`apps/admin`）

- 正文声明 `'Manrope'`、标题声明 `'Inter'`，但**两者尚未经 `@fontsource` 自托管**，当前实际回退系统字体栈。如需正式启用，必须自托管（见 3.3）；否则应从 `font-family` 移除以免误导。

### 3.3 自托管约束

- **禁止**新增 Google Fonts CDN 引用（大陆网络不可达）。
- 新增字体必须走 `@fontsource` 自托管并加入依赖，使用前在 `app.css` 的 `@theme` 中完成映射。

---

## 4. 圆角与间距

### 4.1 前台（规格书，锐利直角）

- 全局 `--radius: 0.25rem`，使 `rounded-md` 锐利化。
- 卡片 / 按钮 / 输入框：`rounded-none` 或继承 `rounded-md`，**禁止**装饰性 `rounded-full` 与 `rounded-xl`。
- 保留的功能性圆形：头像 `avatar`、开关 `switch`、单选 `radio-group`、进度 `progress`、滚动区 `scroll-area`、轮播翻页按钮、抽屉拖拽把手。

### 4.2 后台（shadcn 默认）

- 全局 `--radius: 0.625rem`，沿用 shadcn-svelte 默认圆角体系，不额外收紧。

### 4.3 间距与容器

- 前台容器宽度统一 `max-w-6xl` 或 `max-w-7xl`。
- 前台页边距统一 `px-6 sm:px-10 lg:px-16/20`。
- 后台沿用 shadcn 常规间距体系。

---

## 5. 排版层级

### 5.1 前台（规格书排版）

| 层级 | 规格 | 用途 |
| --- | --- | --- |
| Display | `text-5xl sm:text-7xl lg:text-8xl font-bold tracking-tight text-balance` | 页面主标题 |
| Section 标题 | `text-3xl sm:text-4xl font-bold` + 发丝线 | 板块标题 |
| Body | `text-base sm:text-lg leading-relaxed text-muted-foreground` | 正文 |
| Meta | `font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground` | 标注、序号、元信息、按钮文字、技术名词 |

规格书特征元素：坐标网格背景（`.spec-grid`，单元 56px）、页边出血线、四角裁切标记、大号幽灵序号（`01–09` 等宽 + `text-signal`）、等宽「代码注释式」标注（`// ...`）、发丝线分割。

### 5.2 后台（shadcn 默认）

沿用 shadcn-svelte 标准排版与组件层级，无规格书式大标题与幽灵序号要求。

---

## 6. 动效规范

动效遵循 Emil 设计工程规范，只做「必要且有目的」的动效。

### 6.1 前台动效

- **缓动**：进入用 `cubic-bezier(0.23, 1, 0.32, 1)`（ease-out）；屏内位移用 `cubic-bezier(0.77, 0, 0.175, 1)`（ease-in-out）。
- **时长**：UI 微交互 150–300ms；按钮按压反馈 100–160ms；首屏营销性入场可放宽。
- **入场**：首屏采用阶梯式浮现（30–80ms 间隔），`translateY(16px) + opacity`，不要从 `scale(0)` 出现。
- **按钮按压**：`:active` 时 `scale(0.97)`。
- **过渡属性**：只过渡 `transform` / `opacity` / `color` / `border-color` / `background-color`，**禁止 `transition-all`**。
- **性能**：只动画 `transform` 与 `opacity`，不动画 `width` / `height` / `margin` / `padding`。

### 6.2 后台动效

- 沿用 shadcn 组件自带过渡。
- 后台登录页保留 blob 装饰动画（`animate-blob`，蓝紫粉模糊色块），属历史保留的登录页特色，不属于前台规格书反模式约束范围；后续如需清理应单独立项。

### 6.3 减少动态

所有位移动画必须在 `@media (prefers-reduced-motion: reduce)` 下关闭位移与渐隐。

---

## 7. 组件风格约定

### 7.1 前台（规格书风格）

| 组件 | 规范 |
| --- | --- |
| 主按钮 | `rounded-none bg-signal text-signal-foreground hover:bg-signal/90`，可带 `font-mono` |
| 次按钮 | `rounded-none border border-border hover:border-signal hover:text-signal` |
| 卡片 | `rounded-none border border-border bg-card`，悬停用 `hover:border-signal` 替代位移与重阴影 |
| Badge | `rounded-none`，用 `bg-signal/10 text-signal` 或 `bg-muted text-muted-foreground` |
| 分割线 | `border-t border-border` 或 `h-px bg-border`，禁止渐变分割线 |
| Logo 字母标 | 直角描边方块 + `font-mono text-signal` 字母 |
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

- `bg-gradient-*` / `from-*` / `via-*` / `to-*`（装饰性 CSS 渐变）
- `bg-clip-text text-transparent`（渐变文字）
- `animate-pulse` / `animate-bounce` / `animate-blob`（装饰性动画；骨架屏 `skeleton` 的 pulse 除外）
- emoji 作为 UI 图标（`☕ 📚 🎵 🌱 🎨 🔧 💡 🏃 🎯 🚀 ✨` 等）
- `particles` / canvas 粒子背景
- `rounded-full`（卡片/按钮/徽章；功能性圆形除外）
- `transition-all`
- `glow-*` / `text-blue-*` / 冷蓝、紫、粉残留
- 硬编码色值（应走 token）

### 8.2 后台（相对宽松）

- 后台遵循 shadcn 原始主题，不强制规格书反模式。
- 但**禁止**引入规格书之外的第三种强调色、禁止渐变文字与粒子背景，保持与前台一致的基本克制。

---

## 9. 现状记录与待办

| 项 | 现状 | 建议 |
| --- | --- | --- |
| 前台 `app.css` 的 `.spec-grid` | 登录页归后台后前台无组件引用，为死代码 | 前台如需网格底纹可直接使用；否则待清理 |
| 前台 `app.css` 的 `.animate-blob` / 延迟类 | 前台无组件引用，为死代码 | 随清理项一并移除 |
| 后台 Manrope / Inter 字体 | 已声明于 `font-family` 但未自托管，实际回退系统栈 | 自托管或移除声明 |
| 后台登录页 blob 动画 | 历史保留的登录页特色（蓝紫粉模糊色块） | 保持；如需清理单独立项 |

---

## 10. 开发与校验

1. 取色必须走 token，禁止魔法色值；动效时长、延迟等必须提取为命名常量。
2. 前台规格书元素（幽灵序号、发丝线、等宽标注）须与既有组件（HeroSection / Header / Footer / ContentSection）保持一致的实现方式。
3. 改动后运行对应校验：
   - 前台：`pnpm run check`（或 `cd apps/web && pnpm run check`）
   - 后台：`cd apps/admin && pnpm run check`
   - 全量：`pnpm run quality`
4. 反模式自查：按 8.1 在 `apps/web/src` 下 grep 验证。
5. 注释与对话使用简体中文；注释为完整技术陈述句并以句号结尾，无括号补充、无口语词、无魔法数字。
