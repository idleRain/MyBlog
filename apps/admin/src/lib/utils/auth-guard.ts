// 认证守卫工具

import type { User } from '@myblog/api/modules/user/types'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'

/**
 * 检查用户是否已认证。
 * 本地状态仅承载用户信息，会话 Cookie 是否有效由首次请求的服务端裁决，
 * 认证失效时由请求器的 onAuthFailure 统一引导重新登录。
 */
export function isAuthenticated(): boolean {
  const state = authStore.getCurrentState()
  return state.isAuthenticated && !!state.user
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
    console.log('用户已登录，跳转到首页')
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
  user: User | null
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

  const authenticated = isAuthenticated()

  return {
    isAuthenticated: authenticated,
    user: authStore.getCurrentState().user,
    needsRedirect: !authenticated,
    redirectTo: '/login'
  }
}
