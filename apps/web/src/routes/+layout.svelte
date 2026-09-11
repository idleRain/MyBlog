<script lang="ts">
import Header from '$lib/components/layout/Header.svelte'
import Footer from '$lib/components/layout/Footer.svelte'
import { Toaster } from '$ui/sonner/index.js'
import { ModeWatcher } from 'mode-watcher'
import type { LayoutData } from './$types'
import type { Snippet } from 'svelte'
import '@/app.css'

let { data, children }: { data: LayoutData; children: Snippet } = $props()
</script>

<ModeWatcher />
<Toaster position="top-center" />

<!-- 全局导航栏 -->
<Header friendlyLinks={data.friendlyLinks} />

<!-- 主内容区域 -->
<main class="min-h-screen">
  {@render children()}
</main>

<!-- 全局底部 -->
<Footer categories={data.categories} />
