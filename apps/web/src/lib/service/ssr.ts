// 服务端渲染的请求上下文适配层。
//
// 前台为 SSR 应用，多数页面经 load 在服务端取数。此时请求由 Node 进程发出，
// 浏览器不会自动携带会话 Cookie，后端因此无法识别登录身份，个人化页面只能退化为
// 客户端补拉。本模块把入站请求中的会话 Cookie 转发给后端请求，使服务端取数同样
// 具备登录身份，个人化页面得以回归 load 模式。
//
// 转发范围限定为以 mb_ 为前缀的本站会话 Cookie：后端只解析这两个令牌，
// 转发无关 Cookie 没有收益，反而会把其他站点数据带入后端请求。

import { createHttpClient, type CreateHttpClientOptions } from '@myblog/http'
import type { KyInstance } from 'ky'

// 本站会话 Cookie 的名称前缀，与后端 domain/cookies.go 的常量保持一致。
const SESSION_COOKIE_PREFIX = 'mb_'

// 服务端请求器的基础参数，与应用级客户端保持一致的前缀与超时口径。
const SERVER_REQUEST_BASE_OPTIONS: Pick<CreateHttpClientOptions, 'prefixUrl' | 'timeout'> = {
  prefixUrl: import.meta.env.VITE_PROXY_URL + import.meta.env.VITE_BASE_URL,
  timeout: +import.meta.env.VITE_REQUEST_TIMEOUT || 30000
}

// 后端统一响应信封的成功业务码。
const SUCCESS_CODE = 200

// 后端以业务码 401 表达会话缺失或过期，该判定口径由请求器与页面共同锚定。
const AUTH_FAILURE_CODE = 401

/**
 * 接口工厂签名。
 * 页面传入 @myblog/api 的模块工厂，由本模块以服务端请求器实例化，
 * 使同一次取数使用专属请求器，而应用级单例注册不受影响。
 */
export type ServerApiFactory<T> = (client: KyInstance) => {
  call: () => Promise<{ code: number; message?: string; data?: T }>
}

// 后端统一响应信封中本模块关心的字段。
interface Envelope<T> {
  code: number
  message?: string
  data?: T
}

/**
 * 从入站 Cookie 头中筛出会话相关条目。
 * 入站头缺失或不含会话 Cookie 时返回空字符串。
 */
export function extractSessionCookies(cookieHeader: string | null): string {
  if (!cookieHeader) {
    return ''
  }

  return cookieHeader
    .split(';')
    .map(entry => entry.trim())
    .filter(entry => entry.startsWith(SESSION_COOKIE_PREFIX))
    .join('; ')
}

/**
 * 构造本次请求专属的服务端请求器。
 *
 * 刻意不复用应用级的单例客户端：该单例注册了会话续期与失效跳转回调，
 * 服务端既无法把续期结果写回浏览器 Cookie，也不应触发客户端跳转。
 * 未携带会话 Cookie 时同样返回请求器，由后端按未登录裁决，
 * 使调用方统一经响应码判断登录态，而不依赖前端的提前判定。
 */
export function createServerRequest(cookieHeader: string | null): KyInstance {
  const sessionCookies = extractSessionCookies(cookieHeader)

  return createHttpClient({
    ...SERVER_REQUEST_BASE_OPTIONS,
    ...(sessionCookies ? { headers: { Cookie: sessionCookies } } : {})
  })
}

/**
 * 服务端取数结果。
 * 认证失效以状态标记返回，使页面据此跳转登录页而不必重复判定响应码；
 * 其余业务失败与网络异常统一归一为失败状态。
 */
export type ServerFetchResult<T> =
  { status: 'ok'; data: T } | { status: 'unauthorized' } | { status: 'failed'; message: string }

/**
 * 以指定请求器发起一次受保护接口调用并归一化结果。
 * 页面存在多个端点时可在同一请求器上并发调用本函数。
 */
export async function fetchWithRequest<T>(
  client: KyInstance,
  createApi: ServerApiFactory<T>
): Promise<ServerFetchResult<T>> {
  try {
    const response: Envelope<T> = await createApi(client).call()

    if (response.code === AUTH_FAILURE_CODE) {
      return { status: 'unauthorized' }
    }
    if (response.code !== SUCCESS_CODE || response.data === undefined) {
      return { status: 'failed', message: response.message ?? '' }
    }
    return { status: 'ok', data: response.data }
  } catch {
    return { status: 'failed', message: '' }
  }
}

/**
 * 在服务端发起一次受保护接口调用。
 * 本函数为单端点场景的便捷入口，多端点场景请先取 createServerRequest 再并发调用 fetchWithRequest。
 */
export async function fetchAuthenticated<T>(
  cookieHeader: string | null,
  createApi: ServerApiFactory<T>
): Promise<ServerFetchResult<T>> {
  return fetchWithRequest<T>(createServerRequest(cookieHeader), createApi)
}
