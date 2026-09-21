<script lang="ts">
import type {
  CreateTagRequest,
  Tag,
  TagStatus,
  UpdateTagRequest
} from '@myblog/api/modules/tag/types'
import { Flame, Pencil, Plus, RotateCcw, Search, Tags as TagsIcon, Trash2 } from '@lucide/svelte'
import { dictItemLabel, DICT_TYPE_CODE_TAG_STATUS, findDictItem } from '@myblog/api'
import { Badge, Button, Card, Input, Pagination, Table, ToggleGroup } from '$ui'
import TagFormDialog from '$lib/components/admin/tag/tag-form-dialog.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import { TAG_PAGE_SIZE, TAG_STATUS_CONFIG } from '$lib/constants/tag'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { SITE_NAME_ZH, debounce } from '@myblog/shared'
import type { BadgeVariant } from '$ui/badge'
import { DictAPI, TagAPI } from '$lib/api'
import { onMount } from 'svelte'

let tags = $state<Tag[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let search = $state('')
// 状态筛选使用字符串值：'' 全部，其余为字典项值，与标签整型状态的字符串形式对齐。
let statusFilter = $state('')
// 热门筛选：'' 全部、'1' 仅热门。
let hotFilter = $state('')

// 标签状态字典分组，状态文案与徽标样式经该字典动态读取，加载失败时回退内置配置。
let tagStatusDict = $state<EnabledDictGroup | null>(null)

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<Tag | null>(null)
let deleteTarget = $state<Tag | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// Badge 合法样式集合，用于校验字典扩展元数据中的样式值。
const BADGE_VARIANTS: BadgeVariant[] = ['default', 'secondary', 'destructive', 'outline']

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 状态筛选项，字典可用时按字典项生成并保持排序；停用字典项仍参与筛选，保证存量数据可被过滤。
const statusFilterOptions = $derived.by(() => {
  if (!tagStatusDict) {
    return [
      { value: '1', label: '启用' },
      { value: '0', label: '隐藏' }
    ]
  }
  return [...tagStatusDict.items]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map(item => ({ value: item.value, label: item.label }))
})

// 是否存在生效中的筛选条件，用于空态区分「无数据」与「无匹配」。
const hasActiveFilters = $derived(search.trim() !== '' || statusFilter !== '' || hotFilter !== '')

// 解析标签状态的展示文案：优先字典项显示名，缺失时回退内置配置。
function resolveStatusLabel(status: number): string {
  return dictItemLabel(tagStatusDict, String(status), TAG_STATUS_CONFIG[status]?.label ?? '未知')
}

// 解析标签状态的徽标样式：优先字典扩展元数据，缺失或非法时回退内置配置。
function resolveStatusVariant(status: number): BadgeVariant {
  const fallback = TAG_STATUS_CONFIG[status]?.variant ?? 'secondary'
  const variant = findDictItem(tagStatusDict, String(status))?.extra?.variant
  if (typeof variant === 'string' && BADGE_VARIANTS.includes(variant as BadgeVariant)) {
    return variant as BadgeVariant
  }
  return fallback
}

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
 * 防抖应用搜索词，连续输入只在停顿后请求一次。
 */
const applySearchDebounced = debounce(() => {
  currentPage = 1
  loadTags()
}, SEARCH_DEBOUNCE_MS)

/**
 * 状态筛选变更后立即回到第一页并重新加载。
 */
function handleStatusFilterChange(value: string) {
  statusFilter = value
  currentPage = 1
  loadTags()
}

/**
 * 热门筛选变更后立即回到第一页并重新加载。
 */
function handleHotFilterChange(value: string) {
  hotFilter = value
  currentPage = 1
  loadTags()
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
 * 打开新建标签弹窗。
 */
function openCreateDialog() {
  dialogTarget = null
  isDialogOpen = true
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

// 加载标签状态字典，失败时静默回退内置配置，不阻断列表使用。
async function loadTagStatusDict() {
  try {
    const response = await DictAPI.getByType(DICT_TYPE_CODE_TAG_STATUS)
    if (response.code === 200 && response.data) {
      tagStatusDict = response.data
    }
  } catch (error) {
    console.error('加载标签状态字典失败:', error)
  }
}

onMount(() => {
  loadTags()
  loadTagStatusDict()
})
</script>

<svelte:head>
  <title>标签管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title="标签管理"
  description="管理文章标签，支持热门标记与显示状态控制"
  crumb="标签管理"
>
  {#snippet actions()}
    <Button onclick={openCreateDialog}>
      <Plus data-icon="inline-start" />
      新建标签
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：搜索即输即搜，筛选即点即用，仅保留重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-56 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索标签名称或描述..."
            bind:value={search}
            oninput={applySearchDebounced}
          />
        </div>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={statusFilter}
          onValueChange={handleStatusFilterChange}
        >
          <ToggleGroup.Item value="">全部</ToggleGroup.Item>
          {#each statusFilterOptions as option (option.value)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={hotFilter}
          onValueChange={handleHotFilterChange}
        >
          <ToggleGroup.Item value="">全部</ToggleGroup.Item>
          <ToggleGroup.Item value="1">仅热门</ToggleGroup.Item>
        </ToggleGroup.Root>

        <Button variant="ghost" size="sm" class="ml-auto" onclick={resetAndReload}>
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if tags.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <TagsIcon class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {hasActiveFilters ? '没有匹配的标签' : '暂无标签'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {hasActiveFilters ? '调整或重置筛选条件后再试' : '创建第一个标签，开始组织文章内容'}
            </p>
          </div>
          {#if hasActiveFilters}
            <Button variant="outline" size="sm" onclick={resetAndReload}>
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {:else}
            <Button size="sm" onclick={openCreateDialog}>
              <Plus data-icon="inline-start" />
              新建标签
            </Button>
          {/if}
        </div>
      {:else}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>标签</Table.Head>
              <Table.Head>URL 标识</Table.Head>
              <Table.Head>使用次数</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head>创建时间</Table.Head>
              <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each tags as tag (tag.id)}
              <Table.Row>
                <Table.Cell>
                  <div class="flex flex-col gap-1">
                    <span
                      class="inline-flex w-fit items-center gap-1.5 rounded-md border px-2 py-0.5 text-sm font-medium"
                      style="border-color: color-mix(in oklab, {tag.color} 35%, transparent); background-color: color-mix(in oklab, {tag.color} 10%, transparent);"
                    >
                      <span
                        class="size-2 rounded-full"
                        style="background-color: {tag.color}"
                        aria-hidden="true"
                      ></span>
                      {tag.name}
                      {#if tag.isHot}
                        <Flame class="size-3.5 fill-primary text-primary" aria-hidden="true" />
                        <span class="sr-only">热门</span>
                      {/if}
                    </span>
                    {#if tag.description}
                      <p class="max-w-72 truncate text-xs text-muted-foreground">
                        {tag.description}
                      </p>
                    {/if}
                  </div>
                </Table.Cell>
                <Table.Cell>
                  <code
                    class="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground"
                  >
                    {tag.slug}
                  </code>
                </Table.Cell>
                <Table.Cell>
                  <span
                    class="text-sm tabular-nums {tag.usageCount > 0
                      ? 'text-foreground'
                      : 'text-muted-foreground'}"
                  >
                    {tag.usageCount}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <Badge variant={resolveStatusVariant(tag.status)}>
                    {resolveStatusLabel(tag.status)}
                  </Badge>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground tabular-nums">
                    {new Date(tag.createdAt).toLocaleDateString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-center gap-1">
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

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {total} 个标签</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={total}
            perPage={TAG_PAGE_SIZE}
            page={currentPage}
            onPageChange={handlePageChange}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / TAG_PAGE_SIZE))} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

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
