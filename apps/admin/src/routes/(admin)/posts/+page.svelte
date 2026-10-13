<script lang="ts">
import ArticleFilter from '$lib/components/admin/article/article-filter.svelte'
import { postsPageState, emptyPostsPageData } from './posts-page-state.svelte'
import ArticleTable from '$lib/components/admin/article/article-table.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { ARTICLE_PAGE_SIZE } from '$lib/constants/article'
import { FileText, Plus, RotateCcw } from '@lucide/svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { goto } from '$lib/utils/navigation'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  postsPageState.initialize(data ?? emptyPostsPageData())
})
</script>

<svelte:head>
  <title>文章管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="文章管理" description="管理博客文章，支持搜索、筛选与状态流转" crumb="文章管理">
  {#snippet actions()}
    <Button onclick={() => goto('/posts/new')}>
      <Plus data-icon="inline-start" />
      新建文章
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <ArticleFilter
        bind:search={postsPageState.search}
        bind:status={postsPageState.status}
        bind:sortBy={postsPageState.sortBy}
        bind:order={postsPageState.order}
        statusDisabled={!postsPageState.canManageAllArticles}
        onSearch={postsPageState.applySearchDebounced}
        onFilterChange={() => postsPageState.applyFilterChange()}
        onReset={() => postsPageState.resetFilters()}
      />

      {#if postsPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载文章列表" />
        </div>
      {:else if postsPageState.articles.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <FileText class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {postsPageState.hasActiveFilters ? '没有匹配的文章' : '暂无文章'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {postsPageState.hasActiveFilters ? '调整或重置筛选条件后再试' : '写下第一篇博客文章'}
            </p>
          </div>
          {#if postsPageState.hasActiveFilters}
            <Button variant="outline" size="sm" onclick={() => postsPageState.resetFilters()}>
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {:else}
            <Button size="sm" onclick={() => goto('/posts/new')}>
              <Plus data-icon="inline-start" />
              新建文章
            </Button>
          {/if}
        </div>
      {:else}
        <ArticleTable
          articles={postsPageState.articles}
          onStatusAction={(article, action) => postsPageState.applyStatusAction(article, action)}
          onDelete={article => postsPageState.removeArticle(article)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {postsPageState.total} 篇文章</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={postsPageState.total}
            perPage={ARTICLE_PAGE_SIZE}
            page={postsPageState.currentPage}
            onPageChange={page => postsPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {postsPageState.currentPage} 页，共 {postsPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
</PageHeader>
