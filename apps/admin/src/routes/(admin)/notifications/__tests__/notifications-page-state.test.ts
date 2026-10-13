import type { Notification } from '@myblog/api/modules/notification/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { NOTIFICATION_PAGE_SIZE } from '$lib/constants/notification'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const notificationApiMock = {
  list: vi.fn(),
  markRead: vi.fn(),
  markAllRead: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ NotificationAPI: notificationApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyNotificationsPageData, notificationsPageState } =
  await import('../notifications-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 构造一条通知替身，字段取值只覆盖被测逻辑关心的部分。
function makeNotification(id: number, isRead = false): Notification {
  return {
    id,
    type: 'comment_reply',
    title: `通知${id}`,
    content: null,
    isRead,
    createdAt: '2026-01-01T00:00:00Z'
  } as unknown as Notification
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  notificationsPageState.initialize(emptyNotificationsPageData())
  notificationsPageState.typeFilter = ''
  notificationsPageState.currentPage = 1
}

describe('通知中心页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表、总数与未读数', () => {
    const notifications = [makeNotification(1), makeNotification(2, true)]

    notificationsPageState.initialize({ notifications, total: 2, unreadCount: 1 })

    expect(notificationsPageState.notifications).toEqual(notifications)
    expect(notificationsPageState.total).toBe(2)
    expect(notificationsPageState.unreadCount).toBe(1)
  })

  it('空数据下无生效筛选条件且总页数至少为 1', () => {
    expect(notificationsPageState.notifications).toEqual([])
    expect(notificationsPageState.isLoading).toBe(false)
    expect(notificationsPageState.hasActiveFilters).toBe(false)
    expect(notificationsPageState.totalPages).toBe(1)
  })
})

describe('通知中心页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带分页参数', async () => {
    notificationApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { notifications: [makeNotification(1)], total: 1, unreadCount: 0 }
    })

    await notificationsPageState.loadNotifications()

    expect(notificationApiMock.list).toHaveBeenCalledWith({
      page: 1,
      pageSize: NOTIFICATION_PAGE_SIZE
    })
    expect(notificationsPageState.notifications).toHaveLength(1)
    expect(notificationsPageState.total).toBe(1)
  })

  it('类型筛选生效时按类型查询', async () => {
    notificationApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { notifications: [], total: 0, unreadCount: 0 }
    })
    notificationsPageState.typeFilter = 'system'
    notificationsPageState.currentPage = 3

    await notificationsPageState.loadNotifications()

    expect(notificationApiMock.list).toHaveBeenCalledWith({
      page: 3,
      pageSize: NOTIFICATION_PAGE_SIZE,
      type: 'system'
    })
  })

  it('响应缺失列表字段时按空列表处理', async () => {
    notificationApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: {} })

    await notificationsPageState.loadNotifications()

    expect(notificationsPageState.notifications).toEqual([])
    expect(notificationsPageState.total).toBe(0)
    expect(notificationsPageState.unreadCount).toBe(0)
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    notificationApiMock.list.mockResolvedValue({ code: 500, message: '查询失败' })
    notificationsPageState.initialize({
      notifications: [makeNotification(1)],
      total: 1,
      unreadCount: 1
    })

    await notificationsPageState.loadNotifications()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(notificationsPageState.notifications).toHaveLength(1)
    expect(notificationsPageState.isLoading).toBe(false)
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    notificationApiMock.list.mockRejectedValue(new Error('断网'))

    await notificationsPageState.loadNotifications()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(notificationsPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('通知中心页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('类型筛选变更后回到第一页并重新加载', async () => {
    notificationApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { notifications: [], total: 0, unreadCount: 0 }
    })
    notificationsPageState.currentPage = 4

    notificationsPageState.setTypeFilter('article_like')

    expect(notificationsPageState.typeFilter).toBe('article_like')
    expect(notificationsPageState.currentPage).toBe(1)
    expect(notificationsPageState.hasActiveFilters).toBe(true)
    await vi.waitFor(() => expect(notificationApiMock.list).toHaveBeenCalledTimes(1))
  })

  it('重置筛选清空类型条件并回到第一页', async () => {
    notificationApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { notifications: [], total: 0, unreadCount: 0 }
    })
    notificationsPageState.typeFilter = 'follow'
    notificationsPageState.currentPage = 5

    notificationsPageState.resetFilters()

    expect(notificationsPageState.typeFilter).toBe('')
    expect(notificationsPageState.currentPage).toBe(1)
    expect(notificationsPageState.hasActiveFilters).toBe(false)
    await vi.waitFor(() => expect(notificationApiMock.list).toHaveBeenCalledTimes(1))
  })

  it('切换页码后按目标页加载', async () => {
    notificationApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { notifications: [], total: 50, unreadCount: 0 }
    })

    notificationsPageState.changePage(3)

    expect(notificationsPageState.currentPage).toBe(3)
    await vi.waitFor(() =>
      expect(notificationApiMock.list).toHaveBeenCalledWith({
        page: 3,
        pageSize: NOTIFICATION_PAGE_SIZE
      })
    )
  })

  it('总页数按分页大小向上取整', () => {
    notificationsPageState.initialize({
      notifications: [],
      total: NOTIFICATION_PAGE_SIZE * 2 + 1,
      unreadCount: 0
    })

    expect(notificationsPageState.totalPages).toBe(3)
  })
})

describe('通知中心页状态：标记已读', () => {
  beforeEach(resetState)

  it('已读条目不再发起请求', async () => {
    const target = makeNotification(1, true)

    await notificationsPageState.markRead(target)

    expect(notificationApiMock.markRead).not.toHaveBeenCalled()
  })

  it('标记成功后就地更新条目并递减未读数', async () => {
    notificationApiMock.markRead.mockResolvedValue({ code: SUCCESS_CODE, data: null })
    const target = makeNotification(1)
    notificationsPageState.initialize({ notifications: [target], total: 1, unreadCount: 2 })

    await notificationsPageState.markRead(target)

    expect(notificationApiMock.markRead).toHaveBeenCalledWith(1)
    expect(target.isRead).toBe(true)
    expect(notificationsPageState.unreadCount).toBe(1)
  })

  it('未读数不会因标记操作变为负数', async () => {
    notificationApiMock.markRead.mockResolvedValue({ code: SUCCESS_CODE, data: null })
    const target = makeNotification(1)

    await notificationsPageState.markRead(target)

    expect(notificationsPageState.unreadCount).toBe(0)
  })

  it('业务码非成功时提示后端文案且不改动未读数', async () => {
    notificationApiMock.markRead.mockResolvedValue({ code: 403, message: '无权操作' })
    const target = makeNotification(1)
    notificationsPageState.initialize({ notifications: [target], total: 1, unreadCount: 1 })

    await notificationsPageState.markRead(target)

    expect(toastMock.error).toHaveBeenCalledWith('无权操作')
    expect(target.isRead).toBe(false)
    expect(notificationsPageState.unreadCount).toBe(1)
  })

  it('请求抛错时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    notificationApiMock.markRead.mockRejectedValue(new Error('断网'))
    const target = makeNotification(1)

    await notificationsPageState.markRead(target)

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(target.isRead).toBe(false)
    consoleError.mockRestore()
  })

  it('全部已读成功后清零未读数并同步列表', async () => {
    notificationApiMock.markAllRead.mockResolvedValue({ code: SUCCESS_CODE, data: null })
    notificationsPageState.initialize({
      notifications: [makeNotification(1), makeNotification(2)],
      total: 2,
      unreadCount: 2
    })

    await notificationsPageState.markAllRead()

    expect(notificationsPageState.unreadCount).toBe(0)
    expect(notificationsPageState.notifications.every(item => item.isRead)).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('全部通知已标记为已读')
  })

  it('全部已读失败时提示且保留未读数', async () => {
    notificationApiMock.markAllRead.mockResolvedValue({ code: 500, message: '操作失败' })
    notificationsPageState.initialize({
      notifications: [makeNotification(1)],
      total: 1,
      unreadCount: 1
    })

    await notificationsPageState.markAllRead()

    expect(toastMock.error).toHaveBeenCalledWith('操作失败')
    expect(notificationsPageState.unreadCount).toBe(1)
    expect(notificationsPageState.notifications.every(item => !item.isRead)).toBe(true)
  })
})
