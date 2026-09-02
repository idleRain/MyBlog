// @myblog/api 统一导出：接口模块工厂与响应类型。

export { createUserAPI } from './modules/user/index.ts'
export type { UserAPI } from './modules/user/index.ts'
export * from './modules/user/types.ts'

export { createArticleAPI } from './modules/article/index.ts'
export type { ArticleAPI } from './modules/article/index.ts'
export * from './modules/article/types.ts'

export { createCategoryAPI } from './modules/category/index.ts'
export type { CategoryAPI } from './modules/category/index.ts'
export * from './modules/category/types.ts'

export { createTagAPI } from './modules/tag/index.ts'
export type { TagAPI } from './modules/tag/index.ts'
export * from './modules/tag/types.ts'

export { createCommentAPI } from './modules/comment/index.ts'
export type { CommentAPI } from './modules/comment/index.ts'
export * from './modules/comment/types.ts'

export { createMediaAPI } from './modules/media/index.ts'
export type { MediaAPI } from './modules/media/index.ts'
export * from './modules/media/types.ts'

export { createSettingAPI } from './modules/setting/index.ts'
export type { SettingAPI } from './modules/setting/index.ts'
export * from './modules/setting/types.ts'

export { createFriendlyLinkAPI } from './modules/friendlyLink/index.ts'
export type { FriendlyLinkAPI } from './modules/friendlyLink/index.ts'
export * from './modules/friendlyLink/types.ts'

export { createStatsAPI } from './modules/stats/index.ts'
export type { StatsAPI } from './modules/stats/index.ts'
export * from './modules/stats/types.ts'

export { createNotificationAPI } from './modules/notification/index.ts'
export type { NotificationAPI } from './modules/notification/index.ts'
export * from './modules/notification/types.ts'
