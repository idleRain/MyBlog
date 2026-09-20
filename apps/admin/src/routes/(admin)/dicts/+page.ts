import type { PageLoad } from './$types'

import { DICT_PAGE_SIZE, emptyDictPageData } from './dict-page-state.svelte'
import { DictAPI } from '$lib/api'

// 字典管理页初始数据：类型首页加载后连带首个类型的字典项，字典量级小整页即可覆盖。
export const load: PageLoad = async () => {
  try {
    const typesResponse = await DictAPI.adminListTypes({ page: 1, pageSize: DICT_PAGE_SIZE })
    if (typesResponse.code !== 200 || !typesResponse.data) {
      return emptyDictPageData()
    }

    const types = typesResponse.data.types ?? []
    const firstType = types[0]
    if (!firstType) {
      return {
        types,
        typesTotal: typesResponse.data.total ?? 0,
        selectedTypeId: null,
        items: [],
        itemsTotal: 0
      }
    }

    const itemsResponse = await DictAPI.adminListItems({
      typeId: firstType.id,
      page: 1,
      pageSize: DICT_PAGE_SIZE
    })
    if (itemsResponse.code !== 200 || !itemsResponse.data) {
      return {
        types,
        typesTotal: typesResponse.data.total ?? 0,
        selectedTypeId: firstType.id,
        items: [],
        itemsTotal: 0
      }
    }

    return {
      types,
      typesTotal: typesResponse.data.total ?? 0,
      selectedTypeId: firstType.id,
      items: itemsResponse.data.items ?? [],
      itemsTotal: itemsResponse.data.total ?? 0
    }
  } catch (error) {
    console.error('加载字典管理页数据失败:', error)
    return emptyDictPageData()
  }
}
