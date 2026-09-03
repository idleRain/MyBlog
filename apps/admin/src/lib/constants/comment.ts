// 评论模块常量：状态配置、筛选选项与审核动作映射。

import type { CommentStatus } from '@myblog/api/modules/comment/types'
import type { BadgeVariant } from '$ui/badge'

// 评论审核状态配置
export const COMMENT_STATUS_CONFIG: Record<
  CommentStatus,
  { label: string; variant: BadgeVariant }
> = {
  pending: { label: '待审核', variant: 'secondary' },
  approved: { label: '已通过', variant: 'default' },
  rejected: { label: '已拒绝', variant: 'outline' },
  spam: { label: '垃圾', variant: 'destructive' },
  trash: { label: '回收站', variant: 'outline' }
}

// 评论状态筛选选项，'' 表示全部。
export const COMMENT_STATUS_OPTIONS: Array<{ value: CommentStatus | ''; label: string }> = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待审核' },
  { value: 'approved', label: '已通过' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'spam', label: '垃圾' },
  { value: 'trash', label: '回收站' }
]

// 评论审核动作类型
export type CommentAction = 'approve' | 'reject' | 'spam' | 'trash' | 'delete'

// 各状态下可执行的审核动作
export const COMMENT_ACTIONS: Record<CommentStatus, CommentAction[]> = {
  pending: ['approve', 'reject', 'spam', 'trash'],
  approved: ['reject', 'spam', 'trash'],
  rejected: ['approve', 'spam', 'trash'],
  spam: ['approve', 'reject', 'trash'],
  trash: ['delete']
}

// 审核动作的中文标签
export const COMMENT_ACTION_LABELS: Record<CommentAction, string> = {
  approve: '通过',
  reject: '拒绝',
  spam: '标记垃圾',
  trash: '移入回收站',
  delete: '删除'
}

// 评论默认分页大小
export const COMMENT_PAGE_SIZE = 10
