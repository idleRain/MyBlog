import type { KyInstance } from 'ky'
import type {
  AdminListCommentsRequest,
  CommentActionResponse,
  CommentListResponse
} from './types.ts'

// 评论审核状态流转接口路径后缀。
const ADMIN_COMMENT_ACTION_PATHS = {
  approve: 'admin/comments/approve',
  reject: 'admin/comments/reject',
  spam: 'admin/comments/spam',
  trash: 'admin/comments/trash',
  delete: 'admin/comments/delete'
} as const

/**
 * 创建评论接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createCommentAPI(request: KyInstance) {
  return {
    // 管理端：全量评论列表，按状态与关键词筛选。
    adminList(params: AdminListCommentsRequest): Promise<CommentListResponse> {
      return request.post('admin/comments/list', { json: params }).json()
    },

    // 管理端：审核通过评论。
    approve(id: number): Promise<CommentActionResponse> {
      return request.post(ADMIN_COMMENT_ACTION_PATHS.approve, { json: { id } }).json()
    },

    // 管理端：拒绝评论。
    reject(id: number): Promise<CommentActionResponse> {
      return request.post(ADMIN_COMMENT_ACTION_PATHS.reject, { json: { id } }).json()
    },

    // 管理端：标记为垃圾评论。
    markSpam(id: number): Promise<CommentActionResponse> {
      return request.post(ADMIN_COMMENT_ACTION_PATHS.spam, { json: { id } }).json()
    },

    // 管理端：移入回收站。
    trash(id: number): Promise<CommentActionResponse> {
      return request.post(ADMIN_COMMENT_ACTION_PATHS.trash, { json: { id } }).json()
    },

    // 管理端：删除评论。
    delete(id: number): Promise<CommentActionResponse> {
      return request.post(ADMIN_COMMENT_ACTION_PATHS.delete, { json: { id } }).json()
    }
  }
}

export type CommentAPI = ReturnType<typeof createCommentAPI>

export type { AdminListCommentsRequest, CommentActionResponse, CommentListResponse }
