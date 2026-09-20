import type { ApiResponse } from '@myblog/shared'

// 分类状态枚举，0 表示隐藏，1 表示显示。
export type CategoryStatus = 0 | 1

// 分类单语言翻译行，与后端 model.CategoryTranslation 的 JSON tag 一致，
// 仅管理端全量包请求携带，公共请求按语言输出后不包含该结构。
export interface CategoryTranslation {
  id: number
  categoryId: number
  locale: string
  name: string | null
  description: string | null
  seoTitle: string | null
  seoDescription: string | null
  createdAt: string
  updatedAt: string
}

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
  // 全量翻译行数组，仅 Accept-Language: * 的管理端请求返回。
  translations?: CategoryTranslation[]
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

// 分类单语言翻译补丁，字段可选，提供即更新，未提供保留既有翻译值。
export interface CategoryI18nPatch {
  name?: string
  description?: string
  seoTitle?: string
  seoDescription?: string
}

// 按语言组织的翻译字段包，键为语言标识，缺省语言走主字段不允许出现。
export type CategoryI18nPayload = Record<string, CategoryI18nPatch>

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
  i18n?: CategoryI18nPayload
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
  i18n?: CategoryI18nPayload
}

// 各响应类型
export type CategoryResponse = ApiResponse<Category>
export type CategoryTreeResponse = ApiResponse<CategoryTreeData>
export type CategoryListResponse = ApiResponse<CategoryListData>
export type DeleteCategoryResponse = ApiResponse<null>
