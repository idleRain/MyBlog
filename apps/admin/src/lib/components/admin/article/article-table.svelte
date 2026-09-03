<script lang="ts">
import {
  ARTICLE_ACTION_LABELS,
  ARTICLE_ACTIONS,
  type ArticleStatusAction
} from '$lib/constants/article'
import { Edit, Eye, MessageSquare, MoreHorizontal, Trash2, FileText } from '@lucide/svelte'
import ArticleStatusBadge from '$lib/components/admin/article/article-status-badge.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import type { Article } from '@myblog/api/modules/article/types'
import { Button, Card, DropdownMenu, Table } from '$ui'
import { goto } from '$lib/utils/navigation'

interface Props {
  articles: Article[]
  isLoading: boolean
  onStatusAction: (article: Article, action: ArticleStatusAction) => void
  onDelete: (article: Article) => void
}

let { articles, isLoading, onStatusAction, onDelete }: Props = $props()

// 待删除文章，用于确认对话框；为空时隐藏对话框。
let deleteTarget = $state<Article | null>(null)
let isDeleting = $state(false)

async function confirmDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    await onDelete(deleteTarget)
  } finally {
    isDeleting = false
    deleteTarget = null
  }
}

/**
 * 将时间字符串转为本地化展示，空值返回占位符。
 */
function formatDate(value: string | null): string {
  if (!value) return '—'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<Card.Root>
  <Card.Content class="p-0">
    {#if isLoading}
      <div class="flex h-48 items-center justify-center">
        <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
        ></span>
      </div>
    {:else if articles.length === 0}
      <div class="flex h-48 items-center justify-center">
        <div class="text-center">
          <FileText class="mx-auto size-12 text-muted-foreground" />
          <h3 class="mt-4 text-lg font-medium">暂无文章</h3>
          <p class="text-sm text-muted-foreground">调整筛选条件或创建第一篇博客文章</p>
        </div>
      </div>
    {:else}
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head class="min-w-64">标题</Table.Head>
            <Table.Head>作者</Table.Head>
            <Table.Head>分类</Table.Head>
            <Table.Head>状态</Table.Head>
            <Table.Head>统计</Table.Head>
            <Table.Head>发布时间</Table.Head>
            <Table.Head class="text-right">操作</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each articles as article (article.id)}
            <Table.Row>
              <Table.Cell>
                <div class="max-w-64 space-y-1">
                  <p class="truncate font-medium">{article.title}</p>
                  {#if article.summary}
                    <p class="truncate text-sm text-muted-foreground">{article.summary}</p>
                  {/if}
                </div>
              </Table.Cell>
              <Table.Cell>
                <span class="text-sm"
                  >{article.author?.nickname || article.author?.username || '未知'}</span
                >
              </Table.Cell>
              <Table.Cell>
                <span class="text-sm text-muted-foreground">
                  {article.category?.name || '—'}
                </span>
              </Table.Cell>
              <Table.Cell>
                <ArticleStatusBadge status={article.status} />
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center gap-3 text-sm text-muted-foreground">
                  <span class="inline-flex items-center gap-1">
                    <Eye class="size-3.5" />{article.viewCount}
                  </span>
                  <span class="inline-flex items-center gap-1">
                    <MessageSquare class="size-3.5" />{article.commentCount}
                  </span>
                </div>
              </Table.Cell>
              <Table.Cell>
                <span class="text-sm text-muted-foreground">{formatDate(article.publishedAt)}</span>
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center justify-end gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    aria-label="编辑文章"
                    onclick={() => goto(`/posts/${article.id}`)}
                  >
                    <Edit data-icon="inline-start" />
                    编辑
                  </Button>

                  <DropdownMenu.Root>
                    <DropdownMenu.Trigger>
                      <Button variant="ghost" size="sm" aria-label="更多操作">
                        <MoreHorizontal />
                      </Button>
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content align="end">
                      {#each ARTICLE_ACTIONS[article.status] as action (action)}
                        <DropdownMenu.Item onselect={() => onStatusAction(article, action)}>
                          {ARTICLE_ACTION_LABELS[action]}
                        </DropdownMenu.Item>
                      {/each}
                      <DropdownMenu.Item
                        variant="destructive"
                        onselect={() => (deleteTarget = article)}
                      >
                        <Trash2 data-icon="inline-start" />
                        删除
                      </DropdownMenu.Item>
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

<!-- 删除确认对话框 -->
<ConfirmDialog
  title="删除文章"
  description={deleteTarget ? `确定删除「${deleteTarget.title}」吗？此操作不可恢复。` : ''}
  confirmText="删除"
  destructive
  isLoading={isDeleting}
  open={deleteTarget !== null}
  onOpenChange={open => {
    if (!open && !isDeleting) deleteTarget = null
  }}
  onConfirm={confirmDelete}
/>
