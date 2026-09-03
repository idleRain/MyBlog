<script lang="ts">
import CategoryTagPicker, {
  type CategoryTagSelection
} from '$lib/components/admin/article/category-tag-picker.svelte'
import MarkdownEditor from '$lib/components/admin/markdown-editor.svelte'
import SeoFields from '$lib/components/admin/article/seo-fields.svelte'
import type { Article } from '@myblog/api/modules/article/types'
import { Button, Card, Input, Label, Switch } from '$ui'
import { ArrowLeft, Save, Send } from '@lucide/svelte'
import { goto } from '$lib/utils/navigation'
import { ArticleAPI } from '$lib/api'
import { onMount } from 'svelte'

interface Props {
  article?: Article | null
}

let { article = null }: Props = $props()

// 是否编辑模式由文章快照推导，供模板与保存逻辑复用。
const isEditMode = $derived(article !== null)

// 表单基础字段，初始值在挂载时由文章数据一次性回填。
let title = $state('')
let slug = $state('')
let summary = $state('')
let coverImage = $state('')
let content = $state('')

// 分类与标签选择
let categorySelection = $state<CategoryTagSelection>({
  categoryId: null,
  categoryIds: [],
  tagIds: []
})

// 文章属性开关
let isFeatured = $state(false)
let isTop = $state(false)
let commentEnabled = $state(true)

// SEO 字段
let seoTitle = $state('')
let seoDescription = $state('')
let seoKeywords = $state('')

let isSubmitting = $state(false)

/**
 * 组件挂载后从文章数据回填表单，编辑场景下父组件在数据加载完成后才渲染本组件。
 */
onMount(() => {
  title = article?.title ?? ''
  slug = article?.slug ?? ''
  summary = article?.summary ?? ''
  coverImage = article?.coverImage ?? ''
  content = article?.content ?? ''
  categorySelection = {
    categoryId: article?.categoryId ?? null,
    categoryIds: article?.categories?.map(category => category.id) ?? [],
    tagIds: article?.tags?.map(tag => tag.id) ?? []
  }
  isFeatured = article?.isFeatured ?? false
  isTop = article?.isTop ?? false
  commentEnabled = article?.commentEnabled ?? true
  seoTitle = article?.seoTitle ?? ''
  seoDescription = article?.seoDescription ?? ''
  seoKeywords = article?.seoKeywords ?? ''
})

/**
 * 校验表单必填项，返回错误提示信息，为空表示校验通过。
 */
function validateForm(): string {
  if (!title.trim()) return '请填写文章标题'
  if (!content.trim()) return '请填写文章正文内容'
  return ''
}

/**
 * 组装请求载荷，可选字段仅在非空时携带，适配 exactOptionalPropertyTypes 约束。
 */
function buildPayload(status: 'draft' | 'published') {
  return {
    title: title.trim(),
    content,
    status,
    isFeatured,
    isTop,
    commentEnabled,
    categoryId: categorySelection.categoryId,
    categoryIds: categorySelection.categoryIds,
    tagIds: categorySelection.tagIds,
    ...(slug.trim() ? { slug: slug.trim() } : {}),
    ...(summary.trim() ? { summary: summary.trim() } : {}),
    ...(coverImage.trim() ? { coverImage: coverImage.trim() } : {}),
    ...(seoTitle.trim() ? { seoTitle: seoTitle.trim() } : {}),
    ...(seoDescription.trim() ? { seoDescription: seoDescription.trim() } : {}),
    ...(seoKeywords.trim() ? { seoKeywords: seoKeywords.trim() } : {})
  }
}

/**
 * 保存文章，按目标状态区分草稿与发布，新建与编辑分别调用对应接口。
 * 更新接口由后端按作者或管理员身份统一授权，前端不做角色分派。
 */
async function handleSave(targetStatus: 'draft' | 'published') {
  if (isSubmitting) return
  const validationError = validateForm()
  if (validationError) {
    toast.error(validationError)
    return
  }

  isSubmitting = true
  try {
    const base = buildPayload(targetStatus)
    const response = isEditMode
      ? await ArticleAPI.update({ ...base, id: article!.id })
      : await ArticleAPI.create(base)

    if (response.code === 200 && response.data) {
      toast.success(isEditMode ? '文章更新成功' : '文章创建成功')
      goto('/posts')
    } else {
      toast.error(response.message || '保存文章失败')
    }
  } catch (error) {
    console.error('保存文章失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}
</script>

<div class="mx-auto max-w-4xl space-y-6">
  <!-- 基本信息 -->
  <Card.Root>
    <Card.Header>
      <Card.Title>基本信息</Card.Title>
      <Card.Description>标题、摘要与封面图</Card.Description>
    </Card.Header>
    <Card.Content class="space-y-4">
      <div class="space-y-2">
        <Label.Root for="article-title">标题 *</Label.Root>
        <Input.Root
          id="article-title"
          bind:value={title}
          maxlength={200}
          placeholder="请输入文章标题"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="article-slug">URL 标识</Label.Root>
        <Input.Root
          id="article-slug"
          bind:value={slug}
          maxlength={200}
          placeholder="留空时由标题自动生成"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="article-summary">摘要</Label.Root>
        <Input.Root
          id="article-summary"
          bind:value={summary}
          maxlength={500}
          placeholder="文章摘要，用于列表展示与分享"
          disabled={isSubmitting}
        />
      </div>

      <div class="space-y-2">
        <Label.Root for="article-cover">封面图 URL</Label.Root>
        <Input.Root
          id="article-cover"
          bind:value={coverImage}
          maxlength={500}
          placeholder="https://example.com/cover.jpg"
          disabled={isSubmitting}
        />
      </div>
    </Card.Content>
  </Card.Root>

  <!-- 正文内容 -->
  <Card.Root>
    <Card.Header>
      <Card.Title>正文内容 *</Card.Title>
      <Card.Description>支持 Markdown 语法，可切换编辑与预览模式</Card.Description>
    </Card.Header>
    <Card.Content>
      <MarkdownEditor bind:value={content} />
    </Card.Content>
  </Card.Root>

  <!-- 分类与标签 -->
  <Card.Root>
    <Card.Header>
      <Card.Title>分类与标签</Card.Title>
      <Card.Description>选择主分类、关联分类与标签</Card.Description>
    </Card.Header>
    <Card.Content>
      <CategoryTagPicker bind:value={categorySelection} />
    </Card.Content>
  </Card.Root>

  <!-- 文章设置 -->
  <Card.Root>
    <Card.Header>
      <Card.Title>文章设置</Card.Title>
      <Card.Description>精选、置顶与评论开关</Card.Description>
    </Card.Header>
    <Card.Content>
      <div class="grid gap-4 sm:grid-cols-3">
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-1">
            <Label.Root>精选</Label.Root>
            <p class="text-xs text-muted-foreground">在首页精选展示</p>
          </div>
          <Switch.Switch bind:checked={isFeatured} disabled={isSubmitting} />
        </div>
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-1">
            <Label.Root>置顶</Label.Root>
            <p class="text-xs text-muted-foreground">列表优先展示</p>
          </div>
          <Switch.Switch bind:checked={isTop} disabled={isSubmitting} />
        </div>
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-1">
            <Label.Root>允许评论</Label.Root>
            <p class="text-xs text-muted-foreground">关闭后读者无法评论</p>
          </div>
          <Switch.Switch bind:checked={commentEnabled} disabled={isSubmitting} />
        </div>
      </div>
    </Card.Content>
  </Card.Root>

  <!-- SEO 设置 -->
  <Card.Root>
    <Card.Header>
      <Card.Title>SEO 设置</Card.Title>
      <Card.Description>优化搜索引擎展示效果</Card.Description>
    </Card.Header>
    <Card.Content>
      <SeoFields bind:seoTitle bind:seoDescription bind:seoKeywords />
    </Card.Content>
  </Card.Root>

  <!-- 操作栏 -->
  <div class="flex items-center justify-between gap-2 border-t pt-4">
    <Button variant="ghost" onclick={() => goto('/posts')} disabled={isSubmitting}>
      <ArrowLeft data-icon="inline-start" />
      返回列表
    </Button>
    <div class="flex gap-2">
      <Button variant="outline" onclick={() => handleSave('draft')} disabled={isSubmitting}>
        <Save data-icon="inline-start" />
        保存草稿
      </Button>
      <Button onclick={() => handleSave('published')} disabled={isSubmitting}>
        <Send data-icon="inline-start" />
        {isEditMode ? '更新并发布' : '创建并发布'}
      </Button>
    </div>
  </div>
</div>
