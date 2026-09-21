<script lang="ts">
import {
  COMMENT_ACTION_ICONS,
  COMMENT_ACTION_LABELS,
  COMMENT_ACTIONS,
  COMMENT_PAGE_SIZE,
  COMMENT_STATUS_CONFIG,
  COMMENT_STATUS_OPTIONS,
  type CommentAction
} from '$lib/constants/comment'
import type {
  Comment,
  CommentStatus,
  CommentActionResponse
} from '@myblog/api/modules/comment/types'
import { Badge, Button, Card, DropdownMenu, Input, Pagination, Table, ToggleGroup } from '$ui'
import { Search, MessageSquare, MoreHorizontal, RotateCcw } from '@lucide/svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { SITE_NAME_ZH, debounce } from '@myblog/shared'
import { CommentAPI } from '$lib/api'
import { onMount } from 'svelte'

let comments = $state<Comment[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let keyword = $state('')
let statusFilter = $state<CommentStatus | ''>('')

// 审核动作到接口方法的映射，spam 对应 markSpam 方法名。
const ACTION_METHODS: Record<
  Exclude<CommentAction, 'delete'>,
  (id: number) => Promise<CommentActionResponse>
> = {
  approve: CommentAPI.approve,
  reject: CommentAPI.reject,
  spam: CommentAPI.markSpam,
  trash: CommentAPI.trash
}

let deleteTarget = $state<Comment | null>(null)
let isDeleting = $state(false)

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 是否存在生效中的筛选条件，用于空态区分「无数据」与「无匹配」。
const hasActiveFilters = $derived(keyword.trim() !== '' || statusFilter !== '')

/**
 * 加载评论列表，携带状态与关键词筛选。
 */
async function loadComments() {
  isLoading = true
  try {
    const response = await CommentAPI.adminList({
      page: currentPage,
      pageSize: COMMENT_PAGE_SIZE,
      ...(statusFilter ? { status: statusFilter } : {}),
      ...(keyword.trim() ? { keyword: keyword.trim() } : {})
    })
    if (response.code === 200 && response.data) {
      comments = response.data.comments ?? []
      total = response.data.total ?? 0
    } else {
      toast.error(response.message || '加载评论失败')
    }
  } catch (error) {
    console.error('加载评论失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 防抖应用搜索关键词，连续输入只在停顿后请求一次。
 */
const applyKeywordDebounced = debounce(() => {
  currentPage = 1
  loadComments()
}, SEARCH_DEBOUNCE_MS)

/**
 * 状态筛选变更后立即回到第一页并重新加载。
 */
function handleStatusFilterChange(value: string) {
  statusFilter = value as CommentStatus | ''
  currentPage = 1
  loadComments()
}

/**
 * 重置筛选并回到第一页。
 */
function resetAndReload() {
  keyword = ''
  statusFilter = ''
  currentPage = 1
  loadComments()
}

/**
 * 执行评论审核动作，删除操作走确认对话框。
 */
async function handleAction(comment: Comment, action: CommentAction) {
  if (action === 'delete') {
    deleteTarget = comment
    return
  }
  try {
    const response = await ACTION_METHODS[action](comment.id)
    if (response.code === 200) {
      toast.success(`${COMMENT_ACTION_LABELS[action]}成功`)
      loadComments()
    } else {
      toast.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('评论操作失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

/**
 * 永久删除评论（回收站内）。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await CommentAPI.delete(deleteTarget.id)
    if (response.code === 200) {
      toast.success('评论删除成功')
      deleteTarget = null
      loadComments()
    } else {
      toast.error(response.message || '删除评论失败')
    }
  } catch (error) {
    console.error('删除评论失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

/**
 * 提取评论者展示名称，注册用户优先使用昵称。
 */
function getAuthorName(comment: Comment): string {
  return comment.user?.nickname || comment.user?.username || comment.authorName || '匿名'
}

function handlePageChange(page: number) {
  currentPage = page
  loadComments()
}

onMount(loadComments)
</script>

<svelte:head>
  <title>评论管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="评论管理" description="审核与管理文章评论，覆盖完整审核状态机" crumb="评论管理">
  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索评论内容或评论者..."
            bind:value={keyword}
            oninput={applyKeywordDebounced}
          />
        </div>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={statusFilter}
          onValueChange={handleStatusFilterChange}
        >
          {#each COMMENT_STATUS_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button variant="ghost" size="sm" class="ml-auto" onclick={resetAndReload}>
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if comments.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <MessageSquare class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {hasActiveFilters ? '没有匹配的评论' : '暂无评论'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {hasActiveFilters ? '调整或重置筛选条件后再试' : '文章收到评论后会展示在这里'}
            </p>
          </div>
          {#if hasActiveFilters}
            <Button variant="outline" size="sm" onclick={resetAndReload}>
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {/if}
        </div>
      {:else}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head class="min-w-64">评论内容</Table.Head>
              <Table.Head>评论者</Table.Head>
              <Table.Head>所属文章</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head>评论时间</Table.Head>
              <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each comments as comment (comment.id)}
              <Table.Row>
                <Table.Cell>
                  <p class="line-clamp-2 max-w-64 text-sm">{comment.content}</p>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm">{getAuthorName(comment)}</span>
                </Table.Cell>
                <Table.Cell>
                  <span class="line-clamp-1 max-w-40 text-sm text-muted-foreground">
                    {comment.article?.title ?? `#${comment.articleId}`}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <Badge variant={COMMENT_STATUS_CONFIG[comment.status].variant}>
                    {COMMENT_STATUS_CONFIG[comment.status].label}
                  </Badge>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground tabular-nums">
                    {new Date(comment.createdAt).toLocaleString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-center gap-1">
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <Button variant="ghost" size="sm" aria-label="评论操作">
                          <MoreHorizontal />
                        </Button>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        {#each COMMENT_ACTIONS[comment.status] as action (action)}
                          {@const ActionIcon = COMMENT_ACTION_ICONS[action]}
                          <DropdownMenu.Item
                            variant={action === 'delete' ? 'destructive' : 'default'}
                            onSelect={() => handleAction(comment, action)}
                          >
                            <ActionIcon data-icon="inline-start" />
                            {COMMENT_ACTION_LABELS[action]}
                          </DropdownMenu.Item>
                        {/each}
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  </div>
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {total} 条评论</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={total}
            perPage={COMMENT_PAGE_SIZE}
            page={currentPage}
            onPageChange={handlePageChange}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / COMMENT_PAGE_SIZE))} 页
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
