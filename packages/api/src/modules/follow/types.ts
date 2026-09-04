import type { ApiResponse } from '@myblog/shared'

// 关注关系记录
export interface Follow {
  id: number
  followerId: number
  followingId: number
  createdAt: string
}

// 关注操作请求参数
export interface FollowActionRequest {
  followingId: number
}

// 关注列表查询参数
export interface FollowListParams {
  page?: number
  pageSize?: number
}

// 关注列表响应数据
export interface FollowListData {
  follows: Follow[]
  total: number
  page: number
  pageSize: number
}

export type FollowActionResponse = ApiResponse<null>
export type FollowListResponse = ApiResponse<FollowListData>
