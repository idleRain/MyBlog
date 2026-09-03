<script lang="ts">
import type {
  CreateTagRequest,
  Tag,
  TagStatus,
  UpdateTagRequest
} from '@myblog/api/modules/tag/types'
import { Plus, Search, Trash2, Pencil, Tags as TagsIcon } from '@lucide/svelte'
import TagFormDialog from '$lib/components/admin/tag/tag-form-dialog.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import { TAG_PAGE_SIZE, TAG_STATUS_CONFIG } from '$lib/constants/tag'
import { Badge, Button, Card, Input, Table, ToggleGroup } from '$ui'
import PageHeader from '$lib/components/admin/page-header.svelte'
import Pagination from '$lib/components/admin/pagination.svelte'
import { TagAPI } from '$lib/api'
import { onMount } from 'svelte'

let tags = $state<Tag[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let search = $state('')
// 状态筛选使用字符串值：'' 全部、'1' 启用、'0' 隐藏。
let statusFilter = $state('')
// 热门筛选：'' 全部、'1' 仅热门。
let hotFilter = $state('')

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<Tag | null>(null)
let deleteTarget = $state<Tag | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

/**
 * 加载标签列表，携带搜索、状态与热门筛选。
 */
async function loadTags() {
  isLoading = true
  try {
    const response = await TagAPI.adminList({
      page: currentPage,
      pageSize: TAG_PAGE_SIZE,
      ...(statusFilter !== '' ? { status: Number(statusFilter) as TagStatus } : {}),
      ...(hotFilter === '1' ? { isHot: true } : {}),
      ...(search.trim() ? { search: search.trim() } : {})
    })
    if (response.code === 200 && response.data) {
      tags = response.data.tags ?? []
      total = response.data.total ?? 0
    } else {
      toast.error(response.message || '加载标签失败')
    }
  } catch (error) {
    console.error('加载标签失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 重置筛选并回到第一页。
 */
function resetAndReload() {
  search = ''
  statusFilter = ''
  hotFilter = ''
  currentPage = 1
  loadTags()
}

/**
 * 提交标签表单，新建与编辑分别调用对应接口。
 */
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  try {
    const response = dialogTarget
      ? await TagAPI.update({ id: dialogTarget.id, ...payload } as UpdateTagRequest)
      : await TagAPI.create(payload as unknown as CreateTagRequest)
    if (response.code === 200 && response.data) {
      toast.success(dialogTarget ? '标签更新成功' : '标签创建成功')
      isDialogOpen = false
      loadTags()
    } else {
      toast.error(response.message || '保存标签失败')
    }
  } catch (error) {
    console.error('保存标签失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

/**
 * 删除标签，已挂载文章的标签删除由后端处理。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await TagAPI.delete(deleteTarget.id)
    if (response.code === 200) {
      toast.success('标签删除成功')
      deleteTarget = null
      loadTags()
    } else {
      toast.error(response.message || '删除标签失败')
    }
  } catch (error) {
    console.error('删除标签失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

function handlePageChange(page: number) {
  currentPage = page
  loadTags()
}

onMount(loadTags)
</script>

<svelte:head>
  <title>标签管理 - MyBlog</title>
</svelte:head>

<PageHeader
  title="标签管理"
  description="管理文章标签，支持热门标记与显示状态控制"
  crumb="标签管理"
>
  {#snippet actions()}
    <Button
      onclick={() => {
        dialogTarget = null
        isDialogOpen = true
      }}
    >
      <Plus data-icon="inline-start" />
      新建标签
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <div class="flex flex-wrap items-end gap-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索标签名称或描述..."
            bind:value={search}
            onkeydown={(event: KeyboardEvent) => {
              if (event.key === 'Enter') {
                currentPage = 1
                loadTags()
              }
            }}
          />
        </div>

        <div class="flex flex-col gap-2">
          <ToggleGroup.Root type="single" bind:value={statusFilter} variant="outline" size="sm">
            <ToggleGroup.Item value="">全部</ToggleGroup.Item>
            <ToggleGroup.Item value="1">启用</ToggleGroup.Item>
            <ToggleGroup.Item value="0">隐藏</ToggleGroup.Item>
          </ToggleGroup.Root>

          <ToggleGroup.Root type="single" bind:value={hotFilter} variant="outline" size="sm">
            <ToggleGroup.Item value="">全部标签</ToggleGroup.Item>
            <ToggleGroup.Item value="1">仅热门</ToggleGroup.Item>
          </ToggleGroup.Root>
        </div>

        <div class="flex gap-2">
          <Button
            onclick={() => {
              currentPage = 1
              loadTags()
            }}
          >
            <Search data-icon="inline-start" />
            搜索
          </Button>
          <Button variant="outline" onclick={resetAndReload}>重置</Button>
        </div>
      </div>
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Content class="p-0">
      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if tags.length === 0}
        <div class="flex h-48 items-center justify-center">
          <div class="text-center">
            <TagsIcon class="mx-auto size-12 text-muted-foreground" />
            <h3 class="mt-4 text-lg font-medium">暂无标签</h3>
            <p class="text-sm text-muted-foreground">调整筛选条件或创建第一个标签</p>
          </div>
        </div>
      {:else}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>标签</Table.Head>
              <Table.Head>URL 标识</Table.Head>
              <Table.Head>使用次数</Table.Head>
              <Table.Head>热门</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head>创建时间</Table.Head>
              <Table.Head class="text-right">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each tags as tag (tag.id)}
              <Table.Row>
                <Table.Cell>
                  <div class="flex items-center gap-2">
                    <span class="size-3 rounded-full" style="background-color: {tag.color}"></span>
                    <span class="font-medium">{tag.name}</span>
                  </div>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground">{tag.slug}</span>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm">{tag.usageCount}</span>
                </Table.Cell>
                <Table.Cell>
                  {#if tag.isHot}
                    <Badge variant="outline">热门</Badge>
                  {:else}
                    <span class="text-sm text-muted-foreground">—</span>
                  {/if}
                </Table.Cell>
                <Table.Cell>
                  {@const statusConfig = TAG_STATUS_CONFIG[tag.status]!}
                  <Badge variant={statusConfig.variant}>{statusConfig.label}</Badge>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground">
                    {new Date(tag.createdAt).toLocaleDateString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-end gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      aria-label="编辑标签"
                      onclick={() => {
                        dialogTarget = tag
                        isDialogOpen = true
                      }}
                    >
                      <Pencil />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      aria-label="删除标签"
                      class="text-destructive hover:text-destructive"
                      onclick={() => (deleteTarget = tag)}
                    >
                      <Trash2 />
                    </Button>
                  </div>
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      {/if}
    </Card.Content>
  </Card.Root>

  <Pagination page={currentPage} {total} pageSize={TAG_PAGE_SIZE} onPageChange={handlePageChange} />

  <TagFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除标签"
    description={deleteTarget ? `确定删除标签「${deleteTarget.name}」吗？` : ''}
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
