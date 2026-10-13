// 分类模块常量：状态配置、筛选选项与树形展示的缩进口径。

import type { Category } from '@myblog/api/modules/category/types'
import type { BadgeVariant } from '$ui/badge'

// 分类树扁平化后的展示行，depth 为层级深度，由渲染层换算为缩进量。
export interface CategoryRow {
  category: Category
  depth: number
}

// 分类状态配置，0 隐藏 / 1 显示。
export const CATEGORY_STATUS_CONFIG: Record<number, { label: string; variant: BadgeVariant }> = {
  0: { label: '隐藏', variant: 'secondary' },
  1: { label: '显示', variant: 'default' }
}

// 分类状态筛选选项，取值为状态枚举本身，供展示型筛选使用。
export const CATEGORY_STATUS_OPTIONS: Array<{ value: number | null; label: string }> = [
  { value: null, label: '全部状态' },
  { value: 1, label: '显示' },
  { value: 0, label: '隐藏' }
]

// 状态筛选的不限制取值，空串表示全部状态。
export const CATEGORY_STATUS_FILTER_ALL = ''

// 状态筛选项的界面取值，与 CATEGORY_STATUS_OPTIONS 同序。
// 取值统一为字符串，与切换组的字符串取值口径保持一致。
export const CATEGORY_STATUS_FILTER_OPTIONS: ReadonlyArray<{ value: string; label: string }> =
  CATEGORY_STATUS_OPTIONS.map(option => ({
    value: option.value === null ? CATEGORY_STATUS_FILTER_ALL : String(option.value),
    label: option.label
  }))

// 分类树行的基础左内边距，单位为 rem。
export const CATEGORY_ROW_BASE_PADDING_REM = 1

// 分类树行每深入一层追加的左内边距，单位为 rem，用于呈现父子归属。
export const CATEGORY_ROW_INDENT_PER_LEVEL_REM = 1.5

// 分类默认分页大小，供后续分页形态的分类列表复用。
export const CATEGORY_PAGE_SIZE = 10
