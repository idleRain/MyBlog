<script lang="ts">
import type { FriendlyLink } from '@myblog/api/modules/friendlyLink/types'
import { Button, Dialog, Input, Label, Separator, Switch } from '$ui'

interface Props {
  open: boolean
  // 编辑目标友链，为空表示新建。
  target?: FriendlyLink | null
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let { open, target = null, isSubmitting, onOpenChange, onConfirm }: Props = $props()

const isEditMode = $derived(target !== null)

let name = $state('')
let url = $state('')
let logo = $state('')
let description = $state('')
let contactEmail = $state('')
let sortOrder = $state(0)
let isReciprocal = $state(false)
let formError = $state('')

/**
 * 弹窗打开时重置表单并回填编辑数据。
 */
$effect(() => {
  if (!open) return
  name = target?.name ?? ''
  url = target?.url ?? ''
  logo = target?.logo ?? ''
  description = target?.description ?? ''
  contactEmail = target?.contactEmail ?? ''
  sortOrder = target?.sortOrder ?? 0
  isReciprocal = target?.isReciprocal ?? false
  formError = ''
})

/**
 * 校验并提交表单，名称与链接必填。
 */
function handleSubmit() {
  if (!name.trim()) {
    formError = '请填写站点名称'
    return
  }
  if (!url.trim()) {
    formError = '请填写站点链接'
    return
  }

  const payload: Record<string, unknown> = {
    name: name.trim(),
    url: url.trim(),
    sortOrder,
    isReciprocal
  }
  if (logo.trim()) payload.logo = logo.trim()
  if (description.trim()) payload.description = description.trim()
  if (contactEmail.trim()) payload.contactEmail = contactEmail.trim()

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑友链' : '添加友链'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改友情链接信息' : '登记新的友情链接，默认待审核'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="link-name">站点名称 *</Label.Root>
        <Input.Root
          id="link-name"
          bind:value={name}
          maxlength={50}
          placeholder="请输入站点名称"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="link-url">站点链接 *</Label.Root>
        <Input.Root
          id="link-url"
          bind:value={url}
          maxlength={255}
          placeholder="https://example.com"
          disabled={isSubmitting}
        />
        {#if formError}
          <p class="text-sm text-destructive">{formError}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <Label.Root for="link-logo">站点图标 URL</Label.Root>
        <Input.Root
          id="link-logo"
          bind:value={logo}
          maxlength={500}
          placeholder="站点 logo 或头像地址"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="link-description">站点简介</Label.Root>
        <Input.Root
          id="link-description"
          bind:value={description}
          maxlength={255}
          placeholder="一句话介绍对方站点"
          disabled={isSubmitting}
        />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="link-email">联系邮箱</Label.Root>
          <Input.Root
            id="link-email"
            type="email"
            bind:value={contactEmail}
            maxlength={100}
            placeholder="站长联系邮箱"
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="link-sort">排序权重</Label.Root>
          <Input.Root id="link-sort" type="number" bind:value={sortOrder} disabled={isSubmitting} />
        </div>
      </div>

      <Separator.Root />

      <div class="flex items-center justify-between gap-2">
        <div class="space-y-1">
          <Label.Root>已确认回链</Label.Root>
          <p class="text-xs text-muted-foreground">对方站点已放置本博客链接</p>
        </div>
        <Switch.Switch bind:checked={isReciprocal} disabled={isSubmitting} />
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
        {isEditMode ? '保存修改' : '添加友链'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
