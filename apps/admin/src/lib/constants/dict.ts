// 字典模块常量：生效状态取值与展示配置。
// 字典类型与字典项共用同一状态枚举，1 生效 / 0 停用。

import type { BadgeVariant } from '$ui/badge'

// 字典生效状态取值。
export const DICT_STATUS_ENABLED = 1

// 字典停用状态取值。
export const DICT_STATUS_DISABLED = 0

// 状态展示配置，仅作后端状态文案未下发时的降级来源。
export const DICT_STATUS_CONFIG: Record<number, { label: string; variant: BadgeVariant }> = {
  [DICT_STATUS_ENABLED]: { label: '生效', variant: 'default' },
  [DICT_STATUS_DISABLED]: { label: '停用', variant: 'secondary' }
}

// 查询状态展示配置，未知取值回退为停用样式，避免列表出现无标注的状态格。
export function dictStatusOf(status: number): { label: string; variant: BadgeVariant } {
  return (
    DICT_STATUS_CONFIG[status] ?? {
      label: '未知',
      variant: 'secondary'
    }
  )
}
