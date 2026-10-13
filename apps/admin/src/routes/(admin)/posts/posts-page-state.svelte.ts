import type {
  Article,
  ArticleActionResponse,
  ArticleStatus
} from '@myblog/api/modules/article/types'
import { ARTICLE_PAGE_SIZE, type ArticleStatusAction } from '$lib/constants/article'
import { debounce } from '@myblog/shared'
import { toast } from 'svelte-sonner'
import { ArticleAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 排序字段以字符串承载空值表示默认排序，与后端 GetArticleListRequest 的 oneof 约束对齐。
type ArticleSortField =
  'created_at' | 'updated_at' | 'published_at' | 'view_count' | 'like_count' | ''

// 文章状态操作到接口方法的映射，统一走 /api/articles 端点，权限由后端判定。
const STATUS_ACTION_METHODS: Record<
  ArticleStatusAction,
  (id: number) => Promise<ArticleActionResponse>
> = {
  publish: ArticleAPI.publish,
  unpublish: ArticleAPI.unpublish,
  archive: ArticleAPI.archive,
  private: ArticleAPI.private
}

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface PostsPageData {
  articles: Article[]
  total: number
  canManageAllArticles: boolean
}

// 空页面数据，load 失败时使用。
export function emptyPostsPageData(): PostsPageData {
  return { articles: [], total: 0, canManageAllArticles: false }
}

// 文章管理页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class PostsPageState {
  articles = $state<Article[]>([])
  total = $state(0)
  currentPage = $state(1)
  search = $state('')
  status = $state<ArticleStatus | ''>('')
  sortBy = $state<ArticleSortField>('created_at')
  order = $state<'asc' | 'desc'>('desc')
  // 当前用户是否具备 article:manage 权限，用于控制状态筛选是否可用。
  canManageAllArticles = $state(false)
  isLoading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / ARTICLE_PAGE_SIZE))
  }

  // 是否存在生效中的搜索与状态筛选，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.search.trim() !== '' || this.status !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: PostsPageData) {
    this.articles = data.articles
    this.total = data.total
    this.canManageAllArticles = data.canManageAllArticles
  }

  // 加载文章列表，携带筛选、排序与分页参数。
  // 管理员可查看全部状态，其他角色由服务端强制只返回已发布文章。
  async loadArticles() {
    this.isLoading = true
    try {
      const response = await ArticleAPI.list({
        page: this.currentPage,
        pageSize: ARTICLE_PAGE_SIZE,
        status: this.status,
        sortBy: this.sortBy,
        order: this.order,
        ...(this.search.trim() ? { search: this.search.trim() } : {})
      })
      if (response.code === 200 && response.data) {
        this.articles = response.data.articles ?? []
        this.total = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载文章列表失败')
      }
    } catch (error) {
      console.error('加载文章列表失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 搜索词防抖应用，连续输入只在停顿后请求一次。
  applySearchDebounced = debounce(() => {
    this.currentPage = 1
    void this.loadArticles()
  }, SEARCH_DEBOUNCE_MS)

  // 状态、排序字段与排序方向变更后回到第一页并重新加载。
  applyFilterChange() {
    this.currentPage = 1
    void this.loadArticles()
  }

  // 重置全部筛选条件后重新加载。
  resetFilters() {
    this.search = ''
    this.status = ''
    this.sortBy = 'created_at'
    this.order = 'desc'
    this.currentPage = 1
    void this.loadArticles()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadArticles()
  }

  // 执行文章状态流转并提示结果，成功后刷新当前页。
  async applyStatusAction(article: Article, action: ArticleStatusAction) {
    try {
      const response = await STATUS_ACTION_METHODS[action](article.id)
      if (response.code === 200) {
        toast.success(response.data?.message || '操作成功')
        await this.loadArticles()
      } else {
        toast.error(response.message || '操作失败')
      }
    } catch (error) {
      console.error('文章状态操作失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }

  // 删除文章，由表格组件确认后触发，成功后刷新当前页。
  async removeArticle(article: Article) {
    try {
      const response = await ArticleAPI.delete(article.id)
      if (response.code === 200) {
        toast.success('文章删除成功')
        await this.loadArticles()
      } else {
        toast.error(response.message || '删除文章失败')
      }
    } catch (error) {
      console.error('删除文章失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }
}

// 页面级共享状态实例。
export const postsPageState = new PostsPageState()
