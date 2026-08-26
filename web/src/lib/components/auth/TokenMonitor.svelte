<script lang="ts">
import { onMount, onDestroy } from 'svelte'
import { getAuthStatus, manualRefreshToken, type AuthStatus } from '$lib/utils/jwt'
import { toast } from 'svelte-sonner'

// 是否显示状态信息浮层。
export let showStatus = false
// 是否在令牌临近过期时自动刷新。
export let autoRefresh = true

// 令牌状态检查间隔为 30 秒，每次检查后同步尝试自动刷新。
const TOKEN_CHECK_INTERVAL_MS = 30 * 1000

let interval: NodeJS.Timeout | null = null
// 认证状态类型由 jwt.ts 的 AuthStatus 接口统一约束，令牌缺失时不含时间与用户字段。
let authStatus: AuthStatus = {
  isAuthenticated: false,
  tokenValid: false,
  needsRefresh: false
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

  interval = setInterval(() => {
    updateAuthStatus()
    handleAutoRefresh()
  }, TOKEN_CHECK_INTERVAL_MS)
})

onDestroy(() => {
  if (interval) {
    clearInterval(interval)
  }
})
</script>

{#if showStatus && authStatus.isAuthenticated}
  <!-- 状态浮层：直角实色卡片，状态点为功能性圆形。 -->
  <div
    class="fixed right-4 bottom-4 z-50 rounded-none border border-border bg-background p-3 shadow-md"
  >
    <div class="space-y-1 text-xs">
      <div class="font-medium">令牌状态</div>
      <div class="flex items-center gap-2">
        <div
          class="h-2 w-2 rounded-full {authStatus.tokenValid ? 'bg-signal' : 'bg-destructive'}"
        ></div>
        <span>{authStatus.tokenValid ? '有效' : '已过期'}</span>
      </div>
      {#if authStatus.tokenValid}
        <div class="text-muted-foreground">
          过期时间: {authStatus.expiresAt?.toLocaleTimeString()}
        </div>
      {/if}
      {#if authStatus.needsRefresh}
        <div class="text-signal">需要刷新</div>
      {/if}
    </div>
  </div>
{/if}
