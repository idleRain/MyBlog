<script lang="ts">
import type {
  CreateFriendlyLinkRequest,
  FriendlyLink,
  UpdateFriendlyLinkRequest
} from '@myblog/api/modules/friendlyLink/types'
import LinkFormDialog from '$lib/components/admin/link/link-form-dialog.svelte'
import { linkPageState, emptyLinkPageData } from './link-page-state.svelte'
import { LINK_PAGE_SIZE, LINK_STATUS_OPTIONS } from '$lib/constants/link'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import LinkTable from '$lib/components/admin/link/link-table.svelte'
import { Plus, RotateCcw, Link as LinkIcon } from '@lucide/svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Empty from '$ui/empty'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  linkPageState.initialize(data ?? emptyLinkPageData())
})

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<FriendlyLink | null>(null)
let deleteTarget = $state<FriendlyLink | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 打开友链弹窗，目标为空表示新建。
function openDialog(target: FriendlyLink | null) {
  dialogTarget = target
  isDialogOpen = true
}

// 提交友链表单，新建与编辑分别调用对应写操作。
// 弹窗载荷为动态键值结构，与既有表单弹窗约定一致地经类型断言收窄。
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = dialogTarget
    ? await linkPageState.updateLink({
        id: dialogTarget.id,
        ...payload
      } as UpdateFriendlyLinkRequest)
    : await linkPageState.createLink(payload as unknown as CreateFriendlyLinkRequest)
  isSubmitting = false
  if (success) isDialogOpen = false
}

// 确认后删除目标友链。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await linkPageState.removeLink(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>友情链接 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="友情链接" description="管理互链申请，覆盖申请、审核与展示状态" crumb="友情链接">
  {#snippet actions()}
    <Button onclick={() => openDialog(null)}>
      <Plus data-icon="inline-start" />
      添加友链
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：状态筛选即点即用，仅保留重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={linkPageState.statusFilter}
          onValueChange={value => linkPageState.setStatusFilter(value)}
        >
          {#each LINK_STATUS_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => linkPageState.resetAndReload()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if linkPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载友链" />
        </div>
      {:else if linkPageState.links.length === 0}
        <Empty.Root class="h-64 border-0">
          <Empty.Header>
            <Empty.Media variant="icon">
              <LinkIcon />
            </Empty.Media>
            <Empty.Title>
              {linkPageState.hasActiveFilters ? '没有匹配的友链' : '暂无友链'}
            </Empty.Title>
            <Empty.Description>
              {linkPageState.hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '添加第一条友情链接，与友邻站点互相推荐'}
            </Empty.Description>
          </Empty.Header>
          <Empty.Content>
            {#if linkPageState.hasActiveFilters}
              <Button variant="outline" size="sm" onclick={() => linkPageState.resetAndReload()}>
                <RotateCcw data-icon="inline-start" />
                重置筛选
              </Button>
            {:else}
              <Button size="sm" onclick={() => openDialog(null)}>
                <Plus data-icon="inline-start" />
                添加友链
              </Button>
            {/if}
          </Empty.Content>
        </Empty.Root>
      {:else}
        <LinkTable
          links={linkPageState.links}
          onEdit={link => openDialog(link)}
          onReview={(link, key) => void linkPageState.reviewLink(link, key)}
          onDelete={link => (deleteTarget = link)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {linkPageState.total} 个友链</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={linkPageState.total}
            perPage={LINK_PAGE_SIZE}
            page={linkPageState.currentPage}
            onPageChange={page => linkPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {linkPageState.currentPage} 页，共 {linkPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <LinkFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除友链"
    description={deleteTarget ? `确定删除「${deleteTarget.name}」吗？此操作不可恢复。` : ''}
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
