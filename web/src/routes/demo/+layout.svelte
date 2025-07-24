<script lang="ts">
import '../../app.css'
import { Header } from '$lib/components/layout'

const { children } = $props()
</script>

<Header />

<h1 class="pt-[100px] text-center text-2xl font-bold">i18n 模块化演示</h1>

<main class="min-h-screen">
  {@render children()}
</main>
