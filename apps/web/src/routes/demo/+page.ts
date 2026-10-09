import type { PageLoad } from './$types'
import { dev } from '$app/environment'
import { error } from '@sveltejs/kit'

// i18n 演示沙盒不承载任何业务内容，仅作开发期的取词语法对照。
// 该沙盒不对外发布，因此生产构建下按未找到处理，使路由在生产不可达；
// 开发环境保持原样以便随时对照词表。
export const load: PageLoad = () => {
  if (!dev) {
    throw error(404, '页面不存在')
  }

  return {}
}
