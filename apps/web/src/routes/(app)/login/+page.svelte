<script lang="ts">
import { Eye, EyeOff, LogIn } from '@lucide/svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { UserAPI } from '$lib/api'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 表单状态。
let username = $state('')
let password = $state('')
let showPassword = $state(false)
let submitting = $state(false)

// 已登录用户访问登录页时自动回首页，避免重复登录。
authStore.subscribe(state => {
  if (browser && state.isAuthenticated) {
    void goto('/')
  }
})

// 登录成功后经 authStore 持久化令牌与权限，再跳回首页。
async function handleLogin(event: SubmitEvent) {
  event.preventDefault()
  if (submitting) return

  if (!username.trim() || !password) {
    toast.error('请输入用户名和密码')
    return
  }

  submitting = true
  try {
    const response = await UserAPI.login({
      username: username.trim(),
      password
    })

    if (response.code !== SUCCESS_CODE || !response.data) {
      toast.error(response.message || '登录失败，请检查用户名和密码')
      return
    }

    authStore.login(
      response.data.user,
      response.data.accessToken,
      response.data.refreshToken,
      response.data.expiresIn,
      response.data.permissions
    )
    toast.success('登录成功')
    await goto('/')
  } catch {
    toast.error('网络错误，请稍后重试')
  } finally {
    submitting = false
  }
}
</script>

<svelte:head>
  <title>登录 - {SITE_NAME_ZH}</title>
  <meta name="description" content="登录后即可参与评论、点赞、收藏与关注等互动。" />
</svelte:head>

<section class="texture-grid relative flex min-h-screen items-center overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <div class="mx-auto max-w-sm">
      <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
      <p
        class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
      >
        门户 · Auth
        <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
      </p>

      <h1 class="font-display text-4xl leading-[1.15] font-black tracking-tight">欢迎回来</h1>
      <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
        登录后即可参与评论、点赞、收藏与关注等互动。
      </p>

      <!-- 登录表单：纸深面卡片承载，沿用站内输入框样式。 -->
      <form onsubmit={handleLogin} class="mt-10 border border-line bg-secondary p-6">
        <label class="block">
          <span class="mb-2 block text-sm font-bold">用户名</span>
          <input
            type="text"
            bind:value={username}
            required
            autocomplete="username"
            placeholder="请输入用户名"
            class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>

        <label class="mt-4 block">
          <span class="mb-2 block text-sm font-bold">密码</span>
          <span class="relative block">
            <input
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              required
              autocomplete="current-password"
              placeholder="请输入密码"
              class="w-full border border-line bg-background px-3 py-2.5 pr-10 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
            <button
              type="button"
              onclick={() => (showPassword = !showPassword)}
              class="absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground transition-colors duration-150 hover:text-foreground"
              aria-label={showPassword ? '隐藏密码' : '显示密码'}
            >
              {#if showPassword}
                <EyeOff class="h-4 w-4" aria-hidden="true" />
              {:else}
                <Eye class="h-4 w-4" aria-hidden="true" />
              {/if}
            </button>
          </span>
        </label>

        <button
          type="submit"
          disabled={submitting}
          class="mt-6 inline-flex w-full items-center justify-center gap-2 bg-primary px-6 py-3 text-sm font-bold text-primary-foreground transition-[background-color,color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60"
        >
          <LogIn class="h-4 w-4" aria-hidden="true" />
          {submitting ? '登录中…' : '登录'}
        </button>
      </form>

      <p class="mt-6 font-mono text-xs text-muted-foreground">
        尚无账户时可通过评论区以访客身份留言，或联系站长开通。
      </p>
    </div>
  </div>
</section>
