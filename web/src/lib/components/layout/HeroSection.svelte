<script lang="ts">
import { ArrowUpRight } from '@lucide/svelte'

// 外部链接与站内路径集中声明，避免散落的魔法字符串。
const GITHUB_URL = 'https://github.com/idleRain'
const PORTFOLIO_PATH = '/projects'

// 首屏元素的阶梯浮现延迟，单位毫秒，用于编排进入动画的先后节奏。
const REVEAL_DELAY_MS = {
  eyebrow: 0,
  headline: 90,
  body: 180,
  actions: 270,
  spec: 360
} as const
</script>

<section
  class="relative flex min-h-screen items-center overflow-hidden bg-background text-foreground"
>
  <!-- 坐标网格背景：用发丝线绘制工程图纸式的参考网格。 -->
  <div class="hero-grid absolute inset-0" aria-hidden="true"></div>

  <!-- 右侧出血线：模拟规格书页边处的标注线。 -->
  <div
    class="absolute inset-y-0 right-6 hidden border-l border-border lg:right-12"
    aria-hidden="true"
  ></div>

  <!-- 四角定位标记：强化「图纸 / 规格书」的裁切线质感。 -->
  <div
    class="absolute top-5 left-5 h-4 w-4 border-t border-l border-border"
    aria-hidden="true"
  ></div>
  <div
    class="absolute top-5 right-5 h-4 w-4 border-t border-r border-border"
    aria-hidden="true"
  ></div>
  <div
    class="absolute bottom-5 left-5 h-4 w-4 border-b border-l border-border"
    aria-hidden="true"
  ></div>
  <div
    class="absolute right-5 bottom-5 h-4 w-4 border-r border-b border-border"
    aria-hidden="true"
  ></div>

  <!-- 幽灵序号：右下角的大号索引，仅作装饰。 -->
  <div
    class="pointer-events-none absolute right-10 bottom-16 hidden font-mono text-[9rem] leading-none font-bold text-foreground/5 select-none lg:block"
    aria-hidden="true"
  >
    01
  </div>

  <div class="relative z-10 mx-auto w-full max-w-6xl px-6 sm:px-10 lg:px-20">
    <!-- 顶部标注行 -->
    <div
      class="reveal mb-16 flex items-center justify-between font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase"
      style={`--reveal-delay: ${REVEAL_DELAY_MS.eyebrow}ms`}
    >
      <span><span class="text-signal">//</span> 全栈工程师 · 界面设计师</span>
      <span class="hidden sm:inline">SPEC — 01</span>
    </div>

    <!-- 主标题 -->
    <h1
      class="reveal text-5xl font-bold tracking-tight text-balance sm:text-7xl lg:text-8xl"
      style={`--reveal-delay: ${REVEAL_DELAY_MS.headline}ms`}
    >
      <span class="block">把想法</span>
      <span class="block">
        <span class="text-signal">编译</span>成界面
        <span
          class="ml-4 align-middle font-mono text-sm font-normal tracking-normal text-muted-foreground sm:text-base"
          >// 工程 × 设计</span
        >
      </span>
    </h1>

    <!-- 介绍文案 -->
    <p
      class="reveal mt-8 max-w-xl text-base leading-relaxed text-muted-foreground sm:text-lg"
      style={`--reveal-delay: ${REVEAL_DELAY_MS.body}ms`}
    >
      专注工程与设计的交界。用 Svelte 与 Go，把复杂的问题，做成简单的界面。
    </p>

    <!-- 行动按钮 -->
    <div
      class="reveal mt-12 flex flex-wrap items-center gap-4"
      style={`--reveal-delay: ${REVEAL_DELAY_MS.actions}ms`}
    >
      <a
        href={PORTFOLIO_PATH}
        class="group inline-flex items-center gap-2 rounded-none bg-signal px-6 py-3 font-mono text-sm font-medium text-signal-foreground transition-colors duration-200 hover:bg-signal/90"
      >
        查看作品
        <ArrowUpRight
          class="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5 group-hover:-translate-y-0.5"
        />
      </a>
      <a
        href={GITHUB_URL}
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-2 rounded-none border border-border px-6 py-3 font-mono text-sm text-foreground transition-colors duration-200 hover:border-signal hover:text-signal"
      >
        GitHub
        <ArrowUpRight class="h-4 w-4" />
      </a>
    </div>

    <!-- 底部规格行 -->
    <div
      class="reveal mt-20 border-t border-border"
      style={`--reveal-delay: ${REVEAL_DELAY_MS.spec}ms`}
    >
      <div
        class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 py-5 font-mono text-xs text-muted-foreground"
      >
        <div class="flex items-center gap-6">
          <span>01 / 工程</span>
          <span>02 / 设计</span>
          <span>03 / 开源</span>
        </div>
        <span class="tracking-[0.18em] uppercase">STACK — TS · SVELTE · GO</span>
      </div>
    </div>
  </div>

  <!-- 滚动提示：竖向的等宽标注，替代常见的跳动箭头。 -->
  <div class="absolute bottom-8 left-5 z-10 hidden items-center gap-4 lg:flex">
    <span
      class="font-mono text-[10px] tracking-[0.32em] text-muted-foreground uppercase [writing-mode:vertical-rl]"
    >
      Scroll
    </span>
    <span class="h-14 w-px bg-border" aria-hidden="true"></span>
  </div>
</section>

<style>
/* 用边框色绘制坐标网格，随明暗主题自动切换。 */
.hero-grid {
  opacity: 0.4;
  -webkit-mask-image: radial-gradient(ellipse 90% 80% at 50% 40%, black 30%, transparent 75%);
  mask-image: radial-gradient(ellipse 90% 80% at 50% 40%, black 30%, transparent 75%);
  background-image:
    linear-gradient(to right, var(--border) 1px, transparent 1px),
    linear-gradient(to bottom, var(--border) 1px, transparent 1px);
  background-size: 56px 56px;
}

/* 首屏元素以阶梯方式依次浮现，使用自定义缓动曲线。 */
.reveal {
  opacity: 0;
  animation: hero-reveal 0.7s cubic-bezier(0.23, 1, 0.32, 1) forwards;
  animation-delay: var(--reveal-delay, 0ms);
}

@keyframes hero-reveal {
  from {
    transform: translateY(16px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* 尊重系统减少动态偏好，取消位移与渐隐。 */
@media (prefers-reduced-motion: reduce) {
  .reveal {
    opacity: 1;
    animation: none;
  }
}
</style>
