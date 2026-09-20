import type { ApiResponse } from '@myblog/shared'

// 字典模块接口类型，字段与后端 model.DictType / model.DictItem 的 JSON tag 一致。

// 字典生效状态枚举，0 表示停用，1 表示生效。
export type DictStatus = 0 | 1

// 字典类型翻译行，与后端 model.DictTypeTranslation 的 JSON tag 一致，
// 仅管理端全量包请求（Accept-Language: *）携带。
export interface DictTypeTranslation {
  id: number
  typeId: number
  locale: string
  name: string | null
  description: string | null
  createdAt: string
  updatedAt: string
}

// 字典项翻译行，与后端 model.DictItemTranslation 的 JSON tag 一致。
export interface DictItemTranslation {
  id: number
  itemId: number
  locale: string
  label: string | null
  description: string | null
  createdAt: string
  updatedAt: string
}

// 字典类型信息接口
export interface DictType {
  id: number
  code: string
  name: string
  description: string
  status: DictStatus
  sortOrder: number
  // 预留扩展元数据，如颜色、图标等展示配置
  extra: Record<string, unknown> | null
  createdAt: string
  updatedAt: string
  // 全量翻译行数组，仅 Accept-Language: * 的管理端请求返回。
  translations?: DictTypeTranslation[]
}

// 字典项信息接口
export interface DictItem {
  id: number
  typeId: number
  value: string
  label: string
  description: string
  status: DictStatus
  sortOrder: number
  // 预留扩展元数据，如颜色、图标等展示配置
  extra: Record<string, unknown> | null
  createdAt: string
  updatedAt: string
  // 全量翻译行数组，仅 Accept-Language: * 的管理端请求返回。
  translations?: DictItemTranslation[]
}

// 已生效字典分组，类型字段平铺并附加已生效字典项集合，
// 与后端 service.EnabledDictGroup 的 JSON 输出一致。
export interface EnabledDictGroup extends DictType {
  items: DictItem[]
}

// 字典类型列表查询参数
export interface ListDictTypesRequest {
  page?: number
  pageSize?: number
  status?: DictStatus | null
  search?: string
}

// 字典类型列表数据
export interface DictTypeListData {
  types: DictType[]
  total: number
  page: number
  pageSize: number
}

// 字典项列表查询参数，typeId 必填，管理端按类型维护。
export interface ListDictItemsRequest {
  typeId: number
  page?: number
  pageSize?: number
  status?: DictStatus | null
  search?: string
}

// 字典项列表数据
export interface DictItemListData {
  items: DictItem[]
  total: number
  page: number
  pageSize: number
}

// 全量已生效字典数据
export interface AllDictsData {
  dicts: EnabledDictGroup[]
}

// 字典类型单语言翻译补丁，字段可选，提供即更新，未提供保留既有翻译值。
export interface DictTypeI18nPatch {
  name?: string
  description?: string
}

// 按语言组织的字典类型翻译字段包，键为语言标识，缺省语言走主字段不允许出现。
export type DictTypeI18nPayload = Record<string, DictTypeI18nPatch>

// 字典项单语言翻译补丁
export interface DictItemI18nPatch {
  label?: string
  description?: string
}

// 按语言组织的字典项翻译字段包
export type DictItemI18nPayload = Record<string, DictItemI18nPatch>

// 创建字典类型请求参数
export interface CreateDictTypeRequest {
  code: string
  name: string
  description?: string
  status?: DictStatus | null
  sortOrder?: number | null
  extra?: Record<string, unknown> | null
  i18n?: DictTypeI18nPayload
}

// 更新字典类型请求参数，可选字段显式传入才更新，字典码不可变更。
export interface UpdateDictTypeRequest {
  id: number
  name?: string | null
  description?: string | null
  status?: DictStatus | null
  sortOrder?: number | null
  extra?: Record<string, unknown> | null
  i18n?: DictTypeI18nPayload
}

// 创建字典项请求参数
export interface CreateDictItemRequest {
  typeId: number
  value: string
  label: string
  description?: string
  status?: DictStatus | null
  sortOrder?: number | null
  extra?: Record<string, unknown> | null
  i18n?: DictItemI18nPayload
}

// 更新字典项请求参数，可选字段显式传入才更新，字典项值不可变更。
export interface UpdateDictItemRequest {
  id: number
  label?: string | null
  description?: string | null
  status?: DictStatus | null
  sortOrder?: number | null
  extra?: Record<string, unknown> | null
  i18n?: DictItemI18nPayload
}

// 各响应类型
export type DictTypeResponse = ApiResponse<DictType>
export type DictItemResponse = ApiResponse<DictItem>
export type DictTypeListResponse = ApiResponse<DictTypeListData>
export type DictItemListResponse = ApiResponse<DictItemListData>
export type AllDictsResponse = ApiResponse<AllDictsData>
export type DictByTypeResponse = ApiResponse<EnabledDictGroup>
export type DeleteDictResponse = ApiResponse<null>
