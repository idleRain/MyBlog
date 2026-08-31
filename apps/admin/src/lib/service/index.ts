// HTTP 客户端实例：基于 @myblog/http 工厂创建，认证逻辑在此注入。
// 此文件是应用层与请求器之间的适配层，负责接入认证 store 与界面提示。

import { createHttpClient } from '@myblog/http'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { toast } from 'svelte-sonner'
import ky from 'ky'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

/**
 * 通过刷新令牌获取新访问令牌并更新认证状态。
 */
async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = authStore.getRefreshToken()
  if (!refreshToken) {
    console.warn('没有刷新令牌，无法自动刷新')
    return null
  }

  try {
    const response = await ky
      .post(prefixUrl + '/auth/refresh', {
        json: { refreshToken },
        timeout: 10000,
        retry: 0
      })
      .json<{
        code: number
        message: string
        data: {
          accessToken: string
          refreshToken: string
          expiresIn: number
        }
      }>()

    if (response.code === 200) {
      const { accessToken, refreshToken: newRefreshToken, expiresIn } = response.data
      authStore.updateTokens(accessToken, newRefreshToken, expiresIn)
      return accessToken
    }

    throw new Error(response.message || '刷新令牌失败')
  } catch (error) {
    console.error('令牌刷新失败:', error)
    return null
  }
}

const request = createHttpClient({
  prefixUrl,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  auth: {
    // 从认证 store 读取当前访问令牌。
    getAccessToken: () => {
      const state = authStore.getCurrentState()
      return state.isAuthenticated ? state.accessToken : null
    },

    // 令牌有效时直接返回，接近过期时执行刷新。
    refreshToken: async () => {
      const currentState = authStore.getCurrentState()
      if (!currentState.isAuthenticated) return null
      if (authStore.isTokenValid()) return currentState.accessToken
      if (authStore.shouldRefreshToken()) {
        return await refreshAccessToken()
      }
      return currentState.accessToken
    },

    // 认证失效时清除状态并跳转登录页。
    onAuthFailure: async message => {
      authStore.clearLocalState()
      if (browser) {
        toast.error(message || '登录已过期，请重新登录')
        await goto('/login')
      }
    }
  },
  onError: message => {
    if (browser) {
      toast.error(message)
    }
  }
})

export default request

// 兼容旧导出：供 request 工具与守卫使用。
export { refreshAccessToken }
