<script lang="ts">
import { toAdminPath } from '$lib/utils/navigation'
import { Breadcrumb, Sidebar } from '$ui'

interface Props {
  title: string
  description?: string
  crumb: string
  actions?: import('svelte').Snippet
  children: import('svelte').Snippet
}

let { title, description, crumb, actions, children }: Props = $props()
</script>

<!-- 后台列表页统一头部：面包屑 + 标题 + 右侧操作区 -->
<header class="flex h-16 shrink-0 items-center gap-2 border-b px-6">
  <Sidebar.Trigger />
  <Breadcrumb.Root>
    <Breadcrumb.List>
      <Breadcrumb.Item>
        <Breadcrumb.Link href={toAdminPath('/')}>管理后台</Breadcrumb.Link>
      </Breadcrumb.Item>
      <Breadcrumb.Separator />
      <Breadcrumb.Item>
        <Breadcrumb.Page>{crumb}</Breadcrumb.Page>
      </Breadcrumb.Item>
    </Breadcrumb.List>
  </Breadcrumb.Root>
</header>

<main class="flex-1 space-y-6 p-6">
  <div class="flex flex-wrap items-center justify-between gap-4">
    <div class="space-y-1">
      <h1 class="text-3xl font-bold tracking-tight">{title}</h1>
      {#if description}
        <p class="text-muted-foreground">{description}</p>
      {/if}
    </div>
    {#if actions}
      {@render actions()}
    {/if}
  </div>

  {@render children()}
</main>
