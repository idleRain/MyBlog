<script lang="ts">
import { ModeWatcher } from 'mode-watcher'
import { Button } from '$ui'
import '@/app.css'

let { error, status }: { error: App.Error; status: number } = $props()

// 依据状态码区分文案，500 为服务端故障，其余视为目标资源缺失。
const isServerError = $derived(status === 500)
const errorTitle = $derived(isServerError ? '服务暂时不可用' : '页面不存在或已被移除')
const errorHint = $derived(isServerError ? '请稍后重试或联系系统管理员' : '请检查地址是否输入正确')
</script>

<svelte:head>
  <title>{status} - MyBlog 管理后台</title>
</svelte:head>

<!-- 主题监听器 -->
<ModeWatcher />

<section class="flex min-h-screen items-center justify-center bg-background text-foreground">
  <div class="mx-auto w-full max-w-md px-6 text-center">
    <!-- 等宽标注行，与后台整体视觉语言保持一致。 -->
    <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
      Error / {status}
    </p>

    <!-- 大号状态码作为页面的视觉锚点。 -->
    <h1 class="mt-6 font-mono text-7xl font-bold tracking-tight">{status}</h1>

    <h2 class="mt-6 text-xl font-semibold">{errorTitle}</h2>
    <p class="mt-2 text-sm text-muted-foreground">{errorHint}</p>

    {#if isServerError && error.message}
      <p class="mt-4 font-mono text-xs text-muted-foreground">{error.message}</p>
    {/if}

    <Button href="/manage" class="mt-8">返回管理后台</Button>
  </div>
</section>
