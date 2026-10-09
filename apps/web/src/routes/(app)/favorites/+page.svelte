<script lang="ts">
import { ArticleIndexList, ArticleListSkeleton, PaginationNav } from '$lib/components/article'
import type { Article } from '@myblog/api/modules/article/types'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'
import { navigating } from '$app/state'

// 收藏列表每页数量，与 +page.server.ts 保持一致，用于计算列表跨页编号偏移。
const FAVORITES_PAGE_SIZE = 12

// 骨架行的下限，避免收藏极少时骨架比内容更矮而造成高度跳动。
const MIN_SKELETON_ROWS = 3

// 数据由 +page.server.ts 的 load 在服务端取回，页面只负责渲染。
let { data }: PageProps = $props()

// 页面状态直接派生自 load 结果，因此首屏不存在「先渲染容器再补拉」的空窗。
const articles = $derived<Article[]>(data.list.articles)
const total = $derived(data.list.total)
const currentPage = $derived(data.currentPage)
const totalPages = $derived(Math.max(1, Math.ceil(total / FAVORITES_PAGE_SIZE)))

// 翻页经客户端路由触发，load 在服务端重新求值期间以骨架屏承接等待。
// 仅在目标路径仍为本页时接管显示，跳转其他页面时不清空既有内容。
const isNavigatingHere = $derived(navigating?.to?.url.pathname === '/favorites')

// 骨架行数取当前页规模与下限的较大值，使骨架高度接近加载完成后的版式。
const skeletonRows = $derived(Math.max(MIN_SKELETON_ROWS, articles.length))
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

    {#if isNavigatingHere}
      <ArticleListSkeleton rows={skeletonRows} />
    {:else if articles.length}
      <ArticleIndexList {articles} offset={(currentPage - 1) * FAVORITES_PAGE_SIZE} />

      <PaginationNav {currentPage} {totalPages} basePath="/favorites" />
    {:else}
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">还没有收藏任何文章。</p>
        <p class="mt-2 text-sm text-muted-foreground">在文章页点击收藏按钮，即可在这里找到它们。</p>
      </div>
    {/if}
  </div>
</section>
