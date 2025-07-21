// 通用响应接口
export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
  error?: string
}
