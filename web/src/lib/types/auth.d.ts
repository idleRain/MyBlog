// 认证和权限管理相关类型定义

import type { User, UserRole } from '$lib/api/modules/user/types'

// 认证状态接口
export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null
}

// JWT令牌信息
export interface TokenInfo {
  accessToken: string
  refreshToken: string
  expiresIn: number
  expiresAt: number
}

// 权限检查结果
export interface PermissionCheck {
  hasPermission: boolean
  reason?: string
}

// 角色权限配置
export interface RoleConfig {
  role: UserRole
  name: string
  level: number
  color: 'default' | 'destructive' | 'outline' | 'secondary'
  permissions: string[]
}

// 用户权限上下文
export interface UserPermissionContext {
  user: User
  permissions: string[]
  canAccess: (resource: string, action?: string) => boolean
  canManageUser: (targetRole: UserRole) => boolean
  isAdmin: () => boolean
  isSuperAdmin: () => boolean
  isEditor: () => boolean
}

// 权限验证配置
export interface PermissionGuardConfig {
  requireAuth?: boolean
  requiredRole?: UserRole
  requiredPermissions?: string[]
  requireAll?: boolean // 是否需要所有权限，默认false（需要任意一个）
}

// 路由守卫配置
export interface RouteGuardConfig extends PermissionGuardConfig {
  redirectTo?: string
  fallback?: () => void
}
