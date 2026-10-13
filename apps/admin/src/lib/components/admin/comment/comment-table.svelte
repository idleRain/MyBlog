<script lang="ts">
import {
  COMMENT_ACTION_LABELS,
  COMMENT_ACTIONS,
  COMMENT_STATUS_CONFIG,
  type CommentAction
} from '$lib/constants/comment'
import { ArchiveX, Ban, CircleCheck, CircleX, Trash2, MoreHorizontal } from '@lucide/svelte'
import type { AdminComment } from '@myblog/api/modules/comment/types'
import * as DropdownMenu from '$ui/dropdown-menu'
import type { LucideIcon } from '@lucide/svelte'
import { Button } from '$ui/button'
import * as Table from '$ui/table'
import { Badge } from '$ui/badge'

interface Props {
  comments: AdminComment[]
  onAction: (comment: AdminComment, action: CommentAction) => void
}

let { comments, onAction }: Props = $props()

// 审核动作的图标映射，与 COMMENT_ACTION_LABELS 一一对应。
// 图标只随本组件引入，避免图标桶经常量模块进入页面状态模块与单测的编译图。
const COMMENT_ACTION_ICONS: Record<CommentAction, LucideIcon> = {
  approve: CircleCheck,
  reject: CircleX,
  spam: Ban,
  trash: ArchiveX,
  delete: Trash2
}

// 游客评论与登录用户评论共用列表，两种形态都缺少展示名称时按匿名呈现。
const ANONYMOUS_AUTHOR_NAME = '匿名'

// 提取评论者展示名称，登录用户优先使用昵称，游客回退到提交时填写的姓名。
function getAuthorName(comment: AdminComment): string {
  return (
    comment.user?.nickname || comment.user?.username || comment.authorName || ANONYMOUS_AUTHOR_NAME
  )
}

// 提取归属文章标题，后端未装配文章归属时回退到文章编号。
function getArticleTitle(comment: AdminComment): string {
  return comment.article?.title ?? `#${comment.articleId}`
}

// 评论时间按管理端统一口径只展示日期。
function formatCreatedAt(comment: AdminComment): string {
  return new Date(comment.createdAt).toLocaleDateString('zh-CN')
}
</script>

<!-- 评论表格：所属文章与评论时间在窄屏折叠，评论内容、评论者与状态保持可见 -->
<Table.Root>
  <Table.Header>
    <Table.Row>
      <Table.Head class="min-w-64">评论内容</Table.Head>
      <Table.Head>评论者</Table.Head>
      <Table.Head class="hidden md:table-cell">所属文章</Table.Head>
      <Table.Head>状态</Table.Head>
      <Table.Head class="hidden md:table-cell">评论时间</Table.Head>
      <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {#each comments as comment (comment.id)}
      <Table.Row>
        <Table.Cell>
          <div class="flex flex-col gap-1">
            <p class="line-clamp-2 max-w-64 text-sm">{comment.content}</p>
            <!-- 所属文章与评论时间在窄屏并入评论内容列副行，md 及以上恢复独立列 -->
            <span
              class="flex max-w-64 flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground md:hidden"
            >
              <span class="line-clamp-1 max-w-40">{getArticleTitle(comment)}</span>
              <span aria-hidden="true">·</span>
              <span class="tabular-nums">{formatCreatedAt(comment)}</span>
            </span>
          </div>
        </Table.Cell>
        <Table.Cell>
          <span class="text-sm">{getAuthorName(comment)}</span>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="line-clamp-1 max-w-40 text-sm text-muted-foreground">
            {getArticleTitle(comment)}
          </span>
        </Table.Cell>
        <Table.Cell>
          <Badge variant={COMMENT_STATUS_CONFIG[comment.status].variant}>
            {COMMENT_STATUS_CONFIG[comment.status].label}
          </Badge>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="text-sm text-muted-foreground tabular-nums">
            {formatCreatedAt(comment)}
          </span>
        </Table.Cell>
        <Table.Cell>
          <div class="flex items-center justify-center gap-1">
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`评论操作 ${getAuthorName(comment)}`}
                >
                  <MoreHorizontal />
                </Button>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="end">
                {#each COMMENT_ACTIONS[comment.status] as action (action)}
                  {@const ActionIcon = COMMENT_ACTION_ICONS[action]}
                  <DropdownMenu.Item
                    variant={action === 'delete' ? 'destructive' : 'default'}
                    onSelect={() => onAction(comment, action)}
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
