import type {
  AdminCommentListResponse,
  AdminListCommentsRequest,
  CommentActionResponse,
  CommentListResponse,
  CommentResponse,
  CreateCommentRequest,
  ListCommentsRequest
} from './types.ts'
import type { KyInstance } from 'ky'

/**
 * 创建评论接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createCommentAPI(request: KyInstance) {
  return {
    // 文章评论列表，仅返回已审核通过的评论。
    listByArticle(articleId: number, params: ListCommentsRequest): Promise<CommentListResponse> {
      return request.post('comments/list', { json: { articleId, ...params } }).json()
    },

    // 发表评论，游客需填写姓名，登录用户经令牌绑定身份。
    create(params: CreateCommentRequest): Promise<CommentResponse> {
      return request.post('comments/create', { json: params }).json()
    },

    // 点赞评论，需要登录。
    like(id: number): Promise<CommentActionResponse> {
      return request.post('comments/like', { json: { id } }).json()
    },

    // 取消点赞评论，需要登录。
    unlike(id: number): Promise<CommentActionResponse> {
      return request.post('comments/unlike', { json: { id } }).json()
    },

    // 管理端：全量评论列表，按状态与关键词筛选，响应携带审计字段。
    adminList(params: AdminListCommentsRequest): Promise<AdminCommentListResponse> {
      return request.post('admin/comments/list', { json: params }).json()
    },

    // 管理端：审核通过评论。
    approve(id: number): Promise<CommentActionResponse> {
      return request.post('admin/comments/approve', { json: { id } }).json()
    },

    // 管理端：拒绝评论。
    reject(id: number): Promise<CommentActionResponse> {
      return request.post('admin/comments/reject', { json: { id } }).json()
    },

    // 管理端：标记为垃圾评论。
    markSpam(id: number): Promise<CommentActionResponse> {
      return request.post('admin/comments/spam', { json: { id } }).json()
    },

    // 管理端：移入回收站。
    trash(id: number): Promise<CommentActionResponse> {
      return request.post('admin/comments/trash', { json: { id } }).json()
    },

    // 管理端：删除评论。
    delete(id: number): Promise<CommentActionResponse> {
      return request.post('admin/comments/delete', { json: { id } }).json()
    }
  }
}

export type CommentAPI = ReturnType<typeof createCommentAPI>

export type {
  AdminCommentListResponse,
  AdminListCommentsRequest,
  CommentActionResponse,
  CommentListResponse,
  CommentResponse,
  CreateCommentRequest,
  ListCommentsRequest
}
