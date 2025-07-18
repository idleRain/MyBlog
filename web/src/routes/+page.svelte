<script lang="ts">
import type { PageProps } from './$types'
import { Button } from '$lib/components/ui/button'

let { data }: PageProps = $props()
</script>

<svelte:head>
  <title>主页</title>
  <meta name="description" content="主页" />
</svelte:head>

<section>
  <Button>123</Button>
  <div class="mx-auto">
    <h1 class="flex flex-col text-red-500">数据1：</h1>
    <div>{JSON.stringify(data.data)}</div>
    <div>{1 + 1}</div>
  </div>
</section>
