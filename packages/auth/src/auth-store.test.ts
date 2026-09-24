// 认证 store 的登录持久化、单飞续期调度与跨标签页同步单元测试。
import { afterEach, describe, expect, it, vi } from 'vitest'
import { get } from 'svelte/store'
import type { User } from '@myblog/api/modules/user/types'
import { createAuthStore, type AuthStoreDeps } from './auth-store'
import { installFakeBrowserEnv, AUTH_USER_KEY, UNRELATED_KEY } from './test-fixtures'

// 构造注入依赖的最小替身，浏览器判定恒为 false 以隔离 localStorage。
function createTestStore(overrides?: Partial<AuthStoreDeps>) {
  const logoutApi = vi.fn(async () => {})
  const deps: AuthStoreDeps = {
    isBrowser: () => false,
    logoutApi,
    ...overrides
  }
  return { store: createAuthStore(deps), logoutApi }
}

// 构造接入浏览器环境替身的 store，用于持久化与跨标签页同步相关用例。
function createBrowserStore() {
  const env = installFakeBrowserEnv()
  const logoutApi = vi.fn(async () => {})
  const store = createAuthStore({
    isBrowser: () => true,
    logoutApi
  })
  return { store, env, logoutApi }
}

// 一份可用的测试会话，字段集与后端 UserResponse 形状一致。
const TEST_USER: User = {
  id: 7,
  username: 'tester',
  email: 'tester@example.com',
  nickname: 'tester',
  avatar: '',
  birthday: null,
  role: 'user',
  status: 1,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z'
}
const TEST_PERMISSIONS = ['article:read', 'comment:create']

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('登录与持久化', () => {
  it('登录后记录用户与权限并持久化用户信息', () => {
    const { store } = createBrowserStore()

    store.login(TEST_USER, TEST_PERMISSIONS)

    const state = get(store)
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.id).toBe(7)
    expect(state.permissions).toEqual(TEST_PERMISSIONS)

    const persisted = JSON.parse(window.localStorage.getItem(AUTH_USER_KEY) ?? '{}')
    expect(persisted.username).toBe('tester')
  })

  it('浏览器环境下凭已落盘的用户信息恢复登录态', () => {
    const env = installFakeBrowserEnv()
    env.seedSession({ userId: 3, username: 'user3', permissions: ['article:read'] })

    const store = createAuthStore({
      isBrowser: () => true,
      logoutApi: async () => {}
    })

    const state = get(store)
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.id).toBe(3)
    expect(state.permissions).toEqual(['article:read'])
  })

  it('updateUser 更新状态并同步本地缓存', () => {
    const { store } = createBrowserStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    store.updateUser({ ...TEST_USER, username: 'renamed' })

    const state = get(store)
    expect(state.user?.username).toBe('renamed')
    const persisted = JSON.parse(window.localStorage.getItem(AUTH_USER_KEY) ?? '{}')
    expect(persisted.username).toBe('renamed')
  })
})

describe('登出语义', () => {
  it('登出先调用服务端撤销接口再清除本地状态', async () => {
    const { store, logoutApi, env } = createBrowserStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    await store.logout()

    expect(logoutApi).toHaveBeenCalledTimes(1)
    const state = get(store)
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(window.localStorage.getItem(AUTH_USER_KEY)).toBeNull()
    expect(env).toBeDefined()
  })

  it('skipApiCall 为真时跳过服务端撤销仅清除本地状态', async () => {
    const { store, logoutApi } = createTestStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    await store.logout(true)

    expect(logoutApi).not.toHaveBeenCalled()
    expect(get(store).isAuthenticated).toBe(false)
  })

  it('clearLocalState 清除状态但不触发服务端调用', async () => {
    const { store, logoutApi } = createTestStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    store.clearLocalState()

    expect(logoutApi).not.toHaveBeenCalled()
    expect(get(store).isAuthenticated).toBe(false)
  })
})

describe('单飞续期调度', () => {
  it('并发触发续期时共享同一次请求', async () => {
    const { store } = createTestStore()
    const refreshFn = vi.fn(async () => {
      // 引入异步间隙，模拟真实的网络续期耗时窗口。
      await new Promise(resolve => setTimeout(resolve, 10))
      return 'refreshed'
    })

    const [first, second] = await Promise.all([
      store.refreshSingleFlight(refreshFn),
      store.refreshSingleFlight(refreshFn)
    ])

    expect(refreshFn).toHaveBeenCalledTimes(1)
    expect(first).toBe('refreshed')
    expect(second).toBe('refreshed')
  })

  it('前一次续期完成后再次触发会发起新请求', async () => {
    const { store } = createTestStore()
    const refreshFn = vi.fn(async () => 'refreshed')

    await store.refreshSingleFlight(refreshFn)
    await store.refreshSingleFlight(refreshFn)

    expect(refreshFn).toHaveBeenCalledTimes(2)
  })
})

describe('跨标签页同步', () => {
  it('跨标签页写入用户信息后同步对齐状态', () => {
    const { store, env } = createBrowserStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    env.seedSession({ userId: 9, username: 'other-tab' })
    env.dispatchStorage(AUTH_USER_KEY)

    const state = get(store)
    expect(state.user?.id).toBe(9)
  })

  it('与认证无关的存储变更不触发状态重载', () => {
    const { store, env } = createBrowserStore()
    store.login(TEST_USER, TEST_PERMISSIONS)
    const before = get(store)

    env.seedRaw(UNRELATED_KEY, 'whatever')
    env.dispatchStorage(UNRELATED_KEY)

    expect(get(store)).toEqual(before)
  })

  it('其他标签页登出后本页跟随清除登录态', () => {
    const { store, env } = createBrowserStore()
    store.login(TEST_USER, TEST_PERMISSIONS)

    env.clearSession()
    env.dispatchStorage(null)

    expect(get(store).isAuthenticated).toBe(false)
  })
})

describe('权限查询', () => {
  it('按后端下发的权限列表判断', () => {
    const { store } = createTestStore()
    store.login(TEST_USER, ['article:read'])

    expect(store.hasPermission('article:read')).toBe(true)
    expect(store.hasPermission('article:delete')).toBe(false)
    expect(store.getPermissions()).toEqual(['article:read'])
  })
})
