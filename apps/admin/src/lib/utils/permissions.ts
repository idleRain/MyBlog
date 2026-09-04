// 权限管理工具函数

import { ROLE_PERMISSIONS, ROLE_CONFIG } from '$lib/constants/auth'
import type { User, UserRole, RoleInfo } from '$lib/types'
import { authStore } from '$lib/stores/auth'

/**
 * 检查用户是否拥有指定权限。
 * 权限唯一权威为后端登录下发的权限列表（铁律 A4），未下发时降级使用内置映射。
 */
export function hasPermission(user: User | null, permission: string): boolean {
  if (!user) return false

  const delivered = authStore.getPermissions()
  if (delivered.length > 0) {
    return delivered.includes(permission)
  }

  const rolePermissions = ROLE_PERMISSIONS[user.role] || []
  return rolePermissions.includes(permission)
}

/**
 * 获取角色信息。
 */
export function getRoleInfo(role: UserRole): RoleInfo {
  const config = ROLE_CONFIG[role]
  return {
    name: config.name,
    color: config.color,
    level: config.level
  }
}

/**
 * 获取当前用户可分配的角色列表，超级管理员可分配全部角色，管理员只能分配普通角色。
 */
export function getAssignableRoles(currentUser: User | null): UserRole[] {
  if (!currentUser) return []

  if (currentUser.role === 'superadmin') {
    return ['user', 'editor', 'admin', 'superadmin']
  }

  if (currentUser.role === 'admin') {
    return ['user', 'editor']
  }

  return []
}
