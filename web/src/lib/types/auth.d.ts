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

// 权限常量
export const PERMISSIONS = {
  // 系统管理
  SYSTEM_CONFIG: 'system:config',
  SYSTEM_LOGS: 'system:logs',
  SYSTEM_STATS: 'system:stats',

  // 用户管理
  USER_CREATE: 'user:create',
  USER_READ: 'user:read',
  USER_UPDATE: 'user:update',
  USER_DELETE: 'user:delete',
  USER_LIST: 'user:list',

  // 文章管理
  ARTICLE_CREATE: 'article:create',
  ARTICLE_READ: 'article:read',
  ARTICLE_UPDATE: 'article:update',
  ARTICLE_DELETE: 'article:delete',
  ARTICLE_LIST: 'article:list',
  ARTICLE_PUBLISH: 'article:publish',

  // 分类标签
  CATEGORY_MANAGE: 'category:manage',
  TAG_MANAGE: 'tag:manage',

  // 评论管理
  COMMENT_CREATE: 'comment:create',
  COMMENT_READ: 'comment:read',
  COMMENT_UPDATE: 'comment:update',
  COMMENT_DELETE: 'comment:delete',
  COMMENT_MODERATE: 'comment:moderate',

  // 文件管理
  FILE_UPLOAD: 'file:upload',
  FILE_READ: 'file:read',
  FILE_DELETE: 'file:delete'
} as const

// 角色权限映射
export const ROLE_PERMISSIONS: Record<UserRole, string[]> = {
  user: [
    PERMISSIONS.ARTICLE_READ,
    PERMISSIONS.COMMENT_CREATE,
    PERMISSIONS.COMMENT_READ,
    PERMISSIONS.COMMENT_UPDATE // 只能更新自己的评论
  ],
  editor: [
    PERMISSIONS.USER_READ,
    PERMISSIONS.ARTICLE_CREATE,
    PERMISSIONS.ARTICLE_READ,
    PERMISSIONS.ARTICLE_UPDATE,
    PERMISSIONS.ARTICLE_DELETE,
    PERMISSIONS.ARTICLE_LIST,
    PERMISSIONS.ARTICLE_PUBLISH,
    PERMISSIONS.COMMENT_CREATE,
    PERMISSIONS.COMMENT_READ,
    PERMISSIONS.COMMENT_UPDATE,
    PERMISSIONS.COMMENT_DELETE,
    PERMISSIONS.FILE_UPLOAD,
    PERMISSIONS.FILE_READ
  ],
  admin: [
    PERMISSIONS.SYSTEM_LOGS,
    PERMISSIONS.SYSTEM_STATS,
    PERMISSIONS.USER_CREATE,
    PERMISSIONS.USER_READ,
    PERMISSIONS.USER_UPDATE,
    PERMISSIONS.USER_DELETE,
    PERMISSIONS.USER_LIST,
    PERMISSIONS.ARTICLE_CREATE,
    PERMISSIONS.ARTICLE_READ,
    PERMISSIONS.ARTICLE_UPDATE,
    PERMISSIONS.ARTICLE_DELETE,
    PERMISSIONS.ARTICLE_LIST,
    PERMISSIONS.ARTICLE_PUBLISH,
    PERMISSIONS.CATEGORY_MANAGE,
    PERMISSIONS.TAG_MANAGE,
    PERMISSIONS.COMMENT_CREATE,
    PERMISSIONS.COMMENT_READ,
    PERMISSIONS.COMMENT_UPDATE,
    PERMISSIONS.COMMENT_DELETE,
    PERMISSIONS.COMMENT_MODERATE,
    PERMISSIONS.FILE_UPLOAD,
    PERMISSIONS.FILE_READ,
    PERMISSIONS.FILE_DELETE
  ],
  superadmin: [
    // 超级管理员拥有所有权限
    ...Object.values(PERMISSIONS)
  ]
}

// 角色配置
export const ROLE_CONFIG: Record<UserRole, RoleConfig> = {
  user: {
    role: 'user',
    name: '用户',
    level: 1,
    color: 'outline',
    permissions: ROLE_PERMISSIONS.user
  },
  editor: {
    role: 'editor',
    name: '编辑者',
    level: 2,
    color: 'secondary',
    permissions: ROLE_PERMISSIONS.editor
  },
  admin: {
    role: 'admin',
    name: '管理员',
    level: 3,
    color: 'default',
    permissions: ROLE_PERMISSIONS.admin
  },
  superadmin: {
    role: 'superadmin',
    name: '超级管理员',
    level: 4,
    color: 'destructive',
    permissions: ROLE_PERMISSIONS.superadmin
  }
}
