<script lang="ts">
import {
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  Link as LinkIcon,
  CircleCheck,
  CircleX,
  EyeOff,
  RotateCcw,
  type LucideIcon
} from '@lucide/svelte'
import type {
  CreateFriendlyLinkRequest,
  FriendlyLink,
  LinkStatus,
  UpdateFriendlyLinkRequest
} from '@myblog/api/modules/friendlyLink/types'
import { LINK_PAGE_SIZE, LINK_STATUS_CONFIG, LINK_STATUS_OPTIONS } from '$lib/constants/link'
import { Badge, Button, Card, DropdownMenu, Pagination, Table, ToggleGroup } from '$ui'
import LinkFormDialog from '$lib/components/admin/link/link-form-dialog.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { FriendlyLinkAPI } from '$lib/api'
import { onMount } from 'svelte'

// 各状态下可执行的审核动作，icon 供下拉菜单项渲染语义图标。
const STATUS_ACTIONS: Record<
  LinkStatus,
  Array<{ key: 'approve' | 'hide' | 'reject'; label: string; icon: LucideIcon }>
> = {
  pending: [
    { key: 'approve', label: '通过', icon: CircleCheck },
    { key: 'reject', label: '拒绝', icon: CircleX }
  ],
  active: [{ key: 'hide', label: '下架', icon: EyeOff }],
  hidden: [
    { key: 'approve', label: '重新上架', icon: CircleCheck },
    { key: 'reject', label: '拒绝', icon: CircleX }
  ],
  rejected: [{ key: 'approve', label: '重新上架', icon: CircleCheck }]
}

let links = $state<FriendlyLink[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let statusFilter = $state<LinkStatus | ''>('')

let isDialogOpen = $state(false)
let dialogTarget = $state<FriendlyLink | null>(null)
let deleteTarget = $state<FriendlyLink | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 是否存在生效中的筛选条件，用于空态区分「无数据」与「无匹配」。
const hasActiveFilters = $derived(statusFilter !== '')

/**
 * 加载友链列表。
 */
async function loadLinks() {
  isLoading = true
  try {
    const response = await FriendlyLinkAPI.adminList({
      page: currentPage,
      pageSize: LINK_PAGE_SIZE,
      ...(statusFilter ? { status: statusFilter } : {})
    })
    if (response.code === 200 && response.data) {
      links = response.data.links ?? []
      total = response.data.total ?? 0
    } else {
      toast.error(response.message || '加载友链失败')
    }
  } catch (error) {
    console.error('加载友链失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 状态筛选变更后立即回到第一页并重新加载。
 */
function handleStatusFilterChange(value: string) {
  statusFilter = value as LinkStatus | ''
  currentPage = 1
  loadLinks()
}

/**
 * 重置筛选并回到第一页。
 */
function resetAndReload() {
  statusFilter = ''
  currentPage = 1
  loadLinks()
}

/**
 * 打开添加友链弹窗。
 */
function openCreateDialog() {
  dialogTarget = null
  isDialogOpen = true
}

/**
 * 提交友链表单，新建与编辑分别调用对应接口。
 */
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  try {
    const response = dialogTarget
      ? await FriendlyLinkAPI.update({
          id: dialogTarget.id,
          ...payload
        } as UpdateFriendlyLinkRequest)
      : await FriendlyLinkAPI.create(payload as unknown as CreateFriendlyLinkRequest)
    if (response.code === 200 && response.data) {
      toast.success(dialogTarget ? '友链更新成功' : '友链添加成功')
      isDialogOpen = false
      loadLinks()
    } else {
      toast.error(response.message || '保存友链失败')
    }
  } catch (error) {
    console.error('保存友链失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

/**
 * 执行友链审核动作。
 */
async function handleStatusAction(link: FriendlyLink, key: 'approve' | 'hide' | 'reject') {
  try {
    const response = await FriendlyLinkAPI[key](link.id)
    if (response.code === 200) {
      toast.success(`${STATUS_ACTIONS[link.status].find(item => item.key === key)?.label}成功`)
      loadLinks()
    } else {
      toast.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('友链操作失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

/**
 * 删除友链。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await FriendlyLinkAPI.delete(deleteTarget.id)
    if (response.code === 200) {
      toast.success('友链删除成功')
      deleteTarget = null
      loadLinks()
    } else {
      toast.error(response.message || '删除友链失败')
    }
  } catch (error) {
    console.error('删除友链失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

function handlePageChange(page: number) {
  currentPage = page
  loadLinks()
}

onMount(loadLinks)
</script>

<svelte:head>
  <title>友情链接 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="友情链接" description="管理互链申请，覆盖申请、审核与展示状态" crumb="友情链接">
  {#snippet actions()}
    <Button onclick={openCreateDialog}>
      <Plus data-icon="inline-start" />
      添加友链
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={statusFilter}
          onValueChange={handleStatusFilterChange}
        >
          {#each LINK_STATUS_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
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
      {:else if links.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <LinkIcon class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {hasActiveFilters ? '没有匹配的友链' : '暂无友链'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '添加第一条友情链接，与友邻站点互相推荐'}
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
              添加友链
            </Button>
          {/if}
        </div>
      {:else}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>站点</Table.Head>
              <Table.Head>描述</Table.Head>
              <Table.Head>回链</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head>添加时间</Table.Head>
              <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each links as link (link.id)}
              <Table.Row>
                <Table.Cell>
                  <div class="flex items-center gap-3">
                    {#if link.logo}
                      <img
                        src={link.logo}
                        alt={link.name}
                        class="size-8 rounded-full object-cover"
                      />
                    {/if}
                    <div class="min-w-0">
                      <p class="truncate font-medium">{link.name}</p>
                      <a
                        href={link.url}
                        target="_blank"
                        rel="noreferrer"
                        class="truncate text-xs text-muted-foreground hover:underline"
                      >
                        {link.url}
                      </a>
                    </div>
                  </div>
                </Table.Cell>
                <Table.Cell>
                  <span class="line-clamp-1 max-w-48 text-sm text-muted-foreground">
                    {link.description || '—'}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  {#if link.isReciprocal}
                    <Badge variant="outline">已回链</Badge>
                  {:else}
                    <span class="text-sm text-muted-foreground">—</span>
                  {/if}
                </Table.Cell>
                <Table.Cell>
                  <Badge variant={LINK_STATUS_CONFIG[link.status].variant}>
                    {LINK_STATUS_CONFIG[link.status].label}
                  </Badge>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground tabular-nums">
                    {new Date(link.createdAt).toLocaleDateString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-center gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      aria-label="编辑友链"
                      onclick={() => {
                        dialogTarget = link
                        isDialogOpen = true
                      }}
                    >
                      <Pencil />
                    </Button>
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <Button variant="ghost" size="sm" aria-label="友链操作">
                          <MoreHorizontal />
                        </Button>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        {#each STATUS_ACTIONS[link.status] as action (action.key)}
                          {@const ActionIcon = action.icon}
                          <DropdownMenu.Item onSelect={() => handleStatusAction(link, action.key)}>
                            <ActionIcon data-icon="inline-start" />
                            {action.label}
                          </DropdownMenu.Item>
                        {/each}
                        <DropdownMenu.Item
                          variant="destructive"
                          onSelect={() => (deleteTarget = link)}
                        >
                          <Trash2 data-icon="inline-start" />
                          删除
                        </DropdownMenu.Item>
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  </div>
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {total} 个友链</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={total}
            perPage={LINK_PAGE_SIZE}
            page={currentPage}
            onPageChange={handlePageChange}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / LINK_PAGE_SIZE))} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <LinkFormDialog
    {isSubmitting}
    open={isDialogOpen}
    target={dialogTarget}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除友链"
    description={deleteTarget ? `确定删除「${deleteTarget.name}」吗？此操作不可恢复。` : ''}
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
