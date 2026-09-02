import type { ApiResponse } from '@myblog/shared'
import type { User } from '@myblog/api/modules/user/types'
import type { Category } from '@myblog/api/modules/category/types'
import type { Tag } from '@myblog/api/modules/tag/types'

// 文章状态枚举，与后端 model.ArticleStatus 一致。
export type ArticleStatus = 'draft' | 'published' | 'archived' | 'private'

// 文章来源类型枚举
export type ArticleOriginType = 'original' | 'translation' | 'reprint'

// 文章信息接口，字段与后端 model.Article 的 JSON tag 一致。
export interface Article {
  id: number
  title: string
  slug: string
  summary: string
  content: string
  contentHtml: string
  coverImage: string
  authorId: number
  categoryId: number | null
  status: ArticleStatus
  originType: ArticleOriginType
  sourceUrl: string
  sourceAuthor: string
  isFeatured: boolean
  isTop: boolean
  commentEnabled: boolean
  viewCount: number
  likeCount: number
  bookmarkCount: number
  commentCount: number
  wordCount: number
  readingTime: number
  version: number
  seoTitle: string
  seoDescription: string
  seoKeywords: string
  scheduledAt: string | null
  publishedAt: string | null
  editedAt: string | null
  archivedAt: string | null
  lastCommentAt: string | null
  createdAt: string
  updatedAt: string
  author: User
  category: Category | null
  categories: Category[]
  tags: Tag[]
}

// 文章列表数据，后端分页响应不返回 pages 字段。
export interface ArticleListData {
  page: number
  pageSize: number
  total: number
  articles: Article[]
}

// 文章列表查询参数，排序字段与顺序均受后端 oneof 约束。
export interface GetArticleListRequest {
  page?: number
  pageSize?: number
  status?: ArticleStatus | ''
  authorId?: number
  sortBy?: 'created_at' | 'updated_at' | 'published_at' | 'view_count' | 'like_count' | ''
  order?: 'asc' | 'desc' | ''
  search?: string
}

// 创建文章请求参数，status 仅允许 draft / published / private。
export interface CreateArticleRequest {
  title: string
  slug?: string
  summary?: string
  content: string
  coverImage?: string
  categoryId?: number | null
  categoryIds?: number[]
  tagIds?: number[]
  status?: 'draft' | 'published' | 'private'
  isFeatured?: boolean | null
  isTop?: boolean | null
  commentEnabled?: boolean | null
  seoTitle?: string
  seoDescription?: string
  seoKeywords?: string
}

// 更新文章请求参数，可选字段显式传入才更新。
export interface UpdateArticleRequest {
  id: number
  title: string
  slug?: string | null
  summary?: string | null
  content: string
  coverImage?: string | null
  categoryId?: number | null
  categoryIds?: number[]
  tagIds?: number[]
  status?: ArticleStatus
  isFeatured?: boolean | null
  isTop?: boolean | null
  commentEnabled?: boolean | null
  seoTitle?: string | null
  seoDescription?: string | null
  seoKeywords?: string | null
}

// 文章操作类接口的消息响应
export interface ArticleActionData {
  message: string
}

// 各响应类型
export type ArticleResponse = ApiResponse<Article>
export type ArticleListResponse = ApiResponse<ArticleListData>
export type ArticleActionResponse = ApiResponse<ArticleActionData>
