/// <reference types="vite/client" />

declare interface ImportMetaEnv {
  // 端口号
  readonly VITE_SERVER_PORT: string
  // 代理地址
  readonly VITE_PROXY_URL: string
  // 接口地址
  readonly VITE_BASE_URL: string
  // 请求超时时间
  readonly VITE_REQUEST_TIMEOUT: string
}
