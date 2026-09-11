<script lang="ts">
import { authStore } from '$lib/stores/auth'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { FollowAPI } from '$lib/api'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

interface Props {
  /** 被关注的目标用户 ID。 */
  targetUserId: number
}

let { targetUserId }: Props = $props()

// 关注状态与提交锁，登录后才向服务端拉取真实状态。
let isFollowing = $state(false)
let busy = $state(false)
let loaded = $state(false)

let isAuthenticated = $state(false)
authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
  // 登录态就绪后补拉关注状态，登出时重置为未关注。
  if (isAuthenticated && !loaded) {
    void refreshFollowing()
  }
  if (!isAuthenticated) {
    isFollowing = false
    loaded = false
  }
})

// 拉取当前用户与目标用户的关注关系，失败时保持默认值。
async function refreshFollowing() {
  try {
    const response = await FollowAPI.isFollowing(targetUserId)
    if (response.code === SUCCESS_CODE) {
      isFollowing = response.data?.isFollowing ?? false
      loaded = true
    }
  } catch {
    // 状态拉取失败保持默认，交互时以服务端结果为准。
  }
}

// 未登录时引导前往登录页。
function requireAuth() {
  toast.info('登录后即可关注作者')
  void goto('/login')
}

// 切换关注状态。
async function toggleFollow() {
  if (!isAuthenticated) {
    requireAuth()
    return
  }
  if (busy) return

  busy = true
  try {
    const response = isFollowing
      ? await FollowAPI.unfollow({ followingId: targetUserId })
      : await FollowAPI.follow({ followingId: targetUserId })
    if (response.code !== SUCCESS_CODE) {
      toast.error(response.message || '操作失败，请稍后重试')
      return
    }
    isFollowing = !isFollowing
  } finally {
    busy = false
  }
}
</script>

<button
  type="button"
  onclick={toggleFollow}
  disabled={busy}
  class="inline-flex items-center gap-2 px-6 py-2.5 text-sm font-bold transition-[background-color,color,border-color,transform] duration-200 ease-(--ease-out-strong) active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60 {isFollowing
    ? 'border border-border text-foreground hover:border-signal hover:text-signal'
    : 'bg-primary text-primary-foreground hover:bg-signal hover:text-signal-foreground'}"
>
  {#if busy}
    处理中…
  {:else if isFollowing}
    已关注
  {:else}
    关注
  {/if}
</button>
