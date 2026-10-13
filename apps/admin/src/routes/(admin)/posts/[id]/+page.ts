import type { PageLoad } from './$types'

import { ArticleAPI } from '$lib/api'

// 文章编辑页初始数据：按路由参数加载文章详情供编辑表单回填。
// 加载失败不抛错，改以错误文案返回，由页面渲染失败态并给出返回入口。
export const load: PageLoad = async ({ params }) => {
  const articleId = Number(params.id)
  if (!Number.isFinite(articleId) || articleId <= 0) {
    return { article: null, loadError: '无效的文章标识' }
  }

  try {
    const response = await ArticleAPI.getById(articleId)
    if (response.code === 200 && response.data) {
      return { article: response.data, loadError: '' }
    }
    return { article: null, loadError: response.message || '文章不存在或无权访问' }
  } catch (error) {
    console.error('加载文章失败:', error)
    return { article: null, loadError: '网络错误，请稍后重试' }
  }
}
