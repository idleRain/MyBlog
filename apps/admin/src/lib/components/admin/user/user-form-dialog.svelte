<script lang="ts">
import type { User, UserRole } from '@myblog/api/modules/user/types'
import { Button, Dialog, Input, Label, Select } from '$ui'
import { USER_ROLE_LABELS } from '$lib/constants/user'

interface Props {
  open: boolean
  // 编辑目标用户，为空表示新建。
  target?: User | null
  // 当前用户可分配的角色列表。
  assignableRoles: UserRole[]
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let {
  open,
  target = null,
  assignableRoles,
  isSubmitting,
  onOpenChange,
  onConfirm
}: Props = $props()

const isEditMode = $derived(target !== null)

let username = $state('')
let email = $state('')
let password = $state('')
let nickname = $state('')
let role = $state<UserRole>('user')
let birthday = $state('')
let formErrors = $state<Record<string, string>>({})

/**
 * 弹窗打开时重置表单并回填编辑数据。
 */
$effect(() => {
  if (!open) return
  username = target?.username ?? ''
  email = target?.email ?? ''
  password = ''
  nickname = target?.nickname ?? ''
  role = target?.role ?? 'user'
  birthday = target?.birthday ?? ''
  formErrors = {}
})

/**
 * 校验表单必填项，返回错误信息映射。
 */
function validate(): Record<string, string> {
  const errors: Record<string, string> = {}
  if (!username.trim()) errors.username = '请输入用户名'
  if (!email.trim()) errors.email = '请输入邮箱'
  else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) errors.email = '邮箱格式不正确'
  if (!isEditMode && !password.trim()) errors.password = '请输入密码'
  return errors
}

/**
 * 校验并提交表单，编辑时密码留空不修改。
 */
function handleSubmit() {
  const errors = validate()
  formErrors = errors
  if (Object.keys(errors).length > 0) return

  const payload: Record<string, unknown> = {
    username: username.trim(),
    email: email.trim(),
    role
  }
  if (nickname.trim()) payload.nickname = nickname.trim()
  if (password.trim()) payload.password = password.trim()
  if (birthday) payload.birthday = birthday

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑用户' : '创建用户'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改用户信息，密码留空则不修改' : '填写信息创建新的系统用户'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="user-username">用户名 *</Label.Root>
        <Input.Root
          id="user-username"
          bind:value={username}
          maxlength={20}
          placeholder="请输入用户名"
          disabled={isSubmitting}
        />
        {#if formErrors.username}
          <p class="text-sm text-destructive">{formErrors.username}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <Label.Root for="user-email">邮箱 *</Label.Root>
        <Input.Root
          id="user-email"
          type="email"
          bind:value={email}
          placeholder="请输入邮箱地址"
          disabled={isSubmitting}
        />
        {#if formErrors.email}
          <p class="text-sm text-destructive">{formErrors.email}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <Label.Root for="user-password">
          {isEditMode ? '密码' : '密码 *'}
          {#if isEditMode}
            <span class="text-xs text-muted-foreground">（留空则不修改）</span>
          {/if}
        </Label.Root>
        <Input.Root
          id="user-password"
          type="password"
          bind:value={password}
          placeholder={isEditMode ? '留空则不修改密码' : '请输入密码'}
          disabled={isSubmitting}
        />
        {#if formErrors.password}
          <p class="text-sm text-destructive">{formErrors.password}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <Label.Root for="user-nickname">昵称</Label.Root>
        <Input.Root
          id="user-nickname"
          bind:value={nickname}
          placeholder="请输入昵称"
          disabled={isSubmitting}
        />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root>角色</Label.Root>
          <Select.Root type="single" bind:value={role} disabled={isSubmitting}>
            <Select.Trigger class="w-full">{USER_ROLE_LABELS[role]}</Select.Trigger>
            <Select.Content>
              <Select.Group>
                {#each assignableRoles as assignableRole (assignableRole)}
                  <Select.Item value={assignableRole}
                    >{USER_ROLE_LABELS[assignableRole]}</Select.Item
                  >
                {/each}
              </Select.Group>
            </Select.Content>
          </Select.Root>
        </div>

        <div class="space-y-2">
          <Label.Root for="user-birthday">生日</Label.Root>
          <Input.Root
            id="user-birthday"
            type="date"
            bind:value={birthday}
            disabled={isSubmitting}
          />
        </div>
      </div>
    </div>

    <Dialog.Footer>
      <Button variant="outline" onclick={() => onOpenChange(false)} disabled={isSubmitting}>
        取消
      </Button>
      <Button onclick={handleSubmit} disabled={isSubmitting}>
        {#if isSubmitting}
          <span
            class="mr-2 inline-block size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"
          ></span>
        {/if}
        {isEditMode ? '保存修改' : '创建用户'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
