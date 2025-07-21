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
  token: string
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

// 兼容旧接口
export type Data = UserListData
export type Res = UserListResponse
