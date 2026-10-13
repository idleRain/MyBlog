import type { PageLoad } from './$types'

import { emptySettingsPageData } from './settings-page-state.svelte'
import { SettingAPI } from '$lib/api'

// 系统设置页初始数据：首屏加载全部设置项。
// 加载失败不抛错，回退空数据并由页面状态模块在后续交互中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptySettingsPageData()
  try {
    const response = await SettingAPI.list()
    if (response.code === 200 && response.data) {
      pageData.settings = response.data.settings ?? []
    } else {
      console.error('加载设置失败:', response.message)
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
  return pageData
}
