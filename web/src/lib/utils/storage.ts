import { browser } from '$app/environment'

type StorageType = {
  set: (key: string, value: any) => void
  get: <T = unknown>(key: string) => T | null
  rm: (key: string) => void
  clear: () => void
  isAvailable: () => boolean
}

/**
 * 创建统一的 Storage 操作实例，支持服务端渲染
 * @param getStorage - 获取 Storage 实例的函数
 */
const createStorage = (getStorage: () => Storage | null): StorageType => ({
  // 检查 Storage 是否可用
  isAvailable: (): boolean => {
    return browser && getStorage() !== null
  },

  // 设置
  set: (key: string, value: any): void => {
    if (!browser) return

    const storage = getStorage()
    if (!storage) return

    if (value === void 0) {
      try {
        storage.setItem(key, JSON.stringify(null))
      } catch (error) {
        console.warn(`Failed to set storage item "${key}"`, error)
      }
      return
    }

    try {
      if (typeof value === 'string') {
        storage.setItem(key, value)
      } else {
        const serialized = JSON.stringify(value)
        storage.setItem(key, serialized)
      }
    } catch (error) {
      console.warn(`Failed to serialize data for key "${key}"`, error)
    }
  },

  // 获取
  get: <T = unknown>(key: string): T | null => {
    if (!browser) return null

    const storage = getStorage()
    if (!storage) return null

    try {
      const item = storage.getItem(key)
      if (item === null) return null

      return JSON.parse(item) as T
    } catch (error) {
      return storage.getItem(key) as T | null
    }
  },

  // 删除
  rm: (key: string): void => {
    if (!browser) return

    const storage = getStorage()
    if (!storage) return

    try {
      storage.removeItem(key)
    } catch (error) {
      console.warn(`Failed to remove storage item "${key}"`, error)
    }
  },

  // 清空
  clear: (): void => {
    if (!browser) return

    const storage = getStorage()
    if (!storage) return

    try {
      storage.clear()
    } catch (error) {
      console.warn('Failed to clear storage', error)
    }
  }
})

/**
 * 获取 localStorage
 */
const getLocalStorage = (): Storage | null => {
  try {
    return window.localStorage
  } catch {
    return null
  }
}

/**
 * 获取 sessionStorage
 */
const getSessionStorage = (): Storage | null => {
  try {
    return window.sessionStorage
  } catch {
    return null
  }
}

/**
 * @Description 对 localStorage 的封装，支持服务端渲染
 * @Author IdleRain
 * @Date 2025/6/18 16:45
 * @UpdateDate 2025/7/25 14:56
 */
export const local = createStorage(getLocalStorage)

/**
 * @Description 对 sessionStorage 的封装，支持服务端渲染
 * @Author IdleRain
 * @Date 2025/6/18 16:45
 * @UpdateDate 2025/7/25 14:56
 */
export const session = createStorage(getSessionStorage)

// 向后兼容的导出
export { createStorage }
