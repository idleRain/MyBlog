import {
  isApiSuccess,
  extractApiData,
  safeExtractApiData,
  normalizeError,
  type BaseApiResponse
} from './response.ts'
import ky, { type AfterResponseHook, type BeforeRequestHook, type Options } from 'ky'

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
   * 语言回调，返回值经 Accept-Language 请求头携带，返回 null 时省略请求头。
   * 前台应用注入当前界面语言，管理端可注入通配符表示全量翻译包。
   */
  getLanguage?: () => string | null | Promise<string | null>

  /**
   * 全局错误提示回调，例如 toast。
   */
  onError?: (message: string) => void
}

// 令牌刷新与登录请求的路径标识，用于排除不应触发自动刷新的请求。
const REFRESH_PATH = '/auth/refresh'
const LOGIN_PATH = '/users/login'

// Authorization 请求头名称。
const AUTHORIZATION_HEADER = 'Authorization'

// 请求上下文标记键，记录该请求已携带刷新后的令牌重放过一次。
// 以显式标记判定远比比对令牌代际可靠：刷新回调在本地认为令牌仍有效时会返回原令牌，
// 比对代际会把这种情况误判为"已重放过"，从而跳过刷新直接登出。
const REPLAYED_CONTEXT_KEY = 'myblogReplayed'

// 内容语言协商请求头名称，遵循 HTTP 标准 Accept-Language 语义。
const ACCEPT_LANGUAGE_HEADER = 'Accept-Language'

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
  const { prefixUrl, timeout = 30000, auth, getLanguage, onError } = options

  // 请求拦截器：为请求附加访问令牌与内容语言标识。
  const requestInterceptor: BeforeRequestHook = async request => {
    if (getLanguage) {
      const language = await getLanguage()
      if (language) {
        request.headers.set(ACCEPT_LANGUAGE_HEADER, language)
      }
    }
    if (!auth) return
    const token = await auth.getAccessToken()
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
  }

  // 响应拦截器：以响应体业务码识别认证失效，处理令牌刷新、请求重放与通用错误提示。
  const responseInterceptor: AfterResponseHook = async (request, options, response) => {
    const { code, message } = await parseResponseBody(response)
    const isAuthFailure = code === 401
    const isRefreshRequest = request.url.includes(REFRESH_PATH)
    const isLoginRequest = request.url.includes(LOGIN_PATH)
    // ky 确保 context 恒为对象，可直接按键读取。
    const hasReplayed = options.context[REPLAYED_CONTEXT_KEY] === true

    if (isAuthFailure && auth && !isLoginRequest) {
      try {
        if (!isRefreshRequest && !hasReplayed) {
          const newToken = await auth.refreshToken()
          if (newToken) {
            request.headers.set(AUTHORIZATION_HEADER, `Bearer ${newToken}`)
            // ky 钩子收到的归一化选项已剥离 hooks，重放需显式回传请求与响应拦截器；
            // 内建重试关闭，嵌套调用自身的 401 由其响应拦截器处理。
            // 归一化选项与 Options 在 exactOptionalPropertyTypes 下可选属性类型存在差异，运行时结构一致，此处收窄为 Options。
            const replayOptions = {
              ...options,
              retry: 0,
              context: { ...options.context, [REPLAYED_CONTEXT_KEY]: true },
              hooks: {
                beforeRequest: [requestInterceptor],
                afterResponse: [responseInterceptor]
              }
            } as Options
            return ky(request, replayOptions)
          }
        }

        // 刷新失败、刷新请求自身 401，或重放过一次后仍被拒绝时触发认证失效处理。
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
    // 重试显式关闭：后端业务接口一律使用 POST 且多数为非幂等写。
    // 创建类请求重复提交会产生重复数据，点赞收藏关注为切换语义，重复提交会翻转状态。
    // 自动重试在服务端已处理但响应丢失的场景下必然重放请求体，因此不允许在传输层自动重试。
    // 认证失效后的重放由响应拦截器携带新令牌显式重发承担，与传输层重试无关。
    retry: { limit: 0 }
  })

  return client
}

export { isApiSuccess, extractApiData, safeExtractApiData, normalizeError }
export type { BaseApiResponse }
export type { Options as KyOptions }
