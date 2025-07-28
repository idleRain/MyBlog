// 类型定义统一导出

// 基础类型
export * from './common'

// 认证和权限类型
export * from './auth'

// 管理后台类型
export * from './admin'

// 组件类型
export * from './components'

// API相关类型
export * from './api'
export * from '$lib/api/modules/user/types'

// 重新导出常用类型，提供更好的开发体验
export type {
  User,
  UserRole,
  UserStatus,
  LoginData,
  UserListData
} from '$lib/api/modules/user/types'

export type {
  AuthState,
  PermissionCheck,
  UserPermissionContext,
  PERMISSIONS,
  ROLE_PERMISSIONS,
  ROLE_CONFIG
} from './auth'

export type {
  SidebarMenuItem,
  DashboardCard,
  QuickAction,
  SystemStats,
  UserTableConfig
} from './admin'

export type {
  PaginationParams,
  PaginationData,
  TableColumn,
  TableAction,
  BatchAction,
  NotificationMessage,
  FormState
} from './common'

export type {
  BaseComponentProps,
  ComponentWithChildren,
  UserAvatarProps,
  UserRoleBadgeProps,
  UserStatusBadgeProps,
  PermissionGuardProps,
  DataTableProps,
  DynamicFormProps,
  ConfirmDialogProps,
  PageHeaderProps,
  NotificationProps,
  LoadingProps,
  EmptyStateProps,
  StatsCardProps
} from './components'

export type {
  BaseApiResponse,
  PaginationRequest,
  PaginationResponse,
  BaseListRequest,
  ApiError,
  RequestStatus,
  AsyncState,
  AsyncStateActions,
  ApiResponse,
  ListResponse,
  DetailResponse,
  CreateResponse,
  UpdateResponse,
  DeleteResponse,
  BatchResponse,
  HttpMethod,
  ApiRequestConfig
} from './api'
