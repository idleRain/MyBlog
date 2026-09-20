<script lang="ts">
import type { DictStatus, DictType } from '@myblog/api/modules/dict/types'
import { Button, Dialog, Input, Label, Separator, Switch } from '$ui'

interface Props {
  open: boolean
  // 编辑目标字典类型，为空表示新建。
  target?: DictType | null
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let { open, target = null, isSubmitting, onOpenChange, onConfirm }: Props = $props()

const isEditMode = $derived(target !== null)

// 表单状态，弹窗打开时由目标类型回填；字典码创建后不可变更。
let code = $state('')
let name = $state('')
let description = $state('')
let status = $state<DictStatus>(1)
let sortOrder = $state(0)
// 英文翻译字段，enSnapshot 记录回填值，编辑既有翻译时始终携带补丁以支持清空。
let enName = $state('')
let enDescription = $state('')
let enSnapshot = $state('')
let formError = $state('')

$effect(() => {
  if (!open) return
  code = target?.code ?? ''
  name = target?.name ?? ''
  description = target?.description ?? ''
  status = (target?.status ?? 1) as DictStatus
  sortOrder = target?.sortOrder ?? 0

  // 英文翻译行允许按字段缺失，缺失字段回填为空串。
  const en = target?.translations?.find(row => row.locale === 'en') ?? null
  enName = en?.name ?? ''
  enDescription = en?.description ?? ''
  enSnapshot = JSON.stringify([enName, enDescription])
  formError = ''
})

// 组装英文翻译补丁，存在既有翻译或任一字段非空时携带。
function buildI18nPayload(): Record<string, unknown> | null {
  const patch: Record<string, unknown> = {}
  if (enName.trim()) patch.name = enName.trim()
  if (enDescription.trim()) patch.description = enDescription.trim()
  const hasAnyValue = Object.keys(patch).length > 0
  const hadTranslation = isEditMode && enSnapshot !== JSON.stringify(['', ''])
  if (!hasAnyValue && !hadTranslation) return null
  return { en: patch }
}

// 校验并提交表单，名称必填，字典码仅新建时填写。
function handleSubmit() {
  if (!name.trim()) {
    formError = '请填写字典类型名称'
    return
  }
  if (!isEditMode && !code.trim()) {
    formError = '请填写字典码'
    return
  }

  const payload: Record<string, unknown> = {
    name: name.trim(),
    status,
    sortOrder
  }
  if (description.trim()) payload.description = description.trim()
  if (!isEditMode) payload.code = code.trim()

  // 英文翻译补丁存在时携带，由后端校验语言键并按语言存储。
  const i18n = buildI18nPayload()
  if (i18n) payload.i18n = i18n

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑字典类型' : '新建字典类型'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改字典类型名称与生效状态' : '创建一个新的字典类型'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="dict-type-code">字典码 *</Label.Root>
        <Input.Root
          id="dict-type-code"
          bind:value={code}
          maxlength={50}
          placeholder="如 tag_status，小写字母开头"
          disabled={isSubmitting || isEditMode}
        />
        <p class="text-xs text-muted-foreground">
          {isEditMode ? '字典码创建后不可变更' : '业务侧以字典码定位字典，创建后不可变更'}
        </p>
      </div>

      <div class="space-y-2">
        <Label.Root for="dict-type-name">名称 *</Label.Root>
        <Input.Root
          id="dict-type-name"
          bind:value={name}
          maxlength={50}
          placeholder="请输入字典类型名称"
          disabled={isSubmitting}
        />
        {#if formError}
          <p class="text-sm text-destructive">{formError}</p>
        {/if}
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="dict-type-sort">排序权重</Label.Root>
          <Input.Root
            id="dict-type-sort"
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
        <Label.Root for="dict-type-description">描述</Label.Root>
        <Input.Root
          id="dict-type-description"
          bind:value={description}
          maxlength={200}
          placeholder="字典类型用途说明"
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
          <Label.Root for="dict-type-en-name">英文名称</Label.Root>
          <Input.Root
            id="dict-type-en-name"
            bind:value={enName}
            maxlength={50}
            placeholder="English name"
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="dict-type-en-description">英文描述</Label.Root>
          <Input.Root
            id="dict-type-en-description"
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
        {isEditMode ? '保存修改' : '创建类型'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
