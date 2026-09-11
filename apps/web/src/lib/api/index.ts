// API 接口聚合：基于 @myblog/api 工厂创建，注入应用级 http 客户端实例。

import { createArticleAPI, createCategoryAPI, createFollowAPI, createUserAPI } from '@myblog/api'
import request from '$lib/service'

const UserAPI = createUserAPI(request)
const FollowAPI = createFollowAPI(request)
const ArticleAPI = createArticleAPI(request)
const CategoryAPI = createCategoryAPI(request)

const API = {
  user: UserAPI,
  follow: FollowAPI,
  article: ArticleAPI,
  category: CategoryAPI
}

export { UserAPI, FollowAPI, ArticleAPI, CategoryAPI }

export default API
