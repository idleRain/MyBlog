import type { MediaFile } from '@myblog/api/modules/media/types'
import { toast } from 'svelte-sonner'

import { MEDIA_PAGE_SIZE } from '$lib/constants/media'
import { MediaAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface MediaPageData {
  media: MediaFile[]
  total: number
}

// 空页面数据，load 失败时使用。
export function emptyMediaPageData(): MediaPageData {
  return { media: [], total: 0 }
}

// 媒体库页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class MediaPageState {
  media = $state<MediaFile[]>([])
  total = $state(0)
  currentPage = $state(1)
  // 类型筛选使用 MIME 前缀字符串，空串为全部，与后端 mimeType 参数口径一致。
  typeFilter = $state('')
  isLoading = $state(false)
  isUploading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / MEDIA_PAGE_SIZE))
  }

  // 是否存在生效中的筛选条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.typeFilter !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: MediaPageData) {
    this.media = data.media
    this.total = data.total
  }

  // 加载媒体列表，携带类型筛选与分页参数。
  async loadMedia() {
    this.isLoading = true
    try {
      const response = await MediaAPI.list({
        page: this.currentPage,
        pageSize: MEDIA_PAGE_SIZE,
        ...(this.typeFilter ? { mimeType: this.typeFilter } : {})
      })
      if (response.code === 200 && response.data) {
        this.media = response.data.media ?? []
        this.total = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载媒体失败')
      }
    } catch (error) {
      console.error('加载媒体失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 类型筛选变更后立即回到第一页并重新加载。
  setTypeFilter(value: string) {
    this.typeFilter = value
    this.currentPage = 1
    void this.loadMedia()
  }

  // 重置筛选并回到第一页。
  resetFilters() {
    this.typeFilter = ''
    this.currentPage = 1
    void this.loadMedia()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadMedia()
  }

  // 上传单个文件，成功后回到第一页并刷新列表。
  async uploadFile(file: File): Promise<boolean> {
    if (this.isUploading) return false
    this.isUploading = true
    try {
      const response = await MediaAPI.upload(file)
      if (response.code === 200 && response.data) {
        toast.success('文件上传成功')
        this.currentPage = 1
        void this.loadMedia()
        return true
      }
      toast.error(response.message || '上传失败')
    } catch (error) {
      console.error('上传失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isUploading = false
    }
    return false
  }

  // 复制文件 URL 到剪贴板并提示结果，剪贴板不可用时提示手动复制。
  async copyFileUrl(file: MediaFile) {
    try {
      await navigator.clipboard.writeText(file.fileUrl)
      toast.success('链接已复制')
    } catch (error) {
      console.error('复制链接失败:', error)
      toast.error('复制失败，请手动复制')
    }
  }

  // 删除媒体文件，成功后刷新列表。
  async removeMedia(id: number): Promise<boolean> {
    try {
      const response = await MediaAPI.delete(id)
      if (response.code === 200) {
        toast.success('文件删除成功')
        void this.loadMedia()
        return true
      }
      toast.error(response.message || '删除失败')
    } catch (error) {
      console.error('删除失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }
}

// 页面级共享状态实例。
export const mediaPageState = new MediaPageState()
