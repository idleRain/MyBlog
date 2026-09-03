<script lang="ts">
import type {
  CreateFriendlyLinkRequest,
  FriendlyLink,
  LinkStatus,
  UpdateFriendlyLinkRequest
} from '@myblog/api/modules/friendlyLink/types'
import { LINK_PAGE_SIZE, LINK_STATUS_CONFIG, LINK_STATUS_OPTIONS } from '$lib/constants/link'
import { Plus, MoreHorizontal, Pencil, Trash2, Link as LinkIcon } from '@lucide/svelte'
import LinkFormDialog from '$lib/components/admin/link/link-form-dialog.svelte'
import { Badge, Button, Card, DropdownMenu, Table, ToggleGroup } from '$ui'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import Pagination from '$lib/components/admin/pagination.svelte'
import { FriendlyLinkAPI } from '$lib/api'
import { onMount } from 'svelte'

// 各状态下可执行的审核动作
const STATUS_ACTIONS: Record<
  LinkStatus,
  Array<{ key: 'approve' | 'hide' | 'reject'; label: string }>
> = {
  pending: [
    { key: 'approve', label: '通过' },
    { key: 'reject', label: '拒绝' }
  ],
  active: [{ key: 'hide', label: '下架' }],
  hidden: [
    { key: 'approve', label: '重新上架' },
    { key: 'reject', label: '拒绝' }
  ],
  rejected: [{ key: 'approve', label: '重新上架' }]
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
  <title>友情链接 - MyBlog</title>
</svelte:head>

<PageHeader title="友情链接" description="管理互链申请，覆盖申请、审核与展示状态" crumb="友情链接">
  {#snippet actions()}
    <Button
      onclick={() => {
        dialogTarget = null
        isDialogOpen = true
      }}
    >
      <Plus data-icon="inline-start" />
      添加友链
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <ToggleGroup.Root type="single" bind:value={statusFilter} variant="outline" size="sm">
        {#each LINK_STATUS_OPTIONS as option (option.label)}
          <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
        {/each}
      </ToggleGroup.Root>
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
      {:else if links.length === 0}
        <div class="flex h-48 items-center justify-center">
          <div class="text-center">
            <LinkIcon class="mx-auto size-12 text-muted-foreground" />
            <h3 class="mt-4 text-lg font-medium">暂无友链</h3>
            <p class="text-sm text-muted-foreground">添加第一条友情链接</p>
          </div>
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
              <Table.Head class="text-right">操作</Table.Head>
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
                  <span class="text-sm">{link.isReciprocal ? '已回链' : '—'}</span>
                </Table.Cell>
                <Table.Cell>
                  <Badge variant={LINK_STATUS_CONFIG[link.status].variant}>
                    {LINK_STATUS_CONFIG[link.status].label}
                  </Badge>
                </Table.Cell>
                <Table.Cell>
                  <span class="text-sm text-muted-foreground">
                    {new Date(link.createdAt).toLocaleDateString('zh-CN')}
                  </span>
                </Table.Cell>
                <Table.Cell>
                  <div class="flex items-center justify-end gap-1">
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
                          <DropdownMenu.Item onselect={() => handleStatusAction(link, action.key)}>
                            {action.label}
                          </DropdownMenu.Item>
                        {/each}
                        <DropdownMenu.Item
                          variant="destructive"
                          onselect={() => (deleteTarget = link)}
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
      {/if}
    </Card.Content>
  </Card.Root>

  <Pagination
    page={currentPage}
    {total}
    pageSize={LINK_PAGE_SIZE}
    onPageChange={handlePageChange}
  />

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
