<script lang="ts">
import { Toaster } from '$lib/components/ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import Header from '$lib/components/layout/Header.svelte'
import Footer from '$lib/components/layout/Footer.svelte'
import '@/app.css'

let { children } = $props()
</script>

<ModeWatcher />
<Toaster position="top-center" />

<!-- 全局导航栏 -->
<Header />

<!-- 主内容区域 -->
<main class="min-h-screen">
  {@render children()}
</main>

<!-- 全局底部 -->
<Footer />
