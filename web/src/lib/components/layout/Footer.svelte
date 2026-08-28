<script lang="ts">
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import { Separator } from '$lib/components/ui/separator'
import { Mail } from '@lucide/svelte'

// 当前年份用于版权声明，随时间自动更新。
const currentYear = new Date().getFullYear()

// 社交链接集中声明，悬停统一反馈为 signal 强调色。
const socialLinks = [
  { name: 'GitHub', icon: GithubIcon, href: 'https://github.com/idleRain' },
  { name: 'Email', icon: Mail, href: 'gold.experience@foxmail.com' }
]

// 快捷导航链接集中声明。
const quickLinks = [
  { name: '首页', href: '/' },
  { name: '博客', href: '/blog' },
  { name: '项目', href: '/projects' },
  { name: '关于', href: '/about' },
  { name: '联系', href: '/contact' }
]

// 文章分类链接集中声明。
const categories = [
  { name: '前端开发', href: '/category/frontend' },
  { name: '后端开发', href: '/category/backend' },
  { name: '设计思考', href: '/category/design' },
  { name: '技术分享', href: '/category/tech' },
  { name: '生活随笔', href: '/category/life' }
]
</script>

<footer class="relative border-t border-border bg-muted/40">
  <!-- 主要内容区域 -->
  <div class="mx-auto max-w-7xl px-6 py-12 sm:px-10 lg:px-20">
    <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4">
      <!-- 品牌介绍 -->
      <div class="lg:col-span-2">
        <div class="mb-4 flex items-center space-x-3">
          <div class="relative h-10 w-10">
            <div class="absolute inset-0 border border-border"></div>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="font-mono text-lg font-bold text-signal">M</span>
            </div>
          </div>
          <span class="text-xl font-bold text-foreground">MyBlog</span>
        </div>

        <p class="mb-6 max-w-md text-muted-foreground">
          一个专注于技术分享和创意设计的个人博客，探索现代Web开发的无限可能。
          用代码编织创意，用技术改变世界。
        </p>

        <!-- 社交链接采用直角描边方块，悬停以 signal 色反馈。 -->
        <div class="flex space-x-4">
          {#each socialLinks as link (link.name)}
            {@const IconComponent = link.icon}
            <a
              href={link.href}
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-none border border-border p-2 text-muted-foreground transition-colors duration-200 hover:border-signal hover:text-signal"
              aria-label={link.name}
            >
              <IconComponent class="h-5 w-5" />
            </a>
          {/each}
        </div>
      </div>

      <!-- 快速链接 - 左右两排布局 -->
      <div>
        <h3 class="mb-4 font-semibold text-foreground">快速链接</h3>
        <div class="grid grid-cols-2 gap-x-4 gap-y-2">
          {#each quickLinks as link (link.name)}
            <a
              href={link.href}
              class="text-sm text-muted-foreground transition-colors duration-200 hover:text-signal"
            >
              {link.name}
            </a>
          {/each}
        </div>
      </div>

      <!-- 分类目录 - 左右两排布局 -->
      <div>
        <h3 class="mb-4 font-semibold text-foreground">文章分类</h3>
        <div class="grid grid-cols-2 gap-x-4 gap-y-2">
          {#each categories as category (category.name)}
            <a
              href={category.href}
              class="text-sm text-muted-foreground transition-colors duration-200 hover:text-signal"
            >
              {category.name}
            </a>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <Separator />

  <!-- 版权信息 -->
  <div class="mx-auto max-w-7xl px-6 py-6 sm:px-10 lg:px-20">
    <div class="flex flex-col items-center justify-between gap-4 md:flex-row">
      <!-- 版权行以等宽注释风格呈现，延续规格书的标注语言。 -->
      <p
        class="flex flex-wrap items-center justify-center gap-2 font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase md:justify-start"
      >
        <span>© {currentYear} MyBlog</span>
        <span class="text-signal">//</span>
        <span>crafted with precision</span>
      </p>

      <!-- 技术栈信息属于元信息，统一使用等宽字体。 -->
      <div class="flex flex-wrap items-center gap-4 font-mono text-xs text-muted-foreground">
        <span class="flex items-center gap-1">
          Powered by
          <a
            href="https://kit.svelte.dev"
            target="_blank"
            class="transition-colors hover:text-signal"
          >
            SvelteKit</a
          >
        </span>
        <span class="hidden sm:block">•</span>
        <span class="flex items-center gap-1">
          Styled with
          <a
            href="https://tailwindcss.com"
            target="_blank"
            class="transition-colors hover:text-signal"
          >
            TailwindCSS
          </a>
        </span>
        <span class="hidden sm:block">•</span>
        <span class="flex items-center gap-1">
          Service on
          <a
            href="https://golang.google.cn"
            target="_blank"
            class="transition-colors hover:text-signal"
          >
            Go
          </a>
        </span>
      </div>
    </div>
  </div>

  <!-- 顶部装饰条为实色 signal 发丝线，作为页脚的规格标记。 -->
  <div class="absolute top-0 left-0 h-0.5 w-full bg-signal" aria-hidden="true"></div>
</footer>

<style>
/* 小屏设备底部安全区内边距，避免内容被系统手势条遮挡。 */
@media (max-width: 768px) {
  footer {
    padding-bottom: env(safe-area-inset-bottom);
  }
}
</style>
