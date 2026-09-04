<script lang="ts">
import {
  FileText,
  Send,
  Eye,
  Heart,
  MessageSquare,
  Users,
  FolderTree,
  Tags,
  PlusCircle,
  Edit
} from '@lucide/svelte'
import ViewsTrendChart from '$lib/components/admin/stats/views-trend-chart.svelte'
import StatCard from '$lib/components/admin/dashboard/stat-card.svelte'
import type { StatsOverview } from '@myblog/api/modules/stats/types'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { Article } from '@myblog/api/modules/article/types'
import type { UserRole } from '@myblog/api/modules/user/types'
import { hasPermission } from '$lib/utils/permissions'
import { PERMISSIONS } from '$lib/constants/auth'
import { ArticleAPI, StatsAPI } from '$lib/api'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { Button, Card } from '$ui'
import { onMount } from 'svelte'

// 当前用户角色与昵称，区块裁剪依据后端登录下发的权限列表（铁律 A4）。
let userRole = $state<UserRole>('user')
let canViewStats = $state(false)
let canViewArticles = $state(false)
let canCreateArticle = $state(false)
let canManageUsers = $state(false)
let nickname = $state('')

let overview = $state<StatsOverview | null>(null)
let isLoading = $state(true)
let trendDates = $state<string[]>([])
let trendValues = $state<number[]>([])
let recentArticles = $state<Article[]>([])

/**
 * 加载仪表盘数据。
 * 统计概览与趋势仅具备 system:stats 权限的角色可读取，按下发权限做前端区块裁剪；
 * 最新文章列表对具备 article:list 权限的角色统一走 /api/articles/list，状态可见性由后端判定。
 */
async function loadDashboard() {
  isLoading = true
  try {
    const currentState = authStore.getCurrentState()
    const currentUser = currentState.user
    userRole = currentUser?.role ?? 'user'
    nickname = currentUser?.nickname ?? ''
    // 权限判定与后端各接口判定一致，修改权限表后重新登录即可联动变化。
    canViewStats = hasPermission(currentUser, PERMISSIONS.SYSTEM_STATS)
    canViewArticles = hasPermission(currentUser, PERMISSIONS.ARTICLE_LIST)
    canCreateArticle = hasPermission(currentUser, PERMISSIONS.ARTICLE_CREATE)
    canManageUsers = hasPermission(currentUser, PERMISSIONS.USER_LIST)

    if (canViewStats) {
      const [overviewResult, trendResult] = await Promise.all([
        StatsAPI.getOverview().catch(() => null),
        StatsAPI.getArticleViewsTrend(7).catch(() => null)
      ])

      if (overviewResult?.code === 200 && overviewResult.data) {
        overview = overviewResult.data
      }
      if (trendResult?.code === 200 && trendResult.data) {
        trendDates = trendResult.data.dates ?? []
        trendValues = trendResult.data.values ?? []
      }
    }

    if (canViewArticles) {
      const articleResult = await ArticleAPI.list({ page: 1, pageSize: 5 }).catch(() => null)
      if (articleResult?.code === 200 && articleResult.data) {
        recentArticles = articleResult.data.articles ?? []
      }
    }
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 依据角色计算仪表盘标题。
 */
function welcomeTitle(): string {
  const map: Record<UserRole, string> = {
    superadmin: '超级管理员仪表盘',
    admin: '管理员仪表盘',
    editor: '编辑工作台',
    user: '个人工作台'
  }
  return map[userRole]
}

onMount(loadDashboard)
</script>

<svelte:head>
  <title>管理仪表盘 - MyBlog</title>
</svelte:head>

<PageHeader
  title={welcomeTitle()}
  crumb="仪表盘"
  description={nickname ? `欢迎回来，${nickname}！` : '欢迎使用 MyBlog 管理后台'}
>
  {#if isLoading}
    <div class="flex h-48 items-center justify-center">
      <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
      ></span>
    </div>
  {:else}
    {#if overview}
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="文章总数" value={overview?.articleCount ?? 0} icon={FileText} />
        <StatCard label="已发布文章" value={overview?.publishedCount ?? 0} icon={Send} />
        <StatCard label="总浏览量" value={overview?.totalViews ?? 0} icon={Eye} />
        <StatCard label="总点赞" value={overview?.totalLikes ?? 0} icon={Heart} />
        <StatCard label="评论总数" value={overview?.commentCount ?? 0} icon={MessageSquare} />
        <StatCard label="用户总数" value={overview?.userCount ?? 0} icon={Users} />
        <StatCard label="分类数" value={overview?.categoryCount ?? 0} icon={FolderTree} />
        <StatCard label="标签数" value={overview?.tagCount ?? 0} icon={Tags} />
      </div>
    {/if}

    <div class="grid gap-6 lg:grid-cols-3">
      {#if overview}
        <div class="lg:col-span-2">
          <Card.Root>
            <Card.Header>
              <Card.Title>近 7 天浏览量趋势</Card.Title>
              <Card.Description>每日文章访问量统计</Card.Description>
            </Card.Header>
            <Card.Content>
              <ViewsTrendChart dates={trendDates} values={trendValues} />
            </Card.Content>
          </Card.Root>
        </div>
      {/if}

      <div class="space-y-6 {overview ? '' : 'lg:col-span-3'}">
        <Card.Root>
          <Card.Header>
            <Card.Title>快速操作</Card.Title>
            <Card.Description>常用的管理操作</Card.Description>
          </Card.Header>
          <Card.Content class="space-y-3">
            {#if canCreateArticle}
              <Button
                variant="outline"
                class="h-auto w-full justify-start p-4"
                onclick={() => goto('/posts/new')}
              >
                <div class="flex items-start gap-3">
                  <PlusCircle class="mt-0.5 size-4" />
                  <div class="text-left">
                    <p class="font-medium">发布文章</p>
                    <p class="text-sm text-muted-foreground">创建新的博客文章</p>
                  </div>
                </div>
              </Button>
            {/if}
            {#if canViewArticles}
              <Button
                variant="outline"
                class="h-auto w-full justify-start p-4"
                onclick={() => goto('/posts')}
              >
                <div class="flex items-start gap-3">
                  <Edit class="mt-0.5 size-4" />
                  <div class="text-left">
                    <p class="font-medium">管理文章</p>
                    <p class="text-sm text-muted-foreground">管理已发布的文章</p>
                  </div>
                </div>
              </Button>
            {/if}
            {#if canManageUsers}
              <Button
                variant="outline"
                class="h-auto w-full justify-start p-4"
                onclick={() => goto('/users')}
              >
                <div class="flex items-start gap-3">
                  <Users class="mt-0.5 size-4" />
                  <div class="text-left">
                    <p class="font-medium">用户管理</p>
                    <p class="text-sm text-muted-foreground">管理后台用户与角色</p>
                  </div>
                </div>
              </Button>
            {/if}
          </Card.Content>
        </Card.Root>

        {#if canViewArticles}
          <Card.Root>
            <Card.Header>
              <Card.Title>最新文章</Card.Title>
            </Card.Header>
            <Card.Content class="space-y-3">
              {#each recentArticles as article (article.id)}
                <button
                  type="button"
                  onclick={() => goto(`/posts/${article.id}`)}
                  class="flex w-full items-center justify-between gap-2 text-left transition-colors hover:text-primary"
                >
                  <span class="line-clamp-1 text-sm font-medium">{article.title}</span>
                  <span class="shrink-0 text-xs text-muted-foreground">
                    {new Date(article.createdAt).toLocaleDateString('zh-CN')}
                  </span>
                </button>
              {:else}
                <p class="text-sm text-muted-foreground">暂无文章</p>
              {/each}
            </Card.Content>
          </Card.Root>
        {/if}
      </div>
    </div>
  {/if}
</PageHeader>
