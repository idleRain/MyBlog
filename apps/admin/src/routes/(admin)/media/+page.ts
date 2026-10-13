import type { PageLoad } from './$types'

import { emptyMediaPageData } from './media-page-state.svelte'
import { MEDIA_PAGE_SIZE } from '$lib/constants/media'
import { MediaAPI } from '$lib/api'

// 媒体库页初始数据：首屏加载第一页且不带类型筛选。
// 列表加载失败不抛错，回退空数据并由页面状态模块在后续操作中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptyMediaPageData()
  try {
    const response = await MediaAPI.list({ page: 1, pageSize: MEDIA_PAGE_SIZE })
    if (response.code === 200 && response.data) {
      pageData.media = response.data.media ?? []
      pageData.total = response.data.total ?? 0
    } else {
      console.error('加载媒体列表失败:', response.message)
    }
  } catch (error) {
    console.error('加载媒体列表失败:', error)
  }
  return pageData
}
