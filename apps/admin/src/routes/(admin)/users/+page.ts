import type { PageLoad } from './$types'

import { emptyUserPageData } from './user-page-state.svelte'
import { USER_PAGE_SIZE } from '$lib/constants/user'
import { UserAPI } from '$lib/api'

// 首屏页码固定为第一页，筛选条件留空，与页面初次渲染的可视状态一致。
const FIRST_PAGE = 1

// 用户管理页初始数据：首屏加载第一页且不带关键词的用户列表。
// 加载失败不阻断页面渲染，降级返回空数据并记录日志，后续由页面内交互重新加载。
export const load: PageLoad = async () => {
  try {
    const response = await UserAPI.getUserList(FIRST_PAGE, USER_PAGE_SIZE)
    if (response.code !== 200 || !response.data) {
      console.error('加载用户列表失败:', response.message)
      return emptyUserPageData()
    }

    return {
      users: response.data.users ?? [],
      total: response.data.total ?? 0
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
    return emptyUserPageData()
  }
}
