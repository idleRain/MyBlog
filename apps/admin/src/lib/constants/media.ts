// 媒体模块常量：分页容量、类型筛选项与上传文件约束，供媒体页与网格组件共用。

// 媒体库每页加载数量，与分页组件取值保持一致。
export const MEDIA_PAGE_SIZE = 12

// 图片 MIME 前缀，决定列表项是否走缩略图预览。
export const MEDIA_IMAGE_MIME_PREFIX = 'image/'

// 视频 MIME 前缀，决定列表项使用视频图标。
export const MEDIA_VIDEO_MIME_PREFIX = 'video/'

// 文档 MIME 前缀，未命中图片与视频时按文档图标展示。
export const MEDIA_APPLICATION_MIME_PREFIX = 'application/'

// 媒体类型筛选项，value 为传给后端 mimeType 的 MIME 前缀，空串表示不筛选。
export const MEDIA_TYPE_FILTERS: Array<{ value: string; label: string }> = [
  { value: '', label: '全部' },
  { value: MEDIA_IMAGE_MIME_PREFIX, label: '图片' },
  { value: MEDIA_VIDEO_MIME_PREFIX, label: '视频' },
  { value: MEDIA_APPLICATION_MIME_PREFIX, label: '文档' }
]

// 文件选择框接受的文件类型，与后端媒体上传的白名单保持同一口径。
export const MEDIA_FILE_ACCEPT = 'image/*,video/*,audio/*,application/pdf,text/*'
