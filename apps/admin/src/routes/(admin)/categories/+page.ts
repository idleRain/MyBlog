import type { PageLoad } from './$types'

import { emptyCategoryPageData } from './category-page-state.svelte'
import { CategoryAPI } from '$lib/api'

// 分类管理页初始数据：分类树一次拉取全量，页面扁平化后在客户端筛选，不走服务端分页。
// 加载失败不抛错，回退空数据并由页面状态模块在后续操作中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptyCategoryPageData()
  try {
    const response = await CategoryAPI.getTree()
    if (response.code === 200 && response.data) {
      pageData.tree = response.data.tree ?? []
    } else {
      console.error('加载分类树失败:', response.message)
    }
  } catch (error) {
    console.error('加载分类树失败:', error)
  }
  return pageData
}
