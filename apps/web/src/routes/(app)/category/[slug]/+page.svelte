<script lang="ts">
import { ArticleIndexList, PaginationNav } from '$lib/components/article'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'

let { data }: PageProps = $props()

const category = $derived(data.category)

// 总页数向上取整，无文章时保持 1 以呈现完整导航骨架。
const totalPages = $derived(Math.max(1, Math.ceil(data.list.total / data.list.pageSize)))
</script>

<svelte:head>
  <title>{category.seoTitle || category.name} - 分类目录 - {SITE_NAME_ZH}</title>
  <meta name="description" content={category.seoDescription || category.description} />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 分类标题栏：栏目标注、分类名与总量标注。 -->
    <div class="mb-8 border-b border-line pb-4">
      <p
        class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
      >
        分类 · {category.slug}
        <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
      </p>
      <div class="flex items-end justify-between gap-4">
        <h1 class="font-display text-3xl font-black">{category.name}</h1>
        <span class="shrink-0 font-mono text-sm text-muted-foreground">
          共 {data.list.total} 篇
        </span>
      </div>
      {#if category.description}
        <p class="mt-3 max-w-prose text-sm leading-relaxed text-muted-foreground">
          {category.description}
        </p>
      {/if}
    </div>

    {#if data.list.articles.length === 0}
      <!-- 空状态：以编辑口吻说明现状并引导前往目录页。 -->
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">该分类下暂无已发布的文章。</p>
        <p class="mt-2 text-sm text-muted-foreground">可以去目录页看看其他内容。</p>
      </div>
    {:else}
      <ArticleIndexList
        articles={data.list.articles}
        offset={(data.currentPage - 1) * data.list.pageSize}
      />

      <PaginationNav
        currentPage={data.currentPage}
        {totalPages}
        basePath={`/category/${category.slug}`}
      />
    {/if}
  </div>
</section>
