import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  UserListResponse,
  UserResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  LogoutResponse
} from './types'
import request from '$lib/service'

const UserAPI = {
  // 用户登录
  login(params: LoginRequest): Promise<LoginResponse> {
    return request
      .post('users/login', {
        json: params
      })
      .json()
  },

  // 创建用户（仅管理员）
  createUser(params: RegisterRequest): Promise<RegisterResponse> {
    return request
      .post('users/create', {
        json: params
      })
      .json()
  },

  // 用户注册（已禁用，仅保留代码兼容性）
  register(params: RegisterRequest): Promise<RegisterResponse> {
    // 注册功能已禁用，这个方法仅为保持代码完整性
    // 实际调用会被后端拒绝或重定向
    return request
      .post('users/create', {
        json: params
      })
      .json()
  },

  // 获取用户列表
  getUserList(page = 1, pageSize = 10): Promise<UserListResponse> {
    return request
      .post('users/list', {
        json: {
          page,
          pageSize
        }
      })
      .json()
  },

  // 根据ID获取用户信息 - 统一使用POST方法
  getUserById(id: number): Promise<UserResponse> {
    return request
      .post('users/get', {
        json: { id }
      })
      .json()
  },

  // 更新用户（仅管理员）
  updateUser(params: any): Promise<RegisterResponse> {
    return request
      .post('users/update', {
        json: params
      })
      .json()
  },

  // 删除用户 - 统一使用POST方法
  deleteUser(id: number): Promise<{ code: number; message: string }> {
    return request
      .post('users/delete', {
        json: { id }
      })
      .json()
  },

  // 刷新令牌
  refreshToken(params: RefreshTokenRequest): Promise<RefreshTokenResponse> {
    return request
      .post('auth/refresh', {
        json: params
      })
      .json()
  },

  // 用户登出
  logout(): Promise<LogoutResponse> {
    return request.post('auth/logout').json()
  },

  // 兼容旧接口
  getUser() {
    return this.getUserList(1, 10)
  }
}

export default UserAPI
