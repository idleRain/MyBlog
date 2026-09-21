<script lang="ts">
import {
  ARTICLE_ORDER_OPTIONS,
  ARTICLE_SORT_OPTIONS,
  ARTICLE_STATUS_OPTIONS
} from '$lib/constants/article'
import type { ArticleStatus } from '@myblog/api/modules/article/types'
import { Search, RotateCcw } from '@lucide/svelte'
import { Button, Input, Select } from '$ui'

interface Props {
  search?: string
  status?: ArticleStatus | ''
  sortBy?: string
  order?: 'asc' | 'desc'
  // 编辑者被服务端强制只显示已发布文章，状态筛选无意义时禁用。
  statusDisabled?: boolean
  // 搜索输入防抖回调，由父组件包装防抖逻辑。
  onSearch: () => void
  // 任一下拉筛选变化时的立即应用回调。
  onFilterChange: () => void
  onReset: () => void
}

let {
  search = $bindable(''),
  status = $bindable<ArticleStatus | ''>(''),
  sortBy = $bindable(''),
  order = $bindable<'asc' | 'desc'>('desc'),
  statusDisabled = false,
  onSearch,
  onFilterChange,
  onReset
}: Props = $props()
</script>

<div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
  <!-- 关键词搜索 -->
  <div class="relative min-w-52 flex-1">
    <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
    <Input.Root
      class="pl-9"
      placeholder="搜索标题、摘要或正文..."
      bind:value={search}
      oninput={onSearch}
    />
  </div>

  <!-- 状态筛选 -->
  <Select.Root
    type="single"
    value={status}
    onValueChange={(value: string) => {
      status = value as ArticleStatus | ''
      onFilterChange()
    }}
    disabled={statusDisabled}
  >
    <Select.Trigger
      class="w-32"
      title={statusDisabled ? '当前角色仅可查看已发布文章，状态筛选已禁用' : undefined}
    >
      {status === ''
        ? '全部状态'
        : (ARTICLE_STATUS_OPTIONS.find(option => option.value === status)?.label ?? '全部状态')}
    </Select.Trigger>
    <Select.Content>
      <Select.Group>
        <Select.Item value="">全部状态</Select.Item>
        {#each ARTICLE_STATUS_OPTIONS as option (option.value)}
          <Select.Item value={option.value}>{option.label}</Select.Item>
        {/each}
      </Select.Group>
    </Select.Content>
  </Select.Root>

  <!-- 排序字段 -->
  <Select.Root
    type="single"
    value={sortBy}
    onValueChange={(value: string) => {
      sortBy = value
      onFilterChange()
    }}
  >
    <Select.Trigger class="w-32">
      {ARTICLE_SORT_OPTIONS.find(option => option.value === sortBy)?.label ?? '排序字段'}
    </Select.Trigger>
    <Select.Content>
      <Select.Group>
        {#each ARTICLE_SORT_OPTIONS as option (option.value)}
          <Select.Item value={option.value}>{option.label}</Select.Item>
        {/each}
      </Select.Group>
    </Select.Content>
  </Select.Root>

  <!-- 排序方向 -->
  <Select.Root
    type="single"
    value={order}
    onValueChange={(value: string) => {
      order = value as 'asc' | 'desc'
      onFilterChange()
    }}
  >
    <Select.Trigger class="w-24">
      {ARTICLE_ORDER_OPTIONS.find(option => option.value === order)?.label ?? '降序'}
    </Select.Trigger>
    <Select.Content>
      <Select.Group>
        {#each ARTICLE_ORDER_OPTIONS as option (option.value)}
          <Select.Item value={option.value}>{option.label}</Select.Item>
        {/each}
      </Select.Group>
    </Select.Content>
  </Select.Root>

  <Button variant="ghost" size="sm" class="ml-auto" onclick={onReset}>
    <RotateCcw data-icon="inline-start" />
    重置
  </Button>
</div>
