<script lang="ts">
import type { ArticleArchiveYear } from '@myblog/api/modules/article/types'
import { gsap, MOTION } from '$lib/motion/gsap-setup'
import { formatDate } from '$lib/utils/format-date'
import { scrollReveal } from '$lib/motion/reveal'

// 归档时间线属性：groups 为按年份倒序的归档分组。
interface Props {
  groups: ArticleArchiveYear[]
}

let { groups }: Props = $props()

let root = $state<HTMLElement>()

$effect(() => {
  if (!root) return

  const ctx = gsap.context(() => {
    const media = gsap.matchMedia()

    media.add('(prefers-reduced-motion: no-preference)', () => {
      // 每条年度轨道的朱红进度线随滚动绘制，实现方式与首页创作年轮一致。
      const rails = gsap.utils.toArray<HTMLElement>('[data-archive-rail]')
      rails.forEach(rail => {
        const progress = rail.querySelector<HTMLElement>('[data-archive-rail-progress]')
        if (!progress) return

        gsap.fromTo(
          progress,
          { scaleY: 0 },
          {
            scaleY: 1,
            ease: MOTION.ease.linear,
            scrollTrigger: {
              trigger: rail,
              start: 'top 70%',
              end: 'bottom 55%',
              scrub: true
            }
          }
        )
      })
    })
  }, root)

  return () => ctx.revert()
})
</script>

<div bind:this={root} class="flex flex-col gap-20">
  {#each groups as group (group.year)}
    <div class="grid gap-8 md:grid-cols-[9rem_1fr] md:gap-12">
      <!-- 年度标记：桌面端粘性跟随，滚动时固定在左栏。 -->
      <header class="self-start md:sticky md:top-28" use:scrollReveal={{ y: 18 }}>
        <p class="font-display text-5xl font-black text-foreground">{group.year}</p>
        <p class="mt-2 font-mono text-xs text-muted-foreground">共 {group.total} 篇文章</p>
      </header>

      <!-- 年度轨道：基线为线框色，进度线随滚动以朱红填充。 -->
      <div class="relative ml-1" data-archive-rail>
        <span class="absolute inset-y-0 left-0 w-px bg-line" aria-hidden="true"></span>
        <span
          class="absolute inset-y-0 left-0 w-0.5 origin-top -translate-x-1/2 scale-y-0 bg-signal"
          data-archive-rail-progress
          aria-hidden="true"
        ></span>

        <ol class="flex flex-col gap-10">
          {#each group.months as month (month.month)}
            {#each month.articles as article (article.id)}
              <li class="relative pl-8 sm:pl-10" use:scrollReveal={{ y: 20 }}>
                <span
                  class="absolute top-1.5 left-0 h-3 w-3 -translate-x-1/2 rounded-full border-2 border-signal bg-background"
                  aria-hidden="true"
                ></span>
                <p class="font-mono text-xs font-bold text-signal">
                  {month.month} 月 · {formatDate(article.publishedAt ?? article.createdAt)}
                </p>
                <h3 class="mt-1.5 font-display text-xl leading-snug font-bold">
                  <a
                    href={`/blog/${article.slug}`}
                    class="transition-colors duration-200 hover:text-signal"
                  >
                    {article.title}
                  </a>
                </h3>
                {#if article.summary}
                  <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                    {article.summary}
                  </p>
                {/if}
              </li>
            {/each}
          {/each}
        </ol>
      </div>
    </div>
  {/each}
</div>
