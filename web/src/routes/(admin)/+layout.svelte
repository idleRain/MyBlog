<script lang="ts">
import { onMount } from 'svelte'
import { authStore } from '$lib/stores/auth'
import { Sidebar } from '$ui'
import { AppSidebar } from '$lib/components/admin'
import { Toaster } from '$ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import { requireAuth, checkAuthOnLoad } from '$lib/utils/auth-guard'

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
  <div class="flex h-screen items-center justify-center">
    <div
      class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
    ></div>
  </div>
{:else if isAuthorized}
  <Sidebar.Provider>
    <AppSidebar />
    <Sidebar.Inset>
      {@render children()}
    </Sidebar.Inset>
  </Sidebar.Provider>
{:else}
  <div class="flex h-screen items-center justify-center">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-muted-foreground">请先登录</h1>
      <p class="mt-2 text-sm text-muted-foreground">需要登录后才能访问管理后台</p>
    </div>
  </div>
{/if}
