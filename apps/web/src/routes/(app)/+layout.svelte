<script lang="ts">
import Header from '$lib/components/layout/Header.svelte'
import Footer from '$lib/components/layout/Footer.svelte'

let { children } = $props()
</script>

<!-- 全局导航栏 -->
<Header />

<!-- 主内容区域 -->
<main class="min-h-screen">
  {@render children()}
</main>

<!-- 全局底部 -->
<Footer />
