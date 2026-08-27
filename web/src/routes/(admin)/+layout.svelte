<script lang="ts">
import { onMount } from 'svelte'
import { Sidebar } from '$ui'
import { AppSidebar } from '$lib/components/admin'
import { Toaster } from '$ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import { requireAuth, checkAuthOnLoad } from '$lib/utils/auth-guard'
import { Button } from '$ui/button'
import { Spinner } from '$ui/spinner'

interface Props {
  children: import('svelte').Snippet
}

let { children }: Props = $props()

let isAuthorized = $state(false)
let isLoading = $state(true)

onMount(async () => {
  try {
    // 检查认证状态，包括 token 有效性
    const authResult = await checkAuthOnLoad()

    if (authResult.needsRedirect) {
      await goto(authResult.redirectTo || '/login')
      return
    }

    isAuthorized = authResult.isAuthenticated
    isLoading = false

    // 如果未认证，要求登录
    if (!isAuthorized) {
      await requireAuth()
    }
  } catch (error) {
    console.error('认证检查失败:', error)
    isLoading = false
    await goto('/login')
  }
})
</script>

<svelte:head>
  <title>管理后台 - MyBlog</title>
</svelte:head>

<ModeWatcher />
<Toaster position="top-right" />

{#if isLoading}
  <!-- 加载态：网格底纹加等宽标注，与登录页同一规格书语言。 -->
  <div class="relative flex h-screen items-center justify-center overflow-hidden bg-background">
    <div class="admin-grid absolute inset-0" aria-hidden="true"></div>
    <div class="relative z-10 flex flex-col items-center gap-4">
      <Spinner class="size-6 text-signal" />
      <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
        <span class="text-signal">//</span> LOADING CONSOLE
      </p>
    </div>
  </div>
{:else if isAuthorized}
  <Sidebar.Provider>
    <AppSidebar />
    <Sidebar.Inset>
      {@render children()}
    </Sidebar.Inset>
  </Sidebar.Provider>
{:else}
  <!-- 未授权态：等待认证守卫完成跳转期间的占位提示。 -->
  <div class="relative flex h-screen items-center justify-center overflow-hidden bg-background">
    <div class="admin-grid absolute inset-0" aria-hidden="true"></div>
    <div class="relative z-10 mx-auto flex max-w-sm flex-col items-center gap-3 px-4 text-center">
      <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
        <span class="text-signal">//</span> SESSION REQUIRED
      </p>
      <h1 class="text-2xl font-bold tracking-tight">请先登录</h1>
      <p class="text-sm text-muted-foreground">需要登录后才能访问管理后台。</p>
      <Button
        class="rounded-none bg-signal text-signal-foreground hover:bg-signal/90"
        onclick={() => void goto('/login')}
      >
        前往登录
      </Button>
    </div>
  </div>
{/if}
