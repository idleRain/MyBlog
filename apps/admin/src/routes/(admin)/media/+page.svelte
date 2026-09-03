<script lang="ts">
import { Upload, Image, Film, FileText, Copy, Trash2, Loader2 } from '@lucide/svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import Pagination from '$lib/components/admin/pagination.svelte'
import type { MediaFile } from '@myblog/api/modules/media/types'
import { Button, Card, ToggleGroup } from '$ui'
import { getFileSize } from '@myblog/shared'
import { MediaAPI } from '$lib/api'
import { onMount } from 'svelte'

// 媒体类型筛选选项，mimeType 为前缀过滤值。
const TYPE_FILTERS: Array<{ value: string; label: string }> = [
  { value: '', label: '全部' },
  { value: 'image/', label: '图片' },
  { value: 'video/', label: '视频' },
  { value: 'application/', label: '文档' }
]

let media = $state<MediaFile[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let mimeTypeFilter = $state('')
let isUploading = $state(false)
let fileInput: HTMLInputElement | undefined = $state()

let deleteTarget = $state<MediaFile | null>(null)
let isDeleting = $state(false)

/**
 * 加载媒体列表，携带类型筛选与分页参数。
 */
async function loadMedia() {
  isLoading = true
  try {
    const response = await MediaAPI.list({
      page: currentPage,
      pageSize: 12,
      ...(mimeTypeFilter ? { mimeType: mimeTypeFilter } : {})
    })
    if (response.code === 200 && response.data) {
      media = response.data.media ?? []
      total = response.data.total ?? 0
    } else {
      toast.error(response.message || '加载媒体失败')
    }
  } catch (error) {
    console.error('加载媒体失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 上传单个文件，完成后刷新列表。
 */
async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || isUploading) return

  isUploading = true
  try {
    const response = await MediaAPI.upload(file)
    if (response.code === 200 && response.data) {
      toast.success('文件上传成功')
      currentPage = 1
      loadMedia()
    } else {
      toast.error(response.message || '上传失败')
    }
  } catch (error) {
    console.error('上传失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isUploading = false
    // 清空文件输入以允许重复选择同一文件。
    input.value = ''
  }
}

/**
 * 复制文件 URL 到剪贴板并提示结果。
 */
async function copyUrl(file: MediaFile) {
  try {
    await navigator.clipboard.writeText(file.fileUrl)
    toast.success('链接已复制')
  } catch (error) {
    console.error('复制链接失败:', error)
    toast.error('复制失败，请手动复制')
  }
}

/**
 * 删除媒体文件。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await MediaAPI.delete(deleteTarget.id)
    if (response.code === 200) {
      toast.success('文件删除成功')
      deleteTarget = null
      loadMedia()
    } else {
      toast.error(response.message || '删除失败')
    }
  } catch (error) {
    console.error('删除失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

/**
 * 根据 MIME 类型返回文件类型图标组件。
 */
function typeIcon(file: MediaFile) {
  if (file.mimeType.startsWith('image/')) return Image
  if (file.mimeType.startsWith('video/')) return Film
  return FileText
}

function handlePageChange(page: number) {
  currentPage = page
  loadMedia()
}

onMount(loadMedia)
</script>

<svelte:head>
  <title>媒体管理 - MyBlog</title>
</svelte:head>

<PageHeader title="媒体管理" description="上传与管理图片、视频与文档素材" crumb="媒体管理">
  {#snippet actions()}
    <input
      bind:this={fileInput}
      type="file"
      class="hidden"
      accept="image/*,video/*,audio/*,application/pdf,text/*"
      onchange={handleUpload}
    />
    <Button onclick={() => fileInput?.click()} disabled={isUploading}>
      {#if isUploading}
        <Loader2 data-icon="inline-start" class="animate-spin" />
        上传中...
      {:else}
        <Upload data-icon="inline-start" />
        上传文件
      {/if}
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <ToggleGroup.Root type="single" bind:value={mimeTypeFilter} variant="outline" size="sm">
        {#each TYPE_FILTERS as filter (filter.value)}
          <ToggleGroup.Item value={filter.value}>{filter.label}</ToggleGroup.Item>
        {/each}
      </ToggleGroup.Root>
    </Card.Content>
  </Card.Root>

  {#if isLoading}
    <div class="flex h-48 items-center justify-center">
      <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
      ></span>
    </div>
  {:else if media.length === 0}
    <div class="flex h-48 items-center justify-center">
      <div class="text-center">
        <Image class="mx-auto size-12 text-muted-foreground" />
        <h3 class="mt-4 text-lg font-medium">暂无媒体文件</h3>
        <p class="text-sm text-muted-foreground">点击右上角上传第一个文件</p>
      </div>
    </div>
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {#each media as file (file.id)}
        <Card.Root class="overflow-hidden">
          <div class="flex h-36 items-center justify-center border-b bg-muted/40">
            {#if file.mimeType.startsWith('image/')}
              <img
                src={file.thumbnailUrl || file.fileUrl}
                alt={file.filename}
                class="h-full w-full object-cover"
              />
            {:else}
              {@const IconComponent = typeIcon(file)}
              <IconComponent class="size-10 text-muted-foreground" />
            {/if}
          </div>
          <Card.Content class="space-y-2 p-3">
            <p class="truncate text-sm font-medium" title={file.filename}>{file.filename}</p>
            <p class="text-xs text-muted-foreground">
              {getFileSize(file.fileSize)} · {file.mimeType}
            </p>
            <div class="flex gap-2">
              <Button variant="outline" size="sm" class="flex-1" onclick={() => copyUrl(file)}>
                <Copy data-icon="inline-start" />
                复制链接
              </Button>
              <Button
                variant="outline"
                size="sm"
                aria-label="删除文件"
                class="text-destructive hover:text-destructive"
                onclick={() => (deleteTarget = file)}
              >
                <Trash2 />
              </Button>
            </div>
          </Card.Content>
        </Card.Root>
      {/each}
    </div>

    <Pagination page={currentPage} {total} pageSize={12} onPageChange={handlePageChange} />
  {/if}

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
