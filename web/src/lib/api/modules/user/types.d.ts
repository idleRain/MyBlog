export interface Data {
  page: number
  pageSize: number
  pages: number
  total: number
  users: never[]
}

export interface Res {
  code: number
  message: string
  data: Data
}
