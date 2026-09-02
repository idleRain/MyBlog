import type { KyInstance } from 'ky'
import type {
  CreateTagRequest,
  DeleteTagResponse,
  ListTagsRequest,
  PopularTagsResponse,
  TagListResponse,
  TagResponse,
  UpdateTagRequest
} from './types.ts'

/**
 * 创建标签接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createTagAPI(request: KyInstance) {
  return {
    // 获取标签详情
    getById(id: number): Promise<TagResponse> {
      return request.post('tags/get', { json: { id } }).json()
    },

    // 获取热门标签
    getPopular(limit = 10): Promise<PopularTagsResponse> {
      return request.post('tags/popular', { json: { limit } }).json()
    },

    // 管理端：分页查询标签列表
    adminList(params: ListTagsRequest): Promise<TagListResponse> {
      return request.post('admin/tags/list', { json: params }).json()
    },

    // 管理端：创建标签
    create(params: CreateTagRequest): Promise<TagResponse> {
      return request.post('admin/tags/create', { json: params }).json()
    },

    // 管理端：更新标签
    update(params: UpdateTagRequest): Promise<TagResponse> {
      return request.post('admin/tags/update', { json: params }).json()
    },

    // 管理端：删除标签
    delete(id: number): Promise<DeleteTagResponse> {
      return request.post('admin/tags/delete', { json: { id } }).json()
    }
  }
}

export type TagAPI = ReturnType<typeof createTagAPI>

export type {
  CreateTagRequest,
  DeleteTagResponse,
  ListTagsRequest,
  PopularTagsResponse,
  TagListResponse,
  TagResponse,
  UpdateTagRequest
}
