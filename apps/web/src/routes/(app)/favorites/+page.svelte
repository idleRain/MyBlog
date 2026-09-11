<script lang="ts">
import { ArticleIndexList, PaginationNav } from '$lib/components/article'
import type { Article } from '@myblog/api/modules/article/types'
import { page as pageStore } from '$app/stores'
import { SITE_NAME_ZH } from '@myblog/shared'
import { authStore } from '$lib/stores/auth'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { ArticleAPI } from '$lib/api'

// 收藏列表每页数量，与其他目录页保持一致的版面节奏。
const FAVORITES_PAGE_SIZE = 12

// 后端统一响应的成功业务码。
const RESPONSE_CODE_SUCCESS = 200

// 收藏数据与分页状态。
let articles = $state<Article[]>([])
let total = $state(0)
let currentPage = $state(1)
let loading = $state(true)

// 收藏数据依赖登录令牌，未登录时引导前往登录页。
let isAuthenticated = $state(false)
authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
})

// 分页由 URL 查询参数驱动，分页链接变更时自动重新加载。
const queryPage = $derived(
  Math.max(1, Math.floor(Number($pageStore.url.searchParams.get('page')) || 1))
)

const totalPages = $derived(Math.max(1, Math.ceil(total / FAVORITES_PAGE_SIZE)))

// 登录态或页码变化时拉取收藏列表，登出时清空。
$effect(() => {
  if (!isAuthenticated) {
    articles = []
    total = 0
    currentPage = 1
    void goto('/login')
    return
  }
  void loadFavorites(queryPage)
})

// 拉取指定页的收藏文章。
async function loadFavorites(targetPage: number) {
  loading = true
  try {
    const response = await ArticleAPI.bookmarks({ page: targetPage, pageSize: FAVORITES_PAGE_SIZE })
    if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
      toast.error(response.message || '收藏列表加载失败，请稍后重试')
      return
    }
    articles = response.data.articles
    total = response.data.total
    currentPage = response.data.page
  } catch {
    toast.error('收藏列表加载失败，请稍后重试')
  } finally {
    loading = false
  }
}
</script>

<svelte:head>
  <title>我的收藏 - {SITE_NAME_ZH}</title>
  <meta name="description" content="当前登录用户收藏的文章列表。" />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：与目录页同构。 -->
    <div class="mb-8 flex items-end justify-between gap-4 border-b border-line pb-4">
      <div>
        <p
          class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          收藏 · Bookmarks
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-3xl font-black">我的收藏</h1>
      </div>
      <span class="shrink-0 font-mono text-sm text-muted-foreground">共 {total} 篇</span>
    </div>

    {#if loading}
      <p class="py-10 text-center text-sm text-muted-foreground">正在加载收藏…</p>
    {:else if articles.length === 0}
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">还没有收藏任何文章。</p>
        <p class="mt-2 text-sm text-muted-foreground">在文章页点击收藏按钮，即可在这里找到它们。</p>
      </div>
    {:else}
      <ArticleIndexList {articles} offset={(currentPage - 1) * FAVORITES_PAGE_SIZE} />

      <PaginationNav {currentPage} {totalPages} basePath="/favorites" />
    {/if}
  </div>
</section>
