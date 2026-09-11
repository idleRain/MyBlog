<script lang="ts">
import type { Comment, CommentListData } from '@myblog/api/modules/comment/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { formatDate } from '$lib/utils/format-date'
import { SvelteMap } from 'svelte/reactivity'
import { authStore } from '$lib/stores/auth'
import { toast } from 'svelte-sonner'
import { CommentAPI } from '$lib/api'

// 评论内容的长度上限，与后端 binding 校验保持一致。
const COMMENT_MAX_LENGTH = 2000

// 评论者姓名长度上限，与后端 binding 校验保持一致。
const AUTHOR_NAME_MAX_LENGTH = 50

// 评论加载失败时的降级分页大小。
const FALLBACK_PAGE_SIZE = 20

// 评论线程视图模型：根评论携带回复列表，由平铺评论派生。
interface CommentThread {
  root: Comment
  replies: Comment[]
}

interface Props {
  articleId: number
  initialComments: CommentListData | null
}

let { articleId, initialComments }: Props = $props()

// 评论分页快照：SSR 首屏数据作为初始页，客户端刷新后以本地状态整体接管。
interface CommentPage {
  comments: Comment[]
  total: number
  page: number
  pageSize: number
}

// 评论分页的空值降级，用于首屏数据缺失的场景。
const EMPTY_PAGE: CommentPage = {
  comments: [],
  total: 0,
  page: 1,
  pageSize: FALLBACK_PAGE_SIZE
}

let localPage = $state<CommentPage | null>(null)

// 当前评论分页：本地接管前展示 SSR 首屏数据。
const pageState = $derived<CommentPage>(localPage ?? initialComments ?? EMPTY_PAGE)
const comments = $derived(pageState.comments)
const total = $derived(pageState.total)
const currentPage = $derived(pageState.page)
const pageSize = $derived(pageState.pageSize)

// 表单状态：登录用户仅填写内容，游客补齐姓名与邮箱。
let content = $state('')
let authorName = $state('')
let authorEmail = $state('')
let submitting = $state(false)

// 订阅认证状态，登录用户隐藏游客字段。
let isAuthenticated = $state(false)
authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
})

// 是否存在更多评论页，决定加载更多按钮的展示。
const hasMore = $derived(currentPage * pageSize < total)

// 平铺评论组织为两级线程，根评论在前回复随后。
const threads = $derived.by(() => {
  const result: CommentThread[] = []
  const rootIndex = new SvelteMap<number, number>()
  for (const comment of comments) {
    if (comment.parentId === null) {
      rootIndex.set(comment.id, result.length)
      result.push({ root: comment, replies: [] })
      continue
    }

    const threadId = comment.rootId ?? comment.parentId
    const position = threadId !== null ? rootIndex.get(threadId) : undefined
    if (position !== undefined) {
      result[position]?.replies.push(comment)
    }
  }
  return result
})

// 评论者展示名：注册用户优先昵称，其次用户名，最后游客姓名。
function displayName(comment: Comment): string {
  return comment.user?.nickname || comment.user?.username || comment.authorName
}

// 加载指定页评论并整体接管列表状态，用于发表后的刷新。
async function reloadComments() {
  const response = await CommentAPI.listByArticle(articleId, {
    page: 1,
    pageSize
  })
  if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) return

  localPage = {
    comments: response.data.comments,
    total: response.data.total,
    page: response.data.page,
    pageSize: response.data.pageSize
  }
}

// 追加下一页评论，供加载更多按钮调用。
async function loadMore() {
  const response = await CommentAPI.listByArticle(articleId, {
    page: currentPage + 1,
    pageSize
  })
  if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) return

  localPage = {
    comments: [...comments, ...response.data.comments],
    total: response.data.total,
    page: response.data.page,
    pageSize: response.data.pageSize
  }
}

// 提交评论，登录用户由后端经令牌绑定身份，成功后刷新列表。
async function handleSubmit(event: SubmitEvent) {
  event.preventDefault()
  if (submitting) return

  submitting = true
  try {
    // 游客通道显式携带姓名与邮箱，登录通道不带游客字段以保持契约清晰。
    const response = await CommentAPI.create({
      articleId,
      content,
      ...(isAuthenticated ? {} : { authorName, authorEmail })
    })

    if (response.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(response.message || '评论提交失败，请稍后重试')
      return
    }

    toast.success('评论已提交，审核通过后展示')
    content = ''
    authorName = ''
    authorEmail = ''
    await reloadComments()
  } finally {
    submitting = false
  }
}
</script>

<section id="comments" class="mt-16 border-t border-border pt-10">
  <!-- 板块标题栏：与目录页标题栏同构。 -->
  <div class="mb-8 flex items-end justify-between border-b border-line pb-4">
    <h2 class="font-display text-2xl font-black">评论</h2>
    <span class="font-mono text-sm text-muted-foreground">共 {total} 条</span>
  </div>

  <!-- 发表表单：游客通道要求姓名，登录通道隐藏游客字段。 -->
  <form onsubmit={handleSubmit} class="mb-12">
    {#if !isAuthenticated}
      <div class="mb-4 grid gap-4 sm:grid-cols-2">
        <label class="block">
          <span class="sr-only">姓名</span>
          <input
            type="text"
            bind:value={authorName}
            required
            maxlength={AUTHOR_NAME_MAX_LENGTH}
            placeholder="姓名 *"
            class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
        <label class="block">
          <span class="sr-only">邮箱</span>
          <input
            type="email"
            bind:value={authorEmail}
            placeholder="邮箱（不公开）"
            class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
      </div>
    {/if}
    <label class="block">
      <span class="sr-only">评论内容</span>
      <textarea
        bind:value={content}
        required
        maxlength={COMMENT_MAX_LENGTH}
        rows={4}
        placeholder={isAuthenticated ? '说点什么吧' : '以访客身份说点什么吧'}
        class="w-full resize-y border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
      ></textarea>
    </label>
    <div class="mt-4 flex items-center justify-between gap-4">
      <p class="font-mono text-xs text-muted-foreground">评论提交后需经审核才会展示</p>
      <button
        type="submit"
        disabled={submitting}
        class="shrink-0 bg-signal px-6 py-2.5 text-sm font-bold text-signal-foreground transition-[background-color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal/90 active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60"
      >
        {submitting ? '提交中…' : '发表评论'}
      </button>
    </div>
  </form>

  {#if threads.length === 0}
    <!-- 空状态：以编辑口吻引导参与讨论。 -->
    <div class="border-l-2 border-signal py-4 pl-6">
      <p class="text-sm leading-relaxed text-muted-foreground">还没有评论，来说点什么吧。</p>
    </div>
  {:else}
    <ol class="divide-y divide-line">
      {#each threads as thread (thread.root.id)}
        <li class="py-6">
          <div class="flex items-baseline justify-between gap-4">
            <p class="font-display text-base font-bold">{displayName(thread.root)}</p>
            <span class="shrink-0 font-mono text-xs text-muted-foreground">
              {formatDate(thread.root.createdAt)}
            </span>
          </div>
          <p class="mt-2 text-sm leading-relaxed whitespace-pre-line text-muted-foreground">
            {thread.root.content}
          </p>

          {#if thread.replies.length > 0}
            <ol class="mt-4 space-y-4 border-l border-line pl-5">
              {#each thread.replies as reply (reply.id)}
                <li>
                  <div class="flex items-baseline justify-between gap-4">
                    <p class="text-sm font-bold">{displayName(reply)}</p>
                    <span class="shrink-0 font-mono text-xs text-muted-foreground">
                      {formatDate(reply.createdAt)}
                    </span>
                  </div>
                  <p
                    class="mt-1.5 text-sm leading-relaxed whitespace-pre-line text-muted-foreground"
                  >
                    {reply.content}
                  </p>
                </li>
              {/each}
            </ol>
          {/if}
        </li>
      {/each}
    </ol>

    {#if hasMore}
      <button
        type="button"
        onclick={loadMore}
        class="mt-8 text-sm font-bold text-signal underline-offset-4 hover:underline"
      >
        查看更多评论
      </button>
    {/if}
  {/if}
</section>
