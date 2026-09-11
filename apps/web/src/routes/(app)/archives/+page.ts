import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { ArticleAPI } from '$lib/api'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: PageLoad = async () => {
  const response = await ArticleAPI.archives()

  if (response.code !== SUCCESS_CODE || !response.data) {
    throw error(503, '归档数据加载失败，请稍后重试')
  }

  return { groups: response.data }
}
