<script lang="ts">
import { resolveTagStatusVariant, resolveTagStatusLabel } from '$lib/utils/tag-display'
import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import type { Tag } from '@myblog/api/modules/tag/types'
import { Flame, Pencil, Trash2 } from '@lucide/svelte'
import { Button } from '$ui/button'
import * as Table from '$ui/table'
import { Badge } from '$ui/badge'

interface Props {
  tags: Tag[]
  statusDict: EnabledDictGroup | null
  onEdit: (tag: Tag) => void
  onDelete: (tag: Tag) => void
}

let { tags, statusDict, onEdit, onDelete }: Props = $props()
</script>

<!-- 标签表格：URL 标识与创建时间在窄屏折叠，标签名与使用次数保持可见 -->
<Table.Root>
  <Table.Header>
    <Table.Row>
      <Table.Head>标签</Table.Head>
      <Table.Head class="hidden md:table-cell">URL 标识</Table.Head>
      <Table.Head>使用次数</Table.Head>
      <Table.Head>状态</Table.Head>
      <Table.Head class="hidden md:table-cell">创建时间</Table.Head>
      <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {#each tags as tag (tag.id)}
      <Table.Row>
        <Table.Cell>
          <div class="flex flex-col gap-1">
            <span
              class="inline-flex w-fit items-center gap-1.5 rounded-md border px-2 py-0.5 text-sm font-medium"
              style="border-color: color-mix(in oklab, {tag.color} 35%, transparent); background-color: color-mix(in oklab, {tag.color} 10%, transparent);"
            >
              <span
                class="size-2 rounded-full"
                style="background-color: {tag.color}"
                aria-hidden="true"
              ></span>
              {tag.name}
              {#if tag.isHot}
                <Flame class="size-3.5 fill-primary text-primary" aria-hidden="true" />
                <span class="sr-only">热门</span>
              {/if}
            </span>
            {#if tag.description}
              <p class="max-w-72 truncate text-xs text-muted-foreground">{tag.description}</p>
            {/if}
            <!-- URL 标识在窄屏并入标签列副行，md 及以上恢复独立列 -->
            <code
              class="w-fit rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground md:hidden"
            >
              {tag.slug}
            </code>
          </div>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <code class="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
            {tag.slug}
          </code>
        </Table.Cell>
        <Table.Cell>
          <span
            class="text-sm tabular-nums {tag.usageCount > 0
              ? 'text-foreground'
              : 'text-muted-foreground'}"
          >
            {tag.usageCount}
          </span>
        </Table.Cell>
        <Table.Cell>
          <Badge variant={resolveTagStatusVariant(statusDict, tag.status)}>
            {resolveTagStatusLabel(statusDict, tag.status)}
          </Badge>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="text-sm text-muted-foreground tabular-nums">
            {new Date(tag.createdAt).toLocaleDateString('zh-CN')}
          </span>
        </Table.Cell>
        <Table.Cell>
          <div class="flex items-center justify-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              aria-label={`编辑标签 ${tag.name}`}
              onclick={() => onEdit(tag)}
            >
              <Pencil />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              class="text-destructive hover:text-destructive"
              aria-label={`删除标签 ${tag.name}`}
              onclick={() => onDelete(tag)}
            >
              <Trash2 />
            </Button>
          </div>
        </Table.Cell>
      </Table.Row>
    {/each}
  </Table.Body>
</Table.Root>
