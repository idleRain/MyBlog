<script lang="ts">
import { MEDIA_FILE_ACCEPT, MEDIA_PAGE_SIZE, MEDIA_TYPE_FILTERS } from '$lib/constants/media'
import { mediaPageState, emptyMediaPageData } from './media-page-state.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import MediaGrid from '$lib/components/admin/media/media-grid.svelte'
import { Image, Loader2, RotateCcw, Upload } from '@lucide/svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { MediaFile } from '@myblog/api/modules/media/types'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Empty from '$ui/empty'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  mediaPageState.initialize(data ?? emptyMediaPageData())
})

// 隐藏文件输入框与删除确认状态
let fileInput: HTMLInputElement | undefined = $state()
let deleteTarget = $state<MediaFile | null>(null)
let isDeleting = $state(false)

// 触发隐藏文件输入框，打开系统选择文件对话框。
function openFilePicker() {
  fileInput?.click()
}

// 上传单个文件，无论成功与否都清空输入以允许重复选择同一文件。
async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || mediaPageState.isUploading) return
  try {
    await mediaPageState.uploadFile(file)
  } finally {
    input.value = ''
  }
}

// 确认后删除目标文件，成功后关闭确认弹窗。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await mediaPageState.removeMedia(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>媒体管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="媒体管理" description="上传与管理图片、视频与文档素材" crumb="媒体管理">
  {#snippet actions()}
    <input
      bind:this={fileInput}
      type="file"
      class="hidden"
      accept={MEDIA_FILE_ACCEPT}
      onchange={handleUpload}
    />
    <Button onclick={openFilePicker} disabled={mediaPageState.isUploading}>
      {#if mediaPageState.isUploading}
        <Loader2 data-icon="inline-start" class="animate-spin" />
        上传中...
      {:else}
        <Upload data-icon="inline-start" />
        上传文件
      {/if}
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：类型筛选即点即用，仅保留重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={mediaPageState.typeFilter}
          onValueChange={value => mediaPageState.setTypeFilter(value)}
        >
          {#each MEDIA_TYPE_FILTERS as filter (filter.value)}
            <ToggleGroup.Item value={filter.value}>{filter.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => mediaPageState.resetFilters()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if mediaPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载媒体" />
        </div>
      {:else if mediaPageState.media.length === 0}
        <Empty.Root class="h-64 border-0">
          <Empty.Header>
            <Empty.Media variant="icon">
              <Image />
            </Empty.Media>
            <Empty.Title>
              {mediaPageState.hasActiveFilters ? '没有匹配的文件' : '暂无媒体文件'}
            </Empty.Title>
            <Empty.Description>
              {mediaPageState.hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '上传第一个文件，丰富博客的图片与素材'}
            </Empty.Description>
          </Empty.Header>
          <Empty.Content>
            {#if mediaPageState.hasActiveFilters}
              <Button variant="outline" size="sm" onclick={() => mediaPageState.resetFilters()}>
                <RotateCcw data-icon="inline-start" />
                重置筛选
              </Button>
            {:else}
              <Button size="sm" onclick={openFilePicker} disabled={mediaPageState.isUploading}>
                <Upload data-icon="inline-start" />
                上传文件
              </Button>
            {/if}
          </Empty.Content>
        </Empty.Root>
      {:else}
        <MediaGrid
          media={mediaPageState.media}
          onCopy={file => void mediaPageState.copyFileUrl(file)}
          onDelete={file => (deleteTarget = file)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {mediaPageState.total} 个文件</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={mediaPageState.total}
            perPage={MEDIA_PAGE_SIZE}
            page={mediaPageState.currentPage}
            onPageChange={page => mediaPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {mediaPageState.currentPage} 页，共 {mediaPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <ConfirmDialog
    title="删除文件"
    description={deleteTarget ? `确定删除「${deleteTarget.filename}」吗？此操作不可恢复。` : ''}
    confirmText="删除"
    destructive
    isLoading={isDeleting}
    open={deleteTarget !== null}
    onOpenChange={open => {
      if (!open && !isDeleting) deleteTarget = null
    }}
    onConfirm={handleDelete}
  />
</PageHeader>
