<script lang="ts">
import type { DictItem, DictType } from '@myblog/api/modules/dict/types'
import { ListTree, Pencil, Plus, Trash2 } from '@lucide/svelte'
import { dictStatusOf } from '$lib/constants/dict'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Table from '$ui/table'
import * as Empty from '$ui/empty'
import { Badge } from '$ui/badge'
import * as Card from '$ui/card'

interface Props {
  type: DictType | null
  items: DictItem[]
  isLoading: boolean
  onCreate: () => void
  onEdit: (item: DictItem) => void
  onDelete: (item: DictItem) => void
}

let { type, items, isLoading, onCreate, onEdit, onDelete }: Props = $props()
</script>

<!-- 字典项表格：卡片头部在窄屏换行，描述与排序列按断点折叠，主要字段保持可见 -->
<Card.Root class="lg:col-span-2">
  <Card.Header class="flex flex-wrap items-center justify-between gap-2 pb-3">
    <Card.Title class="flex min-w-0 items-center gap-2 text-base">
      <span class="truncate">字典项</span>
      {#if type}
        <span class="truncate font-mono text-xs text-muted-foreground">{type.code}</span>
      {/if}
    </Card.Title>
    <Button size="sm" onclick={onCreate} disabled={!type}>
      <Plus data-icon="inline-start" />
      新建字典项
    </Button>
  </Card.Header>
  <Card.Content class="p-0">
    {#if !type}
      <div class="flex h-64 items-center justify-center">
        <p class="text-sm text-muted-foreground">请先选择字典类型</p>
      </div>
    {:else if isLoading}
      <div class="flex h-64 items-center justify-center">
        <Spinner class="size-6" aria-label="正在加载字典项" />
      </div>
    {:else if items.length === 0}
      <Empty.Root class="h-64 border-0">
        <Empty.Header>
          <Empty.Media variant="icon">
            <ListTree />
          </Empty.Media>
          <Empty.Title>暂无字典项</Empty.Title>
          <Empty.Description>为当前字典类型添加可选值</Empty.Description>
        </Empty.Header>
        <Empty.Content>
          <Button size="sm" onclick={onCreate}>
            <Plus data-icon="inline-start" />
            新建字典项
          </Button>
        </Empty.Content>
      </Empty.Root>
    {:else}
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>值</Table.Head>
            <Table.Head>显示名</Table.Head>
            <Table.Head class="hidden md:table-cell">描述</Table.Head>
            <Table.Head class="hidden sm:table-cell">排序</Table.Head>
            <Table.Head>状态</Table.Head>
            <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each items as item (item.id)}
            {@const status = dictStatusOf(item.status)}
            <Table.Row>
              <Table.Cell>
                <code
                  class="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground"
                >
                  {item.value}
                </code>
              </Table.Cell>
              <Table.Cell>
                <span class="font-medium">{item.label}</span>
                <!-- 描述在窄屏并入显示名列副行，md 及以上恢复独立列 -->
                {#if item.description}
                  <span class="block text-sm text-muted-foreground md:hidden">
                    {item.description}
                  </span>
                {/if}
              </Table.Cell>
              <Table.Cell class="hidden md:table-cell">
                <span class="text-sm text-muted-foreground">{item.description || '—'}</span>
              </Table.Cell>
              <Table.Cell class="hidden sm:table-cell">
                <span class="text-sm tabular-nums">{item.sortOrder}</span>
              </Table.Cell>
              <Table.Cell>
                <Badge variant={status.variant}>{status.label}</Badge>
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center justify-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`编辑字典项 ${item.label}`}
                    onclick={() => onEdit(item)}
                  >
                    <Pencil />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="text-destructive hover:text-destructive"
                    aria-label={`删除字典项 ${item.label}`}
                    onclick={() => onDelete(item)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    {/if}
  </Card.Content>
</Card.Root>
