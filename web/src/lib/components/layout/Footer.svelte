<script lang="ts">
import { Button } from '$lib/components/ui/button'
import { Separator } from '$lib/components/ui/separator'
import { Github, Mail, Twitter, Linkedin, Heart } from '@lucide/svelte'

const currentYear = new Date().getFullYear()

const socialLinks = [
  {
    name: 'GitHub',
    icon: Github,
    href: 'https://github.com',
    color: 'hover:text-gray-900 dark:hover:text-white'
  },
  { name: 'Twitter', icon: Twitter, href: 'https://twitter.com', color: 'hover:text-blue-400' },
  { name: 'LinkedIn', icon: Linkedin, href: 'https://linkedin.com', color: 'hover:text-blue-600' },
  { name: 'Email', icon: Mail, href: 'mailto:contact@myblog.com', color: 'hover:text-red-500' }
]

const quickLinks = [
  { name: '首页', href: '/' },
  { name: '博客', href: '/blog' },
  { name: '项目', href: '/projects' },
  { name: '关于', href: '/about' },
  { name: '联系', href: '/contact' }
]

const categories = [
  { name: '前端开发', href: '/category/frontend' },
  { name: '后端开发', href: '/category/backend' },
  { name: '设计思考', href: '/category/design' },
  { name: '技术分享', href: '/category/tech' },
  { name: '生活随笔', href: '/category/life' }
]
</script>

<footer
  class="relative border-t border-gray-200/50 bg-gray-50/80 backdrop-blur-sm dark:border-gray-800/50 dark:bg-gray-900/80"
>
  <!-- 主要内容区域 -->
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4">
      <!-- 品牌介绍 -->
      <div class="lg:col-span-2">
        <div class="mb-4 flex items-center space-x-3">
          <div class="relative h-10 w-10">
            <div
              class="absolute inset-0 rotate-3 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600"
            ></div>
            <div
              class="absolute inset-0.5 flex items-center justify-center rounded-lg bg-white dark:bg-gray-950"
            >
              <span
                class="bg-gradient-to-br from-blue-500 to-purple-600 bg-clip-text text-lg font-bold text-transparent"
                >M</span
              >
            </div>
          </div>
          <span class="text-xl font-bold text-gray-900 dark:text-white">MyBlog</span>
        </div>

        <p class="mb-6 max-w-md text-gray-600 dark:text-gray-300">
          一个专注于技术分享和创意设计的个人博客，探索现代Web开发的无限可能。
          用代码编织创意，用技术改变世界。
        </p>

        <!-- 社交媒体链接 -->
        <div class="flex space-x-4">
          {#each socialLinks as link}
            <a
              href={link.href}
              target="_blank"
              rel="noopener noreferrer"
              class="group rounded-lg bg-white/50 p-2 transition-all duration-200 hover:bg-white dark:bg-gray-800/50 dark:hover:bg-gray-800 {link.color}"
              aria-label={link.name}
            >
              <svelte:component this={link.icon} class="h-5 w-5" />
            </a>
          {/each}
        </div>
      </div>

      <!-- 快速链接 - 左右两排布局 -->
      <div>
        <h3 class="mb-4 font-semibold text-gray-900 dark:text-white">快速链接</h3>
        <div class="grid grid-cols-2 gap-x-4 gap-y-2">
          {#each quickLinks as link}
            <a
              href={link.href}
              class="text-sm text-gray-600 transition-colors duration-200 hover:text-blue-600 dark:text-gray-300 dark:hover:text-blue-400"
            >
              {link.name}
            </a>
          {/each}
        </div>
      </div>

      <!-- 分类目录 - 左右两排布局 -->
      <div>
        <h3 class="mb-4 font-semibold text-gray-900 dark:text-white">文章分类</h3>
        <div class="grid grid-cols-2 gap-x-4 gap-y-2">
          {#each categories as category}
            <a
              href={category.href}
              class="text-sm text-gray-600 transition-colors duration-200 hover:text-blue-600 dark:text-gray-300 dark:hover:text-blue-400"
            >
              {category.name}
            </a>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <Separator class="bg-gray-200/50 dark:bg-gray-800/50" />

  <!-- 版权信息 -->
  <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
    <div class="flex flex-col items-center justify-between gap-4 md:flex-row">
      <!-- 版权文字 -->
      <div class="text-center text-sm text-gray-500 md:text-left dark:text-gray-400">
        <p class="flex items-center justify-center gap-1 md:justify-start">
          © {currentYear} MyBlog. Made with
          <Heart class="h-4 w-4 animate-pulse text-red-500" />
          and lots of ☕
        </p>
      </div>

      <!-- 技术栈信息 -->
      <div class="flex flex-wrap items-center gap-4 text-xs text-gray-400 dark:text-gray-500">
        <span class="flex items-center gap-1">
          Powered by
          <a
            href="https://kit.svelte.dev"
            target="_blank"
            class="transition-colors hover:text-blue-600">SvelteKit</a
          >
        </span>
        <span class="hidden sm:block">•</span>
        <span class="flex items-center gap-1">
          Styled with
          <a
            href="https://tailwindcss.com"
            target="_blank"
            class="transition-colors hover:text-blue-600">TailwindCSS</a
          >
        </span>
        <span class="hidden sm:block">•</span>
        <span class="flex items-center gap-1">
          Hosted on
          <a href="https://vercel.com" target="_blank" class="transition-colors hover:text-blue-600"
            >Vercel</a
          >
        </span>
      </div>
    </div>
  </div>

  <!-- 装饰性背景元素 -->
  <div
    class="absolute top-0 left-0 h-1 w-full bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 opacity-50"
  ></div>
</footer>

<style>
/* 确保链接悬停效果流畅 */
a {
  transition: all 0.2s ease-in-out;
}

/* 优化小屏幕显示 */
@media (max-width: 768px) {
  footer {
    padding-bottom: env(safe-area-inset-bottom);
  }
}
</style>
