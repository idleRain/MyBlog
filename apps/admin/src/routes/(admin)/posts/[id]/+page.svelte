<script lang="ts">
import ArticleEditorForm from '$lib/components/admin/article/article-editor-form.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { Article } from '@myblog/api/modules/article/types'
import { ArrowLeft, FileText } from '@lucide/svelte'
import { goto } from '$lib/utils/navigation'
import { ArticleAPI } from '$lib/api'
import { page } from '$app/stores'
import { Button, Card } from '$ui'
import { onMount } from 'svelte'

let article = $state<Article | null>(null)
let isLoading = $state(true)
let loadError = $state('')

/**
 * 根据路由参数加载文章详情，供编辑表单回填。
 */
async function loadArticle() {
  const articleId = Number($page.params.id)
  if (!Number.isFinite(articleId) || articleId <= 0) {
    loadError = '无效的文章标识'
    isLoading = false
    return
  }

  try {
    const response = await ArticleAPI.getById(articleId)
    if (response.code === 200 && response.data) {
      article = response.data
    } else {
      loadError = response.message || '文章不存在或无权访问'
    }
  } catch (error) {
    console.error('加载文章失败:', error)
    loadError = '网络错误，请稍后重试'
  } finally {
    isLoading = false
  }
}

onMount(loadArticle)
</script>

<svelte:head>
  <title>编辑文章 - MyBlog</title>
</svelte:head>

<PageHeader title="编辑文章" description="修改文章内容、分类与发布状态" crumb="编辑文章">
  {#if isLoading}
    <div class="flex h-48 items-center justify-center">
      <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
      ></span>
    </div>
  {:else if loadError}
    <Card.Root>
      <Card.Content class="flex flex-col items-center gap-4 py-16">
        <FileText class="size-12 text-muted-foreground" />
        <div class="text-center">
          <h3 class="text-lg font-medium">无法加载文章</h3>
          <p class="mt-1 text-sm text-muted-foreground">{loadError}</p>
        </div>
        <Button variant="outline" onclick={() => goto('/posts')}>
          <ArrowLeft data-icon="inline-start" />
          返回文章列表
        </Button>
      </Card.Content>
    </Card.Root>
  {:else if article}
    <ArticleEditorForm {article} />
  {/if}
</PageHeader>
