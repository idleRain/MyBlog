<script lang="ts">
import { gsap, MOTION } from '$lib/motion/gsap-setup'
import { ArrowUpRight } from '@lucide/svelte'

// 外部链接与站内锚点集中声明，避免散落的魔法字符串。
const GITHUB_URL = 'https://github.com/idleRain'
const ARTICLES_ANCHOR = '#articles'

let root = $state<HTMLElement>()

$effect(() => {
  const element = root
  if (!element) return

  const ctx = gsap.context(() => {
    const media = gsap.matchMedia()

    media.add('(prefers-reduced-motion: no-preference)', () => {
      // 首屏入场编排： hairline 展开、标题遮罩上升、文案与行动区错峰浮现。
      const timeline = gsap.timeline({ defaults: { ease: MOTION.ease.out } })
      timeline
        .from('.hero-eyebrow-rule', {
          scaleX: 0,
          duration: 0.7,
          transformOrigin: 'left center'
        })
        .from('.hero-eyebrow-label', { autoAlpha: 0, y: 14, duration: 0.5 }, '<0.08')
        .from(
          '.hero-line-inner',
          {
            yPercent: 112,
            duration: MOTION.duration.reveal,
            stagger: 0.12
          },
          '-=0.35'
        )
        .from('.hero-intro', { autoAlpha: 0, y: 22, duration: 0.7 }, '-=0.5')
        .from('.hero-cta', { autoAlpha: 0, y: 18, duration: 0.6, stagger: MOTION.stagger }, '-=0.4')
        .from(
          '.hero-quote-rule',
          {
            scaleY: 0,
            transformOrigin: 'top center',
            duration: 0.7,
            ease: MOTION.ease.inOut
          },
          '-=0.5'
        )
        .from('.hero-quote-body', { autoAlpha: 0, y: 18, duration: 0.7 }, '-=0.4')
        .from('.hero-spec', { autoAlpha: 0, y: 14, duration: 0.6 }, '-=0.3')
        .from('.hero-ghost', { autoAlpha: 0, duration: 1.4 }, '-=1.0')

      // 首屏滚离时的视差退场：内容轻微上移并降低不透明度，为正文让出舞台。
      gsap.to('.hero-content', {
        yPercent: -9,
        autoAlpha: 0.2,
        ease: MOTION.ease.linear,
        scrollTrigger: { trigger: element, start: 'top top', end: 'bottom top', scrub: true }
      })

      // 幽灵刊号以更慢的速率下沉，形成前后景深。
      gsap.to('.hero-ghost', {
        yPercent: 22,
        ease: MOTION.ease.linear,
        scrollTrigger: { trigger: element, start: 'top top', end: 'bottom top', scrub: true }
      })

      // 滚动提示随首次滚动淡出，避免残留在正文区域。
      gsap.to('.hero-scroll-hint', {
        autoAlpha: 0,
        ease: MOTION.ease.linear,
        scrollTrigger: { trigger: element, start: 'top top', end: '12% top', scrub: true }
      })
    })
  }, element)

  return () => ctx.revert()
})
</script>

<section bind:this={root} class="relative flex min-h-screen items-center overflow-hidden pt-16">
  <div class="hero-content relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <div class="grid items-end gap-10 md:grid-cols-12">
      <!-- 刊头主区：大标题与简介构成开篇版式 -->
      <div class="md:col-span-8">
        <p
          class="hero-eyebrow-label mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          手记 · 代码与生活
          <span class="hero-eyebrow-rule h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>

        <h1
          class="font-display text-5xl leading-[1.08] font-black tracking-tight text-balance sm:text-6xl lg:text-7xl"
        >
          <span class="block overflow-hidden">
            <span class="hero-line-inner block">把想法，</span>
          </span>
          <span class="block overflow-hidden">
            <span class="hero-line-inner block">
              编译成<span class="text-signal">界面</span>
              <span
                class="ml-4 align-middle font-mono text-sm font-normal tracking-normal text-muted-foreground sm:text-base"
                >// 工程 × 设计</span
              >
            </span>
          </span>
        </h1>

        <p class="hero-intro mt-8 max-w-xl text-lg leading-relaxed text-muted-foreground">
          记录工程实践的思考、城市漫游的见闻，以及那些深夜读完的书。
          相信缓慢而扎实的积累，偏爱干净的结构与诚实的文字。
        </p>

        <div class="mt-10 flex flex-wrap items-center gap-4">
          <a
            href={ARTICLES_ANCHOR}
            class="hero-cta inline-flex items-center gap-2 bg-primary px-7 py-3.5 text-sm font-bold text-primary-foreground transition-[background-color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97]"
          >
            开始阅读
            <span aria-hidden="true">↓</span>
          </a>
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noopener noreferrer"
            class="hero-cta inline-flex items-center gap-2 border border-border px-7 py-3.5 text-sm font-bold text-foreground transition-[border-color,color,transform] duration-200 ease-(--ease-out-strong) hover:border-signal hover:text-signal active:scale-[0.97]"
          >
            GitHub
            <ArrowUpRight class="h-4 w-4" aria-hidden="true" />
          </a>
        </div>
      </div>

      <!-- 编辑摘录：朱红竖线引出的本月手记 -->
      <div class="md:col-span-4" aria-label="本月摘录">
        <div class="hero-quote-rule border-l-2 border-signal pl-6">
          <blockquote class="hero-quote-body font-display text-xl leading-relaxed font-medium">
            “把想清楚的事写下来，把写下来的事做出来。”
          </blockquote>
          <p class="hero-quote-body mt-4 text-sm text-muted-foreground">—— 创刊编辑手记</p>
        </div>
      </div>
    </div>

    <!-- 底部规格行：延续刊物目录页的标注语言 -->
    <div class="hero-spec mt-20 border-t border-border">
      <div
        class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 py-5 font-mono text-xs text-muted-foreground"
      >
        <div class="flex items-center gap-6">
          <span>01 / 工程</span>
          <span>02 / 设计</span>
          <span>03 / 生活</span>
        </div>
        <span class="tracking-[0.18em] uppercase">MYBLOG — EST. 2024</span>
      </div>
    </div>
  </div>

  <!-- 幽灵刊号：右下角的大号衬线序号，滚动时以更慢速率下沉 -->
  <div
    class="hero-ghost pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    01
  </div>

  <!-- 滚动提示：竖排等宽标注与流动的朱红细线 -->
  <div class="hero-scroll-hint absolute bottom-8 left-6 z-10 hidden items-center gap-4 lg:flex">
    <span
      class="font-mono text-[10px] tracking-[0.32em] text-muted-foreground uppercase [writing-mode:vertical-rl]"
    >
      Scroll
    </span>
    <span class="scroll-track" aria-hidden="true"></span>
  </div>
</section>

<style>
/* 滚动提示轨道：朱红短线自上而下循环流动，示意滚动方向。 */
.scroll-track {
  display: inline-block;
  position: relative;
  background-color: var(--border);
  width: 1px;
  height: 3.5rem;
  overflow: hidden;
}

.scroll-track::after {
  position: absolute;
  inset-inline: 0;
  top: 0;
  animation: scroll-drip 1.8s var(--ease-in-out-strong) infinite;
  background-color: var(--signal);
  height: 40%;
  content: '';
}

@keyframes scroll-drip {
  0% {
    transform: translateY(-100%);
  }
  100% {
    transform: translateY(350%);
  }
}

/* 减少动态偏好下停用流动装饰，保留静态轨道。 */
@media (prefers-reduced-motion: reduce) {
  .scroll-track::after {
    animation: none;
  }
}
</style>
