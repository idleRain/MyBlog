<script lang="ts">
import { ArticleIndexList, PaginationNav } from '$lib/components/article'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'

let { data }: PageProps = $props()

// 总页数向上取整，无文章时保持 1 以呈现完整导航骨架。
const totalPages = $derived(Math.max(1, Math.ceil(data.list.total / data.list.pageSize)))
</script>

<svelte:head>
  <title>博客目录 - {SITE_NAME_ZH}</title>
  <meta name="description" content="全部文章的目录页，按发布时间编排，支持翻页浏览。" />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：栏目标注、刊头标题与总量标注，与首页板块标题同构。 -->
    <div class="mb-8 flex items-end justify-between gap-4 border-b border-line pb-4">
      <div>
        <p
          class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          手记 · 全部刊目
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-3xl font-black">博客目录</h1>
      </div>
      <span class="shrink-0 font-mono text-sm text-muted-foreground">
        共 {data.list.total} 篇 · 持续更新
      </span>
    </div>

    {#if data.list.articles.length === 0}
      <!-- 空状态：以编辑口吻说明现状，避免出现工程化提示。 -->
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">创刊号正在筹备中，文章即将发布。</p>
        <p class="mt-2 text-sm text-muted-foreground">可以先回到首页浏览版面与站点介绍。</p>
      </div>
    {:else}
      <ArticleIndexList
        articles={data.list.articles}
        offset={(data.currentPage - 1) * data.list.pageSize}
      />

      <!-- 分页导航：等宽标注风格，越界端以灰显占位。 -->
      <PaginationNav currentPage={data.currentPage} {totalPages} basePath="/blog" />
    {/if}
  </div>
</section>
