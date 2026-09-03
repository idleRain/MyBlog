<script lang="ts">
import type { Category, CategoryStatus } from '@myblog/api/modules/category/types'
import { Button, Dialog, Input, Label, Select, Separator, Switch } from '$ui'

interface Props {
  open: boolean
  // 编辑目标分类，为空表示新建。
  target?: Category | null
  // 可选父分类列表，用于新建子分类，顶级分类为空。
  categories: Category[]
  isSubmitting: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (payload: Record<string, unknown>) => void
}

let { open, target = null, categories, isSubmitting, onOpenChange, onConfirm }: Props = $props()

const isEditMode = $derived(target !== null)

// 表单状态，弹窗打开时由目标分类回填。
let name = $state('')
let slug = $state('')
let description = $state('')
let coverImage = $state('')
let parentId = $state('')
let sortOrder = $state(0)
let status = $state<CategoryStatus>(1)
let isFeatured = $state(false)
let seoTitle = $state('')
let seoDescription = $state('')
let formError = $state('')

/**
 * 弹窗打开时重置表单并回填编辑数据。
 */
$effect(() => {
  if (!open) return
  name = target?.name ?? ''
  slug = target?.slug ?? ''
  description = target?.description ?? ''
  coverImage = target?.coverImage ?? ''
  parentId = target?.parentId != null ? String(target.parentId) : ''
  sortOrder = target?.sortOrder ?? 0
  status = target?.status ?? 1
  isFeatured = target?.isFeatured ?? false
  seoTitle = target?.seoTitle ?? ''
  seoDescription = target?.seoDescription ?? ''
  formError = ''
})

/**
 * 校验并提交表单，分类名称必填。
 */
function handleSubmit() {
  if (!name.trim()) {
    formError = '请填写分类名称'
    return
  }

  // 新建时支持指定父分类，编辑时父分类不可修改。
  const payload: Record<string, unknown> = {
    name: name.trim(),
    sortOrder,
    status,
    isFeatured
  }
  // 可选字段仅在非空时携带，适配 exactOptionalPropertyTypes。
  if (slug.trim()) payload.slug = slug.trim()
  if (description.trim()) payload.description = description.trim()
  if (coverImage.trim()) payload.coverImage = coverImage.trim()
  if (seoTitle.trim()) payload.seoTitle = seoTitle.trim()
  if (seoDescription.trim()) payload.seoDescription = seoDescription.trim()
  if (!isEditMode && parentId !== '') payload.parentId = Number(parentId)

  onConfirm(payload)
}
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="sm:max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{isEditMode ? '编辑分类' : '新建分类'}</Dialog.Title>
      <Dialog.Description>
        {isEditMode ? '修改分类信息，树结构不可调整' : '创建新的文章分类'}
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="category-name">名称 *</Label.Root>
        <Input.Root
          id="category-name"
          bind:value={name}
          maxlength={50}
          placeholder="请输入分类名称"
          disabled={isSubmitting}
        />
        {#if formError}
          <p class="text-sm text-destructive">{formError}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <Label.Root for="category-slug">URL 标识</Label.Root>
        <Input.Root
          id="category-slug"
          bind:value={slug}
          maxlength={50}
          placeholder="留空时由名称自动生成"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="category-description">描述</Label.Root>
        <Input.Root
          id="category-description"
          bind:value={description}
          maxlength={1000}
          placeholder="分类简介"
          disabled={isSubmitting}
        />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root>父分类</Label.Root>
          <Select.Root type="single" bind:value={parentId} disabled={isEditMode || isSubmitting}>
            <Select.Trigger class="w-full">
              {parentId === ''
                ? '顶级分类'
                : (categories.find(item => String(item.id) === parentId)?.name ?? '顶级分类')}
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                <Select.Item value="">顶级分类</Select.Item>
                {#each categories as category (category.id)}
                  <Select.Item value={String(category.id)}>{category.name}</Select.Item>
                {/each}
              </Select.Group>
            </Select.Content>
          </Select.Root>
          {#if isEditMode}
            <p class="text-xs text-muted-foreground">编辑时不可修改父分类</p>
          {/if}
        </div>

        <div class="space-y-2">
          <Label.Root for="category-sort">排序权重</Label.Root>
          <Input.Root
            id="category-sort"
            type="number"
            bind:value={sortOrder}
            disabled={isSubmitting}
          />
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="category-cover">封面图 URL</Label.Root>
          <Input.Root
            id="category-cover"
            bind:value={coverImage}
            maxlength={255}
            placeholder="分类封面图"
            disabled={isSubmitting}
          />
        </div>

        <div class="space-y-3 pt-2">
          <div class="flex items-center justify-between gap-2">
            <div class="space-y-0.5">
              <Label.Root>显示状态</Label.Root>
              <p class="text-xs text-muted-foreground">隐藏后前台不可见</p>
            </div>
            <Switch.Switch
              checked={status === 1}
              onCheckedChange={(checked: boolean) => {
                status = checked ? 1 : 0
              }}
              disabled={isSubmitting}
            />
          </div>
          <div class="flex items-center justify-between gap-2">
            <div class="space-y-0.5">
              <Label.Root>精选</Label.Root>
              <p class="text-xs text-muted-foreground">作为推荐分类展示</p>
            </div>
            <Switch.Switch bind:checked={isFeatured} disabled={isSubmitting} />
          </div>
        </div>
      </div>

      <Separator.Root />

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label.Root for="category-seo-title">SEO 标题</Label.Root>
          <Input.Root
            id="category-seo-title"
            bind:value={seoTitle}
            maxlength={100}
            disabled={isSubmitting}
          />
        </div>
        <div class="space-y-2">
          <Label.Root for="category-seo-desc">SEO 描述</Label.Root>
          <Input.Root
            id="category-seo-desc"
            bind:value={seoDescription}
            maxlength={255}
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
        {isEditMode ? '保存修改' : '创建分类'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
