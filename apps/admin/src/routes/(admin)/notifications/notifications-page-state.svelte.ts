import type { Notification, NotificationType } from '@myblog/api/modules/notification/types'
import { NOTIFICATION_PAGE_SIZE } from '$lib/constants/notification'
import { NotificationAPI } from '$lib/api'
import { toast } from 'svelte-sonner'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface NotificationsPageData {
  notifications: Notification[]
  total: number
  unreadCount: number
}

// 空页面数据，load 失败时使用。
export function emptyNotificationsPageData(): NotificationsPageData {
  return { notifications: [], total: 0, unreadCount: 0 }
}

// 通知中心页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class NotificationsPageState {
  notifications = $state<Notification[]>([])
  total = $state(0)
  unreadCount = $state(0)
  currentPage = $state(1)
  // 类型筛选使用空串表示全部，与通知类型枚举的取值区间分开。
  typeFilter = $state<NotificationType | ''>('')
  isLoading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / NOTIFICATION_PAGE_SIZE))
  }

  // 是否存在生效中的筛选条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.typeFilter !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: NotificationsPageData) {
    this.notifications = data.notifications
    this.total = data.total
    this.unreadCount = data.unreadCount
  }

  // 加载通知列表并同步未读数。
  async loadNotifications() {
    this.isLoading = true
    try {
      const response = await NotificationAPI.list({
        page: this.currentPage,
        pageSize: NOTIFICATION_PAGE_SIZE,
        ...(this.typeFilter ? { type: this.typeFilter } : {})
      })
      if (response.code === 200 && response.data) {
        this.notifications = response.data.notifications ?? []
        this.total = response.data.total ?? 0
        this.unreadCount = response.data.unreadCount ?? 0
      } else {
        toast.error(response.message || '加载通知失败')
      }
    } catch (error) {
      console.error('加载通知失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 类型筛选变更后回到第一页并重新加载。
  setTypeFilter(value: string) {
    this.typeFilter = value as NotificationType | ''
    this.currentPage = 1
    void this.loadNotifications()
  }

  // 重置筛选条件并回到第一页。
  resetFilters() {
    this.typeFilter = ''
    this.currentPage = 1
    void this.loadNotifications()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadNotifications()
  }

  // 标记单条通知为已读，已读条目不再发起请求。
  async markRead(notification: Notification) {
    if (notification.isRead) return
    try {
      const response = await NotificationAPI.markRead(notification.id)
      if (response.code === 200) {
        notification.isRead = true
        this.unreadCount = Math.max(0, this.unreadCount - 1)
      } else {
        toast.error(response.message || '操作失败')
      }
    } catch (error) {
      console.error('标记已读失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }

  // 标记全部通知为已读，成功后本地列表同步为已读态。
  async markAllRead() {
    try {
      const response = await NotificationAPI.markAllRead()
      if (response.code === 200) {
        this.unreadCount = 0
        this.notifications = this.notifications.map(item => ({ ...item, isRead: true }))
        toast.success('全部通知已标记为已读')
      } else {
        toast.error(response.message || '操作失败')
      }
    } catch (error) {
      console.error('标记全部已读失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }
}

// 页面级共享状态实例。
export const notificationsPageState = new NotificationsPageState()
