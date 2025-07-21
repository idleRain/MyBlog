import ky, { type AfterResponseHook, type BeforeRequestHook } from 'ky'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

// 请求拦截器 - 添加认证token
const requestInterceptor: BeforeRequestHook = (request, options) => {
  if (browser) {
    const token = authStore.getToken()
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
      authStore.logout()
      await goto('/login')
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
