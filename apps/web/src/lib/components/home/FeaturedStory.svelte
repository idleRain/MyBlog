<script lang="ts">
import { featuredStory } from '$lib/data/home-content'
import { gsap, MOTION } from '$lib/motion/gsap-setup'
import { scrollStagger } from '$lib/motion/reveal'

// 占位数据的详情路由尚未建立，链接指向站内锚点，业务接入后替换。
const STORY_LINK = '#articles'

let root = $state<HTMLElement>()

$effect(() => {
  const element = root
  if (!element) return

  const ctx = gsap.context(() => {
    const media = gsap.matchMedia()

    media.add('(prefers-reduced-motion: no-preference)', () => {
      // 封面以 clip-path 自底部展开，刊号同步上升，构成翻开封面的仪式感。
      // 起止值均需显式声明，clip-path 的自然值为 none，无法参与插值。
      gsap.fromTo(
        '.featured-cover',
        { clipPath: 'inset(0% 0% 100% 0%)' },
        {
          clipPath: 'inset(0% 0% 0% 0%)',
          duration: 1,
          ease: MOTION.ease.inOut,
          scrollTrigger: { trigger: element, start: 'top 78%', once: true }
        }
      )

      gsap.from('.featured-issue-no', {
        yPercent: 24,
        autoAlpha: 0,
        duration: MOTION.duration.reveal,
        ease: MOTION.ease.out,
        scrollTrigger: { trigger: element, start: 'top 78%', once: true }
      })
    })
  }, element)

  return () => ctx.revert()
})
</script>

<section bind:this={root} id="featured" class="mx-auto max-w-6xl px-4 pb-16 sm:px-6">
  <article class="grid overflow-hidden border border-line bg-secondary md:grid-cols-2">
    <!-- 封面版面：深色渐变底 + 大号刊号，作为版面的视觉锚点 -->
    <div
      class="featured-cover flex min-h-64 flex-col justify-end bg-gradient-to-br from-[#3a2f26] via-[#6b4f3a] to-[#c83e1d] p-8 text-primary-foreground"
      aria-hidden="true"
    >
      <p class="featured-issue-no font-display text-6xl font-black tracking-tight">
        {featuredStory.issueNo}
      </p>
      <p class="mt-auto font-display text-lg">{featuredStory.category}</p>
    </div>

    <!-- 文章信息区：元信息、标题、摘要与全文入口，进入视口后错峰浮现 -->
    <div class="flex flex-col justify-center gap-6 p-8 sm:p-10" use:scrollStagger>
      <p class="flex items-center gap-3 text-sm text-muted-foreground">
        <span class="bg-signal px-2 py-0.5 text-xs font-bold text-signal-foreground">精选</span>
        {featuredStory.date} · 约 {featuredStory.readMinutes} 分钟
      </p>

      <h2 class="font-display text-3xl leading-snug font-black">
        <a href={STORY_LINK} class="transition-colors duration-200 hover:text-signal">
          {featuredStory.title}
        </a>
      </h2>

      <p class="leading-relaxed text-muted-foreground">{featuredStory.excerpt}</p>

      <a
        href={STORY_LINK}
        class="group inline-flex w-fit items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
      >
        阅读全文
        <span
          class="transition-transform duration-200 ease-(--ease-out-strong) group-hover:translate-x-1"
          aria-hidden="true"
        >
          →
        </span>
      </a>
    </div>
  </article>
</section>
