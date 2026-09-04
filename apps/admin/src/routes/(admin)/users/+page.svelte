<script lang="ts">
import type {
  CreateUserRequest,
  UpdateUserRequest,
  User,
  UserRole
} from '@myblog/api/modules/user/types'
import UserFormDialog from '$lib/components/admin/user/user-form-dialog.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import UserTable from '$lib/components/admin/user/user-table.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { getAssignableRoles } from '$lib/utils/permissions'
import { Button, Card, Input, Pagination } from '$ui'
import { USER_PAGE_SIZE } from '$lib/constants/user'
import { Plus, Search } from '@lucide/svelte'
import { authStore } from '$lib/stores/auth'
import { UserAPI } from '$lib/api'
import { onMount } from 'svelte'

let users = $state<User[]>([])
let isLoading = $state(true)
let total = $state(0)
let currentPage = $state(1)
let searchQuery = $state('')

// 批量选择状态
let selectedIds = $state<number[]>([])

// 弹窗与删除确认状态
let isDialogOpen = $state(false)
let dialogTarget = $state<User | null>(null)
let deleteTarget = $state<User | null>(null)
let isDeleting = $state(false)
let isSubmitting = $state(false)

// 当前用户可分配角色
let assignableRoles = $state<UserRole[]>(['user', 'editor', 'admin', 'superadmin'])

// 批量操作确认状态，用于 AlertDialog 二次确认。
let batchConfirm = $state<{ action: 'enable' | 'disable' | 'delete' } | null>(null)
let isBatchExecuting = $state(false)

/**
 * 加载用户列表，搜索词经 keyword 参数交由后端模糊匹配。
 */
async function loadUsers() {
  isLoading = true
  try {
    const query = searchQuery.trim()
    const response = await UserAPI.getUserList(currentPage, USER_PAGE_SIZE, query)
    if (response.code === 200 && response.data) {
      users = response.data.users ?? []
      total = response.data.total ?? 0
      selectedIds = []
    } else {
      toast.error(response.message || '加载用户列表失败')
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 提交用户表单，新建与编辑分别调用对应接口。
 */
async function handleConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  try {
    const response = dialogTarget
      ? await UserAPI.updateUser({ id: dialogTarget.id, ...payload } as UpdateUserRequest)
      : await UserAPI.createUser(payload as unknown as CreateUserRequest)
    if (response.code === 200 && response.data) {
      toast.success(dialogTarget ? '用户更新成功' : '用户创建成功')
      isDialogOpen = false
      loadUsers()
    } else {
      toast.error(response.message || '保存用户失败')
    }
  } catch (error) {
    console.error('保存用户失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

/**
 * 切换用户启用与禁用状态。
 */
async function handleToggleStatus(user: User) {
  const nextStatus = user.status === 1 ? 0 : 1
  const action = nextStatus === 1 ? '启用' : '禁用'
  try {
    const response = await UserAPI.updateUser({
      id: user.id,
      username: user.username,
      email: user.email,
      nickname: user.nickname || '',
      role: user.role || 'user',
      birthday: user.birthday || '',
      status: nextStatus
    })
    if (response.code === 200) {
      toast.success(`用户${action}成功`)
      loadUsers()
    } else {
      toast.error(response.message || `${action}用户失败`)
    }
  } catch (error) {
    console.error(`切换用户状态失败:`, error)
    toast.error('网络错误，请稍后重试')
  }
}

/**
 * 删除单个用户。
 */
async function handleDelete() {
  if (!deleteTarget || isDeleting) return
  isDeleting = true
  try {
    const response = await UserAPI.deleteUser(deleteTarget.id)
    if (response.code === 200) {
      toast.success('用户删除成功')
      deleteTarget = null
      loadUsers()
    } else {
      toast.error(response.message || '删除用户失败')
    }
  } catch (error) {
    console.error('删除用户失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isDeleting = false
  }
}

/**
 * 请求批量切换状态，弹出确认对话框。
 */
function requestBatchStatus(status: 0 | 1) {
  batchConfirm = { action: status === 1 ? 'enable' : 'disable' }
}

/**
 * 请求批量删除，弹出确认对话框。
 */
function requestBatchDelete() {
  batchConfirm = { action: 'delete' }
}

/**
 * 执行确认后的批量操作，按动作类型分发到状态切换或删除。
 */
async function executeBatch() {
  if (!batchConfirm || isBatchExecuting) return
  isBatchExecuting = true

  let successCount = 0
  if (batchConfirm.action === 'delete') {
    for (const userId of selectedIds) {
      try {
        const response = await UserAPI.deleteUser(userId)
        if (response.code === 200) successCount++
      } catch (error) {
        console.error('批量删除失败:', error)
      }
    }
  } else {
    const targetStatus = batchConfirm.action === 'enable' ? 1 : 0
    for (const userId of selectedIds) {
      const user = users.find(item => item.id === userId)
      if (!user) continue
      try {
        const response = await UserAPI.updateUser({
          id: user.id,
          username: user.username,
          email: user.email,
          nickname: user.nickname || '',
          role: user.role || 'user',
          birthday: user.birthday || '',
          status: targetStatus
        })
        if (response.code === 200) successCount++
      } catch (error) {
        console.error('批量切换状态失败:', error)
      }
    }
  }

  const label =
    batchConfirm.action === 'delete' ? '删除' : batchConfirm.action === 'enable' ? '启用' : '禁用'
  toast.success(`成功${label} ${successCount} 个用户`)
  batchConfirm = null
  isBatchExecuting = false
  loadUsers()
}

function toggleSelectAll() {
  selectedIds = selectedIds.length === users.length ? [] : users.map(user => user.id)
}

function toggleSelect(id: number) {
  selectedIds = selectedIds.includes(id)
    ? selectedIds.filter(item => item !== id)
    : [...selectedIds, id]
}

function handlePageChange(page: number) {
  currentPage = page
  loadUsers()
}

const isAllSelected = $derived(users.length > 0 && selectedIds.length === users.length)

onMount(() => {
  const currentUser = authStore.getCurrentState().user
  assignableRoles = getAssignableRoles(currentUser)
  loadUsers()
})
</script>

<svelte:head>
  <title>用户管理 - MyBlog</title>
</svelte:head>

<PageHeader title="用户管理" description="管理系统用户，包括创建、编辑与删除操作" crumb="用户管理">
  {#snippet actions()}
    <Button
      onclick={() => {
        dialogTarget = null
        isDialogOpen = true
      }}
    >
      <Plus data-icon="inline-start" />
      创建用户
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative min-w-52 flex-1">
          <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input.Root
            class="pl-9"
            placeholder="搜索用户名、邮箱或昵称..."
            bind:value={searchQuery}
            onkeydown={(event: KeyboardEvent) => {
              if (event.key === 'Enter') {
                currentPage = 1
                loadUsers()
              }
            }}
          />
        </div>
        <span class="text-sm text-muted-foreground">共 {total} 个用户</span>
      </div>
    </Card.Content>
  </Card.Root>

  {#if selectedIds.length > 0}
    <Card.Root>
      <Card.Content class="p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="text-sm text-muted-foreground">已选择 {selectedIds.length} 个用户</span>
          <div class="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onclick={() => requestBatchStatus(1)}>
              批量启用
            </Button>
            <Button variant="outline" size="sm" onclick={() => requestBatchStatus(0)}>
              批量禁用
            </Button>
            <Button variant="destructive" size="sm" onclick={requestBatchDelete}>批量删除</Button>
            <Button variant="ghost" size="sm" onclick={() => (selectedIds = [])}>取消选择</Button>
          </div>
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  <UserTable
    {users}
    {selectedIds}
    {isAllSelected}
    {isLoading}
    onToggleSelectAll={toggleSelectAll}
    onToggleSelect={toggleSelect}
    onEdit={user => {
      dialogTarget = user
      isDialogOpen = true
    }}
    onToggleStatus={handleToggleStatus}
    onDelete={user => (deleteTarget = user)}
  />

  <Pagination.Root
    count={total}
    perPage={USER_PAGE_SIZE}
    page={currentPage}
    onPageChange={handlePageChange}
  >
    <Pagination.Content>
      <Pagination.PrevButton />
      <span class="px-2 text-sm text-muted-foreground">
        第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / USER_PAGE_SIZE))} 页
      </span>
      <Pagination.NextButton />
    </Pagination.Content>
  </Pagination.Root>

  <UserFormDialog
    {isSubmitting}
    {assignableRoles}
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
    title={batchConfirm?.action === 'delete' ? '批量删除用户' : '批量切换状态'}
    description={batchConfirm
      ? batchConfirm.action === 'delete'
        ? `确定删除选中的 ${selectedIds.length} 个用户吗？此操作不可恢复。`
        : `确定${batchConfirm.action === 'enable' ? '启用' : '禁用'}选中的 ${selectedIds.length} 个用户吗？`
      : ''}
    confirmText={batchConfirm?.action === 'delete' ? '删除' : '确认'}
    destructive={batchConfirm?.action === 'delete'}
    isLoading={isBatchExecuting}
    open={batchConfirm !== null}
    onOpenChange={open => {
      if (!open && !isBatchExecuting) batchConfirm = null
    }}
    onConfirm={executeBatch}
  />
</PageHeader>
