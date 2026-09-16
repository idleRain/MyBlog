<script lang="ts">
import type { ProfileUser } from '@myblog/api/modules/user/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { SITE_NAME_ZH } from '@myblog/shared'
import { authStore } from '$lib/stores/auth'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { UserAPI } from '$lib/api'

// 资料字段长度上限，与后端 UpdateProfileRequest 的 binding 规则一致。
const NICKNAME_MAX_LENGTH = 50
const AVATAR_MAX_LENGTH = 255
const BIO_MAX_LENGTH = 500
const WEBSITE_MAX_LENGTH = 255

// 密码长度下限与上限，与后端 ChangePasswordRequest 的 binding 规则一致。
const PASSWORD_MIN_LENGTH = 8
const PASSWORD_MAX_LENGTH = 64

// 资料表单状态。
let nickname = $state('')
let avatar = $state('')
let bio = $state('')
let website = $state('')
let savingProfile = $state(false)

// 改密码表单状态。
let oldPassword = $state('')
let newPassword = $state('')
let confirmPassword = $state('')
let savingPassword = $state(false)

// 资料加载状态。
let profileLoading = $state(true)
let loadedProfile: ProfileUser | undefined = $state()

// 收藏列表同款登录闸门：令牌存于 localStorage，SSR load 拿不到身份，按客户端补拉实现。
let isAuthenticated = $state(false)
authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
})

// 登录态变化时拉取一次资料，登出时重置页面状态。
$effect(() => {
  if (!isAuthenticated) {
    loadedProfile = undefined
    profileLoading = false
    void goto('/login')
    return
  }
  if (loadedProfile) return
  void loadProfile()
})

// 拉取当前资料并填充表单，字段以响应数据为准。
async function loadProfile() {
  profileLoading = true
  try {
    const response = await UserAPI.getProfile()
    if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
      toast.error(response.message || '资料加载失败，请稍后重试')
      return
    }
    applyProfile(response.data)
  } catch {
    toast.error('资料加载失败，请稍后重试')
  } finally {
    profileLoading = false
  }
}

// 将响应资料写入页面状态并记为已加载。
function applyProfile(profile: ProfileUser) {
  loadedProfile = profile
  nickname = profile.nickname
  avatar = profile.avatar
  bio = profile.bio
  website = profile.website
}

// 提交资料修改，成功后同步认证 store 中的用户信息。
async function handleProfileSubmit(event: SubmitEvent) {
  event.preventDefault()
  if (savingProfile) return

  savingProfile = true
  try {
    const response = await UserAPI.updateProfile({
      nickname: nickname.trim(),
      avatar: avatar.trim(),
      bio: bio.trim(),
      website: website.trim()
    })

    if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
      toast.error(response.message || '资料保存失败，请稍后重试')
      return
    }

    applyProfile(response.data)
    authStore.updateUser(response.data)
    toast.success('资料已保存')
  } catch {
    toast.error('资料保存失败，请稍后重试')
  } finally {
    savingProfile = false
  }
}

// 提交改密码，两次输入一致才发起请求。
// 成功后服务端撤销该用户全部既有令牌，本地会话随之失效，引导重新登录。
async function handlePasswordSubmit(event: SubmitEvent) {
  event.preventDefault()
  if (savingPassword) return

  if (newPassword !== confirmPassword) {
    toast.error('两次输入的新密码不一致')
    return
  }

  savingPassword = true
  try {
    const response = await UserAPI.changePassword({ oldPassword, newPassword })
    if (response.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(response.message || '密码修改失败，请稍后重试')
      return
    }

    toast.success('密码修改成功，请使用新密码重新登录')
    await authStore.logout(true)
    void goto('/login')
  } catch {
    toast.error('密码修改失败，请稍后重试')
  } finally {
    savingPassword = false
  }
}
</script>

<svelte:head>
  <title>个人资料 - {SITE_NAME_ZH}</title>
  <meta
    name="description"
    content="当前登录用户的资料维护页，支持修改昵称、头像、简介与登录密码。"
  />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 版面标题栏：与收藏页同构。 -->
    <div class="mb-8 border-b border-line pb-4">
      <p
        class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
      >
        资料 · Profile
        <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
      </p>
      <h1 class="font-display text-3xl font-black">个人资料</h1>
    </div>

    {#if profileLoading}
      <p class="py-10 text-center text-sm text-muted-foreground">正在加载资料…</p>
    {:else if loadedProfile}
      <p class="mb-10 font-mono text-sm text-muted-foreground">
        {loadedProfile.username} · {loadedProfile.email}
      </p>

      <div class="grid gap-8 lg:grid-cols-2">
        <!-- 资料表单：纸深面卡片承载，沿用站内输入框样式。 -->
        <form onsubmit={handleProfileSubmit} class="border border-line bg-secondary p-6">
          <h2 class="font-display text-xl font-black">基础资料</h2>
          <p class="mt-1 text-xs text-muted-foreground">昵称与简介会展示在评论与作者主页。</p>

          <label class="mt-6 block">
            <span class="mb-2 block text-sm font-bold">昵称</span>
            <input
              type="text"
              bind:value={nickname}
              required
              maxlength={NICKNAME_MAX_LENGTH}
              placeholder="展示昵称"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <label class="mt-4 block">
            <span class="mb-2 block text-sm font-bold">头像地址</span>
            <input
              type="url"
              bind:value={avatar}
              maxlength={AVATAR_MAX_LENGTH}
              placeholder="https://…"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <label class="mt-4 block">
            <span class="mb-2 block text-sm font-bold">个人简介</span>
            <textarea
              bind:value={bio}
              maxlength={BIO_MAX_LENGTH}
              rows="4"
              placeholder="一句话介绍自己"
              class="w-full resize-y border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            ></textarea>
          </label>

          <label class="mt-4 block">
            <span class="mb-2 block text-sm font-bold">个人网站</span>
            <input
              type="url"
              bind:value={website}
              maxlength={WEBSITE_MAX_LENGTH}
              placeholder="https://…"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <button
            type="submit"
            disabled={savingProfile}
            class="mt-6 inline-flex w-full items-center justify-center bg-primary px-6 py-3 text-sm font-bold text-primary-foreground transition-[background-color,color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60"
          >
            {savingProfile ? '保存中…' : '保存资料'}
          </button>
        </form>

        <!-- 改密码表单：独立于资料表单单独提交。 -->
        <form onsubmit={handlePasswordSubmit} class="h-fit border border-line bg-secondary p-6">
          <h2 class="font-display text-xl font-black">修改密码</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            长度 {PASSWORD_MIN_LENGTH}-{PASSWORD_MAX_LENGTH} 位，需同时包含字母和数字。
          </p>

          <label class="mt-6 block">
            <span class="mb-2 block text-sm font-bold">当前密码</span>
            <input
              type="password"
              bind:value={oldPassword}
              required
              autocomplete="current-password"
              placeholder="请输入当前密码"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <label class="mt-4 block">
            <span class="mb-2 block text-sm font-bold">新密码</span>
            <input
              type="password"
              bind:value={newPassword}
              required
              minlength={PASSWORD_MIN_LENGTH}
              maxlength={PASSWORD_MAX_LENGTH}
              autocomplete="new-password"
              placeholder="请输入新密码"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <label class="mt-4 block">
            <span class="mb-2 block text-sm font-bold">确认新密码</span>
            <input
              type="password"
              bind:value={confirmPassword}
              required
              autocomplete="new-password"
              placeholder="请再次输入新密码"
              class="w-full border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
          </label>

          <button
            type="submit"
            disabled={savingPassword}
            class="mt-6 inline-flex w-full items-center justify-center bg-primary px-6 py-3 text-sm font-bold text-primary-foreground transition-[background-color,color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60"
          >
            {savingPassword ? '提交中…' : '修改密码'}
          </button>
        </form>
      </div>
    {/if}
  </div>
</section>
