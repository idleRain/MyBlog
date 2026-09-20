import type { Category } from '@myblog/api/modules/category/types'
import type { AuthorPublic } from '@myblog/api/modules/user/types'
import type { Tag } from '@myblog/api/modules/tag/types'
import type { ApiResponse } from '@myblog/shared'

// 文章状态枚举，与后端 model.ArticleStatus 一致。
export type ArticleStatus = 'draft' | 'published' | 'archived' | 'private'

// 文章来源类型枚举
export type ArticleOriginType = 'original' | 'translation' | 'reprint'

// 文章单语言翻译行，与后端 model.ArticleTranslation 的 JSON tag 一致，
// 仅管理端全量包请求携带，公共请求按语言输出后不包含该结构。
export interface ArticleTranslation {
  id: number
  articleId: number
  locale: string
  title: string | null
  summary: string | null
  content: string | null
  contentHtml: string | null
  wordCount: number
  seoTitle: string | null
  seoDescription: string | null
  seoKeywords: string | null
  createdAt: string
  updatedAt: string
}

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
  author: AuthorPublic
  category: Category | null
  categories: Category[]
  tags: Tag[]
  // 已填写有效翻译内容的语言列表，轻量翻译状态标记，不含翻译内容。
  translationLocales?: string[]
  // 全量翻译行数组，仅 Accept-Language: * 的管理端请求返回。
  translations?: ArticleTranslation[]
}

// 文章列表数据，后端分页响应不返回 pages 字段。
export interface ArticleListData {
  page: number
  pageSize: number
  total: number
  articles: Article[]
}

// 文章集合响应数据，热门与最新等聚合接口仅返回文章数组。
export interface ArticleCollectionData {
  articles: Article[]
}

// 归档的月份分组，articles 为该月已发布文章，按发布时间倒序。
export interface ArticleArchiveMonth {
  month: number
  articles: Article[]
}

// 归档的年份分组，total 为该年文章总数。
export interface ArticleArchiveYear {
  year: number
  total: number
  months: ArticleArchiveMonth[]
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

// 文章单语言翻译补丁，字段可选，提供即更新，未提供保留既有翻译值。
export interface ArticleI18nPatch {
  title?: string
  summary?: string
  content?: string
  seoTitle?: string
  seoDescription?: string
  seoKeywords?: string
}

// 按语言组织的翻译字段包，键为语言标识，缺省语言走主字段不允许出现。
export type ArticleI18nPayload = Record<string, ArticleI18nPatch>

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
  i18n?: ArticleI18nPayload
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
  i18n?: ArticleI18nPayload
}

// 文章操作类接口的消息响应
export interface ArticleActionData {
  message: string
}

// 点赞状态查询响应数据
export interface ArticleLikeStateData {
  isLiked: boolean
}

// 收藏状态查询响应数据
export interface ArticleBookmarkStateData {
  isBookmarked: boolean
}

// 各响应类型
export type ArticleResponse = ApiResponse<Article>
export type ArticleListResponse = ApiResponse<ArticleListData>
export type ArticleCollectionResponse = ApiResponse<ArticleCollectionData>
export type ArticleArchiveResponse = ApiResponse<ArticleArchiveYear[]>
export type ArticleLikeStateResponse = ApiResponse<ArticleLikeStateData>
export type ArticleBookmarkStateResponse = ApiResponse<ArticleBookmarkStateData>
export type ArticleActionResponse = ApiResponse<ArticleActionData>
