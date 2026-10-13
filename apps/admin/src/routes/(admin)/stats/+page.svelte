<script lang="ts">
import { FileText, Eye, Heart, MessageSquare, Users, FolderTree, Tags, Send } from '@lucide/svelte'
import ViewsTrendChart from '$lib/components/admin/stats/views-trend-chart.svelte'
import { statsPageState, emptyStatsPageData } from './stats-page-state.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { TREND_DAY_OPTIONS } from '$lib/constants/stats'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import * as Card from '$ui/card'

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

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  statsPageState.initialize(data ?? emptyStatsPageData())
})
</script>

<svelte:head>
  <title>站点统计 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="站点统计" description="运营数据分析与文章浏览量趋势" crumb="站点统计">
  {#if statsPageState.isLoading && !statsPageState.overview}
    <div class="flex h-48 items-center justify-center">
      <Spinner class="size-6" aria-label="正在加载统计数据" />
    </div>
  {:else}
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {#each OVERVIEW_CARDS as card (card.key)}
        {@const IconComponent = card.icon}
        <Card.Root>
          <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
            <Card.Title class="text-sm font-medium">{card.label}</Card.Title>
            <IconComponent class="size-4 text-muted-foreground" />
          </Card.Header>
          <Card.Content>
            <div class="text-2xl font-bold tabular-nums">
              {statsPageState.overview ? statsPageState.overview[card.key] : '—'}
            </div>
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
            value={statsPageState.trendDays}
            onValueChange={value => statsPageState.setTrendDays(value)}
            variant="outline"
            size="sm"
          >
            {#each TREND_DAY_OPTIONS as days (days)}
              <ToggleGroup.Item value={days}>{days} 天</ToggleGroup.Item>
            {/each}
          </ToggleGroup.Root>
        </div>
      </Card.Header>
      <Card.Content>
        <ViewsTrendChart dates={statsPageState.trendDates} values={statsPageState.trendValues} />
      </Card.Content>
    </Card.Root>
  {/if}
</PageHeader>
