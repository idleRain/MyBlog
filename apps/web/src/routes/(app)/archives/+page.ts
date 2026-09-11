import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { ArticleAPI } from '$lib/api'

export const load: PageLoad = async () => {
  const response = await ArticleAPI.archives()

  if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
    throw error(503, '归档数据加载失败，请稍后重试')
  }

  return { groups: response.data }
}
