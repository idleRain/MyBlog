// 组件相关类型定义

import type { User, UserRole, UserStatus } from '$lib/api/modules/user/types'
import type { Component } from 'svelte'
import type { Snippet } from 'svelte'

/**
 * 基础组件 Props
 */
export interface BaseComponentProps {
  class?: string
  id?: string
  'data-testid'?: string
}

/**
 * 带子内容的组件 Props
 */
export interface ComponentWithChildren extends BaseComponentProps {
  children: Snippet
}

/**
 * 用户头像组件 Props
 */
export interface UserAvatarProps extends BaseComponentProps {
  user: Pick<User, 'avatar' | 'username' | 'nickname'>
  size?: 'sm' | 'md' | 'lg' | 'xl'
  fallbackIcon?: Component
  showTooltip?: boolean
}

/**
 * 用户角色徽章组件 Props
 */
export interface UserRoleBadgeProps extends BaseComponentProps {
  role: UserRole
  variant?: 'default' | 'outline' | 'secondary'
  showIcon?: boolean
}

/**
 * 用户状态徽章组件 Props
 */
export interface UserStatusBadgeProps extends BaseComponentProps {
  status: UserStatus
  variant?: 'default' | 'outline' | 'secondary'
  showText?: boolean
}

/**
 * 权限守卫组件 Props
 */
export interface PermissionGuardProps {
  user?: User | null
  requiredPermissions?: string[]
  requiredRole?: UserRole
  fallback?: Snippet
  children: Snippet
  onAccessDenied?: (reason: string) => void
}

/**
 * 数据表格列定义
 */
export interface TableColumn<T = any> {
  key: keyof T | string
  title: string
  sortable?: boolean
  width?: string | number
  render?: (value: any, record: T, index: number) => Snippet | string | number
  align?: 'left' | 'center' | 'right'
}

/**
 * 数据表格 Props
 */
export interface DataTableProps<T = any> extends BaseComponentProps {
  data: T[]
  columns: TableColumn<T>[]
  loading?: boolean
  empty?: Snippet
  pagination?: {
    page: number
    limit: number
    total: number
    onPageChange: (page: number) => void
  }
  selection?: {
    selectedKeys: (string | number)[]
    onSelectionChange: (keys: (string | number)[]) => void
    getRowKey: (record: T) => string | number
  }
  actions?: {
    title: string
    render: (record: T, index: number) => Snippet
  }
}

/**
 * 表单字段类型
 */
export type FormFieldType =
  | 'text'
  | 'email'
  | 'password'
  | 'number'
  | 'select'
  | 'textarea'
  | 'checkbox'
  | 'radio'
  | 'date'
  | 'file'

/**
 * 表单字段选项
 */
export interface FormFieldOption {
  label: string
  value: string | number
  disabled?: boolean
}

/**
 * 表单字段定义
 */
export interface FormField {
  name: string
  label: string
  type: FormFieldType
  required?: boolean
  placeholder?: string
  description?: string
  options?: FormFieldOption[]
  validation?: {
    min?: number
    max?: number
    pattern?: RegExp
    custom?: (value: any) => string | null
  }
}

/**
 * 动态表单组件 Props
 */
export interface DynamicFormProps<T = Record<string, any>> extends BaseComponentProps {
  fields: FormField[]
  initialValues?: Partial<T>
  onSubmit: (values: T) => Promise<void> | void
  onCancel?: () => void
  loading?: boolean
  submitText?: string
  cancelText?: string
  layout?: 'horizontal' | 'vertical' | 'inline'
}

/**
 * 确认对话框 Props
 */
export interface ConfirmDialogProps {
  open: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'default' | 'destructive'
  onConfirm: () => Promise<void> | void
  onCancel: () => void
  loading?: boolean
}

/**
 * 页面头部 Props
 */
export interface PageHeaderProps extends BaseComponentProps {
  title: string
  description?: string
  breadcrumbs?: Array<{
    label: string
    href?: string
  }>
  actions?: Snippet
}

/**
 * 侧边栏菜单项组件 Props
 */
export interface SidebarMenuItemProps extends BaseComponentProps {
  item: {
    id: string
    title: string
    icon: Component
    url: string
    badge?: string | number
    children?: Array<{
      id: string
      title: string
      url: string
    }>
  }
  isActive?: boolean
  isExpanded?: boolean
  onToggleExpand?: () => void
  level?: number
}

/**
 * 通知消息类型
 */
export type NotificationType = 'success' | 'error' | 'warning' | 'info'

/**
 * 通知组件 Props
 */
export interface NotificationProps {
  id: string
  type: NotificationType
  title: string
  message?: string
  duration?: number
  closable?: boolean
  onClose?: (id: string) => void
}

/**
 * 加载状态组件 Props
 */
export interface LoadingProps extends BaseComponentProps {
  size?: 'sm' | 'md' | 'lg'
  text?: string
  overlay?: boolean
}

/**
 * 空状态组件 Props
 */
export interface EmptyStateProps extends BaseComponentProps {
  icon?: Component
  title: string
  description?: string
  action?: {
    text: string
    onClick: () => void
  }
}

/**
 * 统计卡片组件 Props
 */
export interface StatsCardProps extends BaseComponentProps {
  title: string
  value: string | number
  icon?: Component
  trend?: {
    value: number
    label: string
    type: 'up' | 'down' | 'neutral'
  }
  color?: string
  loading?: boolean
}
