import type { ApiResponse } from '@myblog/shared'

// 标签状态枚举，0 表示隐藏，1 表示启用。
export type TagStatus = 0 | 1

// 标签单语言翻译行，与后端 model.TagTranslation 的 JSON tag 一致，
// 仅管理端全量包请求携带，公共请求按语言输出后不包含该结构。
export interface TagTranslation {
  id: number
  tagId: number
  locale: string
  name: string | null
  description: string | null
  createdAt: string
  updatedAt: string
}

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
  // 全量翻译行数组，仅 Accept-Language: * 的管理端请求返回。
  translations?: TagTranslation[]
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

// 标签单语言翻译补丁，字段可选，提供即更新，未提供保留既有翻译值。
export interface TagI18nPatch {
  name?: string
  description?: string
}

// 按语言组织的翻译字段包，键为语言标识，缺省语言走主字段不允许出现。
export type TagI18nPayload = Record<string, TagI18nPatch>

// 创建标签请求参数
export interface CreateTagRequest {
  name: string
  slug?: string
  color?: string
  description?: string
  status?: TagStatus | null
  isHot?: boolean | null
  i18n?: TagI18nPayload
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
  i18n?: TagI18nPayload
}

// 各响应类型
export type TagResponse = ApiResponse<Tag>
export type TagListResponse = ApiResponse<TagListData>
export type PopularTagsResponse = ApiResponse<PopularTagsData>
export type DeleteTagResponse = ApiResponse<null>
