import type { KyInstance } from 'ky'
import type {
  CategoryListResponse,
  CategoryResponse,
  CategoryTreeResponse,
  CreateCategoryRequest,
  DeleteCategoryResponse,
  ListCategoriesRequest,
  UpdateCategoryRequest
} from './types.ts'

/**
 * 创建分类接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createCategoryAPI(request: KyInstance) {
  return {
    // 获取分类详情
    getById(id: number): Promise<CategoryResponse> {
      return request.post('categories/get', { json: { id } }).json()
    },

    // 获取分类树，管理端用于树形展示与选择。
    getTree(): Promise<CategoryTreeResponse> {
      return request.post('categories/tree', { json: {} }).json()
    },

    // 管理端：分页查询分类列表
    adminList(params: ListCategoriesRequest): Promise<CategoryListResponse> {
      return request.post('admin/categories/list', { json: params }).json()
    },

    // 管理端：创建分类
    create(params: CreateCategoryRequest): Promise<CategoryResponse> {
      return request.post('admin/categories/create', { json: params }).json()
    },

    // 管理端：更新分类，不支持修改 parentId。
    update(params: UpdateCategoryRequest): Promise<CategoryResponse> {
      return request.post('admin/categories/update', { json: params }).json()
    },

    // 管理端：删除分类，存在子分类时后端拒绝。
    delete(id: number): Promise<DeleteCategoryResponse> {
      return request.post('admin/categories/delete', { json: { id } }).json()
    }
  }
}

export type CategoryAPI = ReturnType<typeof createCategoryAPI>

export type {
  CategoryListResponse,
  CategoryResponse,
  CategoryTreeResponse,
  CreateCategoryRequest,
  DeleteCategoryResponse,
  ListCategoriesRequest,
  UpdateCategoryRequest
}
