// HTTP 客户端实例：基于 @myblog/http 工厂创建，认证逻辑在此注入。
// 此文件是应用层与请求器之间的适配层，负责接入认证 store 与界面提示。

import type { RefreshTokenData } from '@myblog/api/modules/user/types'
import { createTokenRefresher } from '@myblog/auth'
import { createHttpClient } from '@myblog/http'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { toast } from 'svelte-sonner'
import ky from 'ky'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

// 刷新请求超时上限，超时按刷新失败处理并交由 onAuthFailure 引导重新登录。
const REFRESH_TIMEOUT_MS = 10000

/**
 * 向认证端点发起一次真实的令牌刷新请求，失败返回 null。
 * 此处以裸 ky 直连该端点，豁免 A1 禁止页面直连 ky 的规则，避免请求器与认证流程互相依赖形成循环。
 */
async function requestTokenPair(refreshToken: string): Promise<RefreshTokenData | null> {
  try {
    const response = await ky
      .post(prefixUrl + '/auth/refresh', {
        json: { refreshToken },
        timeout: REFRESH_TIMEOUT_MS,
        retry: 0
      })
      .json<{ code: number; message: string; data: RefreshTokenData }>()

    if (response.code !== 200) {
      throw new Error(response.message || '刷新令牌失败')
    }

    return response.data
  } catch (error) {
    console.error('令牌刷新失败:', error)
    return null
  }
}

// 刷新编排由 @myblog/auth 统一提供，覆盖其他标签页已完成旋转与并发失败后重新对齐两种情形。
const refreshAccessToken = createTokenRefresher({ store: authStore, requestTokenPair })

const request = createHttpClient({
  prefixUrl,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  // 管理端请求全量翻译包，编辑表单需要各语言内容；后台界面本身不提供多语言。
  getLanguage: () => '*',
  auth: {
    // 从认证 store 读取当前访问令牌。
    getAccessToken: () => {
      const state = authStore.getCurrentState()
      return state.isAuthenticated ? state.accessToken : null
    },

    // 本回调只在服务端已拒绝访问令牌时被调用，因此必须真正执行刷新。
    // 若以本地有效期判断短路成返回旧令牌，请求会带着同一枚失效令牌重放并直接登出。
    refreshToken: refreshAccessToken,

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
