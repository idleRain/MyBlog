import type { ApiResponse } from '@myblog/shared'
import type { User } from '@myblog/api/modules/user/types'

// 设置值类型枚举
export type SettingType = 'string' | 'number' | 'boolean' | 'json' | 'array'

// 系统设置项接口，字段与后端 model.Setting 的 JSON tag 一致。
export interface Setting {
  id: number
  keyName: string
  label: string
  value: string
  defaultValue: string
  description: string
  type: SettingType
  groupName: string
  isPublic: boolean
  isReadonly: boolean
  isSensitive: boolean
  validationRule: string
  sortOrder: number
  updatedBy: number | null
  createdAt: string
  updatedAt: string
  updatedByUser: User | null
}

// 设置列表响应数据
export interface SettingsData {
  settings: Setting[]
}

// 单个设置项的更新内容
export interface UpdateSettingItem {
  keyName: string
  value: string
}

// 批量更新设置请求参数
export interface UpdateSettingsRequest {
  items: UpdateSettingItem[]
}

// 各响应类型
export type SettingsResponse = ApiResponse<SettingsData>
export type UpdateSettingsResponse = ApiResponse<SettingsData>
