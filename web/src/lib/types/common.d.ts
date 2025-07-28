// 通用类型定义

// 分页参数
export interface PaginationParams {
  page: number
  pageSize: number
}

// 分页响应数据
export interface PaginationData<T = any> {
  page: number
  pageSize: number
  pages: number
  total: number
  data: T[]
}

// 排序参数
export interface SortParams {
  field: string
  order: 'asc' | 'desc'
}

// 搜索参数
export interface SearchParams {
  query?: string
  filters?: Record<string, any>
}

// 列表查询参数
export interface ListParams extends PaginationParams {
  search?: string
  sort?: SortParams
  filters?: Record<string, any>
}

// 操作状态
export type ActionStatus = 'idle' | 'loading' | 'success' | 'error'

// 表单字段错误
export interface FieldError {
  field: string
  message: string
}

// 表单状态
export interface FormState<T = any> {
  data: T
  errors: Record<string, string>
  isSubmitting: boolean
  isValid: boolean
}

// 模态框状态
export interface ModalState {
  isOpen: boolean
  title?: string
  content?: any
  confirmText?: string
  cancelText?: string
  onConfirm?: () => void
  onCancel?: () => void
}

// 通知类型
export type NotificationType = 'success' | 'error' | 'warning' | 'info'

// 通知消息
export interface NotificationMessage {
  id: string
  type: NotificationType
  title?: string
  message: string
  duration?: number
  action?: {
    label: string
    handler: () => void
  }
}

// 主题模式
export type ThemeMode = 'light' | 'dark' | 'system'

// 语言设置
export type Locale = 'zh' | 'en'

// 系统状态
export type SystemStatus = 'normal' | 'warning' | 'error' | 'maintenance'

// 颜色变体（shadcn-svelte）
export type ColorVariant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'

// 尺寸变体
export type SizeVariant = 'default' | 'sm' | 'lg' | 'icon'

// 组件Props基础类型
export interface BaseComponentProps {
  class?: string
  id?: string
}

// 可选的HTML属性
export type OptionalHTMLProps<T extends keyof HTMLElementTagNameMap> = Partial<
  HTMLElementTagNameMap[T]
>

// 表格列定义
export interface TableColumn<T = any> {
  key: keyof T
  title: string
  width?: string | number
  sortable?: boolean
  filterable?: boolean
  render?: (value: any, record: T, index: number) => any
  align?: 'left' | 'center' | 'right'
}

// 表格操作按钮
export interface TableAction<T = any> {
  key: string
  label: string
  icon?: any
  variant?: ColorVariant
  size?: SizeVariant
  disabled?: (record: T) => boolean
  visible?: (record: T) => boolean
  handler: (record: T, index: number) => void
}

// 批量操作
export interface BatchAction<T = any> {
  key: string
  label: string
  icon?: any
  variant?: ColorVariant
  confirmText?: string
  handler: (selectedItems: T[]) => void | Promise<void>
}

// 筛选器配置
export interface FilterConfig {
  key: string
  label: string
  type: 'text' | 'select' | 'date' | 'dateRange' | 'number'
  options?: Array<{ label: string; value: any }>
  placeholder?: string
}

// 文件上传状态
export interface UploadFile {
  id: string
  name: string
  size: number
  type: string
  status: 'pending' | 'uploading' | 'success' | 'error'
  progress?: number
  url?: string
  error?: string
}

// 键值对
export interface KeyValuePair<T = any> {
  key: string
  value: T
  label?: string
}

// 树形数据节点
export interface TreeNode<T = any> {
  id: string | number
  label: string
  children?: TreeNode<T>[]
  data?: T
  expanded?: boolean
  selected?: boolean
  disabled?: boolean
}

// 面包屑项目
export interface BreadcrumbItem {
  label: string
  href?: string
  icon?: any
}
