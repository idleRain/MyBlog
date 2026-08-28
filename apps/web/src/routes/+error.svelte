<script lang="ts">
import Header from '$lib/components/layout/Header.svelte'
import { ModeWatcher } from 'mode-watcher'
import { Button } from '$ui'
import '../app.css'

let { error, status }: { error: App.Error; status: number } = $props()

// 依据状态码区分文案，500 为服务端故障，其余视为目标资源缺失。
const isServerError = $derived(status === 500)
const errorTitle = $derived(isServerError ? '实验出现意外结果' : '探索进入未知领域')
const errorSubtitle = $derived(
  isServerError ? '我们的服务器正在经历技术性阵痛' : '你寻找的页面已消失在数字宇宙中'
)
</script>

<svelte:head>
  <title>{status} - MyBlog</title>
</svelte:head>

<!-- 主题监听器 -->
<ModeWatcher />

<!-- 保留导航栏 -->
<Header />

<section
  class="relative flex min-h-screen items-center overflow-hidden bg-background text-foreground"
>
  <!-- 坐标网格背景：与首屏一致的工程图纸质感。 -->
  <div class="error-grid absolute inset-0" aria-hidden="true"></div>

  <!-- 四角定位标记：强化规格书的裁切线质感。 -->
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

  <div class="relative z-10 mx-auto w-full max-w-2xl px-6 pt-20 pb-16 text-center sm:px-10">
    <!-- 顶部等宽标注行，以注释前缀呼应规格书语言。 -->
    <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
      <span class="text-signal">//</span> Error - {status}
    </p>

    <!-- 大号等宽状态码作为页面的视觉锚点。 -->
    <h1 class="mt-6 font-mono text-8xl font-bold tracking-tight sm:text-9xl">
      {status}
    </h1>

    <!-- signal 短发丝线：强调色以标记形式出现。 -->
    <div class="mx-auto mt-8 h-0.5 w-16 bg-signal" aria-hidden="true"></div>

    <!-- 文案区 -->
    <div class="mt-8 space-y-3">
      <h2 class="text-2xl font-bold text-foreground sm:text-3xl">{errorTitle}</h2>
      <p class="text-base leading-relaxed text-muted-foreground">{errorSubtitle}</p>
      {#if isServerError}
        <p class="text-sm text-muted-foreground">维修团队正在紧急处理中...</p>
      {:else}
        <p class="text-sm text-muted-foreground">带我回家 -></p>
      {/if}
      {#if isServerError && error.message}
        <p class="font-mono text-xs text-muted-foreground">{error.message}</p>
      {/if}
    </div>

    <!-- 操作按钮 -->
    <div class="mt-12 flex flex-wrap items-center justify-center gap-4">
      {#if isServerError}
        <Button
          onclick={() => window.location.reload()}
          class="rounded-none bg-signal px-6 py-3 font-mono text-sm text-signal-foreground transition-colors duration-200 hover:bg-signal/90"
        >
          重试实验
        </Button>
        <Button
          href="/"
          variant="outline"
          class="rounded-none border-border px-6 py-3 font-mono text-sm transition-colors duration-200 hover:border-signal hover:text-signal"
        >
          返回首页
        </Button>
      {:else}
        <Button
          href="/"
          class="rounded-none bg-signal px-8 py-4 font-mono text-sm font-medium text-signal-foreground hover:bg-signal/90!"
        >
          返回安全基地
        </Button>
      {/if}
    </div>
  </div>
</section>

<style>
/* 网格单元边长与首屏 HeroSection 保持一致，形成全站统一的图纸质感。 */
.error-grid {
  opacity: 0.4;
  -webkit-mask-image: radial-gradient(ellipse 90% 80% at 50% 40%, black 30%, transparent 75%);
  mask-image: radial-gradient(ellipse 90% 80% at 50% 40%, black 30%, transparent 75%);
  background-image:
    linear-gradient(to right, var(--border) 1px, transparent 1px),
    linear-gradient(to bottom, var(--border) 1px, transparent 1px);
  background-size: 56px 56px;
}
</style>
