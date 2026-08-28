<script lang="ts">
import { Card, Button, Badge, Breadcrumb, Sidebar } from '$ui'
import { authStore } from '$lib/stores/auth'
import type { User, UserRole, DashboardStats, QuickAction, RecentActivity } from '$lib/types'
import {
  Users,
  FileText,
  Settings,
  TrendingUp,
  Activity,
  Clock,
  Shield,
  PlusCircle,
  Edit
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

// 模拟数据（实际应该从API获取）

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
      color: 'bg-green-500',
      roles: ['editor', 'admin', 'superadmin']
    })
    actions.push({
      id: 'manage-articles',
      title: '管理文章',
      description: '管理已发布的文章',
      icon: Edit,
      action: () => goto('/manage/posts'),
      color: 'bg-blue-500',
      roles: ['editor', 'admin', 'superadmin']
    })
  }

  // 只有管理员可以创建用户
  if (['admin', 'superadmin'].includes(userRole)) {
    actions.push({
      id: 'create-user',
      title: '创建用户',
      description: '添加新的系统用户',
      icon: Users,
      action: () => goto('/manage/users?action=create'),
      color: 'bg-purple-500',
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
      color: 'bg-orange-500',
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
</script>

<svelte:head>
  <title>管理仪表盘 - MyBlog</title>
</svelte:head>

<!-- 头部导航 -->
<header class="flex h-16 shrink-0 items-center gap-2 border-b px-6">
  <Sidebar.Trigger />
  <Breadcrumb.Root>
    <Breadcrumb.List>
      <Breadcrumb.Item>
        <Breadcrumb.Link href="/manage">管理后台</Breadcrumb.Link>
      </Breadcrumb.Item>
      <Breadcrumb.Separator />
      <Breadcrumb.Item>
        <Breadcrumb.Page>仪表盘</Breadcrumb.Page>
      </Breadcrumb.Item>
    </Breadcrumb.List>
  </Breadcrumb.Root>
</header>

<!-- 主内容区域 -->
<main class="flex-1 space-y-6 p-6">
  <!-- 欢迎区域 -->
  <div class="flex flex-col space-y-2">
    <h1 class="text-3xl font-bold tracking-tight">
      {#if userRole === 'superadmin'}
        超级管理员仪表盘
      {:else if userRole === 'admin'}
        管理员仪表盘
      {:else if userRole === 'editor'}
        编辑工作台
      {:else}
        个人工作台
      {/if}
    </h1>
    <p class="text-muted-foreground">
      {#if currentUser}
        欢迎回来，{currentUser.nickname || currentUser.username}！
        {#if userRole === 'superadmin'}
          您拥有系统最高权限，请谨慎操作。
        {:else if userRole === 'admin'}
          您可以管理用户和内容。
        {:else if userRole === 'editor'}
          您可以创建和管理文章内容。
        {:else}
          您可以查看系统概览。
        {/if}
      {:else}
        欢迎使用 MyBlog 管理后台
      {/if}
    </p>
  </div>

  <!-- 统计卡片 -->
  <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
    <!-- 管理员和超级管理员可以看到用户统计 -->
    {#if ['admin', 'superadmin'].includes(userRole)}
      <Card.Root>
        <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
          <Card.Title class="text-sm font-medium">总用户数</Card.Title>
          <Users class="h-4 w-4 text-muted-foreground" />
        </Card.Header>
        <Card.Content>
          <div class="text-2xl font-bold">{dashboardStats.totalUsers}</div>
          <p class="text-xs text-muted-foreground">
            <TrendingUp class="mr-1 inline h-3 w-3" />
            较上月增长 12%
          </p>
        </Card.Content>
      </Card.Root>
    {/if}

    <!-- 编辑及以上角色可以看到文章统计 -->
    {#if ['editor', 'admin', 'superadmin'].includes(userRole)}
      <Card.Root>
        <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
          <Card.Title class="text-sm font-medium">文章总数</Card.Title>
          <FileText class="h-4 w-4 text-muted-foreground" />
        </Card.Header>
        <Card.Content>
          <div class="text-2xl font-bold">{dashboardStats.totalPosts}</div>
          <p class="text-xs text-muted-foreground">
            <TrendingUp class="mr-1 inline h-3 w-3" />
            较上月增长 8%
          </p>
        </Card.Content>
      </Card.Root>
    {/if}

    <!-- 管理员和超级管理员可以看到活跃用户 -->
    {#if ['admin', 'superadmin'].includes(userRole)}
      <Card.Root>
        <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
          <Card.Title class="text-sm font-medium">活跃用户</Card.Title>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </Card.Header>
        <Card.Content>
          <div class="text-2xl font-bold">{dashboardStats.activeUsers}</div>
          <p class="text-xs text-muted-foreground">
            <Clock class="mr-1 inline h-3 w-3" />
            过去24小时
          </p>
        </Card.Content>
      </Card.Root>
    {/if}

    <!-- 所有角色都可以看到系统状态 -->
    <Card.Root>
      <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
        <Card.Title class="text-sm font-medium">系统状态</Card.Title>
        <Shield class="h-4 w-4 text-muted-foreground" />
      </Card.Header>
      <Card.Content>
        <div class="flex items-center space-x-2">
          <Badge variant="default" class="bg-green-500">运行正常</Badge>
        </div>
        <p class="mt-1 text-xs text-muted-foreground">所有服务正常运行</p>
      </Card.Content>
    </Card.Root>
  </div>

  <div class="grid gap-6 lg:grid-cols-3">
    <!-- 快速操作 -->
    <div class="lg:col-span-1">
      <Card.Root>
        <Card.Header>
          <Card.Title>快速操作</Card.Title>
          <Card.Description>常用的管理操作</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-3">
          {#each quickActions as action, index (index)}
            {@const IconComponent = action.icon}
            <Button
              variant="outline"
              class="h-auto w-full justify-start p-4"
              onclick={action.action}
            >
              <div class="flex items-start space-x-3">
                <div class="rounded-md p-2 {action.color} text-white">
                  <IconComponent class="h-4 w-4" />
                </div>
                <div class="text-left">
                  <p class="font-medium">{action.title}</p>
                  <p class="text-sm text-muted-foreground">{action.description}</p>
                </div>
              </div>
            </Button>
          {/each}
        </Card.Content>
      </Card.Root>
    </div>

    <!-- 最近活动 -->
    <div class="lg:col-span-2">
      <Card.Root>
        <Card.Header>
          <Card.Title>最近活动</Card.Title>
          <Card.Description>系统最新的操作记录</Card.Description>
        </Card.Header>
        <Card.Content>
          <div class="space-y-4">
            {#each recentActivities as activity, index (index)}
              <div class="flex items-center space-x-3">
                <div class="h-2 w-2 rounded-full bg-blue-500"></div>
                <div class="flex-1 space-y-1">
                  <p class="text-sm font-medium">
                    {activity.action}
                    <span class="text-muted-foreground">by {activity.user}</span>
                  </p>
                  <p class="text-xs text-muted-foreground">{activity.time}</p>
                </div>
                <Badge variant="outline" class="text-xs">
                  {activity.type}
                </Badge>
              </div>
            {/each}
          </div>
        </Card.Content>
      </Card.Root>
    </div>
  </div>
</main>
