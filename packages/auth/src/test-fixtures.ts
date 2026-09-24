// 认证域测试共享的浏览器环境替身。
// 认证 store 与令牌刷新编排都依赖 window.localStorage 与 storage 事件，
// 测试必须在创建 store 之前完成安装，否则跨标签页同步逻辑不会注册。

import { vi } from 'vitest'

// 持久化键名与 auth-store.ts 保持一致。
// 测试需要绕过 store 直接读写底层存储以模拟其他标签页的落盘结果，因此在此显式列出。
export const AUTH_TOKEN_KEY = 'auth_access_token'
export const AUTH_REFRESH_KEY = 'auth_refresh_token'
export const AUTH_USER_KEY = 'auth_user'
export const AUTH_EXPIRES_KEY = 'auth_expires_at'

// 与认证无关的存储键，用于验证同步逻辑按需触发而非每次存储变更都重载。
export const UNRELATED_KEY = 'unrelated_preference'

/** 一次认证会话的最小描述，用于写入底层存储。 */
export interface SeededSession {
  accessToken: string
  refreshToken: string
  userId: number
  // 相对当前时刻的剩余有效期，单位毫秒。
  expiresInMs: number
}

type StorageListener = (event: { key: string | null }) => void

/** 浏览器环境替身的操作句柄。 */
export interface FakeBrowserEnv {
  /** 写入一份完整认证会话，模拟任一标签页的落盘结果。 */
  seedSession: (session: SeededSession) => void
  /** 写入任意键值，模拟与认证无关的存储变更。 */
  seedRaw: (key: string, value: string) => void
  /** 移除全部认证键，模拟其他标签页登出后的落盘结果。 */
  clearSession: () => void
  /** 派发一次 storage 事件，模拟其他标签页写入后的跨页通知。 */
  dispatchStorage: (key: string | null) => void
}

/**
 * 安装 window 替身并返回操作句柄。
 * 调用方需在用例结束后执行 vi.unstubAllGlobals() 卸载。
 */
export function installFakeBrowserEnv(): FakeBrowserEnv {
  const entries = new Map<string, string>()
  const listeners = new Set<StorageListener>()

  const fakeWindow = {
    localStorage: {
      getItem: (key: string) => entries.get(key) ?? null,
      setItem: (key: string, value: string) => void entries.set(key, value),
      removeItem: (key: string) => void entries.delete(key),
      clear: () => entries.clear()
    },
    addEventListener: (type: string, listener: StorageListener) => {
      if (type === 'storage') listeners.add(listener)
    },
    removeEventListener: (type: string, listener: StorageListener) => {
      if (type === 'storage') listeners.delete(listener)
    }
  }

  vi.stubGlobal('window', fakeWindow)

  return {
    seedSession(session) {
      // 令牌以裸字符串落盘，其余字段按 JSON 序列化，与 @myblog/shared 的 local 封装行为一致。
      entries.set(AUTH_TOKEN_KEY, session.accessToken)
      entries.set(AUTH_REFRESH_KEY, session.refreshToken)
      entries.set(
        AUTH_USER_KEY,
        JSON.stringify({ id: session.userId, username: `user${session.userId}` })
      )
      entries.set(AUTH_EXPIRES_KEY, JSON.stringify(Date.now() + session.expiresInMs))
    },

    seedRaw(key, value) {
      entries.set(key, value)
    },

    clearSession() {
      for (const key of [AUTH_TOKEN_KEY, AUTH_REFRESH_KEY, AUTH_USER_KEY, AUTH_EXPIRES_KEY]) {
        entries.delete(key)
      }
    },

    dispatchStorage(key) {
      for (const listener of listeners) {
        listener({ key })
      }
    }
  }
}
