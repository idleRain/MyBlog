// 文章模块常量：状态配置、筛选选项与排序选项，供列表与编辑页共用。

import type { ArticleStatus } from '@myblog/api/modules/article/types'
import type { BadgeVariant } from '$ui/badge'

// 文章状态徽章配置，变体走 shadcn 语义 token，不硬编码颜色类。
export const ARTICLE_STATUS_CONFIG: Record<
  ArticleStatus,
  { label: string; variant: BadgeVariant }
> = {
  draft: { label: '草稿', variant: 'outline' },
  published: { label: '已发布', variant: 'default' },
  archived: { label: '已归档', variant: 'secondary' },
  private: { label: '私有', variant: 'secondary' }
}

// 文章状态筛选选项，按展示顺序排列。
export const ARTICLE_STATUS_OPTIONS: Array<{ value: ArticleStatus; label: string }> = [
  { value: 'draft', label: '草稿' },
  { value: 'published', label: '已发布' },
  { value: 'archived', label: '已归档' },
  { value: 'private', label: '私有' }
]

// 文章列表排序字段选项，与后端 GetArticleListRequest 的 oneof 约束一致。
export const ARTICLE_SORT_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'created_at', label: '创建时间' },
  { value: 'updated_at', label: '更新时间' },
  { value: 'published_at', label: '发布时间' },
  { value: 'view_count', label: '浏览量' },
  { value: 'like_count', label: '点赞数' }
]

// 排序方向选项
export const ARTICLE_ORDER_OPTIONS: Array<{ value: 'asc' | 'desc'; label: string }> = [
  { value: 'desc', label: '降序' },
  { value: 'asc', label: '升序' }
]

// 文章默认分页大小
export const ARTICLE_PAGE_SIZE = 10

// 文章状态流转操作类型
export type ArticleStatusAction = 'publish' | 'unpublish' | 'archive' | 'private'

// 各状态下可执行的状态操作，与后端状态机一致。
export const ARTICLE_ACTIONS: Record<ArticleStatus, ArticleStatusAction[]> = {
  draft: ['publish', 'private', 'archive'],
  published: ['unpublish', 'archive'],
  archived: ['publish', 'private'],
  private: ['publish', 'archive']
}

// 状态操作的中文标签
export const ARTICLE_ACTION_LABELS: Record<ArticleStatusAction, string> = {
  publish: '发布',
  unpublish: '取消发布',
  archive: '归档',
  private: '设为私有'
}
