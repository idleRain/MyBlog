import type { ApiResponse } from '@myblog/shared'

// 友情链接状态枚举，与后端 model.LinkStatus 一致。
export type LinkStatus = 'pending' | 'active' | 'hidden' | 'rejected'

// 友情链接接口，字段与后端 model.FriendlyLink 的 JSON tag 一致。
export interface FriendlyLink {
  id: number
  name: string
  url: string
  logo: string
  description: string
  contactEmail: string
  sortOrder: number
  status: LinkStatus
  isReciprocal: boolean
  note: string
  createdAt: string
  updatedAt: string
}

// 友情链接列表数据
export interface FriendlyLinkListData {
  page: number
  pageSize: number
  total: number
  links: FriendlyLink[]
}

// 友情链接列表查询参数
export interface ListFriendlyLinksRequest {
  page?: number
  pageSize?: number
  status?: LinkStatus
}

// 创建友情链接请求参数
export interface CreateFriendlyLinkRequest {
  name: string
  url: string
  logo?: string
  description?: string
  contactEmail?: string
  sortOrder?: number | null
  isReciprocal?: boolean | null
}

// 更新友情链接请求参数，可选字段显式传入才更新。
export interface UpdateFriendlyLinkRequest {
  id: number
  name?: string | null
  url?: string | null
  logo?: string | null
  description?: string | null
  contactEmail?: string | null
  sortOrder?: number | null
  isReciprocal?: boolean | null
}

// 各响应类型
export type FriendlyLinkResponse = ApiResponse<FriendlyLink>
export type FriendlyLinkListResponse = ApiResponse<FriendlyLinkListData>
export type FriendlyLinkActionResponse = ApiResponse<{ message: string } | null>
