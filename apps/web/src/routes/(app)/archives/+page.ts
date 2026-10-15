import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { ArticleAPI } from '$lib/api'
import { m } from '$i18n'

export const load: PageLoad = async () => {
  const response = await ArticleAPI.archives()

  if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
    throw error(503, m['blog:error.archivesFailed']())
  }

  return { groups: response.data }
}
