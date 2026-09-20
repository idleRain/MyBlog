// 标签模块常量：状态配置与分页容量。
// 状态展示文案与样式的唯一来源为 tag_status 字典，此内置配置仅作字典不可用时的降级。

import type { BadgeVariant } from '$ui/badge'

// 标签状态配置，0 隐藏 / 1 启用。
export const TAG_STATUS_CONFIG: Record<number, { label: string; variant: BadgeVariant }> = {
  0: { label: '隐藏', variant: 'secondary' },
  1: { label: '启用', variant: 'default' }
}

// 标签默认分页大小
export const TAG_PAGE_SIZE = 10
