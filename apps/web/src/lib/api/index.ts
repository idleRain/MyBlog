// API 接口聚合：基于 @myblog/api 工厂创建，注入应用级 http 客户端实例。

import { createFollowAPI, createUserAPI } from '@myblog/api'
import request from '$lib/service'

const UserAPI = createUserAPI(request)
const FollowAPI = createFollowAPI(request)

const API = {
  user: UserAPI,
  follow: FollowAPI
}

export { UserAPI, FollowAPI }

export default API
