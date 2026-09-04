import type { KyInstance } from 'ky'
import type {
  FollowActionRequest,
  FollowActionResponse,
  FollowListParams,
  FollowListResponse
} from './types.ts'

/**
 * 创建用户关注接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createFollowAPI(request: KyInstance) {
  return {
    // 关注用户，需登录。
    follow(params: FollowActionRequest): Promise<FollowActionResponse> {
      return request.post('users/follow', { json: params }).json()
    },

    // 取消关注，需登录。
    unfollow(params: FollowActionRequest): Promise<FollowActionResponse> {
      return request.post('users/unfollow', { json: params }).json()
    },

    // 分页查询粉丝列表，公开可访问。
    listFollowers(userId: number, params: FollowListParams = {}): Promise<FollowListResponse> {
      return request.post('users/followers', { json: { userId, ...params } }).json()
    },

    // 分页查询关注列表，公开可访问。
    listFollowing(userId: number, params: FollowListParams = {}): Promise<FollowListResponse> {
      return request.post('users/following', { json: { userId, ...params } }).json()
    }
  }
}

export type FollowAPI = ReturnType<typeof createFollowAPI>
