<script lang="ts">
import { ArticleIndexList, PaginationNav } from '$lib/components/article'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'
import { Search } from '@lucide/svelte'

let { data }: PageProps = $props()

// 总页数向上取整，无文章时保持 1 以呈现完整导航骨架。
const totalPages = $derived(Math.max(1, Math.ceil(data.list.total / data.list.pageSize)))

// 翻页时保留当前检索词，空检索词时不携带查询串。
const searchQuery = $derived(data.search ? `search=${encodeURIComponent(data.search)}` : '')
</script>

<svelte:head>
  <title>博客目录 - {SITE_NAME_ZH}</title>
  <meta name="description" content="全部文章的目录页，按发布时间编排，支持关键词检索与翻页浏览。" />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：栏目标注、刊头标题与总量标注，与首页板块标题同构；小屏允许统计换行避免长检索词溢出。 -->
    <div class="mb-8 flex flex-wrap items-end justify-between gap-4 border-b border-line pb-4">
      <div>
        <p
          class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          手记 · 全部刊目
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-3xl font-black">博客目录</h1>
      </div>
      {#if data.search}
        <span class="shrink-0 font-mono text-sm text-muted-foreground">
          检索「{data.search}」· {data.list.total} 篇结果
        </span>
      {:else}
        <span class="shrink-0 font-mono text-sm text-muted-foreground">
          共 {data.list.total} 篇 · 持续更新
        </span>
      {/if}
    </div>

    <!-- 检索区：原生 GET 表单提交 search 参数，复用后端 ngram 全文索引，无脚本参与。 -->
    <form method="get" action="/blog" role="search" class="mb-10 flex items-stretch gap-3">
      <label for="blog-search" class="sr-only">搜索文章</label>
      <div class="relative flex-1">
        <Search
          class="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        <input
          id="blog-search"
          type="search"
          name="search"
          value={data.search}
          placeholder="检索标题与正文关键词…"
          class="w-full border border-line bg-card py-2.5 pr-3 pl-9 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
        />
      </div>
      <button
        type="submit"
        class="shrink-0 bg-signal px-5 text-sm font-bold text-signal-foreground transition-[background-color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal/90 active:scale-[0.97]"
      >
        检索
      </button>
    </form>

    {#if data.list.articles.length === 0}
      {#if data.search}
        <!-- 检索空态：说明无命中并引导清除检索，避免与创刊空态混淆。 -->
        <div class="border-l-2 border-signal py-4 pl-6">
          <p class="font-display text-xl font-medium">未检索到与「{data.search}」相关的文章。</p>
          <p class="mt-2 text-sm text-muted-foreground">
            可以更换关键词重试，或
            <a href="/blog" class="text-signal underline-offset-4 hover:underline">清除检索</a>
            浏览全部刊目。
          </p>
        </div>
      {:else}
        <!-- 空状态：以编辑口吻说明现状，避免出现工程化提示。 -->
        <div class="border-l-2 border-signal py-4 pl-6">
          <p class="font-display text-xl font-medium">创刊号正在筹备中，文章即将发布。</p>
          <p class="mt-2 text-sm text-muted-foreground">可以先回到首页浏览版面与站点介绍。</p>
        </div>
      {/if}
    {:else}
      <ArticleIndexList
        articles={data.list.articles}
        offset={(data.currentPage - 1) * data.list.pageSize}
      />

      <!-- 分页导航：等宽标注风格，越界端以灰显占位。 -->
      <PaginationNav
        currentPage={data.currentPage}
        {totalPages}
        basePath="/blog"
        query={searchQuery}
      />
    {/if}
  </div>
</section>
