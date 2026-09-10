<script lang="ts">
import { timelineYears } from '$lib/data/home-content'
import { gsap, MOTION } from '$lib/motion/gsap-setup'
import { scrollReveal } from '$lib/motion/reveal'

let root = $state<HTMLElement>()

$effect(() => {
  if (!root) return

  const ctx = gsap.context(() => {
    const media = gsap.matchMedia()

    media.add('(prefers-reduced-motion: no-preference)', () => {
      // 每条年度轨道的朱红进度线随滚动绘制，走到哪画到哪。
      const rails = gsap.utils.toArray<HTMLElement>('[data-rail]')
      rails.forEach(rail => {
        const progress = rail.querySelector<HTMLElement>('[data-rail-progress]')
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

<section bind:this={root} id="timeline" class="mx-auto max-w-6xl px-4 pb-24 sm:px-6">
  <div class="mb-14 flex items-end justify-between border-b border-line pb-4" use:scrollReveal>
    <h2 class="font-display text-3xl font-black">创作年轮</h2>
    <span class="font-mono text-sm text-muted-foreground">TIMELINE — 按年归档</span>
  </div>

  <div class="flex flex-col gap-20">
    {#each timelineYears as year (year.year)}
      <div class="grid gap-8 md:grid-cols-[9rem_1fr] md:gap-12">
        <!-- 年度标记：桌面端粘性跟随，滚动时固定在左栏 -->
        <header class="self-start md:sticky md:top-28" use:scrollReveal={{ y: 18 }}>
          <p class="font-display text-5xl font-black text-foreground">{year.year}</p>
          <p class="mt-2 font-mono text-xs text-muted-foreground">共 {year.total} 条记录</p>
        </header>

        <!-- 年度轨道：基线为线框色，进度线随滚动以朱红填充 -->
        <div class="relative ml-1" data-rail>
          <span class="absolute inset-y-0 left-0 w-px bg-line" aria-hidden="true"></span>
          <span
            class="absolute inset-y-0 left-0 w-0.5 origin-top -translate-x-1/2 scale-y-0 bg-signal"
            data-rail-progress
            aria-hidden="true"
          ></span>

          <ol class="flex flex-col gap-10">
            {#each year.entries as entry (entry.id)}
              <li class="relative pl-8 sm:pl-10" use:scrollReveal={{ y: 20 }}>
                <span
                  class="absolute top-1.5 left-0 h-3 w-3 -translate-x-1/2 rounded-full border-2 border-signal bg-background"
                  aria-hidden="true"
                ></span>
                <p class="font-mono text-xs font-bold text-signal">{entry.month} 月</p>
                <h3 class="mt-1.5 font-display text-xl leading-snug font-bold">
                  {entry.title}
                </h3>
                <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">{entry.excerpt}</p>
              </li>
            {/each}
          </ol>
        </div>
      </div>
    {/each}
  </div>
</section>
