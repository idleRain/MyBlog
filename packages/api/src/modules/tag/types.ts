import type { ApiResponse } from '@myblog/shared'

// 标签状态枚举，0 表示隐藏，1 表示启用。
export type TagStatus = 0 | 1

// 标签信息接口，字段与后端 model.Tag 的 JSON tag 一致。
export interface Tag {
  id: number
  name: string
  slug: string
  color: string
  description: string
  status: TagStatus
  usageCount: number
  isHot: boolean
  createdAt: string
  updatedAt: string
}

// 标签列表数据
export interface TagListData {
  page: number
  pageSize: number
  total: number
  tags: Tag[]
}

// 热门标签响应数据
export interface PopularTagsData {
  tags: Tag[]
}

// 标签列表查询参数
export interface ListTagsRequest {
  page?: number
  pageSize?: number
  status?: TagStatus | null
  isHot?: boolean | null
  search?: string
}

// 创建标签请求参数
export interface CreateTagRequest {
  name: string
  slug?: string
  color?: string
  description?: string
  status?: TagStatus | null
  isHot?: boolean | null
}

// 更新标签请求参数，可选字段显式传入才更新。
export interface UpdateTagRequest {
  id: number
  name?: string | null
  slug?: string | null
  color?: string | null
  description?: string | null
  status?: TagStatus | null
  isHot?: boolean | null
}

// 各响应类型
export type TagResponse = ApiResponse<Tag>
export type TagListResponse = ApiResponse<TagListData>
export type PopularTagsResponse = ApiResponse<PopularTagsData>
export type DeleteTagResponse = ApiResponse<null>
