<script lang="ts">
import type { CreateTagRequest, Tag, UpdateTagRequest } from '@myblog/api/modules/tag/types'
import TagFormDialog from '$lib/components/admin/tag/tag-form-dialog.svelte'
import { Plus, RotateCcw, Search, Tags as TagsIcon } from '@lucide/svelte'
import { tagPageState, emptyTagPageData } from './tag-page-state.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import TagTable from '$lib/components/admin/tag/tag-table.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { TAG_PAGE_SIZE } from '$lib/constants/tag'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import { Input } from '$ui/input'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  tagPageState.initialize(data ?? emptyTagPageData())
})

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<Tag | null>(null)
let deleteTarget = $state<Tag | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 打开标签弹窗，目标为空表示新建。
function openDialog(target: Tag | null) {
  dialogTarget = target
  isDialogOpen = true
}

// 提交标签表单，新建与编辑分别调用对应写操作。
// 弹窗载荷为动态键值结构，与既有表单弹窗约定一致地经类型断言收窄。
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = dialogTarget
    ? await tagPageState.updateTag({ id: dialogTarget.id, ...payload } as UpdateTagRequest)
    : await tagPageState.createTag(payload as unknown as CreateTagRequest)
  isSubmitting = false
  if (success) isDialogOpen = false
}

// 确认后删除目标标签，后端负责处理已挂载文章的引用。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await tagPageState.removeTag(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>标签管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title="标签管理"
  description="管理文章标签，支持热门标记与显示状态控制"
  crumb="标签管理"
>
  {#snippet actions()}
    <Button onclick={() => openDialog(null)}>
      <Plus data-icon="inline-start" />
      新建标签
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：搜索即输即搜，筛选即点即用，仅保留重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-56 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            class="pl-9"
            placeholder="搜索标签名称或描述..."
            bind:value={tagPageState.search}
            oninput={tagPageState.applySearchDebounced}
          />
        </div>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={tagPageState.statusFilter}
          onValueChange={value => tagPageState.setStatusFilter(value)}
        >
          <ToggleGroup.Item value="">全部</ToggleGroup.Item>
          {#each tagPageState.statusFilterOptions as option (option.value)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={tagPageState.hotFilter}
          onValueChange={value => tagPageState.setHotFilter(value)}
        >
          <ToggleGroup.Item value="">全部</ToggleGroup.Item>
          <ToggleGroup.Item value="1">仅热门</ToggleGroup.Item>
        </ToggleGroup.Root>

        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => tagPageState.resetFilters()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if tagPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载标签" />
        </div>
      {:else if tagPageState.tags.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <TagsIcon class="size-6 text-muted-foreground" />
          </div>
          <div class="flex flex-col gap-1">
            <h3 class="text-base font-medium">
              {tagPageState.hasActiveFilters ? '没有匹配的标签' : '暂无标签'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {tagPageState.hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '创建第一个标签，开始组织文章内容'}
            </p>
          </div>
          {#if tagPageState.hasActiveFilters}
            <Button variant="outline" size="sm" onclick={() => tagPageState.resetFilters()}>
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {:else}
            <Button size="sm" onclick={() => openDialog(null)}>
              <Plus data-icon="inline-start" />
              新建标签
            </Button>
          {/if}
        </div>
      {:else}
        <TagTable
          tags={tagPageState.tags}
          statusDict={tagPageState.statusDict}
          onEdit={tag => openDialog(tag)}
          onDelete={tag => (deleteTarget = tag)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {tagPageState.total} 个标签</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={tagPageState.total}
            perPage={TAG_PAGE_SIZE}
            page={tagPageState.currentPage}
            onPageChange={page => tagPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {tagPageState.currentPage} 页，共 {tagPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <TagFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除标签"
    description={deleteTarget ? `确定删除标签「${deleteTarget.name}」吗？` : ''}
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
