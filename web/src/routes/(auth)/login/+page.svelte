<script lang="ts">
import { superForm } from 'sveltekit-superforms'
import { zodClient } from 'sveltekit-superforms/adapters'
import { z } from 'zod'
import type { PageData } from './$types'
import { authStore } from '$lib/stores/auth.ts'
import { ThemeToggle } from '$lib/components'
import { Toaster } from '$lib/components/ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import { Input } from '$ui/input'
import { Form, Card } from '$ui'
import { EyeOff, Eye, LogIn } from '@lucide/svelte'
import { UserAPI } from '$lib/api'

export let data: PageData

const loginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码')
})

const form = superForm(data.form, {
  validators: zodClient(loginSchema),
  onUpdated: async ({ form }) => {
    if (form.valid) {
      // 调用 API 登录
      try {
        const response = await UserAPI.login({
          username: form.data.username.trim(),
          password: form.data.password.trim()
        })

        if (response.code === 200 && response.data) {
          authStore.login(response.data.user, response.data.token)
          toast.success('登录成功')
          await goto('/')
        } else {
          toast.error(response.message || '登录失败')
        }
      } catch (error) {
        console.error('Login error:', error)
        toast.error('网络错误，请稍后重试')
      }
    }
  }
})

const { form: formData, enhance, submitting } = form

let showPassword = false

// 如果已登录，重定向到首页
authStore.subscribe(state => {
  if (state.isAuthenticated) {
    goto('/')
  }
})

function togglePasswordVisibility() {
  showPassword = !showPassword
}
</script>

<svelte:head>
  <title>登录 - MyBlog</title>
</svelte:head>

<ModeWatcher />
<Toaster position="top-center" />

<div
  class="relative flex min-h-screen items-center justify-center overflow-hidden bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800"
>
  <!-- 背景装饰 -->
  <div class="absolute inset-0">
    <div
      class="animate-blob absolute top-20 left-20 h-72 w-72 rounded-full bg-blue-300 opacity-20 mix-blend-multiply blur-xl filter dark:bg-blue-500 dark:opacity-10 dark:mix-blend-lighten"
    ></div>
    <div
      class="animate-blob animation-delay-2000 absolute top-40 right-20 h-72 w-72 rounded-full bg-purple-300 opacity-20 mix-blend-multiply blur-xl filter dark:bg-purple-500 dark:opacity-10 dark:mix-blend-lighten"
    ></div>
    <div
      class="animate-blob animation-delay-4000 absolute -bottom-8 left-40 h-72 w-72 rounded-full bg-pink-300 opacity-20 mix-blend-multiply blur-xl filter dark:bg-pink-500 dark:opacity-10 dark:mix-blend-lighten"
    ></div>
  </div>

  <!-- 主题切换按钮 -->
  <ThemeToggle />

  <div class="relative z-10 mx-auto flex w-full max-w-sm flex-col justify-center space-y-6 px-4">
    <!-- Logo and Title -->
    <div class="flex flex-col space-y-2 text-center">
      <h1 class="text-2xl font-semibold tracking-tight">欢迎回来</h1>
      <p class="text-sm text-muted-foreground">请输入您的账户信息来登录</p>
    </div>

    <!-- Login Form -->
    <Card.Root>
      <Card.Content class="pt-6">
        <form method="POST" use:enhance class="space-y-4">
          <!-- Username Field -->
          <Form.Field {form} name="username">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label>用户名或邮箱</Form.Label>
                <Input
                  {...props}
                  type="text"
                  bind:value={$formData.username}
                  disabled={$submitting}
                  placeholder="请输入用户名或邮箱"
                  autocomplete="username"
                />
              {/snippet}
            </Form.Control>
            <Form.FieldErrors class="text-xs" />
          </Form.Field>

          <!-- Password Field -->
          <Form.Field {form} name="password">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label>密码</Form.Label>
                <div class="relative">
                  <Input
                    {...props}
                    type={showPassword ? 'text' : 'password'}
                    bind:value={$formData.password}
                    disabled={$submitting}
                    placeholder="请输入密码"
                    autocomplete="current-password"
                    class="pr-10"
                  />
                  <button
                    type="button"
                    onclick={togglePasswordVisibility}
                    class="absolute top-1/2 right-3 h-4 w-4 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    disabled={$submitting}
                  >
                    {#if showPassword}
                      <EyeOff class="h-4 w-4" />
                    {:else}
                      <Eye class="h-4 w-4" />
                    {/if}
                  </button>
                </div>
              {/snippet}
            </Form.Control>
            <Form.FieldErrors class="text-xs" />
          </Form.Field>

          <!-- Login Button -->
          <Form.Button disabled={$submitting} class="w-full">
            {#if $submitting}
              <div
                class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
              ></div>
              登录中..
            {:else}
              <LogIn class="mr-2 h-4 w-4" />
              登录
            {/if}
          </Form.Button>
        </form>
      </Card.Content>
    </Card.Root>

    <!-- Register Link -->
    <p class="px-8 text-center text-sm text-muted-foreground">
      还没有账户？
      <a href="/register" class="underline underline-offset-4 hover:text-primary"> 立即注册 </a>
    </p>
  </div>
</div>
