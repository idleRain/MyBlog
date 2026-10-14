<script lang="ts">
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { Bookmark, Heart } from '@lucide/svelte'
import { authStore } from '$lib/stores/auth'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { ArticleAPI } from '$lib/api'
import { m } from '$i18n'

interface Props {
  articleId: number
}

let { articleId }: Props = $props()

// 互动状态：仅在登录后向服务端拉取，游客统一视为未互动。
let isLiked = $state(false)
let isBookmarked = $state(false)
let busy = $state(false)

// 认证状态经 store 自动订阅读取，组件卸载时由 Svelte 自动退订，避免重挂载累积订阅。
const isAuthenticated = $derived($authStore.isAuthenticated)

// 登录状态下拉取互动状态，游客不发起状态请求。
$effect(() => {
  if (!isAuthenticated) return
  void refreshStates()
})

// 拉取当前点赞与收藏状态，失败时保持默认值。
async function refreshStates() {
  try {
    const [likeResponse, bookmarkResponse] = await Promise.all([
      ArticleAPI.isLiked(articleId),
      ArticleAPI.isBookmarked(articleId)
    ])
    if (likeResponse.code === RESPONSE_CODE_SUCCESS) {
      isLiked = likeResponse.data?.isLiked ?? false
    }
    if (bookmarkResponse.code === RESPONSE_CODE_SUCCESS) {
      isBookmarked = bookmarkResponse.data?.isBookmarked ?? false
    }
  } catch {
    // 状态拉取失败保持默认，交互时以服务端结果为准。
  }
}

// 未登录时引导前往登录页，登录方案落地前先走占位页。
function requireAuth() {
  toast.info(m['blog:actions.signInRequired']())
  void goto('/login')
}

// 切换点赞状态。
async function toggleLike() {
  if (!isAuthenticated) {
    requireAuth()
    return
  }
  if (busy) return

  busy = true
  try {
    const response = isLiked ? await ArticleAPI.unlike(articleId) : await ArticleAPI.like(articleId)
    if (response.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(response.message || m['blog:actions.failed']())
      return
    }
    isLiked = !isLiked
  } finally {
    busy = false
  }
}

// 切换收藏状态。
async function toggleBookmark() {
  if (!isAuthenticated) {
    requireAuth()
    return
  }
  if (busy) return

  busy = true
  try {
    const response = isBookmarked
      ? await ArticleAPI.unbookmark(articleId)
      : await ArticleAPI.bookmark(articleId)
    if (response.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(response.message || m['blog:actions.failed']())
      return
    }
    isBookmarked = !isBookmarked
  } finally {
    busy = false
  }
}
</script>

<!-- 互动栏：点赞与收藏并排，激活态以朱红标记 -->
<div class="mt-16 flex items-center justify-center gap-4 border-y border-line py-8">
  <button
    type="button"
    onclick={toggleLike}
    disabled={busy}
    class="inline-flex items-center gap-2 border px-6 py-3 text-sm font-bold transition-[border-color,color,transform,background-color] duration-200 ease-(--ease-out-strong) active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60 {isLiked
      ? 'border-signal bg-signal/10 text-signal'
      : 'border-border text-foreground hover:border-signal hover:text-signal'}"
  >
    <Heart class="h-4 w-4" aria-hidden="true" />
    {isLiked ? m['blog:actions.liked']() : m['blog:actions.like']()}
  </button>
  <button
    type="button"
    onclick={toggleBookmark}
    disabled={busy}
    class="inline-flex items-center gap-2 border px-6 py-3 text-sm font-bold transition-[border-color,color,transform,background-color] duration-200 ease-(--ease-out-strong) active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60 {isBookmarked
      ? 'border-signal bg-signal/10 text-signal'
      : 'border-border text-foreground hover:border-signal hover:text-signal'}"
  >
    <Bookmark class="h-4 w-4" aria-hidden="true" />
    {isBookmarked ? m['blog:actions.bookmarked']() : m['blog:actions.bookmark']()}
  </button>
</div>
