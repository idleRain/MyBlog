// 通知模块常量：类型标签映射与筛选选项。

import type { NotificationType } from '@myblog/api/modules/notification/types'

// 通知类型中文标签映射
export const NOTIFICATION_TYPE_LABELS: Record<NotificationType, string> = {
  comment_reply: '评论回复',
  article_like: '文章点赞',
  comment_like: '评论点赞',
  system: '系统通知',
  follow: '用户关注',
  article_new: '新文章'
}

// 通知类型筛选选项，'' 表示全部。
export const NOTIFICATION_TYPE_OPTIONS: Array<{ value: NotificationType | ''; label: string }> = [
  { value: '', label: '全部' },
  { value: 'comment_reply', label: '评论回复' },
  { value: 'article_like', label: '文章点赞' },
  { value: 'comment_like', label: '评论点赞' },
  { value: 'system', label: '系统通知' },
  { value: 'follow', label: '用户关注' },
  { value: 'article_new', label: '新文章' }
]

// 通知默认分页大小
export const NOTIFICATION_PAGE_SIZE = 10
