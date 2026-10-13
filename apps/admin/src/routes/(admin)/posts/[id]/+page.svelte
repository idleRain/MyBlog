<script lang="ts">
import ArticleEditorForm from '$lib/components/admin/article/article-editor-form.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { ArrowLeft, FileText } from '@lucide/svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { goto } from '$lib/utils/navigation'
import type { PageProps } from './$types'
import { Button } from '$ui/button'
import * as Card from '$ui/card'

// 文章详情与失败文案均由 load 注入，组件只做渲染与返回跳转。
let { data }: PageProps = $props()
</script>

<svelte:head>
  <title>编辑文章 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="编辑文章" description="修改文章内容、分类与发布状态" crumb="编辑文章">
  {#if data.loadError}
    <Card.Root>
      <Card.Content class="flex flex-col items-center gap-4 py-16">
        <FileText class="size-12 text-muted-foreground" />
        <div class="text-center">
          <h3 class="text-lg font-medium">无法加载文章</h3>
          <p class="mt-1 text-sm text-muted-foreground">{data.loadError}</p>
        </div>
        <Button variant="outline" onclick={() => goto('/posts')}>
          <ArrowLeft data-icon="inline-start" />
          返回文章列表
        </Button>
      </Card.Content>
    </Card.Root>
  {:else if data.article}
    <ArticleEditorForm article={data.article} />
  {/if}
</PageHeader>
