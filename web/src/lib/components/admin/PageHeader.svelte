<script lang="ts">
import { Breadcrumb, Sidebar } from '$ui'
import type { Snippet } from 'svelte'

// 面包屑条目：href 为空时渲染为当前页，即末级不可点击的 Page。
interface Crumb {
  label: string
  href?: string
}

interface Props {
  crumbs: Crumb[]
  /** 右侧工具区插槽：放置主题切换、主操作按钮等页级动作。 */
  trailing?: Snippet
}

let { crumbs, trailing }: Props = $props()
</script>

<!-- 页头：吸顶发丝线分隔，保持后台与前台一致的克制工程感。 -->
<header
  class="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur sm:px-6"
>
  <Sidebar.Trigger />
  <Breadcrumb.Root>
    <Breadcrumb.List>
      {#each crumbs as crumb, index (index)}
        {#if index > 0}
          <Breadcrumb.Separator />
        {/if}
        <Breadcrumb.Item>
          {#if crumb.href}
            <Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
          {:else}
            <Breadcrumb.Page>{crumb.label}</Breadcrumb.Page>
          {/if}
        </Breadcrumb.Item>
      {/each}
    </Breadcrumb.List>
  </Breadcrumb.Root>
  {#if trailing}
    <div class="ml-auto flex items-center gap-2">
      {@render trailing()}
    </div>
  {/if}
</header>
