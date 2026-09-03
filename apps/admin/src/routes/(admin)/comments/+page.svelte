<script lang="ts">
import {
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
import { Badge, Button, Card, DropdownMenu, Input, Table, ToggleGroup } from '$ui'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import { Search, MessageSquare, MoreHorizontal } from '@lucide/svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import Pagination from '$lib/components/admin/pagination.svelte'
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
  <title>评论管理 - MyBlog</title>
</svelte:head>

<PageHeader title="评论管理" description="审核与管理文章评论，覆盖完整审核状态机" crumb="评论管理">
  <Card.Root>
    <Card.Content class="p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索评论内容或评论者..."
            bind:value={keyword}
            onkeydown={(event: KeyboardEvent) => {
              if (event.key === 'Enter') {
                currentPage = 1
                loadComments()
              }
            }}
          />
        </div>

        <ToggleGroup.Root type="single" bind:value={statusFilter} variant="outline" size="sm">
          {#each COMMENT_STATUS_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button
          onclick={() => {
            currentPage = 1
            loadComments()
          }}
        >
          <Search data-icon="inline-start" />
          搜索
        </Button>
      </div>
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Content class="p-0">
      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if comments.length === 0}
        <div class="flex h-48 items-center justify-center">
          <div class="text-center">
            <MessageSquare class="mx-auto size-12 text-muted-foreground" />
            <h3 class="mt-4 text-lg font-medium">暂无评论</h3>
            <p class="text-sm text-muted-foreground">调整筛选条件后再试</p>
          </div>
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
              <Table.Head class="text-right">操作</Table.Head>
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
                  <span class="text-sm text-muted-foreground">
                    {new Date(comment.createdAt).toLocaleString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-end">
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <Button variant="ghost" size="sm" aria-label="评论操作">
                          <MoreHorizontal />
                        </Button>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        {#each COMMENT_ACTIONS[comment.status] as action (action)}
                          <DropdownMenu.Item
                            variant={action === 'delete' ? 'destructive' : 'default'}
                            onselect={() => handleAction(comment, action)}
                          >
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
      {/if}
    </Card.Content>
  </Card.Root>

  <Pagination
    page={currentPage}
    {total}
    pageSize={COMMENT_PAGE_SIZE}
    onPageChange={handlePageChange}
  />

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
