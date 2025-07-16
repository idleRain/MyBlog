import ky, { type AfterResponseHook, type BeforeRequestHook } from 'ky'

const prefixUrl = import.meta.env.SSR
  ? import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL
  : import.meta.env.VITE_BASE_URL

// 请求拦截器
const requestInterceptor: BeforeRequestHook = (request, options) => {}

// 响应拦截器
const responseInterceptor: AfterResponseHook = async (request, options, response) => {}

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
