<script lang="ts">
import { Button, Dialog, Input, Label, Separator, Switch, Textarea } from '$ui'
import type { Tag, TagStatus } from '@myblog/api/modules/tag/types'

interface Props {
  open: boolean
  // 编辑目标标签，为空表示新建。
  target?: Tag | null
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let { open, target = null, isSubmitting, onOpenChange, onConfirm }: Props = $props()

const isEditMode = $derived(target !== null)

// 表单状态，弹窗打开时由目标标签回填。
let name = $state('')
let slug = $state('')
let color = $state('#808080')
let description = $state('')
let status = $state<TagStatus>(1)
let isHot = $state(false)
// 英文翻译字段，enSnapshot 记录回填值，编辑既有翻译时始终携带补丁以支持清空。
let enName = $state('')
let enDescription = $state('')
let enSnapshot = $state('')
let formError = $state('')

/**
 * 弹窗打开时重置表单并回填编辑数据。
 */
$effect(() => {
  if (!open) return
  name = target?.name ?? ''
  slug = target?.slug ?? ''
  color = target?.color || '#808080'
  description = target?.description ?? ''
  status = target?.status ?? 1
  isHot = target?.isHot ?? false

  // 英文翻译行允许按字段缺失，缺失字段回填为空串。
  const en = target?.translations?.find(row => row.locale === 'en') ?? null
  enName = en?.name ?? ''
  enDescription = en?.description ?? ''
  enSnapshot = JSON.stringify([enName, enDescription])
  formError = ''
})

/**
 * 组装英文翻译补丁，存在既有翻译或任一字段非空时携带。
 */
function buildI18nPayload(): Record<string, unknown> | null {
  const patch: Record<string, unknown> = {}
  if (enName.trim()) patch.name = enName.trim()
  if (enDescription.trim()) patch.description = enDescription.trim()
  const hasAnyValue = Object.keys(patch).length > 0
  const hadTranslation = isEditMode && enSnapshot !== JSON.stringify(['', ''])
  if (!hasAnyValue && !hadTranslation) return null
  return { en: patch }
}

/**
 * 校验并提交表单，标签名称必填。
 */
function handleSubmit() {
  if (!name.trim()) {
    formError = '请填写标签名称'
    return
  }

  const payload: Record<string, unknown> = {
    name: name.trim(),
    color,
    status,
    isHot
  }
  if (slug.trim()) payload.slug = slug.trim()
  if (description.trim()) payload.description = description.trim()

  // 英文翻译补丁存在时携带，由后端校验语言键并按语言存储。
  const i18n = buildI18nPayload()
  if (i18n) payload.i18n = i18n

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑标签' : '新建标签'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改标签名称与属性' : '创建新的文章标签'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="tag-name">
          名称
          <span class="text-destructive" aria-hidden="true">*</span>
          <span class="sr-only">必填</span>
        </Label.Root>
        <Input.Root
          id="tag-name"
          bind:value={name}
          maxlength={30}
          placeholder="请输入标签名称"
          disabled={isSubmitting}
        />
        {#if formError}
          <p class="text-sm text-destructive">{formError}</p>
        {/if}
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="tag-slug">URL 标识</Label.Root>
          <Input.Root
            id="tag-slug"
            bind:value={slug}
            maxlength={30}
            placeholder="留空时由名称自动生成"
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="tag-color">标签颜色</Label.Root>
          <div class="flex items-center gap-2">
            <input
              id="tag-color"
              type="color"
              bind:value={color}
              class="size-8 cursor-pointer rounded border border-input bg-transparent"
              disabled={isSubmitting}
            />
            <Input.Root bind:value={color} maxlength={7} disabled={isSubmitting} />
          </div>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-0.5">
            <Label.Root for="tag-status">启用状态</Label.Root>
            <p class="text-xs text-muted-foreground">隐藏后不再展示</p>
          </div>
          <Switch.Switch
            id="tag-status"
            checked={status === 1}
            onCheckedChange={(checked: boolean) => {
              status = checked ? 1 : 0
            }}
            disabled={isSubmitting}
          />
        </div>
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-0.5">
            <Label.Root for="tag-hot">热门标签</Label.Root>
            <p class="text-xs text-muted-foreground">在热门标签区优先展示</p>
          </div>
          <Switch.Switch id="tag-hot" bind:checked={isHot} disabled={isSubmitting} />
        </div>
      </div>

      <div class="space-y-2">
        <Label.Root for="tag-description">描述</Label.Root>
        <Textarea.Root
          id="tag-description"
          bind:value={description}
          maxlength={200}
          placeholder="标签简介"
          disabled={isSubmitting}
        />
      </div>

      <Separator.Root />

      <div class="space-y-2">
        <p class="text-sm font-medium">英文翻译（可选）</p>
        <p class="text-xs text-muted-foreground">未填写的字段回退展示中文</p>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="tag-en-name">英文名称</Label.Root>
          <Input.Root
            id="tag-en-name"
            bind:value={enName}
            maxlength={30}
            placeholder="English name"
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="tag-en-description">英文描述</Label.Root>
          <Input.Root
            id="tag-en-description"
            bind:value={enDescription}
            maxlength={200}
            placeholder="English description"
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
        {isEditMode ? '保存修改' : '创建标签'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
