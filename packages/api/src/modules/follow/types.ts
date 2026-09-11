import type { ApiResponse } from '@myblog/shared'

// 关注关系中的用户摘要
export interface FollowUserSummary {
  id: number
  username: string
  nickname: string
  avatar: string
}

// 关注关系条目，user 为对方用户的摘要：粉丝列表中为关注者，关注列表中为被关注者
export interface FollowItem {
  id: number
  followerId: number
  followingId: number
  createdAt: string
  user: FollowUserSummary
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
  follows: FollowItem[]
  total: number
  page: number
  pageSize: number
}

// 关注状态查询响应数据
export interface FollowStateData {
  isFollowing: boolean
}

export type FollowActionResponse = ApiResponse<null>
export type FollowListResponse = ApiResponse<FollowListData>
export type FollowStateResponse = ApiResponse<FollowStateData>
