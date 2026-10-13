import type { PageLoad } from './$types'

import { emptyNotificationsPageData } from './notifications-page-state.svelte'
import { NOTIFICATION_PAGE_SIZE } from '$lib/constants/notification'
import { NotificationAPI } from '$lib/api'

// 通知中心页初始数据：首屏加载第一页且不带类型筛选。
// 列表加载失败不抛错，回退空数据并由页面状态模块在后续交互中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptyNotificationsPageData()
  try {
    const response = await NotificationAPI.list({ page: 1, pageSize: NOTIFICATION_PAGE_SIZE })
    if (response.code === 200 && response.data) {
      pageData.notifications = response.data.notifications ?? []
      pageData.total = response.data.total ?? 0
      pageData.unreadCount = response.data.unreadCount ?? 0
    } else {
      console.error('加载通知列表失败:', response.message)
    }
  } catch (error) {
    console.error('加载通知列表失败:', error)
  }
  return pageData
}
