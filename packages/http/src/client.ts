import ky, { type AfterResponseHook, type BeforeRequestHook, type Options } from 'ky'
import {
  isApiSuccess,
  extractApiData,
  safeExtractApiData,
  normalizeError,
  type BaseApiResponse
} from './response.ts'

/**
 * HTTP 客户端认证相关回调集合。
 * 应用层负责实现令牌读取、刷新与失效处理，避免请求器依赖框架或 store。
 */
export interface HttpClientAuthHooks {
  /**
   * 获取当前访问令牌，无令牌时返回 null。
   */
  getAccessToken: () => string | null | Promise<string | null>

  /**
   * 刷新访问令牌，成功返回新令牌，失败返回 null。
   */
  refreshToken: () => Promise<string | null>

  /**
   * 认证失效处理回调，例如跳转登录页并提示用户。
   */
  onAuthFailure?: (message?: string) => void | Promise<void>
}

/**
 * HTTP 客户端创建参数。
 */
export interface CreateHttpClientOptions {
  /**
   * 请求前缀，服务端使用完整代理地址，客户端使用相对路径。
   */
  prefixUrl: string

  /**
   * 请求超时时间，默认 30000 毫秒。
   */
  timeout?: number

  /**
   * 认证相关回调。
   */
  auth?: HttpClientAuthHooks

  /**
   * 全局错误提示回调，例如 toast。
   */
  onError?: (message: string) => void
}

// 令牌刷新与登录请求的路径标识，用于排除不应触发自动刷新的请求。
const REFRESH_PATH = '/auth/refresh'
const LOGIN_PATH = '/users/login'

/**
 * 解析响应体中的业务码与消息，解析失败时返回空值。
 */
async function parseResponseBody(
  response: Response
): Promise<{ code: number | undefined; message: string | undefined }> {
  try {
    const body = (await response.clone().json()) as {
      code?: number
      message?: string
    }
    return { code: body.code, message: body.message }
  } catch {
    return { code: undefined, message: undefined }
  }
}

/**
 * 创建带认证刷新、超时与错误提示的 HTTP 客户端。
 */
export function createHttpClient(options: CreateHttpClientOptions) {
  const { prefixUrl, timeout = 30000, auth, onError } = options

  // 请求拦截器：为请求附加访问令牌。
  const requestInterceptor: BeforeRequestHook = async request => {
    if (!auth) return
    const token = await auth.getAccessToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
  }

  // 响应拦截器：以响应体业务码识别认证失效，处理令牌刷新与通用错误提示。
  const responseInterceptor: AfterResponseHook = async (request, _options, response) => {
    const { code, message } = await parseResponseBody(response)
    const isAuthFailure = code === 401
    const isRefreshRequest = request.url.includes(REFRESH_PATH)
    const isLoginRequest = request.url.includes(LOGIN_PATH)

    if (isAuthFailure && auth && !isLoginRequest) {
      try {
        if (!isRefreshRequest) {
          const newToken = await auth.refreshToken()
          if (newToken) {
            // 令牌刷新成功后返回原响应，调用方依据新令牌重试。
            return response
          }
        }

        // 刷新失败或刷新请求自身 401，触发认证失效处理。
        await auth.onAuthFailure?.(message || '登录已过期，请重新登录')
      } catch {
        await auth.onAuthFailure?.('认证失败，请重新登录')
      }
    }

    // 其他错误响应统一提示。
    if (!response.ok && !isAuthFailure) {
      if (message) {
        onError?.(message)
      } else {
        onError?.(`请求失败: ${response.statusText}`)
      }
    }

    return response
  }

  const client = ky.create({
    prefixUrl,
    timeout,
    // 不设置全局 Content-Type，由 ky 按请求体类型自动生成；
    // 否则 FormData 上传会沿用 application/json 而丢失 multipart 边界。
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

  return client
}

export { isApiSuccess, extractApiData, safeExtractApiData, normalizeError }
export type { BaseApiResponse }
export type { Options as KyOptions }
