<script lang="ts">
import { Toaster } from '$ui/sonner/index.js'
import { ModeWatcher } from 'mode-watcher'
// 字体来源按构建模式注入：本地模式为空数组（CSS 副作用已注册），生产模式输出 CDN 样式表。
import { fontCssLinks } from '#fonts'
import '@/app.css'

let { children }: { children: import('svelte').Snippet } = $props()
</script>

<svelte:head>
  {#each fontCssLinks as href (href)}
    <link rel="stylesheet" {href} />
  {/each}
</svelte:head>

<ModeWatcher />
<Toaster position="top-center" />

{@render children()}
