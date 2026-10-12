import type { AuthorPublic } from '@myblog/api/modules/user/types'
import type { Article } from '@myblog/api/modules/article/types'
import type { ApiResponse } from '@myblog/shared'

// 评论审核状态枚举，与后端 model.CommentStatus 一致。
export type CommentStatus = 'pending' | 'approved' | 'rejected' | 'spam' | 'trash'

// 评论信息接口，字段与后端 model.Comment 的公开 JSON 契约一致；
// 游客邮箱、IP 与 UserAgent 属审计字段，文章归属属管理端字段，公开响应均不含。
export interface Comment {
  id: number
  articleId: number
  userId: number | null
  parentId: number | null
  rootId: number | null
  level: number
  authorName: string
  authorWebsite: string
  content: string
  contentHtml: string
  status: CommentStatus
  likeCount: number
  replyCount: number
  reportedCount: number
  isAuthor: boolean
  isPinned: boolean
  editedAt: string | null
  createdAt: string
  updatedAt: string
  // 评论者公开信息仅注册用户评论携带，游客评论在后端带 omitempty 整体缺省。
  user?: AuthorPublic | null
  // 回复关系仅在装配评论树时输出，公开列表为扁平结构故整体缺省。
  parent?: Comment | null
  root?: Comment | null
  children?: Comment[]
}

// 评论归属文章引用，管理端列表用于展示归属。
// 后端仅预加载文章本体，不含作者与分类标签等关联，故以窄化视图描述而非完整 Article。
export type CommentArticleRef = Omit<Article, 'author' | 'category' | 'categories' | 'tags'>

// 管理端评论视图，与后端 service.AdminCommentView 一致，恢复审计字段与文章归属供审核使用。
export interface AdminComment extends Comment {
  article?: CommentArticleRef
  authorEmail: string
  authorIP: string
  userAgent: string
}

// 评论列表数据
export interface CommentListData {
  page: number
  pageSize: number
  total: number
  comments: Comment[]
}

// 管理端评论列表数据，评论项为恢复审计字段后的视图。
export interface AdminCommentListData {
  page: number
  pageSize: number
  total: number
  comments: AdminComment[]
}

// 创建评论请求参数，游客提交时姓名必填；登录用户经令牌绑定身份，游客字段将被忽略。
export interface CreateCommentRequest {
  articleId: number
  parentId?: number | null
  content: string
  authorName?: string
  authorEmail?: string
  authorWebsite?: string
}

// 文章评论列表查询参数
export interface ListCommentsRequest {
  page?: number
  pageSize?: number
}

// 管理端评论列表查询参数
export interface AdminListCommentsRequest {
  page?: number
  pageSize?: number
  status?: CommentStatus
  keyword?: string
}

// 各响应类型
export type CommentResponse = ApiResponse<Comment>
export type CommentListResponse = ApiResponse<CommentListData>
export type AdminCommentListResponse = ApiResponse<AdminCommentListData>
export type CommentActionResponse = ApiResponse<{ message: string } | null>
