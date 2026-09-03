// 友情链接模块常量：状态配置与筛选选项。

import type { LinkStatus } from '@myblog/api/modules/friendlyLink/types'
import type { BadgeVariant } from '$ui/badge'

// 友链状态配置
export const LINK_STATUS_CONFIG: Record<LinkStatus, { label: string; variant: BadgeVariant }> = {
  pending: { label: '待审核', variant: 'secondary' },
  active: { label: '展示中', variant: 'default' },
  hidden: { label: '已隐藏', variant: 'outline' },
  rejected: { label: '已拒绝', variant: 'destructive' }
}

// 友链状态筛选选项，'' 表示全部。
export const LINK_STATUS_OPTIONS: Array<{ value: LinkStatus | ''; label: string }> = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待审核' },
  { value: 'active', label: '展示中' },
  { value: 'hidden', label: '已隐藏' },
  { value: 'rejected', label: '已拒绝' }
]

// 友链默认分页大小
export const LINK_PAGE_SIZE = 10
