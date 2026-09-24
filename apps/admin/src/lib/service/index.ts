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

// 会话续期请求超时上限，超时按续期失败处理并交由 onAuthFailure 引导重新登录。
const SESSION_TIMEOUT_MS = 10000

/**
 * 发起一次会话续期请求，服务端旋转令牌对并重新写入 Cookie，成功返回 true。
 * 此处以裸 ky 直连该端点，豁免 A1 禁止页面直连 ky 的规则，避免请求器与认证流程互相依赖形成循环。
 * 会话 Cookie 由浏览器自动携带，请求体省略时服务端从 Cookie 读取刷新令牌完成旋转；
 * 旋转结果写入共享的浏览器 Cookie 存储，天然规避多标签页并发旋转的令牌竞争。
 */
async function refreshSession(): Promise<boolean> {
  try {
    const response = await ky
      .post(prefixUrl + '/auth/session', {
        json: {},
        timeout: SESSION_TIMEOUT_MS,
        retry: 0
      })
      .json<{ code: number; message: string }>()

    return response.code === 200
  } catch (error) {
    console.error('会话续期失败:', error)
    return false
  }
}

const request = createHttpClient({
  prefixUrl,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000,
  // 管理端请求全量翻译包，编辑表单需要各语言内容；后台界面本身不提供多语言。
  getLanguage: () => '*',
  auth: {
    // 本回调只在服务端拒绝当前会话时被调用，因此必须真正执行续期；
    // 并发 401 经单飞调度共享同一次续期请求。
    refreshSession: () => authStore.refreshSingleFlight(refreshSession),

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
