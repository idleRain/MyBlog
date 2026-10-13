<script lang="ts">
import type {
  Category,
  CreateCategoryRequest,
  UpdateCategoryRequest
} from '@myblog/api/modules/category/types'
import CategoryFormDialog from '$lib/components/admin/category/category-form-dialog.svelte'
import { categoryPageState, emptyCategoryPageData } from './category-page-state.svelte'
import CategoryTable from '$lib/components/admin/category/category-table.svelte'
import { CATEGORY_STATUS_FILTER_OPTIONS } from '$lib/constants/category'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import { Plus, RotateCcw, Search, FolderTree } from '@lucide/svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Empty from '$ui/empty'
import { Input } from '$ui/input'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  categoryPageState.initialize(data ?? emptyCategoryPageData())
})

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<Category | null>(null)
let deleteTarget = $state<Category | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 打开分类弹窗，目标为空表示新建。
function openDialog(target: Category | null) {
  dialogTarget = target
  isDialogOpen = true
}

// 提交分类表单，新建与编辑分别调用对应写操作。
// 弹窗载荷为动态键值结构，与既有表单弹窗约定一致地经类型断言收窄。
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = dialogTarget
    ? await categoryPageState.updateCategory({
        id: dialogTarget.id,
        ...payload
      } as UpdateCategoryRequest)
    : await categoryPageState.createCategory(payload as unknown as CreateCategoryRequest)
  isSubmitting = false
  if (success) isDialogOpen = false
}

// 确认后删除目标分类，存在子分类时由后端拒绝并回传原因。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await categoryPageState.removeCategory(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>分类管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title="分类管理"
  description="树形管理文章分类，支持层级结构与显示状态控制"
  crumb="分类管理"
>
  {#snippet actions()}
    <Button onclick={() => openDialog(null)}>
      <Plus data-icon="inline-start" />
      新建分类
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：搜索与状态筛选均在客户端完成，分类树一次拉全量，无需服务端分页 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            class="pl-9"
            placeholder="搜索分类名称或 URL 标识..."
            bind:value={categoryPageState.searchQuery}
          />
        </div>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          bind:value={categoryPageState.statusFilter}
        >
          {#each CATEGORY_STATUS_FILTER_OPTIONS as option (option.value)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => categoryPageState.resetFilters()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if categoryPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载分类" />
        </div>
      {:else if categoryPageState.filteredRows.length === 0}
        <Empty.Root class="h-64 border-0">
          <Empty.Header>
            <Empty.Media variant="icon" class="size-12 rounded-xl border bg-muted/50">
              <FolderTree class="size-6 text-muted-foreground" />
            </Empty.Media>
            <Empty.Title>
              {categoryPageState.hasCategories ? '没有匹配的分类' : '暂无分类'}
            </Empty.Title>
            <Empty.Description>
              {categoryPageState.hasCategories
                ? '调整或重置筛选条件后再试'
                : '创建第一个分类来组织文章'}
            </Empty.Description>
          </Empty.Header>
          <Empty.Content>
            {#if categoryPageState.hasCategories}
              <Button variant="outline" size="sm" onclick={() => categoryPageState.resetFilters()}>
                <RotateCcw data-icon="inline-start" />
                重置筛选
              </Button>
            {:else}
              <Button size="sm" onclick={() => openDialog(null)}>
                <Plus data-icon="inline-start" />
                新建分类
              </Button>
            {/if}
          </Empty.Content>
        </Empty.Root>
      {:else}
        <CategoryTable
          rows={categoryPageState.filteredRows}
          onCreateChild={() => openDialog(null)}
          onEdit={category => openDialog(category)}
          onDelete={category => (deleteTarget = category)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {categoryPageState.total} 个分类</p>
          <p class="text-sm text-muted-foreground">
            当前显示 {categoryPageState.filteredRows.length} 个
          </p>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <CategoryFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    categories={categoryPageState.flatCategories}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除分类"
    description={deleteTarget ? `确定删除「${deleteTarget.name}」吗？存在子分类时无法删除。` : ''}
    confirmText="删除"
    destructive
    isLoading={isDeleting}
    open={deleteTarget !== null}
    onOpenChange={open => {
      if (!open && !isDeleting) deleteTarget = null
    }}
    onConfirm={handleDelete}
  />
</PageHeader>
