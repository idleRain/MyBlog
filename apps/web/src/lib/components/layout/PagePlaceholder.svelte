<script lang="ts">
import { SITE_NAME_ZH } from '@myblog/shared'

// 展位页属性：eyebrow 为栏目标注，ghostMark 为右下角幽灵字，title 与 description 构成版面主体并写入文档头。
interface Props {
  eyebrow: string
  ghostMark: string
  title: string
  description: string
}

let { eyebrow, ghostMark, title, description }: Props = $props()
</script>

<svelte:head>
  <title>{title} - {SITE_NAME_ZH}</title>
  <meta name="description" content={description} />
</svelte:head>

<section class="texture-grid relative flex min-h-screen items-center overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与首页刊头的 eyebrow 同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {eyebrow}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-5xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {title}
    </h1>

    <!-- 展位状态行：等宽标注说明当前版面尚未刊出。 -->
    <p class="mt-6 font-mono text-xs text-muted-foreground">版面排期中 · 功能将在后续迭代接入</p>

    <!-- 简介引语：朱红竖线引出，延续编辑手记的引出方式。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">{description}</p>
    </div>

    <!-- 返回入口：朱红粗链接，箭头随悬停轻移。 -->
    <a
      href="/"
      class="group mt-12 inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
    >
      返回首页
      <span
        class="transition-transform duration-200 ease-(--ease-out-strong) group-hover:-translate-x-1"
        aria-hidden="true"
      >
        ←
      </span>
    </a>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    {ghostMark}
  </div>
</section>
