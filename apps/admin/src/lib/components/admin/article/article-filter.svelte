<script lang="ts">
import {
  ARTICLE_ORDER_OPTIONS,
  ARTICLE_SORT_OPTIONS,
  ARTICLE_STATUS_OPTIONS
} from '$lib/constants/article'
import type { ArticleStatus } from '@myblog/api/modules/article/types'
import { Search, RotateCcw } from '@lucide/svelte'
import { Button, Card, Input, Select } from '$ui'

interface Props {
  search?: string
  status?: ArticleStatus | ''
  sortBy?: string
  order?: 'asc' | 'desc'
  onSearch: () => void
  onReset: () => void
}

let {
  search = $bindable(''),
  status = $bindable<ArticleStatus | ''>(''),
  sortBy = $bindable(''),
  order = $bindable<'asc' | 'desc'>('desc'),
  onSearch,
  onReset
}: Props = $props()
</script>

<Card.Root>
  <Card.Content class="p-4">
    <div class="flex flex-wrap items-end gap-3">
      <!-- 关键词搜索 -->
      <div class="relative min-w-52 flex-1">
        <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input.Root
          class="pl-9"
          placeholder="搜索标题、摘要或正文..."
          bind:value={search}
          onkeydown={(event: KeyboardEvent) => {
            if (event.key === 'Enter') onSearch()
          }}
        />
      </div>

      <!-- 状态筛选 -->
      <Select.Root type="single" bind:value={status}>
        <Select.Trigger class="w-32">
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
      <Select.Root type="single" bind:value={sortBy}>
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
      <Select.Root type="single" bind:value={order}>
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

      <div class="flex gap-2">
        <Button onclick={onSearch}>
          <Search data-icon="inline-start" />
          搜索
        </Button>
        <Button variant="outline" onclick={onReset}>
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>
    </div>
  </Card.Content>
</Card.Root>
