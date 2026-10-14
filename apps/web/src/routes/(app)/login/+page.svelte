<script lang="ts">
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { Eye, EyeOff, LogIn } from '@lucide/svelte'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { UserAPI } from '$lib/api'
import { m } from '$i18n'

// 表单状态。
let username = $state('')
let password = $state('')
let showPassword = $state(false)
let submitting = $state(false)

// 标题中的站名取自词表，以随语言切换。
const pageTitle = m['account:login.title']({ site: m['ui:site.name']() })

// 认证状态经 store 自动订阅读取，组件卸载时由 Svelte 自动退订，避免重挂载累积订阅。
const isAuthenticated = $derived($authStore.isAuthenticated)

// 已登录用户访问登录页时自动回首页，避免重复登录。
$effect(() => {
  if (browser && isAuthenticated) {
    void goto('/')
  }
})

// 登录成功后经 authStore 持久化令牌与权限，再跳回首页。
async function handleLogin(event: SubmitEvent) {
  event.preventDefault()
  if (submitting) return

  if (!username.trim() || !password) {
    toast.error(m['account:login.validationFailed']())
    return
  }

  submitting = true
  try {
    const response = await UserAPI.login({
      username: username.trim(),
      password
    })

    if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
      toast.error(response.message || m['account:login.failed']())
      return
    }

    // 凭登录响应中的刷新令牌建立会话，服务端将令牌对写入 HttpOnly Cookie。
    const session = await UserAPI.createSession({ refreshToken: response.data.refreshToken })
    if (session.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(session.message || m['account:login.sessionFailed']())
      return
    }

    authStore.login(response.data.user, response.data.permissions)
    toast.success(m['account:login.success']())
    await goto('/')
  } catch {
    toast.error(m['account:login.networkFailed']())
  } finally {
    submitting = false
  }
}
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['account:login.description']()} />
</svelte:head>

<section class="texture-grid relative flex min-h-screen items-center overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <div class="mx-auto max-w-sm">
      <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
      <p
        class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
      >
        {m['account:login.eyebrow']()}
        <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
      </p>

      <h1 class="font-display text-4xl leading-[1.15] font-black tracking-tight">
        {m['account:login.heading']()}
      </h1>
      <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
        {m['account:login.description']()}
      </p>

      <!-- 登录表单：纸深面卡片承载，沿用站内输入框样式。 -->
      <form onsubmit={handleLogin} class="mt-10 border border-line bg-secondary p-6">
        <label class="block">
          <span class="mb-2 block text-sm font-bold">{m['account:login.usernameLabel']()}</span>
          <input
            type="text"
            bind:value={username}
            required
            autocomplete="username"
            placeholder={m['account:login.usernamePlaceholder']()}
            class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>

        <label class="mt-4 block">
          <span class="mb-2 block text-sm font-bold">{m['account:login.passwordLabel']()}</span>
          <span class="relative block">
            <input
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              required
              autocomplete="current-password"
              placeholder={m['account:login.passwordPlaceholder']()}
              class="w-full border border-line bg-background px-3 py-2.5 pr-10 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
            <button
              type="button"
              onclick={() => (showPassword = !showPassword)}
              class="absolute top-1/2 right-0.5 flex h-9 w-9 -translate-y-1/2 items-center justify-center text-muted-foreground transition-colors duration-150 hover:text-foreground"
              aria-label={showPassword
                ? m['account:login.hidePassword']()
                : m['account:login.showPassword']()}
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
          {submitting ? m['account:login.submitting']() : m['account:login.submit']()}
        </button>
      </form>

      <p class="mt-6 font-mono text-xs text-muted-foreground">
        {m['account:login.guestNote']()}
      </p>
    </div>
  </div>
</section>
