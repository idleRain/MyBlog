<script lang="ts">
import Header from '$lib/components/layout/Header.svelte'
import { ModeWatcher } from 'mode-watcher'
import { page } from '$app/state'
import { m } from '$i18n'
import '../app.css'

// 404 等加载与路由错误不经过渲染边界流程，错误页组件的 props 中不含状态码与错误对象，必须从 page 对象读取。
const status = $derived(page.status)
const error = $derived(page.error)

// 依据状态码区分文案，500 为服务端故障，其余视为目标资源缺失。
const isServerError = $derived(status === 500)
const errorTitle = $derived(
  isServerError ? m['ui:error.serverErrorTitle']() : m['ui:error.notFoundTitle']()
)
const errorSubtitle = $derived(
  isServerError ? m['ui:error.serverErrorSubtitle']() : m['ui:error.notFoundSubtitle']()
)
</script>

<svelte:head>
  <title>{status} - {m['ui:site.name']()}</title>
</svelte:head>

<!-- 主题监听器 -->
<ModeWatcher />

<!-- 保留导航栏 -->
<Header />

<section class="texture-grid relative flex min-h-screen items-center overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {m['ui:error.eyebrow']()} · {status}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {errorTitle}
    </h1>

    <!-- 说明引语：朱红竖线引出，保持编辑手记的口吻。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">{errorSubtitle}</p>
      {#if isServerError && error}
        <p class="mt-3 font-mono text-xs text-muted-foreground">{error.message}</p>
      {/if}
    </div>

    <!-- 操作区：500 提供重试与返回双入口，其余状态仅保留返回。 -->
    <div class="mt-12 flex flex-wrap items-center gap-4">
      {#if isServerError}
        <button
          type="button"
          onclick={() => window.location.reload()}
          class="inline-flex items-center gap-2 bg-primary px-7 py-3.5 text-sm font-bold text-primary-foreground transition-[background-color,color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97]"
        >
          {m['ui:error.retry']()}
        </button>
        <a
          href="/"
          class="inline-flex items-center gap-2 border border-border px-7 py-3.5 text-sm font-bold text-foreground transition-[border-color,color,transform] duration-200 ease-(--ease-out-strong) hover:border-signal hover:text-signal active:scale-[0.97]"
        >
          {m['ui:error.backHome']()}
        </a>
      {:else}
        <a
          href="/"
          class="group inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
        >
          {m['ui:error.backHome']()}
          <span
            class="transition-transform duration-200 ease-(--ease-out-strong) group-hover:-translate-x-1"
            aria-hidden="true"
          >
            ←
          </span>
        </a>
      {/if}
    </div>
  </div>

  <!-- 幽灵状态码：右下角的大号衬线数字，作为版面前景装饰。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    {status}
  </div>
</section>
