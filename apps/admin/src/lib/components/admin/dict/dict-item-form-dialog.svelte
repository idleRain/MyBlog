<script lang="ts">
import type { DictItem, DictStatus } from '@myblog/api/modules/dict/types'
import { Button, Dialog, Input, Label, Separator, Switch } from '$ui'

interface Props {
  open: boolean
  // 编辑目标字典项，为空表示新建。
  target?: DictItem | null
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let { open, target = null, isSubmitting, onOpenChange, onConfirm }: Props = $props()

const isEditMode = $derived(target !== null)

// 表单状态，弹窗打开时由目标字典项回填；字典项值创建后不可变更。
let value = $state('')
let label = $state('')
let description = $state('')
let status = $state<DictStatus>(1)
let sortOrder = $state(0)
// 英文翻译字段，enSnapshot 记录回填值，编辑既有翻译时始终携带补丁以支持清空。
let enLabel = $state('')
let enDescription = $state('')
let enSnapshot = $state('')
let formError = $state('')

$effect(() => {
  if (!open) return
  value = target?.value ?? ''
  label = target?.label ?? ''
  description = target?.description ?? ''
  status = (target?.status ?? 1) as DictStatus
  sortOrder = target?.sortOrder ?? 0

  // 英文翻译行允许按字段缺失，缺失字段回填为空串。
  const en = target?.translations?.find(row => row.locale === 'en') ?? null
  enLabel = en?.label ?? ''
  enDescription = en?.description ?? ''
  enSnapshot = JSON.stringify([enLabel, enDescription])
  formError = ''
})

// 组装英文翻译补丁，存在既有翻译或任一字段非空时携带。
function buildI18nPayload(): Record<string, unknown> | null {
  const patch: Record<string, unknown> = {}
  if (enLabel.trim()) patch.label = enLabel.trim()
  if (enDescription.trim()) patch.description = enDescription.trim()
  const hasAnyValue = Object.keys(patch).length > 0
  const hadTranslation = isEditMode && enSnapshot !== JSON.stringify(['', ''])
  if (!hasAnyValue && !hadTranslation) return null
  return { en: patch }
}

// 校验并提交表单，显示名必填，字典项值仅新建时填写。
function handleSubmit() {
  if (!label.trim()) {
    formError = '请填写字典项显示名'
    return
  }
  if (!isEditMode && !value.trim()) {
    formError = '请填写字典项值'
    return
  }

  const payload: Record<string, unknown> = {
    label: label.trim(),
    status,
    sortOrder
  }
  if (description.trim()) payload.description = description.trim()
  if (!isEditMode) payload.value = value.trim()

  // 英文翻译补丁存在时携带，由后端校验语言键并按语言存储。
  const i18n = buildI18nPayload()
  if (i18n) payload.i18n = i18n

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑字典项' : '新建字典项'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改字典项显示名与生效状态' : '为当前字典类型新增一个可选值'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="dict-item-value">字典项值 *</Label.Root>
        <Input.Root
          id="dict-item-value"
          bind:value
          maxlength={50}
          placeholder="业务存储值，如 1"
          disabled={isSubmitting || isEditMode}
        />
        <p class="text-xs text-muted-foreground">
          {isEditMode ? '字典项值创建后不可变更' : '业务侧按该值匹配字典项，创建后不可变更'}
        </p>
      </div>

      <div class="space-y-2">
        <Label.Root for="dict-item-label">显示名 *</Label.Root>
        <Input.Root
          id="dict-item-label"
          bind:value={label}
          maxlength={50}
          placeholder="请输入字典项显示名"
          disabled={isSubmitting}
        />
        {#if formError}
          <p class="text-sm text-destructive">{formError}</p>
        {/if}
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="dict-item-sort">排序权重</Label.Root>
          <Input.Root
            id="dict-item-sort"
            type="number"
            bind:value={sortOrder}
            min={0}
            max={9999}
            disabled={isSubmitting}
          />
        </div>
        <div class="flex items-center justify-between gap-2 pt-2">
          <div class="space-y-0.5">
            <Label.Root>是否生效</Label.Root>
            <p class="text-xs text-muted-foreground">停用后 app 端不再下发</p>
          </div>
          <Switch.Switch
            checked={status === 1}
            onCheckedChange={(checked: boolean) => {
              status = checked ? 1 : 0
            }}
            disabled={isSubmitting}
          />
        </div>
      </div>

      <Separator.Root />

      <div class="space-y-2">
        <Label.Root for="dict-item-description">描述</Label.Root>
        <Input.Root
          id="dict-item-description"
          bind:value={description}
          maxlength={200}
          placeholder="字典项用途说明"
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
          <Label.Root for="dict-item-en-label">英文显示名</Label.Root>
          <Input.Root
            id="dict-item-en-label"
            bind:value={enLabel}
            maxlength={50}
            placeholder="English label"
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="dict-item-en-description">英文描述</Label.Root>
          <Input.Root
            id="dict-item-en-description"
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
        {isEditMode ? '保存修改' : '创建字典项'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
