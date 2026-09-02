import type { KyInstance } from 'ky'
import type {
  CreateFriendlyLinkRequest,
  FriendlyLinkActionResponse,
  FriendlyLinkListResponse,
  FriendlyLinkResponse,
  ListFriendlyLinksRequest,
  UpdateFriendlyLinkRequest
} from './types.ts'

// 友链状态流转接口路径后缀。
const ADMIN_LINK_ACTION_PATHS = {
  approve: 'admin/friendly-links/approve',
  hide: 'admin/friendly-links/hide',
  reject: 'admin/friendly-links/reject'
} as const

/**
 * 创建友情链接接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createFriendlyLinkAPI(request: KyInstance) {
  return {
    // 管理端：分页查询友链列表
    adminList(params: ListFriendlyLinksRequest): Promise<FriendlyLinkListResponse> {
      return request.post('admin/friendly-links/list', { json: params }).json()
    },

    // 管理端：创建友链，新链接默认待审核。
    create(params: CreateFriendlyLinkRequest): Promise<FriendlyLinkResponse> {
      return request.post('admin/friendly-links/create', { json: params }).json()
    },

    // 管理端：更新友链
    update(params: UpdateFriendlyLinkRequest): Promise<FriendlyLinkResponse> {
      return request.post('admin/friendly-links/update', { json: params }).json()
    },

    // 管理端：删除友链
    delete(id: number): Promise<FriendlyLinkActionResponse> {
      return request.post('admin/friendly-links/delete', { json: { id } }).json()
    },

    // 管理端：审核通过，进入展示状态。
    approve(id: number): Promise<FriendlyLinkActionResponse> {
      return request.post(ADMIN_LINK_ACTION_PATHS.approve, { json: { id } }).json()
    },

    // 管理端：下架友链但保留数据。
    hide(id: number): Promise<FriendlyLinkActionResponse> {
      return request.post(ADMIN_LINK_ACTION_PATHS.hide, { json: { id } }).json()
    },

    // 管理端：拒绝该互链申请。
    reject(id: number): Promise<FriendlyLinkActionResponse> {
      return request.post(ADMIN_LINK_ACTION_PATHS.reject, { json: { id } }).json()
    }
  }
}

export type FriendlyLinkAPI = ReturnType<typeof createFriendlyLinkAPI>

export type {
  CreateFriendlyLinkRequest,
  FriendlyLinkActionResponse,
  FriendlyLinkListResponse,
  FriendlyLinkResponse,
  ListFriendlyLinksRequest,
  UpdateFriendlyLinkRequest
}
