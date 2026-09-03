<script lang="ts">
import type {
  Category,
  CategoryTreeNode,
  CreateCategoryRequest,
  UpdateCategoryRequest
} from '@myblog/api/modules/category/types'
import CategoryFormDialog from '$lib/components/admin/category/category-form-dialog.svelte'
import { FolderTree, Plus, MoreHorizontal, Trash2, Pencil, Search } from '@lucide/svelte'
import { Button, Card, Badge, DropdownMenu, Input, ToggleGroup } from '$ui'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { CATEGORY_STATUS_CONFIG } from '$lib/constants/category'
import { CategoryAPI } from '$lib/api'
import { onMount } from 'svelte'

// 带缩进深度的分类行，用于表格化展示树形结构。
interface CategoryRow {
  category: Category
  depth: number
}

let rows = $state<CategoryRow[]>([])
let flatCategories = $state<Category[]>([])
let isLoading = $state(true)

// 客户端筛选状态，'' 表示全部，'1' 仅显示，'0' 仅隐藏。
let searchQuery = $state('')
let statusFilter = $state('')

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<Category | null>(null)
let deleteTarget = $state<Category | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

/**
 * 按名称/URL 标识与显示状态过滤分类树行，树数据为全量，无需服务端分页。
 */
const filteredRows = $derived(
  rows.filter(row => {
    const query = searchQuery.trim().toLowerCase()
    const nameHit = !query || row.category.name.toLowerCase().includes(query)
    const slugHit = !query || (row.category.slug ?? '').toLowerCase().includes(query)
    const statusHit = statusFilter === '' || row.category.status === Number(statusFilter)
    return (nameHit || slugHit) && statusHit
  })
)

/**
 * 重置筛选条件并清空搜索结果。
 */
function resetFilters() {
  searchQuery = ''
  statusFilter = ''
}

/**
 * 将分类树扁平化为带缩进深度的行序列。
 */
function flattenTree(nodes: CategoryTreeNode[], depth = 0): CategoryRow[] {
  return nodes.flatMap(node => [
    { category: node, depth },
    ...flattenTree(node.children ?? [], depth + 1)
  ])
}

/**
 * 加载分类树并生成扁平行与可选项列表。
 */
async function loadCategories() {
  isLoading = true
  try {
    const response = await CategoryAPI.getTree()
    if (response.code === 200 && response.data) {
      const tree = response.data.tree ?? []
      rows = flattenTree(tree)
      flatCategories = flattenTree(tree).map(row => row.category)
    } else {
      toast.error(response.message || '加载分类失败')
    }
  } catch (error) {
    console.error('加载分类失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 打开新建弹窗。
 */
function openCreate() {
  dialogTarget = null
  isDialogOpen = true
}

/**
 * 提交分类表单，新建与编辑分别调用对应接口。
 */
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  try {
    const response = dialogTarget
      ? await CategoryAPI.update({ id: dialogTarget.id, ...payload } as UpdateCategoryRequest)
      : await CategoryAPI.create(payload as unknown as CreateCategoryRequest)
    if (response.code === 200 && response.data) {
      toast.success(dialogTarget ? '分类更新成功' : '分类创建成功')
      isDialogOpen = false
      loadCategories()
    } else {
      toast.error(response.message || '保存分类失败')
    }
  } catch (error) {
    console.error('保存分类失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

/**
 * 删除分类，存在子分类时后端会拒绝。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await CategoryAPI.delete(deleteTarget.id)
    if (response.code === 200) {
      toast.success('分类删除成功')
      deleteTarget = null
      loadCategories()
    } else {
      toast.error(response.message || '删除分类失败')
    }
  } catch (error) {
    console.error('删除分类失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

onMount(loadCategories)
</script>

<svelte:head>
  <title>分类管理 - MyBlog</title>
</svelte:head>

<PageHeader
  title="分类管理"
  description="树形管理文章分类，支持层级结构与显示状态控制"
  crumb="分类管理"
>
  {#snippet actions()}
    <Button onclick={() => openCreate()}>
      <Plus data-icon="inline-start" />
      新建分类
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索分类名称或 URL 标识..."
            bind:value={searchQuery}
          />
        </div>

        <ToggleGroup.Root type="single" bind:value={statusFilter} variant="outline" size="sm">
          <ToggleGroup.Item value="">全部状态</ToggleGroup.Item>
          <ToggleGroup.Item value="1">显示</ToggleGroup.Item>
          <ToggleGroup.Item value="0">隐藏</ToggleGroup.Item>
        </ToggleGroup.Root>

        <Button variant="outline" onclick={resetFilters}>重置</Button>
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
      {:else if filteredRows.length === 0}
        <div class="flex h-48 items-center justify-center">
          <div class="text-center">
            <FolderTree class="mx-auto size-12 text-muted-foreground" />
            <h3 class="mt-4 text-lg font-medium">
              {rows.length === 0 ? '暂无分类' : '没有匹配的分类'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {rows.length === 0 ? '创建第一个分类来组织文章' : '调整筛选条件后再试'}
            </p>
          </div>
        </div>
      {:else}
        <div class="divide-y">
          {#each filteredRows as row (row.category.id)}
            {@const statusConfig = CATEGORY_STATUS_CONFIG[row.category.status]!}
            <div
              class="flex items-center gap-4 px-6 py-3"
              style="padding-left: {1 + row.depth * 1.5}rem"
            >
              <div class="flex min-w-0 flex-1 items-center gap-3">
                {#if row.depth > 0}
                  <span class="text-muted-foreground">└</span>
                {/if}
                <span class="font-medium">{row.category.name}</span>
                <span class="text-sm text-muted-foreground">/slug: {row.category.slug}</span>
                <Badge variant={statusConfig.variant}>{statusConfig.label}</Badge>
                {#if row.category.isFeatured}
                  <Badge variant="outline">精选</Badge>
                {/if}
              </div>
              <span class="shrink-0 text-sm text-muted-foreground">
                {row.category.articleCount} 篇文章
              </span>

              <DropdownMenu.Root>
                <DropdownMenu.Trigger>
                  <Button variant="ghost" size="sm" aria-label="分类操作">
                    <MoreHorizontal />
                  </Button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  <DropdownMenu.Item onselect={() => openCreate()}>新建子分类</DropdownMenu.Item>
                  <DropdownMenu.Item
                    onselect={() => {
                      dialogTarget = row.category
                      isDialogOpen = true
                    }}
                  >
                    <Pencil data-icon="inline-start" />
                    编辑
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    variant="destructive"
                    onselect={() => (deleteTarget = row.category)}
                  >
                    <Trash2 data-icon="inline-start" />
                    删除
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
          {/each}
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <CategoryFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    categories={flatCategories}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除分类"
    description={deleteTarget ? `确定删除「${deleteTarget.name}」吗？存在子分类时无法删除。` : ''}
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
