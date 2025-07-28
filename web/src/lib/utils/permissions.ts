// 权限管理工具函数

import type { User, UserRole, UserPermissionContext, PermissionCheck, RoleInfo } from '$lib/types'
import { ROLE_PERMISSIONS, ROLE_CONFIG } from '$lib/types/auth'

/**
 * 检查用户是否拥有指定权限
 */
export function hasPermission(user: User | null, permission: string): boolean {
  if (!user) return false

  const rolePermissions = ROLE_PERMISSIONS[user.role] || []
  return rolePermissions.includes(permission)
}

/**
 * 检查用户是否拥有任意一个权限
 */
export function hasAnyPermission(user: User | null, permissions: string[]): boolean {
  if (!user || !permissions.length) return false

  return permissions.some(permission => hasPermission(user, permission))
}

/**
 * 检查用户是否拥有所有权限
 */
export function hasAllPermissions(user: User | null, permissions: string[]): boolean {
  if (!user || !permissions.length) return false

  return permissions.every(permission => hasPermission(user, permission))
}

/**
 * 检查用户角色级别是否足够
 */
export function hasRoleLevel(user: User | null, minLevel: number): boolean {
  if (!user) return false

  const userLevel = ROLE_CONFIG[user.role]?.level || 0
  return userLevel >= minLevel
}

/**
 * 检查用户是否为指定角色或更高级别
 */
export function hasRoleOrAbove(user: User | null, role: UserRole): boolean {
  if (!user) return false

  const requiredLevel = ROLE_CONFIG[role]?.level || 0
  return hasRoleLevel(user, requiredLevel)
}

/**
 * 检查用户是否为管理员或更高级别
 */
export function isAdmin(user: User | null): boolean {
  return hasRoleOrAbove(user, 'admin')
}

/**
 * 检查用户是否为超级管理员
 */
export function isSuperAdmin(user: User | null): boolean {
  return user?.role === 'superadmin'
}

/**
 * 检查用户是否为编辑者或更高级别
 */
export function isEditor(user: User | null): boolean {
  return hasRoleOrAbove(user, 'editor')
}

/**
 * 检查用户是否可以管理指定角色的用户
 */
export function canManageUserRole(manager: User | null, targetRole: UserRole): boolean {
  if (!manager) return false

  // 超级管理员可以管理所有用户
  if (manager.role === 'superadmin') return true

  // 管理员可以管理编辑者和普通用户
  if (manager.role === 'admin') {
    return targetRole === 'editor' || targetRole === 'user'
  }

  // 其他角色不能管理用户
  return false
}

/**
 * 获取角色信息
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
 * 获取用户的所有权限
 */
export function getUserPermissions(user: User | null): string[] {
  if (!user) return []
  return ROLE_PERMISSIONS[user.role] || []
}

/**
 * 创建用户权限上下文
 */
export function createUserPermissionContext(user: User): UserPermissionContext {
  const permissions = getUserPermissions(user)

  return {
    user,
    permissions,
    canAccess: (resource: string, action?: string) => {
      const permission = action ? `${resource}:${action}` : resource
      return hasPermission(user, permission)
    },
    canManageUser: (targetRole: UserRole) => canManageUserRole(user, targetRole),
    isAdmin: () => isAdmin(user),
    isSuperAdmin: () => isSuperAdmin(user),
    isEditor: () => isEditor(user)
  }
}

/**
 * 权限检查装饰器（用于组件属性）
 */
export function requirePermission(permission: string) {
  return (user: User | null): PermissionCheck => {
    const hasAccess = hasPermission(user, permission)
    return {
      hasPermission: hasAccess,
      reason: hasAccess ? undefined : `需要权限: ${permission}`
    } as PermissionCheck
  }
}

/**
 * 角色检查装饰器
 */
export function requireRole(role: UserRole) {
  return (user: User | null): PermissionCheck => {
    const hasAccess = hasRoleOrAbove(user, role)
    return {
      hasPermission: hasAccess,
      reason: hasAccess ? undefined : `需要角色: ${getRoleInfo(role).name} 或更高级别`
    } as PermissionCheck
  }
}

/**
 * 检查用户状态是否正常
 */
export function isUserActive(user: User | null): boolean {
  return user?.status === 1
}

/**
 * 获取角色列表（按级别排序）
 */
export function getRoleList(): Array<{ role: UserRole; name: string; level: number }> {
  return Object.entries(ROLE_CONFIG)
    .map(([role, config]) => ({
      role: role as UserRole,
      name: config.name,
      level: config.level
    }))
    .sort((a, b) => a.level - b.level)
}

/**
 * 获取可分配的角色列表（基于当前用户权限）
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

/**
 * 检查是否可以访问指定路由
 */
export function canAccessRoute(
  user: User | null,
  routePermissions?: string[],
  requiredRole?: UserRole
): boolean {
  if (!user || !isUserActive(user)) return false

  // 检查角色要求
  if (requiredRole && !hasRoleOrAbove(user, requiredRole)) {
    return false
  }

  // 检查权限要求
  if (routePermissions && routePermissions.length > 0) {
    return hasAnyPermission(user, routePermissions)
  }

  return true
}
