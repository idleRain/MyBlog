import type { KyInstance } from 'ky'
import type {
  AllDictsResponse,
  CreateDictItemRequest,
  CreateDictTypeRequest,
  DeleteDictResponse,
  DictByTypeResponse,
  DictItemListResponse,
  DictItemResponse,
  DictTypeListResponse,
  DictTypeResponse,
  ListDictItemsRequest,
  ListDictTypesRequest,
  UpdateDictItemRequest,
  UpdateDictTypeRequest
} from './types.ts'

/**
 * 字典接口模块，依赖注入的 http 客户端由调用方提供。
 * 公开接口供 app 端读取已生效字典，管理端接口需字典管理权限。
 */
export function createDictAPI(request: KyInstance) {
  return {
    // app 端：全量已生效字典，按类型分组输出
    getAllEnabled(): Promise<AllDictsResponse> {
      return request.post('dicts/all', { json: {} }).json()
    },

    // app 端：按字典码查询单个已生效字典列表
    getByType(type: string): Promise<DictByTypeResponse> {
      return request.post(`dicts/${type}`, { json: {} }).json()
    },

    // 管理端：分页查询字典类型列表
    adminListTypes(params: ListDictTypesRequest): Promise<DictTypeListResponse> {
      return request.post('admin/dicts/types/list', { json: params }).json()
    },

    // 管理端：创建字典类型
    adminCreateType(params: CreateDictTypeRequest): Promise<DictTypeResponse> {
      return request.post('admin/dicts/types/create', { json: params }).json()
    },

    // 管理端：更新字典类型，字典码不可变更
    adminUpdateType(params: UpdateDictTypeRequest): Promise<DictTypeResponse> {
      return request.post('admin/dicts/types/update', { json: params }).json()
    },

    // 管理端：删除字典类型，级联删除其全部字典项
    adminDeleteType(id: number): Promise<DeleteDictResponse> {
      return request.post('admin/dicts/types/delete', { json: { id } }).json()
    },

    // 管理端：分页查询字典项列表，按类型维护
    adminListItems(params: ListDictItemsRequest): Promise<DictItemListResponse> {
      return request.post('admin/dicts/items/list', { json: params }).json()
    },

    // 管理端：创建字典项
    adminCreateItem(params: CreateDictItemRequest): Promise<DictItemResponse> {
      return request.post('admin/dicts/items/create', { json: params }).json()
    },

    // 管理端：更新字典项，字典项值不可变更
    adminUpdateItem(params: UpdateDictItemRequest): Promise<DictItemResponse> {
      return request.post('admin/dicts/items/update', { json: params }).json()
    },

    // 管理端：删除字典项
    adminDeleteItem(id: number): Promise<DeleteDictResponse> {
      return request.post('admin/dicts/items/delete', { json: { id } }).json()
    }
  }
}

export type DictAPI = ReturnType<typeof createDictAPI>

export { dictItemLabel, findDictItem } from './helpers.ts'

export type {
  AllDictsResponse,
  CreateDictItemRequest,
  CreateDictTypeRequest,
  DeleteDictResponse,
  DictByTypeResponse,
  DictItemListResponse,
  DictItemResponse,
  DictTypeListResponse,
  DictTypeResponse,
  ListDictItemsRequest,
  ListDictTypesRequest,
  UpdateDictItemRequest,
  UpdateDictTypeRequest
} from './types.ts'
