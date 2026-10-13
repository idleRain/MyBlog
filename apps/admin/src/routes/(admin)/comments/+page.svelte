<script lang="ts">
import {
  COMMENT_PAGE_SIZE,
  COMMENT_STATUS_OPTIONS,
  type CommentAction
} from '$lib/constants/comment'
import { commentPageState, emptyCommentPageData } from './comment-page-state.svelte'
import CommentTable from '$lib/components/admin/comment/comment-table.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import type { AdminComment } from '@myblog/api/modules/comment/types'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { MessageSquare, RotateCcw, Search } from '@lucide/svelte'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
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
  commentPageState.initialize(data ?? emptyCommentPageData())
})

// 删除确认状态，deleteTarget 非空表示对话框已打开。
let deleteTarget = $state<AdminComment | null>(null)
let isDeleting = $state(false)

// 审核动作分发，永久删除需经确认对话框，其余动作立即提交。
function handleAction(comment: AdminComment, action: CommentAction) {
  if (action === 'delete') {
    deleteTarget = comment
    return
  }
  void commentPageState.runAction(comment, action)
}

// 确认后永久删除目标评论，失败时保留对话框以便重试。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await commentPageState.deleteComment(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>评论管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="评论管理" description="审核与管理文章评论，覆盖完整审核状态机" crumb="评论管理">
  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：搜索即输即搜，状态筛选即点即用，仅保留重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            class="pl-9"
            placeholder="搜索评论内容或评论者..."
            bind:value={commentPageState.keyword}
            oninput={commentPageState.applyKeywordDebounced}
          />
        </div>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={commentPageState.statusFilter}
          onValueChange={value => commentPageState.setStatusFilter(value)}
        >
          {#each COMMENT_STATUS_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => commentPageState.resetFilters()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if commentPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载评论" />
        </div>
      {:else if commentPageState.comments.length === 0}
        <Empty.Root class="h-64 border-0">
          <Empty.Header>
            <Empty.Media variant="icon">
              <MessageSquare />
            </Empty.Media>
            <Empty.Title>
              {commentPageState.hasActiveFilters ? '没有匹配的评论' : '暂无评论'}
            </Empty.Title>
            <Empty.Description>
              {commentPageState.hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '文章收到评论后会展示在这里'}
            </Empty.Description>
          </Empty.Header>
          {#if commentPageState.hasActiveFilters}
            <Empty.Content>
              <Button variant="outline" size="sm" onclick={() => commentPageState.resetFilters()}>
                <RotateCcw data-icon="inline-start" />
                重置筛选
              </Button>
            </Empty.Content>
          {/if}
        </Empty.Root>
      {:else}
        <CommentTable comments={commentPageState.comments} onAction={handleAction} />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {commentPageState.total} 条评论</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={commentPageState.total}
            perPage={COMMENT_PAGE_SIZE}
            page={commentPageState.currentPage}
            onPageChange={page => commentPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {commentPageState.currentPage} 页，共 {commentPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <ConfirmDialog
    title="删除评论"
    description={deleteTarget ? '确定永久删除这条评论吗？此操作不可恢复。' : ''}
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
