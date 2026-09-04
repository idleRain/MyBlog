// API 接口聚合：基于 @myblog/api 工厂创建，注入应用级 http 客户端实例。

import {
  createArticleAPI,
  createCategoryAPI,
  createCommentAPI,
  createFollowAPI,
  createFriendlyLinkAPI,
  createMediaAPI,
  createNotificationAPI,
  createSettingAPI,
  createStatsAPI,
  createTagAPI,
  createUserAPI
} from '@myblog/api'
import request from '$lib/service'

const UserAPI = createUserAPI(request)
const ArticleAPI = createArticleAPI(request)
const CategoryAPI = createCategoryAPI(request)
const TagAPI = createTagAPI(request)
const CommentAPI = createCommentAPI(request)
const MediaAPI = createMediaAPI(request)
const SettingAPI = createSettingAPI(request)
const FriendlyLinkAPI = createFriendlyLinkAPI(request)
const StatsAPI = createStatsAPI(request)
const NotificationAPI = createNotificationAPI(request)
const FollowAPI = createFollowAPI(request)

const API = {
  user: UserAPI,
  article: ArticleAPI,
  category: CategoryAPI,
  tag: TagAPI,
  comment: CommentAPI,
  media: MediaAPI,
  setting: SettingAPI,
  friendlyLink: FriendlyLinkAPI,
  stats: StatsAPI,
  notification: NotificationAPI,
  follow: FollowAPI
}

export {
  UserAPI,
  ArticleAPI,
  CategoryAPI,
  TagAPI,
  CommentAPI,
  MediaAPI,
  SettingAPI,
  FriendlyLinkAPI,
  StatsAPI,
  NotificationAPI,
  FollowAPI
}

export default API
