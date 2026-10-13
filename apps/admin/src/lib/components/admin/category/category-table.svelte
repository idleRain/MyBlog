<script lang="ts">
import {
  CATEGORY_ROW_BASE_PADDING_REM,
  CATEGORY_ROW_INDENT_PER_LEVEL_REM,
  CATEGORY_STATUS_CONFIG,
  type CategoryRow
} from '$lib/constants/category'
import { FolderPlus, MoreHorizontal, Pencil, Trash2 } from '@lucide/svelte'
import type { Category } from '@myblog/api/modules/category/types'
import * as DropdownMenu from '$ui/dropdown-menu'
import { Button } from '$ui/button'
import { Badge } from '$ui/badge'

interface Props {
  rows: CategoryRow[]
  onCreateChild: (category: Category) => void
  onEdit: (category: Category) => void
  onDelete: (category: Category) => void
}

let { rows, onCreateChild, onEdit, onDelete }: Props = $props()
</script>

<!-- 分类列表：层级缩进呈现树形归属，URL 标识在窄屏折叠 -->
<div class="divide-y">
  {#each rows as row (row.category.id)}
    {@const statusConfig = CATEGORY_STATUS_CONFIG[row.category.status]!}
    <div
      class="flex flex-wrap items-center gap-4 px-4 py-3"
      style="padding-left: {CATEGORY_ROW_BASE_PADDING_REM +
        row.depth * CATEGORY_ROW_INDENT_PER_LEVEL_REM}rem"
    >
      <div class="flex min-w-0 flex-1 items-center gap-3">
        {#if row.depth > 0}
          <span class="text-muted-foreground">└</span>
        {/if}
        <!-- 名称允许截断收缩，slug 在窄屏隐藏，避免行内内容撑出可视区 -->
        <span class="min-w-0 truncate font-medium">{row.category.name}</span>
        <span class="hidden font-mono text-xs text-muted-foreground md:inline"
          >/{row.category.slug}</span
        >
        <Badge variant={statusConfig.variant}>{statusConfig.label}</Badge>
        {#if row.category.isFeatured}
          <Badge variant="outline">精选</Badge>
        {/if}
      </div>
      <span class="shrink-0 text-sm text-muted-foreground tabular-nums">
        {row.category.articleCount} 篇文章
      </span>

      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          <Button variant="ghost" size="icon" aria-label={`分类操作 ${row.category.name}`}>
            <MoreHorizontal />
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end">
          <DropdownMenu.Item onSelect={() => onCreateChild(row.category)}>
            <FolderPlus data-icon="inline-start" />
            新建子分类
          </DropdownMenu.Item>
          <DropdownMenu.Item onSelect={() => onEdit(row.category)}>
            <Pencil data-icon="inline-start" />
            编辑
          </DropdownMenu.Item>
          <DropdownMenu.Item variant="destructive" onSelect={() => onDelete(row.category)}>
            <Trash2 data-icon="inline-start" />
            删除
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  {/each}
</div>
