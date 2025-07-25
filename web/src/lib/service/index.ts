import ky, { type AfterResponseHook, type BeforeRequestHook } from 'ky'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

// 令牌刷新状态
let isRefreshing = false
let refreshPromise: Promise<boolean> | null = null

// 刷新令牌函数
async function refreshTokenIfNeeded(): Promise<boolean> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  if (!authStore.shouldRefreshToken()) {
    return true
  }

  const refreshToken = authStore.getRefreshToken()
  if (!refreshToken) {
    authStore.logout()
    if (browser) {
      await goto('/login')
    }
    return false
  }

  isRefreshing = true
  refreshPromise = performTokenRefresh(refreshToken)

  try {
    const success = await refreshPromise
    return success
  } finally {
    isRefreshing = false
    refreshPromise = null
  }
}

// 执行令牌刷新
async function performTokenRefresh(refreshToken: string): Promise<boolean> {
  try {
    const response = await ky
      .post(prefixUrl + 'auth/refresh', {
        json: { refreshToken: refreshToken },
        timeout: 10000
      })
      .json<{
        code: number
        data: {
          accessToken: string
          refreshToken: string
          expiresIn: number
        }
      }>()

    if (response.code === 200) {
      authStore.updateTokens(
        response.data.accessToken,
        response.data.refreshToken,
        response.data.expiresIn
      )
      return true
    }
  } catch (error) {
    console.error('令牌刷新失败:', error)
  }

  // 刷新失败，清除认证状态
  authStore.logout()
  if (browser) {
    await goto('/login')
  }
  return false
}

// 请求拦截器 - 添加认证token和处理令牌刷新
const requestInterceptor: BeforeRequestHook = async (request, options) => {
  if (browser) {
    // 检查是否需要刷新令牌
    if (authStore.shouldRefreshToken()) {
      await refreshTokenIfNeeded()
    }

    const token = authStore.getAccessToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
  }
}

// 响应拦截器 - 处理认证失败
const responseInterceptor: AfterResponseHook = async (request, options, response) => {
  // 处理401未认证错误
  if (response.status === 401) {
    if (browser) {
      // 尝试刷新令牌
      const refreshSuccess = await refreshTokenIfNeeded()
      if (!refreshSuccess) {
        authStore.logout()
        await goto('/login')
      }
    }
  }

  // 处理其他错误
  if (!response.ok) {
    try {
      const errorData = await response.json()
      console.error('API Error:', errorData)
    } catch (e) {
      console.error('API Error:', response.statusText)
    }
  }
}

const request = ky.create({
  prefixUrl,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT,
  headers: {
    'Content-Type': 'application/json'
  },
  hooks: {
    // 请求前置钩子
    beforeRequest: [requestInterceptor],
    // 请求后置钩子
    afterResponse: [responseInterceptor]
  }
})

export default request
