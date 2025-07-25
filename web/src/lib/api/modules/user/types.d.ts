import { ApiResponse } from '#/index'

// 用户信息接口
export interface User {
  id: number
  username: string
  email: string
  nickname: string
  avatar: string
  birthday: string
  status: number
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

// 注册请求参数
export interface RegisterRequest {
  username: string
  email: string
  password: string
  nickname?: string
  birthday?: string
}

// 各类型的响应接口
export type UserListResponse = ApiResponse<UserListData>
export type LoginResponse = ApiResponse<LoginData>
export type RegisterResponse = ApiResponse<User>
export type UserResponse = ApiResponse<User>
export type RefreshTokenResponse = ApiResponse<RefreshTokenData>
export type LogoutResponse = ApiResponse<null>

// 兼容旧接口
export type Data = UserListData
export type Res = UserListResponse
