import type { KyInstance } from 'ky'
import type {
  DeleteMediaResponse,
  ListMediaRequest,
  MediaListResponse,
  MediaResponse
} from './types.ts'

/**
 * 创建媒体接口模块，依赖注入的 http 客户端由调用方提供。
 * 上传接口使用 multipart/form-data，表单字段名为 file。
 */
export function createMediaAPI(request: KyInstance) {
  return {
    // 上传文件，返回去重后的媒体记录。
    upload(file: File): Promise<MediaResponse> {
      const formData = new FormData()
      formData.append('file', file)
      return request.post('media/upload', { body: formData }).json()
    },

    // 获取媒体文件详情
    getById(id: number): Promise<MediaResponse> {
      return request.post('media/get', { json: { id } }).json()
    },

    // 媒体列表，非管理员仅返回自己上传的文件。
    list(params: ListMediaRequest): Promise<MediaListResponse> {
      return request.post('media/list', { json: params }).json()
    },

    // 删除媒体文件，仅上传者本人或管理员可删。
    delete(id: number): Promise<DeleteMediaResponse> {
      return request.post('media/delete', { json: { id } }).json()
    }
  }
}

export type MediaAPI = ReturnType<typeof createMediaAPI>

export type { DeleteMediaResponse, ListMediaRequest, MediaListResponse, MediaResponse }
