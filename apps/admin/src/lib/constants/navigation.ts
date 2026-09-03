// 后台路由的角色访问规则，供 (admin) 布局做路由级守卫。
// 与 AppSidebar 的导航分组保持一致，作为直达 URL 时的后端 403 前置拦截。

import type { UserRole } from '@myblog/api/modules/user/types'

// 单个路由规则，prefix 为路径前缀，roles 为允许访问的角色集合。
export interface RouteRoleRule {
  prefix: string
  roles: UserRole[]
}

// 全部后台路由的角色规则，顺序不影响匹配，匹配时取最长前缀。
export const ADMIN_ROUTE_RULES: RouteRoleRule[] = [
  { prefix: '/notifications', roles: ['user', 'editor', 'admin', 'superadmin'] },
  { prefix: '/posts', roles: ['editor', 'admin', 'superadmin'] },
  { prefix: '/categories', roles: ['admin', 'superadmin'] },
  { prefix: '/tags', roles: ['admin', 'superadmin'] },
  { prefix: '/comments', roles: ['admin', 'superadmin'] },
  { prefix: '/media', roles: ['editor', 'admin', 'superadmin'] },
  { prefix: '/links', roles: ['superadmin'] },
  { prefix: '/stats', roles: ['admin', 'superadmin'] },
  { prefix: '/users', roles: ['admin', 'superadmin'] },
  { prefix: '/settings', roles: ['superadmin'] },
  { prefix: '/', roles: ['user', 'editor', 'admin', 'superadmin'] }
]

// 全部角色常量，供未命中规则的路径使用。
const ALL_ROLES: UserRole[] = ['user', 'editor', 'admin', 'superadmin']

/**
 * 解析路径允许访问的角色集合，按最长前缀匹配，未命中时默认放行。
 */
export function resolveAllowedRoles(pathname: string): UserRole[] {
  const matched = ADMIN_ROUTE_RULES.filter(
    rule => pathname === rule.prefix || pathname.startsWith(`${rule.prefix}/`)
  )
  if (matched.length === 0) return ALL_ROLES
  matched.sort((a, b) => b.prefix.length - a.prefix.length)
  // 长度已判空，取最长前缀匹配的规则。
  return matched[0]!.roles
}
