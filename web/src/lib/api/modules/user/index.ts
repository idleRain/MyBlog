import request from '$lib/service'
import type { Res } from './types'

const UserAPI = {
  getUser() {
    return request
      .post<Res>('users/list', {
        json: {
          page: 1,
          pageSize: 10
        }
      })
      .json()
  }
}

export default UserAPI
