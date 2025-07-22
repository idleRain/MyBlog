<script lang="ts">
import { Button } from '$lib/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '$lib/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '$lib/components/ui/dialog'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger
} from '$lib/components/ui/sheet'
import ThemeToggle from '$lib/components/theme-toggle.svelte'
import { authStore } from '$lib/stores/auth'
import { Globe, User, ExternalLink, Menu, LogIn, Settings } from '@lucide/svelte'
import { goto } from '$app/navigation'
import { onMount } from 'svelte'

let isScrolled = $state(false)
let isMobileMenuOpen = $state(false)

// 订阅认证状态
let isAuthenticated = $state(false)
let currentUser = $state(null)

authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
  currentUser = state.user
})

// 登出功能
function handleLogout() {
  authStore.logout()
  goto('/')
}

// 监听滚动事件
onMount(() => {
  const scrollHandler = () => {
    isScrolled = window.scrollY > 20
  }

  window.addEventListener('scroll', scrollHandler)

  return () => {
    window.removeEventListener('scroll', scrollHandler)
  }
})
</script>

<header
  class="fixed top-0 right-0 left-0 z-50 transition-all duration-300 {isScrolled
    ? 'border-b border-gray-200/20 bg-white/80 shadow-lg backdrop-blur-md dark:border-gray-800/20 dark:bg-gray-950/80'
    : 'bg-transparent'}"
>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
    <div class="flex h-16 items-center justify-between">
      <!-- Logo左侧 -->
      <div class="flex items-center">
        <a href="/" class="group flex items-center space-x-3">
          <!-- 抽象几何Logo -->
          <div class="relative h-10 w-10">
            <div
              class="absolute inset-0 rotate-3 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 transition-transform duration-300 group-hover:rotate-6"
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
          <span class="hidden text-lg font-bold text-gray-900 sm:block dark:text-white">MyBlog</span
          >
        </a>
      </div>

      <!-- 中部导航 -->
      <nav class="hidden items-center space-x-8 md:flex">
        <a
          href="/"
          class="group relative text-gray-700 transition-colors duration-200 hover:text-blue-600 dark:text-gray-300 dark:hover:text-blue-400"
        >
          博客
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-blue-600 transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
        <a
          href="/projects"
          class="group relative text-gray-700 transition-colors duration-200 hover:text-blue-600 dark:text-gray-300 dark:hover:text-blue-400"
        >
          项目
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-blue-600 transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
        <a
          href="/about"
          class="group relative text-gray-700 transition-colors duration-200 hover:text-blue-600 dark:text-gray-300 dark:hover:text-blue-400"
        >
          关于
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-blue-600 transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
      </nav>

      <!-- 右侧功能区 -->
      <div class="flex items-center space-x-3">
        <!-- 桌面端功能按钮 -->
        <div class="hidden items-center space-x-3 md:flex">
          <!-- 语言切换 -->
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" class="h-9 w-9">
                <Globe class="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              class="bg-white/90 backdrop-blur-md dark:bg-gray-950/90"
            >
              <DropdownMenuItem>简体中文</DropdownMenuItem>
              <DropdownMenuItem>English</DropdownMenuItem>
              <DropdownMenuItem>日本語</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <!-- 主题切换 -->
          <ThemeToggle />

          <!-- 个人介绍 -->
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" class="h-9 w-9">
                <User class="h-4 w-4" />
              </Button>
            </DialogTrigger>
            <DialogContent class="bg-white/95 backdrop-blur-md sm:max-w-md dark:bg-gray-950/95">
              <DialogHeader>
                <DialogTitle>关于作者</DialogTitle>
                <DialogDescription>
                  一个热爱编程和设计的创造者，用代码编织创意，用技术改变世界。
                </DialogDescription>
              </DialogHeader>
              <div class="flex flex-col space-y-3">
                <div class="flex items-center space-x-3">
                  <div
                    class="flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600"
                  >
                    <span class="text-lg font-bold text-white">M</span>
                  </div>
                  <div>
                    <h3 class="font-semibold text-gray-900 dark:text-white">开发者</h3>
                    <p class="text-sm text-gray-600 dark:text-gray-400">全栈开发工程师</p>
                  </div>
                </div>
              </div>
            </DialogContent>
          </Dialog>

          <!-- 友情链接 -->
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" class="h-9 w-9">
                <ExternalLink class="h-4 w-4" />
              </Button>
            </DialogTrigger>
            <DialogContent class="bg-white/95 backdrop-blur-md sm:max-w-md dark:bg-gray-950/95">
              <DialogHeader>
                <DialogTitle>友情链接</DialogTitle>
              </DialogHeader>
              <div class="space-y-3">
                <a
                  href="https://github.com"
                  target="_blank"
                  class="flex items-center justify-between rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-700"
                >
                  <span class="font-medium">GitHub</span>
                  <ExternalLink class="h-4 w-4 text-gray-500" />
                </a>
                <a
                  href="https://svelte.dev"
                  target="_blank"
                  class="flex items-center justify-between rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-700"
                >
                  <span class="font-medium">SvelteKit</span>
                  <ExternalLink class="h-4 w-4 text-gray-500" />
                </a>
                <a
                  href="https://tailwindcss.com"
                  target="_blank"
                  class="flex items-center justify-between rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-700"
                >
                  <span class="font-medium">TailwindCSS</span>
                  <ExternalLink class="h-4 w-4 text-gray-500" />
                </a>
              </div>
            </DialogContent>
          </Dialog>
        </div>

        <!-- 登录/登出和后台按钮 -->
        <div
          class="ml-3 hidden items-center space-x-3 border-l border-gray-200 pl-3 md:flex dark:border-gray-700"
        >
          {#if isAuthenticated}
            <!-- 后台管理按钮 -->
            <Button
              variant="outline"
              size="sm"
              onclick={() => goto('/admin')}
              class="border-blue-200 text-blue-600 hover:bg-blue-50 dark:border-blue-700 dark:text-blue-400 dark:hover:bg-blue-950"
            >
              <Settings class="mr-1 h-4 w-4" />
              后台管理
            </Button>

            <!-- 用户信息下拉菜单 -->
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" class="h-9 w-9">
                  <div
                    class="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600"
                  >
                    <span class="text-xs font-bold text-white">
                      {currentUser?.username?.charAt(0)?.toUpperCase() || 'U'}
                    </span>
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="end"
                class="w-48 bg-white/90 backdrop-blur-md dark:bg-gray-950/90"
              >
                <div class="border-b border-gray-200 px-3 py-2 dark:border-gray-700">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {currentUser?.username || '用户'}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {currentUser?.email || ''}
                  </p>
                </div>
                <DropdownMenuItem
                  onclick={handleLogout}
                  class="text-red-600 hover:text-red-700 dark:text-red-400"
                >
                  登出
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          {:else}
            <!-- 登录按钮 -->
            <Button
              size="sm"
              onclick={() => goto('/login')}
              class="bg-gradient-to-r from-blue-600 to-purple-600 text-white hover:from-blue-700 hover:to-purple-700"
            >
              <LogIn class="mr-1 h-4 w-4" />
              登录
            </Button>
          {/if}
        </div>

        <!-- 移动端菜单 -->
        <Sheet bind:open={isMobileMenuOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" class="h-9 w-9 md:hidden">
              <Menu class="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent
            side="right"
            class="w-80 bg-white/95 p-6 backdrop-blur-md dark:bg-gray-950/95"
          >
            <SheetHeader class="mb-6 text-left">
              <SheetTitle class="text-xl font-bold">菜单</SheetTitle>
            </SheetHeader>

            <!-- 移动端导航链接 -->
            <nav class="space-y-1">
              <a
                href="/"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-gray-900 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:text-white dark:hover:bg-blue-950/50 dark:hover:text-blue-400"
                onclick={() => (isMobileMenuOpen = false)}
              >
                博客
              </a>
              <a
                href="/projects"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-gray-900 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:text-white dark:hover:bg-blue-950/50 dark:hover:text-blue-400"
                onclick={() => (isMobileMenuOpen = false)}
              >
                项目
              </a>
              <a
                href="/about"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-gray-900 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:text-white dark:hover:bg-blue-950/50 dark:hover:text-blue-400"
                onclick={() => (isMobileMenuOpen = false)}
              >
                关于
              </a>
            </nav>

            <!-- 移动端功能按钮 -->
            <div class="mt-8 space-y-6">
              <div
                class="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-2 dark:bg-gray-800/50"
              >
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">主题切换</span>
                <ThemeToggle />
              </div>

              <div class="space-y-3">
                <h4 class="px-4 text-sm font-semibold text-gray-900 dark:text-white">语言选择</h4>
                <div class="space-y-1">
                  <button
                    class="flex w-full items-center rounded-lg px-4 py-2 text-left text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800/50"
                  >
                    简体中文
                  </button>
                  <button
                    class="flex w-full items-center rounded-lg px-4 py-2 text-left text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800/50"
                  >
                    English
                  </button>
                  <button
                    class="flex w-full items-center rounded-lg px-4 py-2 text-left text-sm transition-colors hover:bg-gray-100 dark:hover:bg-gray-800/50"
                  >
                    日本語
                  </button>
                </div>
              </div>
            </div>

            <!-- 移动端登录/后台区域 -->
            <div class="mt-8 border-t border-gray-200/50 pt-6 dark:border-gray-700/50">
              {#if isAuthenticated}
                <div class="space-y-3">
                  <div
                    class="rounded-lg bg-gradient-to-r from-blue-50 to-purple-50 px-4 py-3 dark:from-blue-950/30 dark:to-purple-950/30"
                  >
                    <p class="text-sm font-semibold text-gray-900 dark:text-white">
                      {currentUser?.username || '用户'}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                      {currentUser?.email || ''}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start border-blue-200 text-blue-600 hover:bg-blue-50 dark:border-blue-700 dark:text-blue-400 dark:hover:bg-blue-950/50"
                    onclick={() => {
                      isMobileMenuOpen = false
                      goto('/admin')
                    }}
                  >
                    <Settings class="mr-2 h-4 w-4" />
                    后台管理
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start border-red-200 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-950/50"
                    onclick={() => {
                      isMobileMenuOpen = false
                      handleLogout()
                    }}
                  >
                    登出
                  </Button>
                </div>
              {:else}
                <Button
                  size="sm"
                  class="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg hover:from-blue-700 hover:to-purple-700"
                  onclick={() => {
                    isMobileMenuOpen = false
                    goto('/login')
                  }}
                >
                  <LogIn class="mr-2 h-4 w-4" />
                  登录
                </Button>
              {/if}
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </div>
  </div>
</header>

<style>
/* 确保导航栏渐变玻璃效果 */
header {
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}
</style>
