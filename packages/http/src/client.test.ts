// HTTP 请求器认证刷新与重放行为的单元测试。
// 覆盖 401 之后"刷新一次、重放一次、再失败即判定认证失效"的完整语义。
import { createHttpClient, type HttpClientAuthHooks } from './client'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AfterResponseHook } from 'ky'

// 响应拦截器的归一化选项类型，从钩子签名推导以避免依赖 ky 未导出的内部类型。
type ResponseHookOptions = Parameters<AfterResponseHook>[1]

// 测试只关心请求、选项与响应三个参数。
// ky 的钩子签名还带第四个内部状态参数，本实现未使用该参数，因此按三参数签名收窄以便直接调用。
type TestableResponseHook = (
  request: Request,
  options: ResponseHookOptions,
  response: Response
) => ReturnType<AfterResponseHook>

// ky 替身：捕获 createHttpClient 注册的钩子，并记录每一次重放调用。
const kyMock = vi.hoisted(() => {
  const state = {
    hooks: null as { afterResponse: unknown[] } | null,
    replayCalls: [] as Array<[Request, ResponseHookOptionsLike]>
  }

  const ky = Object.assign(
    vi.fn((request: Request, options: ResponseHookOptionsLike) => {
      state.replayCalls.push([request, options])
      return Promise.resolve(new Response('{}'))
    }),
    {
      create: vi.fn((options: { hooks: { afterResponse: unknown[] } }) => {
        state.hooks = options.hooks
        return { mockedClient: true }
      })
    }
  )

  return { state, ky }
})

// 重放选项在替身中仅按普通对象记录，断言时再收窄为钩子签名对应的类型。
type ResponseHookOptionsLike = Record<string, unknown> & { context?: Record<string, unknown> }

vi.mock('ky', () => ({ default: kyMock.ky }))

// 后端认证失败以 HTTP 200 加业务码表达，此处复刻统一响应信封。
function envelopeResponse(body: { code: number; message?: string }): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' }
  })
}

// 取回 createHttpClient 注册的响应拦截器。
function getResponseHook(): TestableResponseHook {
  const hooks = kyMock.state.hooks
  if (!hooks) throw new Error('响应拦截器尚未注册')
  return hooks.afterResponse[0] as TestableResponseHook
}

// 带调用记录的认证回调集合。
interface AuthSpies {
  hooks: HttpClientAuthHooks
  refreshToken: ReturnType<typeof vi.fn>
  onAuthFailure: ReturnType<typeof vi.fn>
}

// 构造认证回调替身，刷新结果可配置。
function createAuthSpies(refreshResult: string | null = 'fresh-token'): AuthSpies {
  const refreshToken = vi.fn(async () => refreshResult)
  const onAuthFailure = vi.fn(async (_message?: string) => {})

  return {
    hooks: { getAccessToken: () => 'stale-token', refreshToken, onAuthFailure },
    refreshToken,
    onAuthFailure
  }
}

// 以统一响应信封触发一次业务码 401，返回被拒的请求与拦截器返回值。
async function triggerAuthFailure(
  auth: HttpClientAuthHooks,
  init: { url?: string; body?: { code: number; message?: string } } = {}
) {
  createHttpClient({ prefixUrl: 'http://localhost/api', auth })

  const request = new Request(init.url ?? 'http://localhost/api/articles/list', {
    method: 'POST',
    headers: { Authorization: 'Bearer stale-token' }
  })
  const options = { context: {} } as unknown as ResponseHookOptions
  const result = await getResponseHook()(
    request,
    options,
    envelopeResponse(init.body ?? { code: 401 })
  )

  return { request, result }
}

beforeEach(() => {
  kyMock.state.hooks = null
  kyMock.state.replayCalls.length = 0
})

describe('401 刷新与重放', () => {
  it('刷新返回新令牌时携带新令牌重放原请求', async () => {
    const auth = createAuthSpies('fresh-token')

    const { request, result } = await triggerAuthFailure(auth.hooks)

    expect(auth.refreshToken).toHaveBeenCalledTimes(1)
    expect(kyMock.state.replayCalls).toHaveLength(1)
    expect(request.headers.get('Authorization')).toBe('Bearer fresh-token')
    expect(result).toBeInstanceOf(Response)
    expect(auth.onAuthFailure).not.toHaveBeenCalled()
  })

  it('刷新回调返回与原令牌相同的令牌时仍然重放一次', async () => {
    // 刷新回调在本地认为令牌仍有效时会返回原令牌，重放守卫不得据此判定为已重放过。
    const auth = createAuthSpies('stale-token')

    await triggerAuthFailure(auth.hooks)

    expect(kyMock.state.replayCalls).toHaveLength(1)
    expect(auth.onAuthFailure).not.toHaveBeenCalled()
  })

  it('重放后仍返回 401 时不再重放并触发认证失效', async () => {
    const auth = createAuthSpies('fresh-token')

    const { request } = await triggerAuthFailure(auth.hooks)
    const replayedOptions = kyMock.state.replayCalls[0]?.[1] as unknown as ResponseHookOptions

    await getResponseHook()(request, replayedOptions, envelopeResponse({ code: 401 }))

    expect(auth.refreshToken).toHaveBeenCalledTimes(1)
    expect(kyMock.state.replayCalls).toHaveLength(1)
    expect(auth.onAuthFailure).toHaveBeenCalledTimes(1)
  })

  it('刷新失败时不重放并触发认证失效', async () => {
    const auth = createAuthSpies(null)

    await triggerAuthFailure(auth.hooks)

    expect(kyMock.state.replayCalls).toHaveLength(0)
    expect(auth.onAuthFailure).toHaveBeenCalledTimes(1)
  })
})

describe('401 语义边界', () => {
  it('登录请求返回 401 时不刷新也不触发认证失效', async () => {
    const auth = createAuthSpies()

    await triggerAuthFailure(auth.hooks, { url: 'http://localhost/api/users/login' })

    expect(auth.refreshToken).not.toHaveBeenCalled()
    expect(auth.onAuthFailure).not.toHaveBeenCalled()
  })

  it('刷新请求自身返回 401 时不重放并触发认证失效', async () => {
    const auth = createAuthSpies()

    await triggerAuthFailure(auth.hooks, { url: 'http://localhost/api/auth/refresh' })

    expect(auth.refreshToken).not.toHaveBeenCalled()
    expect(kyMock.state.replayCalls).toHaveLength(0)
    expect(auth.onAuthFailure).toHaveBeenCalledTimes(1)
  })

  it('HTTP 层失败不触发刷新并提示错误', async () => {
    const auth = createAuthSpies()
    const onError = vi.fn()

    createHttpClient({ prefixUrl: 'http://localhost/api', auth: auth.hooks, onError })
    const request = new Request('http://localhost/api/articles/list', { method: 'POST' })
    const options = { context: {} } as unknown as ResponseHookOptions

    await getResponseHook()(
      request,
      options,
      new Response(JSON.stringify({ code: 500, message: '服务器内部错误' }), { status: 500 })
    )

    expect(auth.refreshToken).not.toHaveBeenCalled()
    expect(auth.onAuthFailure).not.toHaveBeenCalled()
    expect(onError).toHaveBeenCalledWith('服务器内部错误')
  })
})
