import ky, { type AfterResponseHook, type BeforeRequestHook } from 'ky'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

// 令牌刷新状态管理
let isRefreshing = false
let refreshPromise: Promise<boolean> | null = null
let failedQueue: Array<{
  resolve: (token: string | null) => void
  reject: (error: any) => void
}> = []

// 处理队列中的请求
function processQueue(error: any = null, token: string | null = null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token)
    }
  })

  failedQueue = []
}

// 刷新令牌函数
async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = authStore.getRefreshToken()

  if (!refreshToken) {
    console.warn('没有刷新令牌，无法自动刷新')
    return null
  }

  if (isRefreshing) {
    // 如果正在刷新，将请求加入队列等待
    return new Promise((resolve, reject) => {
      failedQueue.push({ resolve, reject })
    })
  }

  isRefreshing = true

  try {
    console.log('开始刷新访问令牌...')

    const response = await ky
      .post(prefixUrl + 'auth/refresh', {
        json: { refreshToken },
        timeout: 10000,
        retry: 0 // 刷新请求不重试
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

      // 更新存储的令牌
      authStore.updateTokens(accessToken, newRefreshToken, expiresIn)

      console.log('令牌刷新成功')
      processQueue(null, accessToken)

      return accessToken
    } else {
      throw new Error(response.message || '刷新令牌失败')
    }
  } catch (error) {
    console.error('令牌刷新失败:', error)

    // 刷新失败，清除认证状态并跳转到登录页（跳过 API 调用避免循环）
    authStore.clearLocalState()
    processQueue(error, null)

    if (browser) {
      toast.error('登录已过期，请重新登录')
      await goto('/login')
    }

    return null
  } finally {
    isRefreshing = false
  }
}

// 检查并刷新令牌（如果需要）
async function ensureValidToken(): Promise<string | null> {
  const currentState = authStore.getCurrentState()

  if (!currentState.isAuthenticated) {
    return null
  }

  // 如果 token 仍然有效，直接返回
  if (authStore.isTokenValid()) {
    return currentState.accessToken
  }

  // 如果需要刷新，执行刷新
  if (authStore.shouldRefreshToken()) {
    return await refreshAccessToken()
  }

  return currentState.accessToken
}

// 请求拦截器 - 添加认证token
const requestInterceptor: BeforeRequestHook = async (request, options) => {
  if (browser) {
    const token = await ensureValidToken()

    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
  }
}

// 响应拦截器 - 处理认证失败和其他错误
const responseInterceptor: AfterResponseHook = async (request, options, response) => {
  // 处理 401 未认证错误
  if (response.status === 401) {
    console.warn('收到 401 响应，令牌可能无效')

    try {
      // 尝试解析错误信息
      const errorData = await response.clone().json()
      console.error('401 错误详情:', errorData)

      // 检查是否是令牌相关的错误
      const tokenErrors = ['无效的认证令牌', 'token expired', 'invalid token', 'unauthorized']

      const isTokenError = tokenErrors.some(error =>
        errorData.message?.toLowerCase().includes(error.toLowerCase())
      )

      if (isTokenError) {
        // 如果不是在刷新令牌的请求中，尝试刷新令牌
        const isRefreshRequest = request.url.includes('/auth/refresh')

        if (!isRefreshRequest && browser) {
          console.log('尝试刷新令牌...')
          const newToken = await refreshAccessToken()

          if (newToken) {
            // 令牌刷新成功，重新发起原始请求
            console.log('令牌刷新成功，重新发起请求')
            const newRequest = request.clone()
            newRequest.headers.set('Authorization', `Bearer ${newToken}`)

            // 这里不直接返回新响应，让调用方处理重试逻辑
            return response
          }
        } else {
          // 刷新令牌请求失败，或者已经在刷新过程中，清除认证状态
          console.error('刷新令牌失败，清除认证状态')
          authStore.clearLocalState()

          if (browser) {
            toast.error('登录已过期，请重新登录')
            await goto('/login')
          }
        }
      }
    } catch (parseError) {
      console.error('解析 401 错误响应失败:', parseError)

      // 无法解析错误，直接清除认证状态
      authStore.clearLocalState()

      if (browser) {
        toast.error('认证失败，请重新登录')
        await goto('/login')
      }
    }
  }

  // 处理其他客户端和服务器错误
  if (!response.ok && response.status !== 401) {
    try {
      const errorData = await response.clone().json()
      console.error(`HTTP ${response.status} 错误:`, errorData)

      // 显示用户友好的错误消息
      if (browser && errorData.message) {
        toast.error(errorData.message)
      }
    } catch (parseError) {
      console.error(`HTTP ${response.status} 错误:`, response.statusText)

      if (browser) {
        toast.error(`请求失败: ${response.statusText}`)
      }
    }
  }

  return response
}

// 创建 HTTP 客户端实例
const request = ky.create({
  prefixUrl,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  hooks: {
    beforeRequest: [requestInterceptor],
    afterResponse: [responseInterceptor]
  },
  retry: {
    limit: 2,
    methods: ['get', 'put', 'head', 'delete', 'options', 'trace'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  }
})

export default request

// 导出刷新令牌函数供外部使用
export { refreshAccessToken }
