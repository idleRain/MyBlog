<script lang="ts">
import { ArchiveTimeline } from '$lib/components/article'
import type { PageProps } from './$types'
import { m } from '$i18n'

let { data }: PageProps = $props()

// 标题在多处头部标签复用，站名取自词表以随语言切换。
const pageTitle = m['blog:archives.title']({ site: m['ui:site.name']() })
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['blog:archives.description']()} />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：与目录页同构的栏目标注与刊头。 -->
    <div class="mb-14 flex items-end justify-between border-b border-line pb-4">
      <div>
        <p
          class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          {m['blog:archives.eyebrow']()}
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-3xl font-black">{m['blog:archives.heading']()}</h1>
      </div>
      <span class="font-mono text-sm text-muted-foreground">{m['blog:archives.orderNote']()}</span>
    </div>

    {#if data.groups.length === 0}
      <!-- 空状态：暂无已发布文章时保持编辑口吻。 -->
      <div class="border-l-2 border-signal py-4 pl-6">
        <p class="font-display text-xl font-medium">{m['blog:archives.emptyTitle']()}</p>
        <p class="mt-2 text-sm text-muted-foreground">{m['ui:empty.backHome']()}</p>
      </div>
    {:else}
      <ArchiveTimeline groups={data.groups} />
    {/if}
  </div>
</section>
