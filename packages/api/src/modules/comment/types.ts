import type { ApiResponse } from '@myblog/shared'
import type { User } from '@myblog/api/modules/user/types'
import type { Article } from '@myblog/api/modules/article/types'

// 评论审核状态枚举，与后端 model.CommentStatus 一致。
export type CommentStatus = 'pending' | 'approved' | 'rejected' | 'spam' | 'trash'

// 评论信息接口，字段与后端 model.Comment 的 JSON tag 一致。
export interface Comment {
  id: number
  articleId: number
  userId: number | null
  parentId: number | null
  rootId: number | null
  level: number
  authorName: string
  authorEmail: string
  authorWebsite: string
  authorIP: string
  content: string
  contentHtml: string
  status: CommentStatus
  likeCount: number
  replyCount: number
  reportedCount: number
  userAgent: string
  isAuthor: boolean
  isPinned: boolean
  editedAt: string | null
  createdAt: string
  updatedAt: string
  article: Article | null
  user: User | null
  parent: Comment | null
  root: Comment | null
  children: Comment[]
}

// 评论列表数据
export interface CommentListData {
  page: number
  pageSize: number
  total: number
  comments: Comment[]
}

// 创建评论请求参数，游客提交时姓名必填。
export interface CreateCommentRequest {
  articleId: number
  parentId?: number | null
  content: string
  authorName?: string
  authorEmail?: string
  authorWebsite?: string
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
export type CommentActionResponse = ApiResponse<{ message: string } | null>
