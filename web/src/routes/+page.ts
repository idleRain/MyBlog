import type { PageLoad } from './$types'
import request from '../service/index'

export const prerender = false

export interface Data {
  page: number
  pageSize: number
  pages: number
  total: number
  users: never[]
}

export interface Res {
  code: number
  message: string
  data: Data
}

export const load: PageLoad = async () => {
  return await request
    .post<Res>('users/list', {
      json: {
        page: 1,
        pageSize: 10
      }
    })
    .json()
}
