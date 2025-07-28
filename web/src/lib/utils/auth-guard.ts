// 认证守卫工具

import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { get } from 'svelte/store'

/**
 * 检查用户是否已认证
 */
export function isAuthenticated(): boolean {
  const state = authStore.getCurrentState()
  return state.isAuthenticated && !!state.user && authStore.isTokenValid()
}

/**
 * 要求用户登录
 * 如果用户未登录，跳转到登录页面
 */
export async function requireAuth(redirectTo: string = '/login'): Promise<boolean> {
  if (!browser) return true // SSR 时不检查

  if (!isAuthenticated()) {
    console.log('用户未认证，跳转到登录页面')
    await goto(redirectTo)
    return false
  }

  return true
}

/**
 * 要求用户未登录
 * 如果用户已登录，跳转到首页或指定页面
 */
export async function requireGuest(redirectTo: string = '/'): Promise<boolean> {
  if (!browser) return true // SSR 时不检查

  if (isAuthenticated()) {
    console.log('用户已认证，跳转到首页')
    await goto(redirectTo)
    return false
  }

  return true
}

/**
 * 要求特定角色
 * 检查用户是否具有指定角色或更高权限
 */
export async function requireRole(
  requiredRole: string,
  redirectTo: string = '/403'
): Promise<boolean> {
  if (!browser) return true // SSR 时不检查

  if (!isAuthenticated()) {
    await goto('/login')
    return false
  }

  const state = authStore.getCurrentState()
  const userRole = state.user?.role

  // 角色权限级别映射
  const roleLevel: Record<string, number> = {
    user: 1,
    editor: 2,
    admin: 3,
    superadmin: 4
  }

  const userLevel = roleLevel[userRole || ''] || 0
  const requiredLevel = roleLevel[requiredRole] || 0

  if (userLevel < requiredLevel) {
    console.log(`用户权限不足，需要 ${requiredRole}，当前 ${userRole}`)
    await goto(redirectTo)
    return false
  }

  return true
}

/**
 * 页面加载时的认证检查
 * 确保在页面完全加载前完成认证检查
 */
export async function checkAuthOnLoad(): Promise<{
  isAuthenticated: boolean
  user: any
  needsRedirect: boolean
  redirectTo?: string
}> {
  if (!browser) {
    return {
      isAuthenticated: false,
      user: null,
      needsRedirect: false
    }
  }

  const state = authStore.getCurrentState()
  const authenticated = state.isAuthenticated && !!state.user

  // 检查 token 是否有效
  if (authenticated && !authStore.isTokenValid()) {
    // Token 无效，尝试刷新
    try {
      const { refreshAccessToken } = await import('$lib/service')
      const newToken = await refreshAccessToken()

      if (!newToken) {
        // 刷新失败，清除本地状态，需要重新登录
        authStore.clearLocalState()
        return {
          isAuthenticated: false,
          user: null,
          needsRedirect: true,
          redirectTo: '/login'
        }
      }
    } catch (error) {
      console.error('Token 刷新失败:', error)
      authStore.clearLocalState()
      return {
        isAuthenticated: false,
        user: null,
        needsRedirect: true,
        redirectTo: '/login'
      }
    }
  }

  return {
    isAuthenticated: authenticated,
    user: state.user,
    needsRedirect: false
  }
}

/**
 * 创建认证状态响应式监听器
 */
export function createAuthWatcher(callback: (authenticated: boolean) => void) {
  return authStore.subscribe(state => {
    const authenticated = state.isAuthenticated && !!state.user && authStore.isTokenValid()
    callback(authenticated)
  })
}

/**
 * 安全地获取当前用户信息
 */
export function getCurrentUser() {
  if (!isAuthenticated()) return null
  return authStore.getCurrentState().user
}

/**
 * 检查用户是否有指定权限
 */
export function hasPermission(permission: string): boolean {
  const user = getCurrentUser()
  if (!user) return false

  // 根据角色检查权限（简化版，实际项目中可能需要更复杂的权限系统）
  const rolePermissions: Record<string, string[]> = {
    user: ['read'],
    editor: ['read', 'write'],
    admin: ['read', 'write', 'manage'],
    superadmin: ['read', 'write', 'manage', 'admin']
  }

  const userPermissions = rolePermissions[user.role] || []
  return userPermissions.includes(permission)
}

/**
 * 清理认证状态
 * 用于页面卸载或应用关闭时清理
 */
export function cleanupAuth() {
  // 这里可以添加清理逻辑，比如取消定时刷新等
  console.log('清理认证状态')
}
