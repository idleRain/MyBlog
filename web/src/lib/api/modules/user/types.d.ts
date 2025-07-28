import { ApiResponse } from '#/index'

// 用户角色枚举
export type UserRole = 'user' | 'editor' | 'admin' | 'superadmin'

// 用户状态枚举
export type UserStatus = 0 | 1 // 0-禁用, 1-正常

// 用户角色信息
export interface RoleInfo {
  name: string
  color: 'default' | 'destructive' | 'outline' | 'secondary'
  level: number
}

// 用户信息接口
export interface User {
  id: number
  username: string
  email: string
  nickname: string
  avatar: string
  birthday: string
  role: UserRole
  status: UserStatus
  createdAt: string
  updatedAt: string
}

// 用户列表数据
export interface UserListData {
  page: number
  pageSize: number
  pages: number
  total: number
  users: User[]
}

// 登录请求参数
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应数据
export interface LoginData {
  user: User
  accessToken: string
  refreshToken: string
  expiresIn: number
}

// 获取用户信息请求参数
export interface GetUserByIdRequest {
  id: number
}

// 删除用户请求参数
export interface DeleteUserRequest {
  id: number
}

// 刷新令牌请求参数
export interface RefreshTokenRequest {
  refreshToken: string
}

// 刷新令牌响应数据
export interface RefreshTokenData {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

// 创建用户请求参数
export interface CreateUserRequest {
  username: string
  email: string
  password: string
  nickname?: string
  birthday?: string
  role?: UserRole
}

// 更新用户请求参数
export interface UpdateUserRequest {
  id: number
  username: string
  email: string
  password?: string // 可选，留空则不更新
  nickname?: string
  birthday?: string
  role?: UserRole
  status?: UserStatus
}

// 注册请求参数
export interface RegisterRequest {
  username: string
  email: string
  password: string
  nickname?: string
  birthday?: string
  role?: UserRole
}

// 各类型的响应接口
export type UserListResponse = ApiResponse<UserListData>
export type LoginResponse = ApiResponse<LoginData>
export type RegisterResponse = ApiResponse<User>
export type UserResponse = ApiResponse<User>
export type CreateUserResponse = ApiResponse<User>
export type UpdateUserResponse = ApiResponse<User>
export type DeleteUserResponse = ApiResponse<null>
export type RefreshTokenResponse = ApiResponse<RefreshTokenData>
export type LogoutResponse = ApiResponse<null>

// 权限相关类型
export interface Permission {
  id: string
  name: string
  resource: string
  action: string
}

// 角色权限映射
export interface RolePermissions {
  role: UserRole
  permissions: Permission[]
}

// 用户统计信息
export interface UserStats {
  totalUsers: number
  activeUsers: number
  newUsersToday: number
  disabledUsers: number
}

// 仪表盘统计数据
export interface DashboardStats {
  totalUsers: number
  totalPosts: number
  activeUsers: number
  systemStatus: 'normal' | 'warning' | 'error'
}

// 快速操作项
export interface QuickAction {
  title: string
  description: string
  icon: any // Lucide图标组件
  action: () => void
  color: string
  roles: UserRole[]
}

// 导航菜单项
export interface NavigationItem {
  title: string
  icon: any // Lucide图标组件
  url: string
  roles: UserRole[]
}

// 兼容旧接口
export type Data = UserListData
export type Res = UserListResponse
