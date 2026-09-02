import type { ApiResponse } from '@myblog/shared'

// 分类状态枚举，0 表示隐藏，1 表示显示。
export type CategoryStatus = 0 | 1

// 分类信息接口，字段与后端 model.Category 的 JSON tag 一致。
export interface Category {
  id: number
  name: string
  slug: string
  description: string
  coverImage: string
  parentId: number | null
  rootId: number | null
  level: number
  path: string
  sortOrder: number
  status: CategoryStatus
  articleCount: number
  isFeatured: boolean
  seoTitle: string
  seoDescription: string
  createdAt: string
  updatedAt: string
}

// 分类树节点，在分类字段基础上递归携带子节点。
export interface CategoryTreeNode extends Category {
  children: CategoryTreeNode[]
}

// 分类树响应数据
export interface CategoryTreeData {
  tree: CategoryTreeNode[]
}

// 分类列表数据
export interface CategoryListData {
  page: number
  pageSize: number
  total: number
  categories: Category[]
}

// 分类列表查询参数
export interface ListCategoriesRequest {
  page?: number
  pageSize?: number
  status?: CategoryStatus | null
  search?: string
}

// 创建分类请求参数
export interface CreateCategoryRequest {
  name: string
  slug?: string
  description?: string
  coverImage?: string
  parentId?: number | null
  sortOrder?: number | null
  status?: CategoryStatus | null
  isFeatured?: boolean | null
  seoTitle?: string
  seoDescription?: string
}

// 更新分类请求参数，可选字段显式传入才更新。
export interface UpdateCategoryRequest {
  id: number
  name?: string | null
  slug?: string | null
  description?: string | null
  coverImage?: string | null
  sortOrder?: number | null
  status?: CategoryStatus | null
  isFeatured?: boolean | null
  seoTitle?: string | null
  seoDescription?: string | null
}

// 各响应类型
export type CategoryResponse = ApiResponse<Category>
export type CategoryTreeResponse = ApiResponse<CategoryTreeData>
export type CategoryListResponse = ApiResponse<CategoryListData>
export type DeleteCategoryResponse = ApiResponse<null>
