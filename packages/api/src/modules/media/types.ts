import type { UploaderPublic } from '@myblog/api/modules/user/types'
import type { ApiResponse } from '@myblog/shared'

// 媒体文件状态枚举
export type MediaStatus = 'active' | 'processing' | 'failed' | 'lost'

// 存储类型枚举
export type StorageType = 'local' | 'oss' | 's3' | 'cos'

// 媒体文件接口，字段与后端 model.MediaFile 的 JSON tag 一致。
export interface MediaFile {
  id: number
  filename: string
  storedName: string
  filePath: string
  fileUrl: string
  thumbnailUrl: string
  mimeType: string
  fileSize: number
  fileHash: string
  width: number | null
  height: number | null
  durationSeconds: number
  altText: string
  status: MediaStatus
  processedAt: string | null
  uploaderId: number
  storageType: StorageType
  folder: string
  usageCount: number
  downloadCount: number
  isPublic: boolean
  createdAt: string
  updatedAt: string
  // 上传者经窄化视图输出；后端 omitempty 语义下未预加载关联时键整体缺省。
  uploader?: UploaderPublic
}

// 媒体列表数据
export interface MediaListData {
  page: number
  pageSize: number
  total: number
  media: MediaFile[]
}

// 媒体列表查询参数
export interface ListMediaRequest {
  page?: number
  pageSize?: number
  folder?: string
  mimeType?: string
}

// 各响应类型
export type MediaResponse = ApiResponse<MediaFile>
export type MediaListResponse = ApiResponse<MediaListData>
export type DeleteMediaResponse = ApiResponse<null>
