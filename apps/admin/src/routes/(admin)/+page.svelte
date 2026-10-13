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
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { UserRole } from '@myblog/api/modules/user/types'
import { SITE_NAME_ZH } from '@myblog/shared'
import { goto } from '$lib/utils/navigation'
import type { PageProps } from './$types'
import { Button } from '$ui/button'
import * as Card from '$ui/card'

// 仪表盘区块可见性与各区块首屏数据均由 load 注入，组件只做渲染与跳转分发。
let { data }: PageProps = $props()

// 角色到仪表盘标题的映射，标题跟随后端下发的角色。
const WELCOME_TITLES: Record<UserRole, string> = {
  superadmin: '超级管理员仪表盘',
  admin: '管理员仪表盘',
  editor: '编辑工作台',
  user: '个人工作台'
}
</script>

<svelte:head>
  <title>管理仪表盘 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title={WELCOME_TITLES[data.userRole]}
  crumb="仪表盘"
  description={data.nickname ? `欢迎回来，${data.nickname}！` : `欢迎使用 ${SITE_NAME_ZH} 管理后台`}
>
  {#if data.overview}
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard label="文章总数" value={data.overview.articleCount} icon={FileText} />
      <StatCard label="已发布文章" value={data.overview.publishedCount} icon={Send} />
      <StatCard label="总浏览量" value={data.overview.totalViews} icon={Eye} />
      <StatCard label="总点赞" value={data.overview.totalLikes} icon={Heart} />
      <StatCard label="评论总数" value={data.overview.commentCount} icon={MessageSquare} />
      <StatCard label="用户总数" value={data.overview.userCount} icon={Users} />
      <StatCard label="分类数" value={data.overview.categoryCount} icon={FolderTree} />
      <StatCard label="标签数" value={data.overview.tagCount} icon={Tags} />
    </div>
  {/if}

  <div class="grid gap-6 lg:grid-cols-3">
    {#if data.overview}
      <div class="lg:col-span-2">
        <Card.Root>
          <Card.Header>
            <Card.Title>近 7 天浏览量趋势</Card.Title>
            <Card.Description>每日文章访问量统计</Card.Description>
          </Card.Header>
          <Card.Content>
            <ViewsTrendChart dates={data.trendDates} values={data.trendValues} />
          </Card.Content>
        </Card.Root>
      </div>
    {/if}

    <div class="space-y-6 {data.overview ? '' : 'lg:col-span-3'}">
      <Card.Root>
        <Card.Header>
          <Card.Title>快速操作</Card.Title>
          <Card.Description>常用的管理操作</Card.Description>
        </Card.Header>
        <Card.Content class="space-y-3">
          {#if data.canCreateArticle}
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
          {#if data.canViewArticles}
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
          {#if data.canManageUsers}
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

      {#if data.canViewArticles}
        <Card.Root>
          <Card.Header>
            <Card.Title>最新文章</Card.Title>
          </Card.Header>
          <Card.Content class="divide-y">
            {#each data.recentArticles as article (article.id)}
              <button
                type="button"
                onclick={() => goto(`/posts/${article.id}`)}
                class="flex w-full items-center justify-between gap-2 py-2.5 text-left transition-colors first:pt-0 last:pb-0 hover:text-primary"
              >
                <span class="line-clamp-1 text-sm font-medium">{article.title}</span>
                <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
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
</PageHeader>
