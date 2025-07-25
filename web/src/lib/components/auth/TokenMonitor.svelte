<script lang="ts">
import { onMount, onDestroy } from 'svelte'
import { authStore } from '$lib/stores/auth'
import { getAuthStatus, manualRefreshToken } from '$lib/utils/jwt'
import { toast } from 'svelte-sonner'

export let showStatus = false // 是否显示状态信息
export let autoRefresh = true // 是否自动刷新令牌

let interval: NodeJS.Timeout | null = null
let authStatus = {
  isAuthenticated: false,
  tokenValid: false,
  needsRefresh: false,
  expiresAt: new Date(),
  user: null
}

function updateAuthStatus() {
  authStatus = getAuthStatus()
}

async function handleAutoRefresh() {
  if (!autoRefresh || !authStatus.isAuthenticated) return

  if (authStatus.needsRefresh && authStatus.tokenValid) {
    try {
      const success = await manualRefreshToken()
      if (success && showStatus) {
        toast.success('令牌已自动刷新')
      }
    } catch (error) {
      console.error('自动刷新令牌失败:', error)
      if (showStatus) {
        toast.error('令牌自动刷新失败')
      }
    }
  }
}

onMount(() => {
  updateAuthStatus()

  // 每30秒检查一次令牌状态
  interval = setInterval(() => {
    updateAuthStatus()
    handleAutoRefresh()
  }, 30 * 1000)
})

onDestroy(() => {
  if (interval) {
    clearInterval(interval)
  }
})
</script>

{#if showStatus && authStatus.isAuthenticated}
  <div
    class="fixed right-4 bottom-4 z-50 rounded-lg border bg-background/80 p-3 shadow-md backdrop-blur-sm"
  >
    <div class="space-y-1 text-xs">
      <div class="font-medium">令牌状态</div>
      <div class="flex items-center gap-2">
        <div
          class="h-2 w-2 rounded-full {authStatus.tokenValid ? 'bg-green-500' : 'bg-red-500'}"
        ></div>
        <span>{authStatus.tokenValid ? '有效' : '已过期'}</span>
      </div>
      {#if authStatus.tokenValid}
        <div class="text-muted-foreground">
          过期时间: {authStatus.expiresAt.toLocaleTimeString()}
        </div>
      {/if}
      {#if authStatus.needsRefresh}
        <div class="text-orange-500">需要刷新</div>
      {/if}
    </div>
  </div>
{/if}

<!-- 这个组件主要用于监控，不渲染可见内容 -->
<style>
/* 隐藏组件，仅用于逻辑 */
</style>
