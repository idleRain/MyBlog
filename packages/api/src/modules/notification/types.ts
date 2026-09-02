import type { ApiResponse } from '@myblog/shared'
import type { User } from '@myblog/api/modules/user/types'

// 通知类型枚举，与后端 model.Notification 的 type 取值一致。
export type NotificationType =
  'comment_reply' | 'article_like' | 'comment_like' | 'system' | 'follow' | 'article_new'

// 通知信息接口，字段与后端 model.Notification 的 JSON tag 一致。
export interface Notification {
  id: number
  userId: number
  senderId: number | null
  type: NotificationType
  title: string
  content: string | null
  actionUrl: string
  relatedType: string | null
  relatedId: number | null
  isRead: boolean
  readAt: string | null
  createdAt: string
  updatedAt: string
  sender: User | null
}

// 通知列表数据，额外附带未读数。
export interface NotificationListData {
  page: number
  pageSize: number
  total: number
  unreadCount: number
  notifications: Notification[]
}

// 通知列表查询参数
export interface ListNotificationsRequest {
  page?: number
  pageSize?: number
  type?: NotificationType
  isRead?: boolean | null
}

// 未读数响应数据
export interface UnreadCountData {
  unreadCount: number
}

// 各响应类型
export type NotificationListResponse = ApiResponse<NotificationListData>
export type UnreadCountResponse = ApiResponse<UnreadCountData>
export type NotificationActionResponse = ApiResponse<null>
