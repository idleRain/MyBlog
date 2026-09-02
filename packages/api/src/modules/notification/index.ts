import type { KyInstance } from 'ky'
import type {
  ListNotificationsRequest,
  NotificationActionResponse,
  NotificationListResponse,
  UnreadCountResponse
} from './types.ts'

/**
 * 创建通知接口模块，依赖注入的 http 客户端由调用方提供。
 * 通知接口仅能操作当前登录用户本人的数据。
 */
export function createNotificationAPI(request: KyInstance) {
  return {
    // 分页查询当前用户通知，响应附带未读数。
    list(params: ListNotificationsRequest): Promise<NotificationListResponse> {
      return request.post('notifications/list', { json: params }).json()
    },

    // 获取当前用户未读通知数量。
    getUnreadCount(): Promise<UnreadCountResponse> {
      return request.post('notifications/unread-count', { json: {} }).json()
    },

    // 标记单条通知为已读。
    markRead(id: number): Promise<NotificationActionResponse> {
      return request.post('notifications/read', { json: { id } }).json()
    },

    // 标记全部通知为已读。
    markAllRead(): Promise<NotificationActionResponse> {
      return request.post('notifications/read-all', { json: {} }).json()
    }
  }
}

export type NotificationAPI = ReturnType<typeof createNotificationAPI>

export type {
  ListNotificationsRequest,
  NotificationActionResponse,
  NotificationListResponse,
  UnreadCountResponse
}
