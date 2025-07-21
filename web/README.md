# MyBlog Frontend

基于 SvelteKit 的现代化博客前端应用，采用现代极简主义设计风格。

## 技术栈

- **框架**: SvelteKit + TypeScript
- **样式**: TailwindCSS + 现代极简主义风格
- **组件库**: shadcn-svelte
- **图标**: lucide-svelte
- **动画**: GSAP
- **构建工具**: Vite
- **包管理**: Bun

## 项目结构

```
web/
├── src/
│   ├── routes/          # 页面路由和布局
│   ├── lib/             # 可复用组件
│   ├── service/         # API 调用和数据管理
│   ├── utils/           # 工具函数
│   └── app.html         # HTML 模板
├── static/              # 静态资源
├── tests/               # 测试文件
└── package.json         # 依赖配置
```

## 开发指南

### 启动开发服务器

```bash
# 从根目录启动（推荐）
bun run dev:web

# 或在 web 目录下启动
cd web && bun run dev
```

### 代码规范

#### 样式规范
- UI 样式风格使用 **现代极简主义风格**
- 使用 TailwindCSS 框架，代码中尽量使用 TailwindCSS 的类名
- UI 组件库使用 shadcn-svelte，优先使用组件库提供的类名
- 常用的 CSS 类名或变量统一配置到 `tailwind.config.js` 中，方便后续维护

#### 组件开发
- 图标使用 lucide-svelte 图标库
- 动画效果使用 GSAP 库
- 组件命名使用 PascalCase
- 文件命名使用 kebab-case

#### TypeScript 规范
- 严格的类型检查，避免使用 `any`
- 接口定义放在 `src/types/` 目录下
- 使用类型导入：`import type { ... } from '...'`

#### 自动导入规范
项目已配置 `unplugin-auto-import`，以下内容会自动导入，无需手动 import：

**核心模块（自动导入）**：
- **SvelteKit**: `goto`, `page`, `navigating`, `browser`, `dev`
- **Svelte**: `onMount`, `onDestroy`, `writable`, `readable`, `derived`, `get`
- **UI 组件**: `Button`, `Sonner`, `Toaster`
- **API**: `userService` (用户相关 API)
- **状态管理**: `authStore` (认证状态)
- **工具**: `toast` (消息提示)
- **图标**: `Mail`, `Lock`, `Eye`, `EyeOff`, `LogIn`, `UserPlus`, `Calendar` 等常用图标

**手动导入的组件**：
- shadcn 复合组件：`Card`, `Input`, `Label` 等需要手动导入
- `User` 图标需要手动导入 (避免与用户类型冲突)
- `mode` store 需要手动导入 (Svelte store 规范)
```svelte
import * as Card from '$lib/components/ui/card'
import { Input } from '$lib/components/ui/input'
import { Label } from '$lib/components/ui/label'
import { User } from 'lucide-svelte'
import { mode } from 'mode-watcher'
```

**类型定义文件**：
- 自动生成：`src/auto-imports.d.ts`（提供自动导入的类型支持）

### 构建和测试

```bash
# 类型检查
bun run check

# 类型检查（监听模式）
bun run check:watch

# 构建项目
bun run build

# 运行测试
bun run test

# 预览构建结果
bun run preview
```

### API 调用规范

- 所有 API 调用统一放在 `src/service/` 目录下
- 使用统一的错误处理机制
- API 接口使用 POST 方法，参数使用 camelCase 命名
- 响应数据格式：`{ code, message, data }`

### 环境配置

环境变量配置在 `.env` 文件中：

```env
VITE_API_URL=http://localhost:3000
VITE_SERVER_PORT=5173
```

## 部署

```bash
# 构建生产版本
bun run build

# 生产环境预览
bun run preview
```

构建产物位于 `build/` 目录，可直接部署到任何支持 Node.js 的服务器。
