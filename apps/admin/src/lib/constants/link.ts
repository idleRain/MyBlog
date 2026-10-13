// 友情链接模块常量：状态配置、筛选选项与审核动作。

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

// 友链审核动作键，与友情链接接口的同名审核方法一致。
export type LinkStatusAction = 'approve' | 'hide' | 'reject'

// 各状态下可执行的审核动作；同一动作在不同状态下的文案不同，故文案随状态声明。
export const LINK_STATUS_ACTIONS: Record<
  LinkStatus,
  Array<{ key: LinkStatusAction; label: string }>
> = {
  pending: [
    { key: 'approve', label: '通过' },
    { key: 'reject', label: '拒绝' }
  ],
  active: [{ key: 'hide', label: '下架' }],
  hidden: [
    { key: 'approve', label: '重新上架' },
    { key: 'reject', label: '拒绝' }
  ],
  rejected: [{ key: 'approve', label: '重新上架' }]
}

// 友链默认分页大小
export const LINK_PAGE_SIZE = 10
