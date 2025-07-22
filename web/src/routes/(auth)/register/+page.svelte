<script lang="ts">
import { superForm } from 'sveltekit-superforms'
import { zodClient } from 'sveltekit-superforms/adapters'
import { z } from 'zod'
import type { PageData } from './$types'
import { ThemeToggle } from '$lib/components'
import { Toaster } from '$lib/components/ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import { Form, Card } from '$ui'
import { Input } from '$ui/input'
import { Eye, EyeOff, Lock, Mail, User, UserPlus } from '@lucide/svelte'
import { authStore } from '$lib/stores/auth.ts'
import { UserAPI } from '$lib/api'
import '../../../app.css'

export let data: PageData

const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, '用户名至少需要3个字符')
      .max(20, '用户名不能超过20个字符')
      .regex(/^[a-zA-Z0-9_]+$/, '用户名只能包含字母、数字和下划线'),
    email: z.string().email('请输入有效的邮箱地址'),
    password: z.string().min(6, '密码至少需要6个字符').max(50, '密码不能超过50个字符'),
    confirmPassword: z.string()
  })
  .refine(data => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword']
  })

const form = superForm(data.form, {
  validators: zodClient(registerSchema),
  onUpdated: async ({ form }) => {
    if (form.valid) {
      // 调用 API 注册
      try {
        const registerData = {
          username: form.data.username.trim(),
          email: form.data.email.trim(),
          password: form.data.password
        }

        const response = await UserAPI.register(registerData)

        if (response.code === 200) {
          toast.success('注册成功，请登录')
          goto('/login')
        } else {
          toast.error(response.message || '注册失败')
        }
      } catch (error) {
        console.error('Register error:', error)
        toast.error('网络错误，请稍后重试')
      }
    }
  }
})

const { form: formData, enhance, submitting } = form

let showPassword = false
let showConfirmPassword = false

// 如果已登录，重定向到首页
authStore.subscribe(state => {
  if (state.isAuthenticated) {
    goto('/')
  }
})

function togglePasswordVisibility(field: 'password' | 'confirmPassword') {
  if (field === 'password') {
    showPassword = !showPassword
  } else {
    showConfirmPassword = !showConfirmPassword
  }
}
</script>

<svelte:head>
  <title>注册 - MyBlog</title>
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
      <h1 class="text-2xl font-semibold tracking-tight">创建账户</h1>
      <p class="text-sm text-muted-foreground">快速注册，开始您的博客之旅</p>
    </div>

    <!-- Register Form. -->
    <Card.Root>
      <Card.Content class="pt-6">
        <form method="POST" use:enhance class="space-y-4">
          <!-- Username Field -->
          <Form.Field {form} name="username">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label class="flex items-center gap-2">
                  <User class="h-4 w-4" />
                  用户名 *
                </Form.Label>
                <Input
                  {...props}
                  type="text"
                  bind:value={$formData.username}
                  disabled={$submitting}
                  placeholder="请输入用户名"
                  autocomplete="username"
                />
              {/snippet}
            </Form.Control>
            <Form.FieldErrors class="text-xs" />
          </Form.Field>

          <!-- Email Field -->
          <Form.Field {form} name="email">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label class="flex items-center gap-2">
                  <Mail class="h-4 w-4" />
                  邮箱地址 *
                </Form.Label>
                <Input
                  {...props}
                  type="email"
                  bind:value={$formData.email}
                  disabled={$submitting}
                  placeholder="请输入邮箱地址"
                  autocomplete="email"
                />
              {/snippet}
            </Form.Control>
            <Form.FieldErrors class="text-xs" />
          </Form.Field>

          <!-- Password Field -->
          <Form.Field {form} name="password">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label class="flex items-center gap-2">
                  <Lock class="h-4 w-4" />
                  密码 *
                </Form.Label>
                <div class="relative">
                  <Input
                    {...props}
                    type={showPassword ? 'text' : 'password'}
                    bind:value={$formData.password}
                    disabled={$submitting}
                    placeholder="请输入密码（至少6位）"
                    autocomplete="new-password"
                    class="pr-10"
                  />
                  <button
                    type="button"
                    on:click={() => togglePasswordVisibility('password')}
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

          <!-- Confirm Password Field -->
          <Form.Field {form} name="confirmPassword">
            <Form.Control>
              {#snippet children({ props })}
                <Form.Label class="flex items-center gap-2">
                  <Lock class="h-4 w-4" />
                  确认密码 *
                </Form.Label>
                <div class="relative">
                  <Input
                    {...props}
                    type={showConfirmPassword ? 'text' : 'password'}
                    bind:value={$formData.confirmPassword}
                    disabled={$submitting}
                    placeholder="请再次输入密码"
                    autocomplete="new-password"
                    class="pr-10"
                  />
                  <button
                    type="button"
                    on:click={() => togglePasswordVisibility('confirmPassword')}
                    class="absolute top-1/2 right-3 h-4 w-4 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    disabled={$submitting}
                  >
                    {#if showConfirmPassword}
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

          <!-- Register Button -->
          <Form.Button disabled={$submitting} class="w-full">
            {#if $submitting}
              <div
                class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
              ></div>
              注册中...
            {:else}
              <UserPlus class="mr-2 h-4 w-4" />
              注册账户
            {/if}
          </Form.Button>
        </form>
      </Card.Content>
    </Card.Root>

    <!-- Login Link -->
    <p class="px-8 text-center text-sm text-muted-foreground">
      已有账户？
      <a href="/login" class="underline underline-offset-4 hover:text-primary"> 立即登录 </a>
    </p>
  </div>
</div>
