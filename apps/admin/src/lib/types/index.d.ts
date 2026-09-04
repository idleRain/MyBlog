// 类型定义统一导出

// 基础类型
export * from './common'

// 认证和权限类型
export * from './auth'

// 管理后台类型
export * from './admin'

// 接口类型唯一来源为 @myblog/api，禁止应用层定义同构影子类型（铁律 A2）。
export * from '@myblog/api/modules/user/types'

// 重新导出常用类型，提供更好的开发体验
export type {
  User,
  UserRole,
  UserStatus,
  LoginData,
  UserListData
} from '@myblog/api/modules/user/types'

export type { AuthState, PermissionCheck, UserPermissionContext } from './auth'

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
