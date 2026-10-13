import type {
  AdminComment,
  CommentActionResponse,
  CommentStatus
} from '@myblog/api/modules/comment/types'
import {
  COMMENT_ACTION_LABELS,
  COMMENT_PAGE_SIZE,
  type CommentAction
} from '$lib/constants/comment'
import { debounce } from '@myblog/shared'
import { toast } from 'svelte-sonner'

import { CommentAPI } from '$lib/api'

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 审核动作到接口方法的映射，spam 对应 markSpam 方法名。
const ACTION_METHODS: Record<
  Exclude<CommentAction, 'delete'>,
  (id: number) => Promise<CommentActionResponse>
> = {
  approve: CommentAPI.approve,
  reject: CommentAPI.reject,
  spam: CommentAPI.markSpam,
  trash: CommentAPI.trash
}

// 除永久删除外的审核动作，删除需经页面确认对话框后另行提交。
type CommentReviewAction = Exclude<CommentAction, 'delete'>

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface CommentPageData {
  comments: AdminComment[]
  total: number
}

// 空页面数据，load 失败时使用。
export function emptyCommentPageData(): CommentPageData {
  return { comments: [], total: 0 }
}

// 评论管理页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class CommentPageState {
  // 管理端列表在公开契约之上携带审计字段与文章归属，列表项按 AdminComment 声明。
  comments = $state<AdminComment[]>([])
  total = $state(0)
  currentPage = $state(1)
  keyword = $state('')
  // 状态筛选使用审核状态枚举值：空串为全部，与筛选选项中的空值项对应。
  statusFilter = $state<CommentStatus | ''>('')
  isLoading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径；无评论时仍按一页展示。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / COMMENT_PAGE_SIZE))
  }

  // 是否存在生效中的筛选条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.keyword.trim() !== '' || this.statusFilter !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: CommentPageData) {
    this.comments = data.comments
    this.total = data.total
  }

  // 加载评论列表，携带状态与关键词筛选。
  async loadComments() {
    this.isLoading = true
    try {
      const response = await CommentAPI.adminList({
        page: this.currentPage,
        pageSize: COMMENT_PAGE_SIZE,
        ...(this.statusFilter ? { status: this.statusFilter } : {}),
        ...(this.keyword.trim() ? { keyword: this.keyword.trim() } : {})
      })
      if (response.code === 200 && response.data) {
        this.comments = response.data.comments ?? []
        this.total = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载评论失败')
      }
    } catch (error) {
      console.error('加载评论失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 搜索词防抖应用，连续输入只在停顿后请求一次。
  applyKeywordDebounced = debounce(() => {
    this.currentPage = 1
    void this.loadComments()
  }, SEARCH_DEBOUNCE_MS)

  // 状态筛选变更后回到第一页并重新加载。
  setStatusFilter(value: string) {
    this.statusFilter = value as CommentStatus | ''
    this.currentPage = 1
    void this.loadComments()
  }

  // 重置全部筛选条件并回到第一页。
  resetFilters() {
    this.keyword = ''
    this.statusFilter = ''
    this.currentPage = 1
    void this.loadComments()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadComments()
  }

  // 提交审核动作，成功后按动作文案提示并刷新列表。
  async runAction(comment: AdminComment, action: CommentReviewAction) {
    try {
      const response = await ACTION_METHODS[action](comment.id)
      if (response.code === 200) {
        toast.success(`${COMMENT_ACTION_LABELS[action]}成功`)
        void this.loadComments()
      } else {
        toast.error(response.message || '操作失败')
      }
    } catch (error) {
      console.error('评论操作失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }

  // 永久删除评论，仅回收站内的评论可达该操作，成功后刷新当前页。
  async deleteComment(id: number): Promise<boolean> {
    try {
      const response = await CommentAPI.delete(id)
      if (response.code === 200) {
        toast.success('评论删除成功')
        void this.loadComments()
        return true
      }
      toast.error(response.message || '删除评论失败')
    } catch (error) {
      console.error('删除评论失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }
}

// 页面级共享状态实例。
export const commentPageState = new CommentPageState()
