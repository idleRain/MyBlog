<script lang="ts">
import type { PageProps } from './$types'

let { data }: PageProps = $props()
</script>

<svelte:head>
  <title>主页</title>
  <meta name="description" content="主页" />
</svelte:head>

<section>
  <div class="mx-auto">
    <h1 class="flex flex-col text-red-500">数据：</h1>
    <div>{JSON.stringify(data.data)}</div>
    <div>{1 + 1}</div>
  </div>
</section>
