// 用户模块常量：状态配置与角色选择。

import type { UserRole, UserStatus } from '@myblog/api/modules/user/types'
import type { BadgeVariant } from '$ui/badge'

// 用户状态配置，0 禁用 / 1 正常 / 2 锁定。
export const USER_STATUS_CONFIG: Record<UserStatus, { label: string; variant: BadgeVariant }> = {
  0: { label: '禁用', variant: 'secondary' },
  1: { label: '正常', variant: 'default' },
  2: { label: '锁定', variant: 'destructive' }
}

// 角色中文名称映射
export const USER_ROLE_LABELS: Record<UserRole, string> = {
  superadmin: '超级管理员',
  admin: '管理员',
  editor: '编辑者',
  user: '用户'
}

// 用户默认分页大小
export const USER_PAGE_SIZE = 10

// 搜索时单次抓取的用户数，取后端 pageSize 上限以覆盖更多用户。
export const USER_SEARCH_PAGE_SIZE = 100

// 搜索时最多抓取的页数，避免超大用户量时发起过多请求。
export const USER_SEARCH_MAX_PAGES = 10
