<script lang="ts">
import { Card, Badge } from '$ui'
import { PageHeader } from '$lib/components/admin'
import { authStore } from '$lib/stores/auth'
import type { User, UserRole, DashboardStats, QuickAction, RecentActivity } from '$lib/types'
import {
  Activity,
  ArrowUpRight,
  Clock,
  Edit,
  FileText,
  PlusCircle,
  Settings,
  Shield,
  TrendingUp,
  Users as UsersIcon
} from '@lucide/svelte'

// 用户角色和权限
let currentUser = $state<User | null>(null)
let userRole = $state<UserRole>('user')

// 仪表盘数据
let dashboardStats = $derived<DashboardStats>({
  totalUsers: 128,
  totalPosts: 456,
  activeUsers: 32,
  systemStatus: 'normal'
})

// 获取用户信息
$effect(() => {
  authStore.subscribe(state => {
    if (state.isAuthenticated && state.user) {
      currentUser = state.user
      userRole = state.user.role || 'user'
    }
  })
})

// 统计卡片规格行文案：按角色拼装欢迎语，避免模板内多层条件分支。
let roleTitle = $derived.by(() => {
  switch (userRole) {
    case 'superadmin':
      return '超级管理员仪表盘'
    case 'admin':
      return '管理员仪表盘'
    case 'editor':
      return '编辑工作台'
    default:
      return '个人工作台'
  }
})

let roleHint = $derived.by(() => {
  switch (userRole) {
    case 'superadmin':
      return '您拥有系统最高权限，请谨慎操作。'
    case 'admin':
      return '您可以管理用户和内容。'
    case 'editor':
      return '您可以创建和管理文章内容。'
    default:
      return '您可以查看系统概览。'
  }
})

// 根据用户角色显示不同的快速操作
let quickActions = $derived.by((): QuickAction[] => {
  const actions: QuickAction[] = []

  // 所有用户都可以发布文章（如果角色是editor及以上）
  if (['editor', 'admin', 'superadmin'].includes(userRole)) {
    actions.push({
      id: 'create-article',
      title: '发布文章',
      description: '创建新的博客文章',
      icon: PlusCircle,
      action: () => goto('/manage/posts?action=create'),
      roles: ['editor', 'admin', 'superadmin']
    })
    actions.push({
      id: 'manage-articles',
      title: '管理文章',
      description: '管理已发布的文章',
      icon: Edit,
      action: () => goto('/manage/posts'),
      roles: ['editor', 'admin', 'superadmin']
    })
  }

  // 只有管理员可以创建用户
  if (['admin', 'superadmin'].includes(userRole)) {
    actions.push({
      id: 'create-user',
      title: '创建用户',
      description: '添加新的系统用户',
      icon: UsersIcon,
      action: () => goto('/manage/users?action=create'),
      roles: ['admin', 'superadmin']
    })
  }

  // 只有超级管理员可以访问系统设置
  if (userRole === 'superadmin') {
    actions.push({
      id: 'system-settings',
      title: '系统设置',
      description: '管理系统配置',
      icon: Settings,
      action: () => goto('/manage/settings'),
      roles: ['superadmin']
    })
  }

  return actions
})

const recentActivities: RecentActivity[] = [
  { id: '1', action: '用户登录', user: 'admin', time: '2分钟前', type: 'login' },
  { id: '2', action: '创建文章', user: 'editor', time: '15分钟前', type: 'create' },
  { id: '3', action: '用户注册', user: 'system', time: '1小时前', type: 'register' },
  { id: '4', action: '修改设置', user: 'admin', time: '2小时前', type: 'update' }
]

// 统计卡的规格编号：以两位等宽序号对应图纸条目编号。
const STAT_CARD_WIDTH_CLASS = 'md:grid-cols-2 xl:grid-cols-4'
</script>

<svelte:head>
  <title>管理仪表盘 - MyBlog</title>
</svelte:head>

<!-- 头部导航 -->
<PageHeader crumbs={[{ label: '管理后台', href: '/manage' }, { label: '仪表盘' }]} />

<!-- 主内容区域：网格底纹仅覆盖首屏，向下渐隐以保持数据区可读。 -->
<div class="relative flex-1 overflow-y-auto">
  <div class="admin-grid pointer-events-none absolute inset-0" aria-hidden="true"></div>
  <div class="relative z-10 flex flex-col gap-6 p-4 sm:p-6">
    <!-- 欢迎区域：等宽眉标加主标题，沿用前台规格书排版。 -->
    <div class="flex flex-col gap-2">
      <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
        <span class="text-signal">//</span> DASHBOARD - OVERVIEW
      </p>
      <h1 class="text-2xl font-bold tracking-tight sm:text-3xl">{roleTitle}</h1>
      <p class="text-sm text-muted-foreground">
        {#if currentUser}
          欢迎回来，{currentUser.nickname || currentUser.username}。{roleHint}
        {:else}
          欢迎使用 MyBlog 管理后台。
        {/if}
      </p>
    </div>

    <!-- 统计卡片：mono 序号加 tabular 数字，构成规格书数据行。 -->
    <div class="grid grid-cols-1 gap-4 {STAT_CARD_WIDTH_CLASS}">
      {#if ['admin', 'superadmin'].includes(userRole)}
        <Card.Root class="rounded-none border border-border ring-0">
          <Card.Header class="flex-row items-start justify-between">
            <Card.Title class="text-sm font-medium text-muted-foreground">总用户数</Card.Title>
            <span class="font-mono text-xs text-muted-foreground/60">/ 01</span>
          </Card.Header>
          <Card.Content>
            <div class="text-3xl font-bold tracking-tight tabular-nums">
              {dashboardStats.totalUsers}
            </div>
            <p class="mt-2 flex items-center gap-1.5 font-mono text-xs text-muted-foreground">
              <TrendingUp class="size-3 text-signal" />
              较上月增长 12%
            </p>
          </Card.Content>
        </Card.Root>
      {/if}

      {#if ['editor', 'admin', 'superadmin'].includes(userRole)}
        <Card.Root class="rounded-none border border-border ring-0">
          <Card.Header class="flex-row items-start justify-between">
            <Card.Title class="text-sm font-medium text-muted-foreground">文章总数</Card.Title>
            <span class="font-mono text-xs text-muted-foreground/60">/ 02</span>
          </Card.Header>
          <Card.Content>
            <div class="text-3xl font-bold tracking-tight tabular-nums">
              {dashboardStats.totalPosts}
            </div>
            <p class="mt-2 flex items-center gap-1.5 font-mono text-xs text-muted-foreground">
              <TrendingUp class="size-3 text-signal" />
              较上月增长 8%
            </p>
          </Card.Content>
        </Card.Root>
      {/if}

      {#if ['admin', 'superadmin'].includes(userRole)}
        <Card.Root class="rounded-none border border-border ring-0">
          <Card.Header class="flex-row items-start justify-between">
            <Card.Title class="text-sm font-medium text-muted-foreground">活跃用户</Card.Title>
            <span class="font-mono text-xs text-muted-foreground/60">/ 03</span>
          </Card.Header>
          <Card.Content>
            <div class="text-3xl font-bold tracking-tight tabular-nums">
              {dashboardStats.activeUsers}
            </div>
            <p class="mt-2 flex items-center gap-1.5 font-mono text-xs text-muted-foreground">
              <Clock class="size-3 text-signal" />
              过去 24 小时
            </p>
          </Card.Content>
        </Card.Root>
      {/if}

      <!-- 所有角色都可以看到系统状态 -->
      <Card.Root class="rounded-none border border-border ring-0">
        <Card.Header class="flex-row items-start justify-between">
          <Card.Title class="text-sm font-medium text-muted-foreground">系统状态</Card.Title>
          <span class="font-mono text-xs text-muted-foreground/60">/ 04</span>
        </Card.Header>
        <Card.Content>
          <div class="flex h-9 items-center gap-2.5">
            <!-- 状态点：signal 呼吸点与文字共同指示状态，不依赖单一颜色。 -->
            <span class="relative flex size-2 shrink-0">
              <span
                class="absolute inline-flex size-full animate-ping rounded-full bg-signal opacity-60"
              ></span>
              <span class="relative inline-flex size-2 rounded-full bg-signal"></span>
            </span>
            <span class="text-lg font-semibold">运行正常</span>
          </div>
          <p class="mt-1 font-mono text-xs text-muted-foreground">所有服务正常运行</p>
        </Card.Content>
      </Card.Root>
    </div>

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <!-- 快速操作 -->
      <Card.Root class="rounded-none border border-border ring-0">
        <Card.Header>
          <Card.Title>快速操作</Card.Title>
          <Card.Description>常用的管理操作</Card.Description>
        </Card.Header>
        <Card.Content class="flex flex-col gap-2">
          {#each quickActions as action (action.id)}
            {@const IconComponent = action.icon}
            <button
              type="button"
              class="group flex w-full cursor-pointer items-center gap-3 border border-border bg-card p-3 text-left transition-colors duration-200 outline-none hover:border-signal/50 hover:bg-signal/5 focus-visible:ring-2 focus-visible:ring-ring"
              onclick={action.action}
            >
              <span
                class="flex size-8 shrink-0 items-center justify-center bg-signal/10 text-signal [&_svg]:size-4"
              >
                <IconComponent />
              </span>
              <span class="flex min-w-0 flex-1 flex-col">
                <span class="text-sm font-medium">{action.title}</span>
                <span class="truncate text-xs text-muted-foreground">{action.description}</span>
              </span>
              <ArrowUpRight
                class="size-3.5 shrink-0 text-muted-foreground transition-all duration-200 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-signal"
              />
            </button>
          {/each}
        </Card.Content>
      </Card.Root>

      <!-- 最近活动：序号加时间的日志式排版，对应规格书修订记录。 -->
      <Card.Root class="rounded-none border border-border ring-0 lg:col-span-2">
        <Card.Header>
          <Card.Title>最近活动</Card.Title>
          <Card.Description>系统最新的操作记录</Card.Description>
        </Card.Header>
        <Card.Content>
          <div class="flex flex-col">
            {#each recentActivities as activity, index (activity.id)}
              <div class="flex items-center gap-4 border-b border-border py-3 last:border-b-0">
                <span class="w-8 shrink-0 font-mono text-xs text-muted-foreground/70">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm">
                    {activity.action}
                    <span class="text-muted-foreground"> · {activity.user}</span>
                  </p>
                  <p class="mt-0.5 font-mono text-xs text-muted-foreground">{activity.time}</p>
                </div>
                <Badge variant="outline" class="rounded-none font-mono text-xs">
                  {activity.type}
                </Badge>
              </div>
            {/each}
          </div>
        </Card.Content>
      </Card.Root>
    </div>

    <!-- 底部规格行：与首屏 Hero 的底部标注行同构，收束页面。 -->
    <div
      class="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 border-t border-border pt-4 font-mono text-xs text-muted-foreground"
    >
      <span>MyBlog ADMIN CONSOLE</span>
      <span class="flex items-center gap-2">
        <Activity class="size-3 text-signal" />
        STATUS: <span class="text-foreground">OK</span>
        <FileText class="ml-4 size-3 text-signal" />
        RENDERED BY <span class="text-foreground">SVELTE 5</span>
        <Shield class="ml-4 size-3 text-signal" />
        RBAC: <span class="text-foreground">ON</span>
      </span>
    </div>
  </div>
</div>
