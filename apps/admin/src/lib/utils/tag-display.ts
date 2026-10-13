// 标签状态展示解析：优先读取标签状态字典的扩展元数据，缺失或非法时回退内置配置。
// 页面与表格组件共用同一解析口径，避免两处各自维护降级规则。

import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import { findDictItem, dictItemLabel } from '@myblog/api'
import type { BadgeVariant } from '$ui/badge'

import { TAG_STATUS_CONFIG } from '$lib/constants/tag'

// Badge 合法样式集合，用于校验字典扩展元数据中的样式取值。
const BADGE_VARIANTS: BadgeVariant[] = ['default', 'secondary', 'destructive', 'outline']

// 解析标签状态的展示文案。
export function resolveTagStatusLabel(statusDict: EnabledDictGroup | null, status: number): string {
  return dictItemLabel(statusDict, String(status), TAG_STATUS_CONFIG[status]?.label ?? '未知')
}

// 解析标签状态的徽标样式，扩展元数据取值非法时回退内置配置。
export function resolveTagStatusVariant(
  statusDict: EnabledDictGroup | null,
  status: number
): BadgeVariant {
  const fallback = TAG_STATUS_CONFIG[status]?.variant ?? 'secondary'
  const variant = findDictItem(statusDict, String(status))?.extra?.variant
  if (typeof variant === 'string' && BADGE_VARIANTS.includes(variant as BadgeVariant)) {
    return variant as BadgeVariant
  }
  return fallback
}
