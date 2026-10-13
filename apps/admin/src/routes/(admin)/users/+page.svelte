<script lang="ts">
import type { CreateUserRequest, UpdateUserRequest, User } from '@myblog/api/modules/user/types'
import { USER_PAGE_SIZE, USER_STATUS_DISABLED, USER_STATUS_ENABLED } from '$lib/constants/user'
import { userPageState, emptyUserPageData, type BatchAction } from './user-page-state.svelte'
import UserFormDialog from '$lib/components/admin/user/user-form-dialog.svelte'
import { Plus, RotateCcw, Search, Users as UsersIcon } from '@lucide/svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import UserTable from '$lib/components/admin/user/user-table.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Empty from '$ui/empty'
import { Input } from '$ui/input'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  userPageState.initialize(data ?? emptyUserPageData())
})

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<User | null>(null)
let deleteTarget = $state<User | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 批量操作确认状态，用于 AlertDialog 二次确认。
let batchConfirm = $state<BatchAction | null>(null)
let isBatchExecuting = $state(false)

// 打开用户弹窗，目标为空表示新建。
function openDialog(target: User | null) {
  dialogTarget = target
  isDialogOpen = true
}

// 提交用户表单，新建与编辑分别调用对应写操作。
// 弹窗载荷为动态键值结构，与既有表单弹窗约定一致地经类型断言收窄。
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = dialogTarget
    ? await userPageState.updateUser({ id: dialogTarget.id, ...payload } as UpdateUserRequest)
    : await userPageState.createUser(payload as unknown as CreateUserRequest)
  isSubmitting = false
  if (success) isDialogOpen = false
}

// 确认后删除目标用户，列表刷新由状态模块在写操作成功后触发。
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  const success = await userPageState.removeUser(deleteTarget.id)
  isDeleting = false
  if (success) deleteTarget = null
}

// 请求批量切换状态，弹出确认对话框。
function requestBatchStatus(status: typeof USER_STATUS_DISABLED | typeof USER_STATUS_ENABLED) {
  batchConfirm = status === USER_STATUS_ENABLED ? 'enable' : 'disable'
}

// 请求批量删除，弹出确认对话框。
function requestBatchDelete() {
  batchConfirm = 'delete'
}

// 执行确认后的批量操作，成功后关闭确认对话框。
async function handleBatchConfirm() {
  if (!batchConfirm || isBatchExecuting) return
  const action = batchConfirm
  isBatchExecuting = true
  const success = await userPageState.executeBatch(action)
  isBatchExecuting = false
  if (success) batchConfirm = null
}
</script>

<svelte:head>
  <title>用户管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="用户管理" description="管理系统用户，包括创建、编辑与删除操作" crumb="用户管理">
  {#snippet actions()}
    <Button onclick={() => openDialog(null)}>
      <Plus data-icon="inline-start" />
      创建用户
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <!-- 工具栏：搜索即输即搜，重置作为整体撤销入口 -->
      <div class="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            class="pl-9"
            placeholder="搜索用户名、邮箱或昵称..."
            bind:value={userPageState.keyword}
            oninput={userPageState.applySearchDebounced}
          />
        </div>
        <Button
          variant="ghost"
          size="sm"
          class="ml-auto"
          onclick={() => userPageState.resetAndReload()}
        >
          <RotateCcw data-icon="inline-start" />
          重置
        </Button>
      </div>

      <!-- 批量操作栏：仅在存在选中项时出现，操作均经确认对话框二次确认 -->
      {#if userPageState.selectedIds.length > 0}
        <div
          class="flex flex-wrap items-center justify-between gap-3 border-b bg-muted/40 px-4 py-2.5"
        >
          <span class="text-sm text-muted-foreground">
            已选择 {userPageState.selectedIds.length} 个用户
          </span>
          <div class="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onclick={() => requestBatchStatus(USER_STATUS_ENABLED)}
            >
              批量启用
            </Button>
            <Button
              variant="outline"
              size="sm"
              onclick={() => requestBatchStatus(USER_STATUS_DISABLED)}
            >
              批量禁用
            </Button>
            <Button variant="destructive" size="sm" onclick={requestBatchDelete}>批量删除</Button>
            <Button variant="ghost" size="sm" onclick={() => userPageState.clearSelection()}>
              取消选择
            </Button>
          </div>
        </div>
      {/if}

      {#if userPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载用户" />
        </div>
      {:else if userPageState.users.length === 0}
        <Empty.Root class="h-64 border-0">
          <Empty.Header>
            <Empty.Media variant="icon">
              <UsersIcon />
            </Empty.Media>
            <Empty.Title>
              {userPageState.hasActiveFilters ? '没有匹配的用户' : '暂无用户'}
            </Empty.Title>
            <Empty.Description>
              {userPageState.hasActiveFilters ? '调整或重置搜索条件后再试' : '创建第一个用户账号'}
            </Empty.Description>
          </Empty.Header>
          <Empty.Content>
            {#if userPageState.hasActiveFilters}
              <Button variant="outline" size="sm" onclick={() => userPageState.resetAndReload()}>
                <RotateCcw data-icon="inline-start" />
                重置搜索
              </Button>
            {:else}
              <Button size="sm" onclick={() => openDialog(null)}>
                <Plus data-icon="inline-start" />
                创建用户
              </Button>
            {/if}
          </Empty.Content>
        </Empty.Root>
      {:else}
        <UserTable
          users={userPageState.users}
          selectedIds={userPageState.selectedIds}
          isAllSelected={userPageState.isAllSelected}
          onToggleSelectAll={() => userPageState.toggleSelectAll()}
          onToggleSelect={id => userPageState.toggleSelect(id)}
          onEdit={user => openDialog(user)}
          onToggleStatus={user => void userPageState.toggleUserStatus(user)}
          onDelete={user => (deleteTarget = user)}
        />

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {userPageState.total} 个用户</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={userPageState.total}
            perPage={USER_PAGE_SIZE}
            page={userPageState.currentPage}
            onPageChange={page => userPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {userPageState.currentPage} 页，共 {userPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <UserFormDialog
    {isSubmitting}
    assignableRoles={userPageState.assignableRoles}
    open={isDialogOpen}
    target={dialogTarget}
    onOpenChange={open => (isDialogOpen = open)}
    onConfirm={handleConfirm}
  />

  <ConfirmDialog
    title="删除用户"
    description={deleteTarget ? `确定删除用户「${deleteTarget.username}」吗？此操作不可恢复。` : ''}
    confirmText="删除"
    destructive
    isLoading={isDeleting}
    open={deleteTarget !== null}
    onOpenChange={open => {
      if (!open && !isDeleting) deleteTarget = null
    }}
    onConfirm={handleDelete}
  />

  <ConfirmDialog
    title={batchConfirm === 'delete' ? '批量删除用户' : '批量切换状态'}
    description={batchConfirm
      ? batchConfirm === 'delete'
        ? `确定删除选中的 ${userPageState.selectedIds.length} 个用户吗？此操作不可恢复。`
        : `确定${batchConfirm === 'enable' ? '启用' : '禁用'}选中的 ${userPageState.selectedIds.length} 个用户吗？`
      : ''}
    confirmText={batchConfirm === 'delete' ? '删除' : '确认'}
    destructive={batchConfirm === 'delete'}
    isLoading={isBatchExecuting}
    open={batchConfirm !== null}
    onOpenChange={open => {
      if (!open && !isBatchExecuting) batchConfirm = null
    }}
    onConfirm={handleBatchConfirm}
  />
</PageHeader>
