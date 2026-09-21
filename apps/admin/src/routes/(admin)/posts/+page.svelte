<script lang="ts">
import type {
  Article,
  ArticleActionResponse,
  ArticleStatus
} from '@myblog/api/modules/article/types'
import { ARTICLE_PAGE_SIZE, type ArticleStatusAction } from '$lib/constants/article'
import ArticleFilter from '$lib/components/admin/article/article-filter.svelte'
import ArticleTable from '$lib/components/admin/article/article-table.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { FileText, Plus, RotateCcw } from '@lucide/svelte'
import { SITE_NAME_ZH, debounce } from '@myblog/shared'
import { hasPermission } from '$lib/utils/permissions'
import { PERMISSIONS } from '$lib/constants/auth'
import { Button, Card, Pagination } from '$ui'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { ArticleAPI } from '$lib/api'
import { onMount } from 'svelte'

// 筛选与分页状态，sortBy 类型与后端 GetArticleListRequest 的 oneof 约束对齐。
let articles = $state<Article[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let search = $state('')
let status = $state<ArticleStatus | ''>('')
let sortBy = $state<
  'created_at' | 'updated_at' | 'published_at' | 'view_count' | 'like_count' | ''
>('created_at')
let order = $state<'asc' | 'desc'>('desc')

// 当前用户是否具备 article:manage 权限，用于控制状态筛选是否可用。
let canManageAllArticles = $state(false)

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 是否存在生效中的搜索与状态筛选，用于空态区分「无数据」与「无匹配」。
const hasActiveFilters = $derived(search.trim() !== '' || status !== '')

// 文章状态操作到接口方法的映射，统一走 /api/articles 端点，权限由后端判定。
const STATUS_ACTION_METHODS: Record<
  ArticleStatusAction,
  (id: number) => Promise<ArticleActionResponse>
> = {
  publish: ArticleAPI.publish,
  unpublish: ArticleAPI.unpublish,
  archive: ArticleAPI.archive,
  private: ArticleAPI.private
}

/**
 * 加载文章列表，携带筛选、排序与分页参数。
 * 管理员可查看全部状态，其他角色由服务端强制只返回已发布文章。
 */
async function loadArticles() {
  isLoading = true
  try {
    const response = await ArticleAPI.list({
      page: currentPage,
      pageSize: ARTICLE_PAGE_SIZE,
      status,
      sortBy,
      order,
      ...(search.trim() ? { search: search.trim() } : {})
    })
    if (response.code === 200 && response.data) {
      articles = response.data.articles ?? []
      total = response.data.total ?? 0
    } else {
      toast.error(response.message || '加载文章列表失败')
    }
  } catch (error) {
    console.error('加载文章列表失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 防抖应用搜索词，连续输入只在停顿后请求一次。
 */
const applySearchDebounced = debounce(() => {
  currentPage = 1
  loadArticles()
}, SEARCH_DEBOUNCE_MS)

/**
 * 下拉筛选变更后立即回到第一页并重新加载。
 */
function handleFilterChange() {
  currentPage = 1
  loadArticles()
}

/**
 * 重置全部筛选条件后重新加载。
 */
function handleReset() {
  search = ''
  status = ''
  sortBy = 'created_at'
  order = 'desc'
  currentPage = 1
  loadArticles()
}

/**
 * 执行文章状态流转并提示结果。
 */
async function handleStatusAction(article: Article, action: ArticleStatusAction) {
  try {
    const response = await STATUS_ACTION_METHODS[action](article.id)
    if (response.code === 200) {
      toast.success(response.data?.message || '操作成功')
      loadArticles()
    } else {
      toast.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('文章状态操作失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

/**
 * 删除文章，由表格组件确认后触发。
 */
async function handleDelete(article: Article) {
  try {
    const response = await ArticleAPI.delete(article.id)
    if (response.code === 200) {
      toast.success('文章删除成功')
      loadArticles()
    } else {
      toast.error(response.message || '删除文章失败')
    }
  } catch (error) {
    console.error('删除文章失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

function handlePageChange(page: number) {
  currentPage = page
  loadArticles()
}

onMount(() => {
  const currentUser = authStore.getCurrentState().user
  canManageAllArticles = hasPermission(currentUser, PERMISSIONS.ARTICLE_MANAGE)
  loadArticles()
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
        bind:search
        bind:status
        bind:sortBy
        bind:order
        statusDisabled={!canManageAllArticles}
        onSearch={applySearchDebounced}
        onFilterChange={handleFilterChange}
        onReset={handleReset}
      />

      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if articles.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <FileText class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {hasActiveFilters ? '没有匹配的文章' : '暂无文章'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {hasActiveFilters ? '调整或重置筛选条件后再试' : '写下第一篇博客文章'}
            </p>
          </div>
          {#if hasActiveFilters}
            <Button variant="outline" size="sm" onclick={handleReset}>
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
        <ArticleTable {articles} onStatusAction={handleStatusAction} onDelete={handleDelete} />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {total} 篇文章</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={total}
            perPage={ARTICLE_PAGE_SIZE}
            page={currentPage}
            onPageChange={handlePageChange}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / ARTICLE_PAGE_SIZE))} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
</PageHeader>
