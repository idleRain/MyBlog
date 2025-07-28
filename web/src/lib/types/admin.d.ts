// 管理后台相关类型定义

import type { TreeNode, TableColumn, TableAction, BatchAction } from './common'
import type { User, UserRole } from '$lib/api/modules/user/types'

// 侧边栏菜单项
export interface SidebarMenuItem {
  id: string
  title: string
  icon: any // Lucide图标组件
  url: string
  roles: UserRole[]
  badge?: string
  children?: SidebarMenuItem[]
}

// 仪表盘卡片数据
export interface DashboardCard {
  id: string
  title: string
  value: number | string
  icon: any
  trend?: {
    value: number
    type: 'up' | 'down'
    text: string
  }
  color?: string
}

// 快速操作配置
export interface QuickAction {
  id: string
  title: string
  description: string
  icon: any
  action: () => void
  color: string
  roles: UserRole[]
}

// 最近活动项
export interface RecentActivity {
  id: string
  action: string
  user: string
  time: string
  type: 'login' | 'create' | 'update' | 'delete' | 'register'
  details?: Record<string, any>
}

// 系统统计信息
export interface SystemStats {
  totalUsers: number
  totalPosts: number
  totalComments: number
  totalCategories: number
  totalTags: number
  activeUsers: number
  todayVisits: number
  systemStatus: 'normal' | 'warning' | 'error' | 'maintenance'
  cpuUsage?: number
  memoryUsage?: number
  diskUsage?: number
}

// 用户管理表格配置
export interface UserTableConfig {
  columns: TableColumn<User>[]
  actions: TableAction<User>[]
  batchActions: BatchAction<User>[]
  filters: {
    roles: UserRole[]
    statuses: Array<{ label: string; value: number }>
  }
}

// 权限管理配置
export interface PermissionConfig {
  resource: string
  actions: string[]
  description: string
}

// 角色管理配置
export interface RoleManagementConfig {
  role: UserRole
  name: string
  description: string
  permissions: string[]
  userCount: number
  canEdit: boolean
  canDelete: boolean
}

// 系统设置项
export interface SystemSetting {
  key: string
  title: string
  description?: string
  type: 'text' | 'number' | 'boolean' | 'select' | 'textarea' | 'file'
  value: any
  options?: Array<{ label: string; value: any }>
  required?: boolean
  validation?: {
    min?: number
    max?: number
    pattern?: string
    message?: string
  }
  group: string
}

// 系统设置分组
export interface SettingGroup {
  key: string
  title: string
  description?: string
  icon?: any
  settings: SystemSetting[]
}

// 日志条目
export interface LogEntry {
  id: string
  timestamp: string
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  module: string
  userId?: number
  username?: string
  ip?: string
  userAgent?: string
  details?: Record<string, any>
}

// 日志查询参数
export interface LogQueryParams {
  level?: string
  module?: string
  userId?: number
  startTime?: string
  endTime?: string
  keyword?: string
  page: number
  pageSize: number
}

// 监控指标
export interface MonitoringMetric {
  name: string
  value: number
  unit: string
  timestamp: string
  trend?: 'up' | 'down' | 'stable'
  threshold?: {
    warning: number
    critical: number
  }
}

// 性能监控数据
export interface PerformanceData {
  cpu: MonitoringMetric[]
  memory: MonitoringMetric[]
  disk: MonitoringMetric[]
  network: MonitoringMetric[]
  responseTime: MonitoringMetric[]
  errorRate: MonitoringMetric[]
}

// 备份信息
export interface BackupInfo {
  id: string
  name: string
  type: 'full' | 'incremental' | 'database' | 'files'
  size: number
  createdAt: string
  status: 'creating' | 'completed' | 'failed'
  downloadUrl?: string
}

// 通知设置
export interface NotificationSetting {
  type: 'email' | 'sms' | 'webhook'
  enabled: boolean
  events: string[]
  config: Record<string, any>
}

// 安全设置
export interface SecuritySetting {
  key: string
  title: string
  enabled: boolean
  description: string
  config?: Record<string, any>
}

// 插件信息
export interface PluginInfo {
  id: string
  name: string
  version: string
  description: string
  author: string
  enabled: boolean
  configurable: boolean
  dependencies?: string[]
  config?: Record<string, any>
}

// 主题配置
export interface ThemeConfig {
  name: string
  primaryColor: string
  secondaryColor: string
  backgroundColor: string
  textColor: string
  borderRadius: number
  fontSize: number
  fontFamily: string
}

// 管理员操作记录
export interface AdminAction {
  id: string
  adminId: number
  adminName: string
  action: string
  target: string
  targetId?: string
  details: Record<string, any>
  timestamp: string
  ip: string
  userAgent: string
}

// 工作流状态
export interface WorkflowStatus {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  startTime: string
  endTime?: string
  result?: any
  error?: string
}
