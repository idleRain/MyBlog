// 认证相关工具函数

import type { User, UserRole, AuthState, UserStatus } from '$lib/types'
import { authStore } from '$lib/stores/auth'
import { get } from 'svelte/store'

/**
 * 认证状态检查结果
 */
export interface AuthCheckResult {
  isValid: boolean
  reason?: string
}

/**
 * Token 验证结果
 */
export interface TokenValidationResult {
  isValid: boolean
  expiresIn?: number
  shouldRefresh: boolean
}

/**
 * 获取当前认证状态
 */
export function getCurrentAuthState(): AuthState {
  return get(authStore)
}

/**
 * 获取当前用户信息
 */
export function getCurrentUser(): User | null {
  const state = getCurrentAuthState()
  return state.user
}

/**
 * 获取当前用户角色
 */
export function getCurrentUserRole(): UserRole {
  const user = getCurrentUser()
  return user?.role ?? 'user'
}

/**
 * 检查是否已登录
 */
export function isAuthenticated(): boolean {
  const state = getCurrentAuthState()
  return state.isAuthenticated && !!state.user
}

/**
 * 检查用户是否活跃
 */
export function isUserActive(user?: User | null): boolean {
  const targetUser = user ?? getCurrentUser()
  return targetUser?.status === 1
}

/**
 * 获取用户状态信息
 */
export function getUserStatus(user?: User | null): { status: UserStatus; statusText: string } {
  const targetUser = user ?? getCurrentUser()
  const status: UserStatus = targetUser?.status ?? 0

  const statusMap: Record<UserStatus, string> = {
    0: '禁用',
    1: '正常'
  }

  return {
    status,
    statusText: statusMap[status] ?? '未知'
  }
}

/**
 * 检查token是否有效
 */
export function isTokenValid(): boolean {
  const state = getCurrentAuthState()
  if (!state.accessToken || !state.expiresAt) return false

  // 检查是否过期（提前5分钟判断）
  return Date.now() < state.expiresAt - 5 * 60 * 1000
}

/**
 * 检查是否需要刷新token
 */
export function shouldRefreshToken(): boolean {
  const state = getCurrentAuthState()
  if (!state.accessToken || !state.expiresAt) return false

  // 如果在5分钟内过期，需要刷新
  return Date.now() >= state.expiresAt - 5 * 60 * 1000
}

/**
 * 获取token验证结果（包含详细信息）
 */
export function getTokenValidationResult(): TokenValidationResult {
  const state = getCurrentAuthState()

  if (!state.accessToken || !state.expiresAt) {
    return {
      isValid: false,
      shouldRefresh: false
    }
  }

  const now = Date.now()
  const expiresIn = state.expiresAt - now
  const refreshThreshold = 5 * 60 * 1000 // 5分钟

  return {
    isValid: expiresIn > refreshThreshold,
    expiresIn: Math.max(0, expiresIn),
    shouldRefresh: expiresIn <= refreshThreshold && expiresIn > 0
  }
}

/**
 * 获取访问令牌
 */
export function getAccessToken(): string | null {
  const state = getCurrentAuthState()
  return state.accessToken
}

/**
 * 获取刷新令牌
 */
export function getRefreshToken(): string | null {
  const state = getCurrentAuthState()
  return state.refreshToken
}

/**
 * 获取Authorization头
 */
export function getAuthorizationHeader(): Record<string, string> {
  const token = getAccessToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

/**
 * 检查是否可以访问管理后台
 */
export function canAccessAdmin(user?: User | null): AuthCheckResult {
  const targetUser = user ?? getCurrentUser()

  if (!isAuthenticated()) {
    return {
      isValid: false,
      reason: '用户未登录'
    }
  }

  if (!isUserActive(targetUser)) {
    return {
      isValid: false,
      reason: '用户账户已被禁用'
    }
  }

  if (!isTokenValid()) {
    return {
      isValid: false,
      reason: 'Token已过期，请重新登录'
    }
  }

  return {
    isValid: true
  }
}

/**
 * 获取用户显示名称
 */
export function getUserDisplayName(user?: User | null): string {
  const currentUser = user ?? getCurrentUser()
  if (!currentUser) return '未知用户'

  return currentUser.nickname || currentUser.username || '未知用户'
}

/**
 * 安全地获取用户显示名称（不会返回空字符串）
 */
export function getSafeUserDisplayName(user?: User | null): string {
  const displayName = getUserDisplayName(user)
  return displayName.trim() || '匿名用户'
}

/**
 * 格式化用户完整信息
 */
export interface UserInfo {
  id: number
  displayName: string
  username: string
  email: string
  role: UserRole
  roleName: string
  avatar?: string
  isActive: boolean
}

export function getUserInfo(user?: User | null): UserInfo | null {
  const currentUser = user ?? getCurrentUser()
  if (!currentUser) return null

  return {
    id: currentUser.id,
    displayName: getSafeUserDisplayName(currentUser),
    username: currentUser.username,
    email: currentUser.email,
    role: currentUser.role,
    roleName: getRoleDisplayName(currentUser.role),
    avatar: currentUser.avatar,
    isActive: currentUser.status === 1
  }
}

/**
 * 获取角色显示名称
 */
function getRoleDisplayName(role: UserRole): string {
  const roleNames: Record<UserRole, string> = {
    user: '用户',
    editor: '编辑者',
    admin: '管理员',
    superadmin: '超级管理员'
  }
  return roleNames[role] || '未知角色'
}

/**
 * 创建认证状态监听器
 */
export function createAuthWatcher(callback: (state: AuthState) => void): () => void {
  return authStore.subscribe(callback)
}

/**
 * 等待认证状态初始化
 */
export function waitForAuthInit(timeout = 5000): Promise<AuthState> {
  return new Promise((resolve, reject) => {
    let timeoutId: NodeJS.Timeout | null = null

    const unsubscribe = authStore.subscribe(state => {
      // 假设初始化完成的标志是token存在或明确为null
      if (state.accessToken !== undefined) {
        if (timeoutId) clearTimeout(timeoutId)
        unsubscribe()
        resolve(state)
      }
    })

    // 设置超时处理
    timeoutId = setTimeout(() => {
      unsubscribe()
      reject(new Error('认证状态初始化超时'))
    }, timeout)
  })
}

/**
 * 创建类型安全的认证状态监听器
 */
export function createTypedAuthWatcher<T>(
  selector: (state: AuthState) => T,
  callback: (value: T, previousValue?: T) => void
): () => void {
  let previousValue: T | undefined

  return authStore.subscribe(state => {
    const currentValue = selector(state)
    if (currentValue !== previousValue) {
      callback(currentValue, previousValue)
      previousValue = currentValue
    }
  })
}

/**
 * 批量检查认证状态
 */
export interface AuthStateCheck {
  isAuthenticated: boolean
  isTokenValid: boolean
  isUserActive: boolean
  userRole: UserRole
  canAccessAdmin: AuthCheckResult
}

export function getAuthStateCheck(user?: User | null): AuthStateCheck {
  return {
    isAuthenticated: isAuthenticated(),
    isTokenValid: isTokenValid(),
    isUserActive: isUserActive(user),
    userRole: getCurrentUserRole(),
    canAccessAdmin: canAccessAdmin(user)
  }
}
