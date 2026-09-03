// 分类模块常量：状态配置与筛选选项。

import type { BadgeVariant } from '$ui/badge'

// 分类状态配置，0 隐藏 / 1 显示。
export const CATEGORY_STATUS_CONFIG: Record<number, { label: string; variant: BadgeVariant }> = {
  0: { label: '隐藏', variant: 'secondary' },
  1: { label: '显示', variant: 'default' }
}

// 分类状态筛选选项
export const CATEGORY_STATUS_OPTIONS: Array<{ value: number | null; label: string }> = [
  { value: null, label: '全部状态' },
  { value: 1, label: '显示' },
  { value: 0, label: '隐藏' }
]

// 分类默认分页大小
export const CATEGORY_PAGE_SIZE = 10
