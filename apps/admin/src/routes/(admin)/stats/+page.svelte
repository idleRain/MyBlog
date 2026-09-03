<script lang="ts">
import { FileText, Eye, Heart, MessageSquare, Users, FolderTree, Tags, Send } from '@lucide/svelte'
import ViewsTrendChart from '$lib/components/admin/stats/views-trend-chart.svelte'
import type { StatsOverview } from '@myblog/api/modules/stats/types'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { Card, ToggleGroup } from '$ui'
import { StatsAPI } from '$lib/api'
import { onMount } from 'svelte'

// 概览卡片配置，统一图标与语义色。
const OVERVIEW_CARDS = [
  { key: 'articleCount', label: '文章总数', icon: FileText },
  { key: 'publishedCount', label: '已发布', icon: Send },
  { key: 'totalViews', label: '总浏览量', icon: Eye },
  { key: 'totalLikes', label: '总点赞', icon: Heart },
  { key: 'commentCount', label: '评论总数', icon: MessageSquare },
  { key: 'userCount', label: '用户总数', icon: Users },
  { key: 'categoryCount', label: '分类数', icon: FolderTree },
  { key: 'tagCount', label: '标签数', icon: Tags }
] as const

let overview = $state<StatsOverview | null>(null)
let isLoading = $state(true)
// 趋势天数以字符串承载，适配 ToggleGroup 的字符串值约束。
let trendDays = $state('7')
let trendDates = $state<string[]>([])
let trendValues = $state<number[]>([])

/**
 * 加载统计概览与浏览量趋势。
 */
async function loadStats() {
  isLoading = true
  try {
    const [overviewResult, trendResult] = await Promise.all([
      StatsAPI.getOverview(),
      StatsAPI.getArticleViewsTrend(Number(trendDays))
    ])
    if (overviewResult.code === 200 && overviewResult.data) {
      overview = overviewResult.data
    }
    if (trendResult.code === 200 && trendResult.data) {
      trendDates = trendResult.data.dates ?? []
      trendValues = trendResult.data.values ?? []
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 切换趋势天数后重新加载。
 */
function handleDaysChange() {
  loadStats()
}

onMount(loadStats)
</script>

<svelte:head>
  <title>站点统计 - MyBlog</title>
</svelte:head>

<PageHeader title="站点统计" description="运营数据分析与文章浏览量趋势" crumb="站点统计">
  {#if isLoading && !overview}
    <div class="flex h-48 items-center justify-center">
      <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
      ></span>
    </div>
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {#each OVERVIEW_CARDS as card (card.key)}
        {@const IconComponent = card.icon}
        <Card.Root>
          <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
            <Card.Title class="text-sm font-medium">{card.label}</Card.Title>
            <IconComponent class="size-4 text-muted-foreground" />
          </Card.Header>
          <Card.Content>
            <div class="text-2xl font-bold">{overview ? overview[card.key] : '—'}</div>
          </Card.Content>
        </Card.Root>
      {/each}
    </div>

    <Card.Root>
      <Card.Header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <Card.Title>文章浏览量趋势</Card.Title>
            <Card.Description>按日统计文章访问量，缺失日期补零</Card.Description>
          </div>
          <ToggleGroup.Root
            type="single"
            bind:value={trendDays}
            onValueChange={handleDaysChange}
            variant="outline"
            size="sm"
          >
            <ToggleGroup.Item value="7">7 天</ToggleGroup.Item>
            <ToggleGroup.Item value="14">14 天</ToggleGroup.Item>
            <ToggleGroup.Item value="30">30 天</ToggleGroup.Item>
          </ToggleGroup.Root>
        </div>
      </Card.Header>
      <Card.Content>
        <ViewsTrendChart dates={trendDates} values={trendValues} />
      </Card.Content>
    </Card.Root>
  {/if}
</PageHeader>
