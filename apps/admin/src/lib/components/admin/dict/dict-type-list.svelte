<script lang="ts">
import { BookText, ListTree, Pencil, Plus, Trash2 } from '@lucide/svelte'
import type { DictType } from '@myblog/api/modules/dict/types'
import { dictStatusOf } from '$lib/constants/dict'
import { Spinner } from '$ui/spinner'
import { cn } from '@myblog/shared'
import { Button } from '$ui/button'
import * as Empty from '$ui/empty'
import { Badge } from '$ui/badge'
import * as Card from '$ui/card'

interface Props {
  types: DictType[]
  selectedTypeId: number | null
  isLoading: boolean
  onSelect: (id: number) => void
  onEdit: (type: DictType) => void
  onDelete: (type: DictType) => void
  onCreate: () => void
}

let { types, selectedTypeId, isLoading, onSelect, onEdit, onDelete, onCreate }: Props = $props()
</script>

<!-- 字典类型列表：选择区与行内操作区为并列的可聚焦元素，窄屏下点击区域互不重叠 -->
<Card.Root class="lg:col-span-1">
  <Card.Header class="pb-3">
    <Card.Title class="flex items-center gap-2 text-base">
      <ListTree class="size-4" />
      字典类型
    </Card.Title>
  </Card.Header>
  <Card.Content class="p-0">
    {#if isLoading}
      <div class="flex h-64 items-center justify-center">
        <Spinner class="size-6" aria-label="正在加载字典类型" />
      </div>
    {:else if types.length === 0}
      <Empty.Root class="h-64 border-0">
        <Empty.Header>
          <Empty.Media variant="icon">
            <BookText />
          </Empty.Media>
          <Empty.Title>暂无字典类型</Empty.Title>
          <Empty.Description>创建第一个字典类型开始使用</Empty.Description>
        </Empty.Header>
        <Empty.Content>
          <Button size="sm" onclick={onCreate}>
            <Plus data-icon="inline-start" />
            新建字典类型
          </Button>
        </Empty.Content>
      </Empty.Root>
    {:else}
      <ul class="max-h-[32rem] divide-y overflow-y-auto">
        {#each types as type (type.id)}
          {@const isSelected = type.id === selectedTypeId}
          {@const status = dictStatusOf(type.status)}
          <li class={cn('transition-colors', isSelected ? 'bg-muted' : 'hover:bg-muted/60')}>
            <div class="flex items-stretch gap-1 pr-3">
              <button
                type="button"
                class="flex min-w-0 flex-1 flex-col gap-1 px-4 py-3 text-left outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                aria-current={isSelected ? 'true' : undefined}
                onclick={() => onSelect(type.id)}
              >
                <span class="truncate font-medium">{type.name}</span>
                <span class="flex min-w-0 items-center gap-2">
                  <span class="truncate font-mono text-xs text-muted-foreground">{type.code}</span>
                  <Badge variant={status.variant} class="shrink-0">{status.label}</Badge>
                </span>
              </button>
              <div class="flex shrink-0 items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`编辑字典类型 ${type.name}`}
                  onclick={() => onEdit(type)}
                >
                  <Pencil />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  class="text-destructive hover:text-destructive"
                  aria-label={`删除字典类型 ${type.name}`}
                  onclick={() => onDelete(type)}
                >
                  <Trash2 />
                </Button>
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </Card.Content>
</Card.Root>
