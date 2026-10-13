import type { PageLoad } from './$types'

import { emptyTagPageData } from './tag-page-state.svelte'
import { DICT_TYPE_CODE_TAG_STATUS } from '@myblog/api'
import { TAG_PAGE_SIZE } from '$lib/constants/tag'
import { DictAPI, TagAPI } from '$lib/api'

// 标签管理页初始数据：首屏标签列表与标签状态字典并行加载。
// 字典加载失败不阻断页面，状态展示回退内置配置，故失败仅记录日志。
export const load: PageLoad = async () => {
  const [tagsResult, dictResult] = await Promise.allSettled([
    TagAPI.adminList({ page: 1, pageSize: TAG_PAGE_SIZE }),
    DictAPI.getByType(DICT_TYPE_CODE_TAG_STATUS)
  ])

  const pageData = emptyTagPageData()

  if (tagsResult.status === 'fulfilled' && tagsResult.value.code === 200 && tagsResult.value.data) {
    pageData.tags = tagsResult.value.data.tags ?? []
    pageData.total = tagsResult.value.data.total ?? 0
  } else if (tagsResult.status === 'rejected') {
    console.error('加载标签列表失败:', tagsResult.reason)
  }

  if (dictResult.status === 'fulfilled' && dictResult.value.code === 200 && dictResult.value.data) {
    pageData.statusDict = dictResult.value.data
  } else if (dictResult.status === 'rejected') {
    console.error('加载标签状态字典失败:', dictResult.reason)
  }

  return pageData
}
