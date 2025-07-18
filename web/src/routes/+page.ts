import type { PageLoad } from './$types'
import API from '$lib/api'

export const prerender = false

export const load: PageLoad = async () => {
  return API.user.getUser()
}
