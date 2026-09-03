import type { KyInstance } from 'ky'
import type {
  ArticleActionResponse,
  ArticleListResponse,
  ArticleResponse,
  CreateArticleRequest,
  GetArticleListRequest,
  UpdateArticleRequest
} from './types.ts'

// 文章状态流转接口路径后缀，全部挂在 /api/articles 下，由后端按角色授权。
const ARTICLE_ACTION_PATHS = {
  publish: 'articles/publish',
  unpublish: 'articles/unpublish',
  archive: 'articles/archive',
  private: 'articles/private'
} as const

/**
 * 创建文章接口模块，依赖注入的 http 客户端由调用方提供。
 * 接口全部挂 /api/articles 下，作者或管理员均可操作，权限由后端服务层统一判定。
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

    // 更新文章，作者或管理员可更新。
    update(params: UpdateArticleRequest): Promise<ArticleResponse> {
      return request.post('articles/update', { json: params }).json()
    },

    // 删除文章，作者或管理员可删除。
    delete(id: number): Promise<ArticleActionResponse> {
      return request.post('articles/delete', { json: { id } }).json()
    },

    // 文章列表，管理员可查看全部状态，其他角色被服务端强制为已发布状态。
    list(params: GetArticleListRequest): Promise<ArticleListResponse> {
      return request.post('articles/list', { json: params }).json()
    },

    // 发布文章。
    publish(id: number): Promise<ArticleActionResponse> {
      return request.post(ARTICLE_ACTION_PATHS.publish, { json: { id } }).json()
    },

    // 取消发布文章。
    unpublish(id: number): Promise<ArticleActionResponse> {
      return request.post(ARTICLE_ACTION_PATHS.unpublish, { json: { id } }).json()
    },

    // 归档文章。
    archive(id: number): Promise<ArticleActionResponse> {
      return request.post(ARTICLE_ACTION_PATHS.archive, { json: { id } }).json()
    },

    // 设置文章为私有。
    private(id: number): Promise<ArticleActionResponse> {
      return request.post(ARTICLE_ACTION_PATHS.private, { json: { id } }).json()
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
