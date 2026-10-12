import type { ApiResponse } from '@myblog/shared'

// 友情链接状态枚举，与后端 model.LinkStatus 一致。
export type LinkStatus = 'pending' | 'active' | 'hidden' | 'rejected'

// 友情链接接口，字段与后端 model.FriendlyLink 的 JSON tag 一致，仅用于管理端接口。
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

// 公开友链窄化视图，与后端 service.FriendlyLinkPublic 一致。
// 公开端点不输出站长联系邮箱与管理员备注等非展示字段。
export interface FriendlyLinkPublic {
  id: number
  name: string
  url: string
  logo: string
  description: string
  sortOrder: number
  isReciprocal: boolean
}

// 访客提交友链申请的请求
export interface ApplyFriendlyLinkRequest {
  name: string
  url: string
  description?: string
  contactEmail: string
}

// 友情链接列表数据
export interface FriendlyLinkListData {
  page: number
  pageSize: number
  total: number
  links: FriendlyLink[]
}

// 公开友链列表数据，条目为窄化视图。
export interface FriendlyLinkPublicListData {
  links: FriendlyLinkPublic[]
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
export type FriendlyLinkPublicResponse = ApiResponse<FriendlyLinkPublic>
export type FriendlyLinkPublicListResponse = ApiResponse<FriendlyLinkPublicListData>
export type FriendlyLinkActionResponse = ApiResponse<{ message: string } | null>
