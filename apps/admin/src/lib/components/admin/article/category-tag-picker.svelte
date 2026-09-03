<script lang="ts">
import type { Category, CategoryTreeNode } from '@myblog/api/modules/category/types'
import { Checkbox, Label, ScrollArea, Select, Skeleton } from '$ui'
import type { Tag } from '@myblog/api/modules/tag/types'
import { CategoryAPI, TagAPI } from '$lib/api'
import { onMount } from 'svelte'

// 文章的分类与标签选择结果，主分类、多分类与多标签三者独立维护。
export interface CategoryTagSelection {
  categoryId: number | null
  categoryIds: number[]
  tagIds: number[]
}

interface Props {
  value: CategoryTagSelection
}

let { value = $bindable<CategoryTagSelection>() }: Props = $props()

let categories = $state<Category[]>([])
let tags = $state<Tag[]>([])
let isLoading = $state(true)

// 将分类树扁平化，按层级缩进展示名称。
function flattenTree(nodes: CategoryTreeNode[], depth = 0): Category[] {
  return nodes.flatMap(node => [
    { ...node, name: `${'　'.repeat(depth)}${node.name}` },
    ...flattenTree(node.children ?? [], depth + 1)
  ])
}

/**
 * 并行加载分类树与标签列表，任一失败时按空数组处理。
 */
async function loadOptions() {
  const [categoryResult, tagResult] = await Promise.all([
    CategoryAPI.getTree().catch(() => null),
    TagAPI.adminList({ page: 1, pageSize: 100, search: '' }).catch(() => null)
  ])
  categories = categoryResult?.data?.tree ? flattenTree(categoryResult.data.tree) : []
  tags = tagResult?.data?.tags ?? []
  isLoading = false
}

function toggleCategory(id: number, checked: boolean) {
  const categoryIds = checked
    ? [...value.categoryIds, id]
    : value.categoryIds.filter(item => item !== id)
  value = { ...value, categoryIds }
}

function toggleTag(id: number, checked: boolean) {
  const tagIds = checked ? [...value.tagIds, id] : value.tagIds.filter(item => item !== id)
  value = { ...value, tagIds }
}

function handlePrimaryCategoryChange(raw: string) {
  const categoryId = raw === '' ? null : Number(raw)
  value = { ...value, categoryId }
}

onMount(loadOptions)
</script>

<div class="grid gap-6 md:grid-cols-3">
  <!-- 主分类 -->
  <div class="space-y-2">
    <Label.Root>主分类</Label.Root>
    {#if isLoading}
      <Skeleton.Skeleton class="h-8 w-full" />
    {:else}
      <Select.Root
        type="single"
        value={value.categoryId === null ? '' : String(value.categoryId)}
        onValueChange={handlePrimaryCategoryChange}
      >
        <Select.Trigger class="w-full">
          {categories.find(item => item.id === value.categoryId)?.name.trim() ?? '无主分类'}
        </Select.Trigger>
        <Select.Content>
          <Select.Group>
            <Select.Item value="">无主分类</Select.Item>
            {#each categories as category (category.id)}
              <Select.Item value={String(category.id)}>{category.name.trim()}</Select.Item>
            {/each}
          </Select.Group>
        </Select.Content>
      </Select.Root>
    {/if}
  </div>

  <!-- 关联分类多选 -->
  <div class="space-y-2">
    <Label.Root>关联分类</Label.Root>
    {#if isLoading}
      <Skeleton.Skeleton class="h-24 w-full" />
    {:else}
      <ScrollArea.Root class="h-24 rounded-md border">
        <div class="grid gap-1 p-2">
          {#each categories as category (category.id)}
            <label class="flex cursor-pointer items-center gap-2 text-sm">
              <Checkbox.Root
                checked={value.categoryIds.includes(category.id)}
                onCheckedChange={(checked: boolean) => toggleCategory(category.id, checked)}
              />
              <span class="truncate">{category.name.trim()}</span>
            </label>
          {/each}
          {#if categories.length === 0}
            <p class="p-2 text-sm text-muted-foreground">暂无分类</p>
          {/if}
        </div>
      </ScrollArea.Root>
    {/if}
  </div>

  <!-- 标签多选 -->
  <div class="space-y-2">
    <Label.Root>标签</Label.Root>
    {#if isLoading}
      <Skeleton.Skeleton class="h-24 w-full" />
    {:else}
      <ScrollArea.Root class="h-24 rounded-md border">
        <div class="grid gap-1 p-2">
          {#each tags as tag (tag.id)}
            <label class="flex cursor-pointer items-center gap-2 text-sm">
              <Checkbox.Root
                checked={value.tagIds.includes(tag.id)}
                onCheckedChange={(checked: boolean) => toggleTag(tag.id, checked)}
              />
              <span class="truncate">{tag.name}</span>
            </label>
          {/each}
          {#if tags.length === 0}
            <p class="p-2 text-sm text-muted-foreground">暂无标签</p>
          {/if}
        </div>
      </ScrollArea.Root>
    {/if}
  </div>
</div>
