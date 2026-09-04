import type { KyInstance } from 'ky'
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  UserListResponse,
  UserResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  LogoutResponse,
  UpdateUserRequest,
  CreateUserRequest
} from './types.ts'

/**
 * 创建用户接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createUserAPI(request: KyInstance) {
  return {
    // 用户登录
    login(params: LoginRequest): Promise<LoginResponse> {
      return request.post('users/login', { json: params }).json()
    },

    // 创建用户（仅管理员）
    createUser(params: CreateUserRequest): Promise<RegisterResponse> {
      return request.post('users/create', { json: params }).json()
    },

    // 获取用户列表，keyword 非空时按用户名、邮箱或昵称模糊匹配。
    getUserList(page = 1, pageSize = 10, keyword = ''): Promise<UserListResponse> {
      return request.post('users/list', { json: { page, pageSize, keyword } }).json()
    },

    // 根据 ID 获取用户信息，统一使用 POST 方法。
    getUserById(id: number): Promise<UserResponse> {
      return request.post('users/get', { json: { id } }).json()
    },

    // 更新用户（仅管理员）
    updateUser(params: UpdateUserRequest): Promise<RegisterResponse> {
      return request.post('users/update', { json: params }).json()
    },

    // 删除用户，统一使用 POST 方法。
    deleteUser(id: number): Promise<{ code: number; message: string }> {
      return request.post('users/delete', { json: { id } }).json()
    },

    // 刷新令牌
    refreshToken(params: RefreshTokenRequest): Promise<RefreshTokenResponse> {
      return request.post('auth/refresh', { json: params }).json()
    },

    // 用户登出
    logout(): Promise<LogoutResponse> {
      return request.post('auth/logout').json()
    },

    // 兼容旧接口
    getUser(): Promise<UserListResponse> {
      return this.getUserList(1, 10)
    }
  }
}

export type UserAPI = ReturnType<typeof createUserAPI>

export type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  UserListResponse,
  UserResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  LogoutResponse,
  UpdateUserRequest,
  CreateUserRequest
}
