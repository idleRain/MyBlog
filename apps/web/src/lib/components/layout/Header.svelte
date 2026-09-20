<script lang="ts">
import { Globe, User, ExternalLink, Menu, LogIn, Settings, Ellipsis } from '@lucide/svelte'
import FriendlyLinkDialog from '$lib/components/layout/FriendlyLinkDialog.svelte'
import NotificationBell from '$lib/components/layout/NotificationBell.svelte'
import type { FriendlyLink } from '@myblog/api/modules/friendlyLink/types'
import type { User as UserType } from '@myblog/api/modules/user/types'
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import ThemeToggle from '$lib/components/theme-toggle.svelte'
import { setLocale, getLocale } from '$lib/paraglide/runtime'
import { Button, DropdownMenu, Dialog, Sheet } from '$ui'
import { authStore } from '$lib/stores/auth'
import { goto } from '$app/navigation'
import { m } from '$i18n'

// 展示中的友情链接，由根布局的全站数据提供；错误页等无数据场景降级为空列表。
interface Props {
  friendlyLinks?: FriendlyLink[]
}

let { friendlyLinks = [] }: Props = $props()

let isMobileMenuOpen = $state(false)
// 更多菜单收纳的低频入口弹窗状态，触发权在菜单项上。
let isAuthorDialogOpen = $state(false)
let isFriendlyLinkDialogOpen = $state(false)

// 订阅认证状态
let isAuthenticated = $state(false)
let currentUser = $state<UserType>()

authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
  currentUser = state.user!
})

// 登出功能
function handleLogout() {
  authStore.logout()
  goto('/')
}

// 切换语言：写入语言 cookie 后整页刷新，确保 SSR 数据与界面文案同步切换。
const setLanguage = (lang: 'zh' | 'en') => {
  if (getLocale() === lang) return
  setLocale(lang)
  location.reload()
}

// 后台管理地址，开发环境为独立端口，生产环境为同源子路径。
const adminUrl = import.meta.env.VITE_ADMIN_URL || '/admin'

// GitHub 主页地址，桌面更多菜单与移动抽屉共用同一外链来源。
const GITHUB_URL = 'https://github.com/idleRain'

// 外部链接统一经用户手势打开新标签页，并禁用 opener 反向引用。
function openExternal(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}

// 打开后台管理控制台
function openAdminConsole() {
  openExternal(adminUrl)
}
</script>

<header
  class="fixed top-0 right-0 left-0 z-50 border-b border-line bg-background/85 backdrop-blur-md"
>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
    <div class="flex h-16 items-center">
      <!-- Logo左侧 -->
      <div class="flex items-center">
        <a href="/" class="group flex items-center space-x-3">
          <!-- 抽象几何Logo -->
          <div class="relative h-10 w-10">
            <div
              class="absolute inset-0 border border-border transition-colors duration-200 group-hover:border-signal"
            ></div>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="font-mono text-lg font-bold text-signal">M</span>
            </div>
          </div>
          <span class="hidden font-display text-lg font-black text-foreground sm:block">
            {m['ui:site.name']()}
          </span>
        </a>
      </div>

      <!-- 中部导航 - 绝对居中；lg 以下隐藏以避免与右侧功能区在中屏重叠，导航入口由抽屉承接 -->
      <nav
        class="absolute top-1/2 left-1/2 hidden -translate-x-1/2 -translate-y-1/2 transform items-center space-x-8 lg:flex"
      >
        <a
          href="/"
          class="group relative text-muted-foreground transition-colors duration-200 hover:text-signal"
        >
          {m['ui:header.nav.blog']()}
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-signal transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
        <a
          href="/projects"
          class="group relative text-muted-foreground transition-colors duration-200 hover:text-signal"
        >
          {m['ui:header.nav.projects']()}
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-signal transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
        <a
          href="/about"
          class="group relative text-muted-foreground transition-colors duration-200 hover:text-signal"
        >
          {m['ui:header.nav.about']()}
          <span
            class="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-signal transition-transform duration-200 group-hover:scale-x-100"
          ></span>
        </a>
      </nav>

      <!-- 右侧功能区 -->
      <div class="ml-auto flex items-center">
        <!-- 移动端通知入口：与桌面铃铛共用组件，未登录时组件内部不渲染 -->
        <div class="md:hidden">
          <NotificationBell />
        </div>

        <!-- 桌面端功能按钮：高频开关常驻，低频信息类入口收纳进更多菜单 -->
        <div class="hidden items-center space-x-1.5 md:flex">
          <!-- 通知铃铛：登录后展示未读数与最近通知 -->
          <NotificationBell />

          <!-- 语言切换 -->
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <Button
                variant="ghost"
                size="icon"
                class="h-9 w-9"
                aria-label={m['ui:header.switchLanguage']()}
              >
                <Globe class="h-4 w-4" />
              </Button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end" class="w-max bg-popover/90 backdrop-blur-md">
              <DropdownMenu.Item onclick={() => setLanguage('zh')}>简体中文</DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => setLanguage('en')}>English</DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>

          <!-- 主题切换 -->
          <ThemeToggle variant="ghost" label={m['ui:header.themeToggle']()} />

          <!-- 更多：GitHub、关于作者与友情链接等低频入口 -->
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <Button
                variant="ghost"
                size="icon"
                class="h-9 w-9"
                aria-label={m['ui:header.more']()}
              >
                <Ellipsis class="h-4 w-4" />
              </Button>
            </DropdownMenu.Trigger>
            <!-- w-max 覆盖 $ui 默认的锚点宽度绑定，图标触发器下菜单按内容自适应不换行 -->
            <DropdownMenu.Content align="end" class="w-max bg-popover/90 backdrop-blur-md">
              <DropdownMenu.Item onclick={() => openExternal(GITHUB_URL)}>
                <GithubIcon class="h-4 w-4" />
                GitHub
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => (isAuthorDialogOpen = true)}>
                <User class="h-4 w-4" />
                {m['ui:header.aboutAuthor']()}
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => (isFriendlyLinkDialogOpen = true)}>
                <ExternalLink class="h-4 w-4" />
                {m['ui:linkDialog.title']()}
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>

          <!-- 关于作者弹窗：由更多菜单触发 -->
          <Dialog.Root bind:open={isAuthorDialogOpen}>
            <Dialog.Content class="bg-popover/95 backdrop-blur-md sm:max-w-md">
              <Dialog.Header>
                <Dialog.Title>{m['ui:header.aboutAuthor']()}</Dialog.Title>
                <Dialog.Description>{m['ui:header.aboutAuthorDesc']()}</Dialog.Description>
              </Dialog.Header>
              <div class="flex flex-col space-y-3">
                <div class="flex items-center space-x-3">
                  <div class="flex h-12 w-12 items-center justify-center rounded-none bg-signal">
                    <span class="text-lg font-bold text-signal-foreground">M</span>
                  </div>
                  <div>
                    <h3 class="font-semibold text-foreground">{m['ui:header.authorName']()}</h3>
                    <p class="text-sm text-muted-foreground">{m['ui:header.authorTitle']()}</p>
                  </div>
                </div>
              </div>
            </Dialog.Content>
          </Dialog.Root>

          <!-- 友情链接：列表展示与互换申请，由更多菜单触发 -->
          <FriendlyLinkDialog {friendlyLinks} bind:open={isFriendlyLinkDialogOpen} />
        </div>

        <!-- 登录/登出和后台按钮 -->
        <div class="ml-3 hidden items-center space-x-3 border-l border-border pl-3 md:flex">
          {#if isAuthenticated}
            <!-- 后台管理按钮 -->
            <Button
              variant="outline"
              size="sm"
              onclick={openAdminConsole}
              class="border-signal/30 text-signal hover:bg-signal/10 dark:border-signal/30 dark:text-signal dark:hover:bg-signal/10"
            >
              <Settings class="mr-1 h-4 w-4" />
              {m['ui:header.adminConsole']()}
            </Button>

            <!-- 用户信息下拉菜单 -->
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-9 w-9"
                  aria-label={m['ui:header.userMenu']()}
                >
                  <div class="flex h-6 w-6 items-center justify-center rounded-none bg-signal">
                    <span class="text-xs font-bold text-signal-foreground">
                      {currentUser?.username?.charAt(0)?.toUpperCase() || 'U'}
                    </span>
                  </div>
                </Button>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="end" class="min-w-48 bg-popover/90 backdrop-blur-md">
                <div class="border-b border-border px-3 py-2">
                  <p class="text-sm font-medium text-foreground">
                    {currentUser?.username || m['ui:header.userFallback']()}
                  </p>
                  <p class="text-xs text-muted-foreground">
                    {currentUser?.email || ''}
                  </p>
                </div>
                <DropdownMenu.Item onclick={() => goto('/profile')}>
                  {m['ui:header.myProfile']()}
                </DropdownMenu.Item>
                <DropdownMenu.Item onclick={() => goto('/favorites')}>
                  {m['ui:header.myFavorites']()}
                </DropdownMenu.Item>
                <DropdownMenu.Item
                  onclick={handleLogout}
                  class="text-destructive hover:text-destructive/80"
                >
                  {m['ui:header.logout']()}
                </DropdownMenu.Item>
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          {:else}
            <!-- 登录按钮 -->
            <Button
              size="sm"
              onclick={() => goto('/login')}
              class="bg-signal text-signal-foreground hover:bg-signal/90"
            >
              <LogIn class="mr-1 h-4 w-4" />
              {m['ui:header.login']()}
            </Button>
          {/if}
        </div>

        <!-- 移动端菜单 -->
        <Sheet.Root bind:open={isMobileMenuOpen}>
          <Sheet.Trigger>
            <Button
              variant="ghost"
              size="icon"
              class="h-9 w-9 lg:hidden"
              aria-label={m['ui:header.openMenu']()}
            >
              <Menu class="h-5 w-5" />
            </Button>
          </Sheet.Trigger>
          <Sheet.Content side="right" class="w-80 bg-popover/95 p-6 backdrop-blur-md">
            <Sheet.Header class="mb-6 text-left">
              <Sheet.Title class="text-xl font-bold">{m['ui:header.menu']()}</Sheet.Title>
            </Sheet.Header>

            <!-- 移动端导航链接 -->
            <nav class="space-y-1">
              <a
                href="/"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-foreground transition-colors hover:bg-signal/10 hover:text-signal"
                onclick={() => (isMobileMenuOpen = false)}
              >
                {m['ui:header.nav.blog']()}
              </a>
              <a
                href="/projects"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-foreground transition-colors hover:bg-signal/10 hover:text-signal"
                onclick={() => (isMobileMenuOpen = false)}
              >
                {m['ui:header.nav.projects']()}
              </a>
              <a
                href="/about"
                class="flex items-center rounded-lg px-4 py-3 text-lg font-medium text-foreground transition-colors hover:bg-signal/10 hover:text-signal"
                onclick={() => (isMobileMenuOpen = false)}
              >
                {m['ui:header.nav.about']()}
              </a>
            </nav>

            <!-- 移动端功能按钮 -->
            <div class="mt-8 space-y-6">
              <div class="flex items-center justify-between bg-secondary px-4 py-2">
                <span class="text-sm font-medium text-muted-foreground">
                  {m['ui:header.themeToggle']()}
                </span>
                <ThemeToggle label={m['ui:header.themeToggle']()} />
              </div>

              <!-- GitHub链接 -->
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center justify-between bg-secondary px-4 py-2 transition-colors hover:bg-accent"
              >
                <span class="text-sm font-medium text-muted-foreground">GitHub</span>
                <GithubIcon class="h-4 w-4 text-muted-foreground" />
              </a>

              <div class="space-y-3">
                <h4 class="px-4 text-sm font-semibold text-foreground">
                  {m['ui:header.languageSection']()}
                </h4>
                <div class="space-y-1">
                  <button
                    class="flex w-full items-center rounded-lg px-4 py-2 text-left text-sm transition-colors hover:bg-accent"
                    onclick={() => setLanguage('zh')}
                  >
                    简体中文
                  </button>
                  <button
                    class="flex w-full items-center rounded-lg px-4 py-2 text-left text-sm transition-colors hover:bg-accent"
                    onclick={() => setLanguage('en')}
                  >
                    English
                  </button>
                </div>
              </div>
            </div>

            <!-- 移动端登录/后台区域 -->
            <div class="mt-8 border-t border-border pt-6">
              {#if isAuthenticated}
                <div class="space-y-3">
                  <div class="bg-signal/5 px-4 py-3">
                    <p class="text-sm font-semibold text-foreground">
                      {currentUser?.username || m['ui:header.userFallback']()}
                    </p>
                    <p class="text-xs text-muted-foreground">
                      {currentUser?.email || ''}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start"
                    onclick={() => {
                      isMobileMenuOpen = false
                      goto('/profile')
                    }}
                  >
                    {m['ui:header.myProfile']()}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start border-signal/30 text-signal hover:bg-signal/10 dark:border-signal/30 dark:text-signal dark:hover:bg-signal/10"
                    onclick={() => {
                      isMobileMenuOpen = false
                      openAdminConsole()
                    }}
                  >
                    <Settings class="mr-2 h-4 w-4" />
                    {m['ui:header.adminConsole']()}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start border-destructive/30 text-destructive hover:bg-destructive/10"
                    onclick={() => {
                      isMobileMenuOpen = false
                      handleLogout()
                    }}
                  >
                    {m['ui:header.logout']()}
                  </Button>
                </div>
              {:else}
                <Button
                  size="sm"
                  class="w-full bg-signal text-signal-foreground hover:bg-signal/90"
                  onclick={() => {
                    isMobileMenuOpen = false
                    goto('/login')
                  }}
                >
                  <LogIn class="mr-2 h-4 w-4" />
                  {m['ui:header.login']()}
                </Button>
              {/if}
            </div>
          </Sheet.Content>
        </Sheet.Root>
      </div>
    </div>
  </div>
</header>

<style>
/* 中部导航采用绝对定位居中，父容器需要相对定位作为定位上下文。 */
header > div > div {
  position: relative;
}
</style>
