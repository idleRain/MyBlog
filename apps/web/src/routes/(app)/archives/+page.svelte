<script lang="ts">
import { ArchiveTimeline } from '$lib/components/article'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'

let { data }: PageProps = $props()
</script>

<svelte:head>
  <title>归档 - {SITE_NAME_ZH}</title>
  <meta name="description" content="全部文章的编年归档，按年份与月份编排。" />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：与目录页同构的栏目标注与刊头。 -->
    <div class="mb-14 flex items-end justify-between border-b border-line pb-4">
      <div>
        <p
          class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          编年 · Archives
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-3xl font-black">归档</h1>
      </div>
      <span class="font-mono text-sm text-muted-foreground">按年份编排</span>
    </div>

    {#if data.groups.length === 0}
      <!-- 空状态：暂无已发布文章时保持编辑口吻。 -->
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">时间线正在等待第一批文章。</p>
        <p class="mt-2 text-sm text-muted-foreground">可以先回到首页浏览版面与站点介绍。</p>
      </div>
    {:else}
      <ArchiveTimeline groups={data.groups} />
    {/if}
  </div>
</section>
