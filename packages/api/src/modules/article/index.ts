import type { KyInstance } from 'ky'
import type {
  ArticleActionResponse,
  ArticleListResponse,
  ArticleResponse,
  CreateArticleRequest,
  GetArticleListRequest,
  UpdateArticleRequest
} from './types.ts'

// 文章状态流转接口路径后缀，管理端与编辑者端点分别挂载。
const ADMIN_ARTICLE_ACTION_PATHS = {
  publish: 'admin/articles/publish',
  unpublish: 'admin/articles/unpublish',
  archive: 'admin/articles/archive',
  private: 'admin/articles/private'
} as const

// 编辑者状态流转端点路径，调用方需具备文章创建或发布权限。
const EDITOR_ARTICLE_ACTION_PATHS = {
  publish: 'articles/publish',
  unpublish: 'articles/unpublish',
  archive: 'articles/archive',
  private: 'articles/private'
} as const

/**
 * 创建文章接口模块，依赖注入的 http 客户端由调用方提供。
 * 编辑者端点挂 /api/articles 下，管理端点挂 /api/admin/articles 下。
 */
export function createArticleAPI(request: KyInstance) {
  return {
    // 获取文章详情，未发布文章仅作者或管理员可见。
    getById(id: number): Promise<ArticleResponse> {
      return request.post('articles/get', { json: { id } }).json()
    },

    // 创建文章，需文章创建权限。
    create(params: CreateArticleRequest): Promise<ArticleResponse> {
      return request.post('articles/create', { json: params }).json()
    },

    // 编辑者更新自己的文章。
    update(params: UpdateArticleRequest): Promise<ArticleResponse> {
      return request.post('articles/update', { json: params }).json()
    },

    // 编辑者删除自己的文章。
    delete(id: number): Promise<ArticleActionResponse> {
      return request.post('articles/delete', { json: { id } }).json()
    },

    // 文章列表，非管理员被服务端强制为已发布状态。
    list(params: GetArticleListRequest): Promise<ArticleListResponse> {
      return request.post('articles/list', { json: params }).json()
    },

    // 编辑者发布自己的文章。
    publish(id: number): Promise<ArticleActionResponse> {
      return request.post(EDITOR_ARTICLE_ACTION_PATHS.publish, { json: { id } }).json()
    },

    // 编辑者取消发布自己的文章。
    unpublish(id: number): Promise<ArticleActionResponse> {
      return request.post(EDITOR_ARTICLE_ACTION_PATHS.unpublish, { json: { id } }).json()
    },

    // 编辑者归档自己的文章。
    archive(id: number): Promise<ArticleActionResponse> {
      return request.post(EDITOR_ARTICLE_ACTION_PATHS.archive, { json: { id } }).json()
    },

    // 编辑者设置自己的文章为私有。
    private(id: number): Promise<ArticleActionResponse> {
      return request.post(EDITOR_ARTICLE_ACTION_PATHS.private, { json: { id } }).json()
    },

    // 管理端：全量文章列表，可含全部状态并按状态筛选。
    adminList(params: GetArticleListRequest): Promise<ArticleListResponse> {
      return request.post('admin/articles/list', { json: params }).json()
    },

    // 管理端：更新任意文章。
    adminUpdate(params: UpdateArticleRequest): Promise<ArticleResponse> {
      return request.post('admin/articles/update', { json: params }).json()
    },

    // 管理端：删除任意文章。
    adminDelete(id: number): Promise<ArticleActionResponse> {
      return request.post('admin/articles/delete', { json: { id } }).json()
    },

    // 管理端：发布文章。
    adminPublish(id: number): Promise<ArticleActionResponse> {
      return request.post(ADMIN_ARTICLE_ACTION_PATHS.publish, { json: { id } }).json()
    },

    // 管理端：取消发布文章。
    adminUnpublish(id: number): Promise<ArticleActionResponse> {
      return request.post(ADMIN_ARTICLE_ACTION_PATHS.unpublish, { json: { id } }).json()
    },

    // 管理端：归档文章。
    adminArchive(id: number): Promise<ArticleActionResponse> {
      return request.post(ADMIN_ARTICLE_ACTION_PATHS.archive, { json: { id } }).json()
    },

    // 管理端：设置文章为私有。
    adminPrivate(id: number): Promise<ArticleActionResponse> {
      return request.post(ADMIN_ARTICLE_ACTION_PATHS.private, { json: { id } }).json()
    }
  }
}

export type ArticleAPI = ReturnType<typeof createArticleAPI>

export type {
  ArticleActionResponse,
  ArticleListResponse,
  ArticleResponse,
  CreateArticleRequest,
  GetArticleListRequest,
  UpdateArticleRequest
}
