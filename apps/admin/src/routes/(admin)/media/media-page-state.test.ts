import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { MediaFile } from '@myblog/api/modules/media/types'

import {
  MEDIA_IMAGE_MIME_PREFIX,
  MEDIA_PAGE_SIZE,
  MEDIA_VIDEO_MIME_PREFIX
} from '$lib/constants/media'

// 接口、提示与剪贴板均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与浏览器接口。
const mediaApiMock = {
  upload: vi.fn(),
  list: vi.fn(),
  delete: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

const clipboardMock = {
  writeText: vi.fn()
}

vi.mock('$lib/api', () => ({ MediaAPI: mediaApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyMediaPageData, mediaPageState } = await import('./media-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 网络异常的统一提示文案，与页面状态模块内的取值一致。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 构造一个媒体文件替身，字段取值只覆盖被测逻辑关心的部分。
function makeMedia(id: number, mimeType = 'image/png'): MediaFile {
  return {
    id,
    filename: `文件${id}.png`,
    fileUrl: `https://cdn.example.com/${id}.png`,
    thumbnailUrl: '',
    mimeType,
    fileSize: 1024
  } as unknown as MediaFile
}

// 构造一个上传文件替身，Node 环境不依赖真实的 File 实现。
function makeUploadFile(): File {
  return { name: 'photo.png', type: 'image/png', size: 1024 } as unknown as File
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  mediaApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { media: [], total: 0 } })
  mediaPageState.initialize(emptyMediaPageData())
  mediaPageState.typeFilter = ''
  mediaPageState.currentPage = 1
  mediaPageState.isLoading = false
  mediaPageState.isUploading = false
}

describe('媒体页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表与总数', () => {
    const media = [makeMedia(1), makeMedia(2)]

    mediaPageState.initialize({ media, total: 2 })

    expect(mediaPageState.media).toEqual(media)
    expect(mediaPageState.total).toBe(2)
  })

  it('空数据下无生效筛选条件且总页数为 1', () => {
    expect(mediaPageState.hasActiveFilters).toBe(false)
    expect(mediaPageState.totalPages).toBe(1)
  })
})

describe('媒体页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带分页参数', async () => {
    mediaApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { media: [makeMedia(1)], total: 1 }
    })

    await mediaPageState.loadMedia()

    expect(mediaApiMock.list).toHaveBeenCalledWith({ page: 1, pageSize: MEDIA_PAGE_SIZE })
    expect(mediaPageState.media).toHaveLength(1)
    expect(mediaPageState.total).toBe(1)
  })

  it('类型筛选生效时按 MIME 前缀查询', async () => {
    mediaPageState.typeFilter = MEDIA_IMAGE_MIME_PREFIX
    mediaPageState.currentPage = 3

    await mediaPageState.loadMedia()

    expect(mediaApiMock.list).toHaveBeenCalledWith({
      page: 3,
      pageSize: MEDIA_PAGE_SIZE,
      mimeType: MEDIA_IMAGE_MIME_PREFIX
    })
    expect(mediaPageState.hasActiveFilters).toBe(true)
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    mediaApiMock.list.mockResolvedValue({ code: 500, message: '查询失败' })
    mediaPageState.initialize({ media: [makeMedia(1)], total: 1 })

    await mediaPageState.loadMedia()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(mediaPageState.media).toHaveLength(1)
  })

  it('请求抛错时提示网络错误并结束加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    mediaApiMock.list.mockRejectedValue(new Error('断网'))

    await mediaPageState.loadMedia()

    expect(toastMock.error).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE)
    expect(mediaPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('媒体页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('类型筛选变更后回到第一页并重新加载', async () => {
    mediaPageState.currentPage = 4

    mediaPageState.setTypeFilter(MEDIA_VIDEO_MIME_PREFIX)

    expect(mediaPageState.typeFilter).toBe(MEDIA_VIDEO_MIME_PREFIX)
    expect(mediaPageState.currentPage).toBe(1)
    await vi.waitFor(() =>
      expect(mediaApiMock.list).toHaveBeenCalledWith({
        page: 1,
        pageSize: MEDIA_PAGE_SIZE,
        mimeType: MEDIA_VIDEO_MIME_PREFIX
      })
    )
  })

  it('重置筛选清空条件并回到第一页', async () => {
    mediaPageState.typeFilter = MEDIA_IMAGE_MIME_PREFIX
    mediaPageState.currentPage = 5

    mediaPageState.resetFilters()

    expect(mediaPageState.typeFilter).toBe('')
    expect(mediaPageState.currentPage).toBe(1)
    expect(mediaPageState.hasActiveFilters).toBe(false)
    await vi.waitFor(() =>
      expect(mediaApiMock.list).toHaveBeenCalledWith({ page: 1, pageSize: MEDIA_PAGE_SIZE })
    )
  })

  it('切换页码后按目标页加载', async () => {
    mediaApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { media: [], total: 25 } })

    mediaPageState.changePage(3)

    await vi.waitFor(() =>
      expect(mediaApiMock.list).toHaveBeenCalledWith({ page: 3, pageSize: MEDIA_PAGE_SIZE })
    )
    expect(mediaPageState.currentPage).toBe(3)
  })

  it('总页数按分页容量向上取整，无数据时为 1 页', async () => {
    expect(mediaPageState.totalPages).toBe(1)

    mediaApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { media: [], total: 25 } })
    await mediaPageState.loadMedia()

    expect(mediaPageState.totalPages).toBe(3)
  })
})

describe('媒体页状态：上传', () => {
  beforeEach(resetState)

  it('上传成功后提示、回到第一页并刷新列表', async () => {
    const file = makeUploadFile()
    mediaApiMock.upload.mockResolvedValue({ code: SUCCESS_CODE, data: makeMedia(9) })
    mediaPageState.currentPage = 4

    const uploaded = await mediaPageState.uploadFile(file)

    expect(uploaded).toBe(true)
    expect(mediaApiMock.upload).toHaveBeenCalledWith(file)
    expect(toastMock.success).toHaveBeenCalledWith('文件上传成功')
    expect(mediaPageState.currentPage).toBe(1)
    expect(mediaPageState.isUploading).toBe(false)
    await vi.waitFor(() =>
      expect(mediaApiMock.list).toHaveBeenCalledWith({ page: 1, pageSize: MEDIA_PAGE_SIZE })
    )
  })

  it('上传失败时提示后端文案且不刷新列表', async () => {
    mediaApiMock.upload.mockResolvedValue({ code: 400, message: '文件类型不支持' })

    const uploaded = await mediaPageState.uploadFile(makeUploadFile())

    expect(uploaded).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('文件类型不支持')
    expect(mediaApiMock.list).not.toHaveBeenCalled()
    expect(mediaPageState.isUploading).toBe(false)
  })

  it('上传请求抛错时提示网络错误并复位上传态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    mediaApiMock.upload.mockRejectedValue(new Error('断网'))

    const uploaded = await mediaPageState.uploadFile(makeUploadFile())

    expect(uploaded).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE)
    expect(mediaPageState.isUploading).toBe(false)
    consoleError.mockRestore()
  })

  it('上传进行中重复提交直接忽略', async () => {
    mediaPageState.isUploading = true

    const uploaded = await mediaPageState.uploadFile(makeUploadFile())

    expect(uploaded).toBe(false)
    expect(mediaApiMock.upload).not.toHaveBeenCalled()
    expect(mediaApiMock.list).not.toHaveBeenCalled()
  })
})

describe('媒体页状态：复制链接', () => {
  beforeEach(() => {
    resetState()
    vi.stubGlobal('navigator', { clipboard: clipboardMock })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('复制成功后提示链接已复制', async () => {
    clipboardMock.writeText.mockResolvedValue(undefined)

    await mediaPageState.copyFileUrl(makeMedia(1))

    expect(clipboardMock.writeText).toHaveBeenCalledWith('https://cdn.example.com/1.png')
    expect(toastMock.success).toHaveBeenCalledWith('链接已复制')
  })

  it('复制失败时提示手动复制', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    clipboardMock.writeText.mockRejectedValue(new Error('剪贴板不可用'))

    await mediaPageState.copyFileUrl(makeMedia(1))

    expect(toastMock.error).toHaveBeenCalledWith('复制失败，请手动复制')
    consoleError.mockRestore()
  })
})

describe('媒体页状态：删除', () => {
  beforeEach(resetState)

  it('删除成功后提示并刷新列表', async () => {
    mediaApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })

    const removed = await mediaPageState.removeMedia(1)

    expect(removed).toBe(true)
    expect(mediaApiMock.delete).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('文件删除成功')
    await vi.waitFor(() => expect(mediaApiMock.list).toHaveBeenCalledTimes(1))
  })

  it('删除失败时提示后端文案且不刷新列表', async () => {
    mediaApiMock.delete.mockResolvedValue({ code: 400, message: '文件不存在' })

    const removed = await mediaPageState.removeMedia(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('文件不存在')
    expect(mediaApiMock.list).not.toHaveBeenCalled()
  })

  it('删除请求抛错时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    mediaApiMock.delete.mockRejectedValue(new Error('断网'))

    const removed = await mediaPageState.removeMedia(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE)
    consoleError.mockRestore()
  })
})
