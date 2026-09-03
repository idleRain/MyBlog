<script lang="ts">
import ArticleEditorForm from '$lib/components/admin/article/article-editor-form.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
</script>

<svelte:head>
  <title>新建文章 - MyBlog</title>
</svelte:head>

<PageHeader title="新建文章" description="撰写一篇新的博客文章，可先存草稿再发布" crumb="新建文章">
  <ArticleEditorForm />
</PageHeader>
