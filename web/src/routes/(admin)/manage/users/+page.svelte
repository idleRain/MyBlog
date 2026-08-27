<script lang="ts">
import { onMount } from 'svelte'
import { authStore } from '$lib/stores/auth'
import { Avatar, Badge, Button, Card, Checkbox, Dialog, Input, Select, Skeleton, Table } from '$ui'
import * as Field from '$ui/field'
import * as Empty from '$ui/empty'
import { Spinner } from '$ui/spinner'
import { ConfirmDialog, PageHeader } from '$lib/components/admin'
import { getRoleInfo, getRoleList } from '$lib/utils/permissions'
import {
  ChevronLeft,
  ChevronRight,
  Edit,
  Plus,
  Search,
  Shield,
  Trash2,
  User as UserIcon
} from '@lucide/svelte'
import { UserAPI } from '$lib/api'
import type { UpdateUserRequest, User, UserRole } from '$lib/api/modules/user/types'

// 后端统一成功响应码，与 pkg/response 的 CodeSuccess 对齐。
const RESPONSE_CODE_SUCCESS = 200

// 用户状态枚举值：与后端模型约定一致，1 表示启用，0 表示禁用。
const USER_STATUS = {
  active: 1,
  disabled: 0
} as const

// 用户列表每页条数，与 loadUsers 的分页请求保持一致。
const USER_PAGE_SIZE = 10

// 加载骨架屏行数：覆盖表格首屏高度，避免内容跳变。
const SKELETON_ROW_COUNT = 6

// 角色下拉选项从角色配置派生，中文名与权限模块保持同源。
const roleOptions = getRoleList()

// 确认对话框状态：暂存待执行动作的描述与回调，确认后统一执行。
interface ConfirmAction {
  title: string
  description: string
  confirmText: string
  destructive: boolean
  run: () => Promise<void>
}

// 权限检查
let userRole = $state('user')
let hasPermission = $state(false)

// 用户列表状态
let users = $state<User[]>([])
let isLoading = $state(true)
let searchQuery = $state('')
let isCreateModalOpen = $state(false)
let isEditModalOpen = $state(false)
let selectedUser = $state<User | null>(null)

// 表单状态；role 受 UserRole 枚举约束，避免自由字符串流入 API 请求。
let userForm = $state({
  username: '',
  email: '',
  password: '',
  nickname: '',
  role: 'user' as UserRole,
  birthday: ''
})

let isSubmitting = $state(false)
let formErrors = $state<Record<string, string>>({})

// 分页状态
let currentPage = $state(1)
let totalPages = $state(1)
let totalUsers = $state(0)

// 批量操作状态
let selectedUserIds = $state<number[]>([])
let isAllSelected = $state(false)

// 确认对话框状态
let confirmState = $state<ConfirmAction | null>(null)
let isConfirmOpen = $state(false)

// 加载用户列表
async function loadUsers() {
  try {
    isLoading = true
    const response = await UserAPI.getUserList(currentPage, USER_PAGE_SIZE)

    if (response.code === RESPONSE_CODE_SUCCESS && response.data) {
      users = response.data.users || []
      totalPages = response.data.pages || 1
      totalUsers = response.data.total || 0
    } else {
      toast.error(response.message || '加载用户列表失败')
    }
  } catch (error) {
    console.error('Load users error:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

// 创建用户
async function createUser() {
  if (isSubmitting) return

  // 重置错误
  formErrors = {}

  // 简单验证
  if (!userForm.username.trim()) {
    formErrors.username = '请输入用户名'
    return
  }
  if (!userForm.email.trim()) {
    formErrors.email = '请输入邮箱'
    return
  }
  if (!userForm.password.trim()) {
    formErrors.password = '请输入密码'
    return
  }

  try {
    isSubmitting = true
    const response = await UserAPI.createUser({
      username: userForm.username.trim(),
      email: userForm.email.trim(),
      password: userForm.password.trim(),
      nickname: userForm.nickname.trim(),
      role: userForm.role,
      birthday: userForm.birthday
    })

    if (response.code === RESPONSE_CODE_SUCCESS) {
      toast.success('用户创建成功')
      isCreateModalOpen = false
      resetForm()
      await loadUsers()
    } else {
      toast.error(response.message || '创建用户失败')
    }
  } catch (error) {
    console.error('Create user error:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

// 更新用户
async function updateUser() {
  if (isSubmitting || !selectedUser) return

  // 重置错误
  formErrors = {}

  // 简单验证
  if (!userForm.username.trim()) {
    formErrors.username = '请输入用户名'
    return
  }
  if (!userForm.email.trim()) {
    formErrors.email = '请输入邮箱'
    return
  }

  try {
    isSubmitting = true
    // 更新载荷按接口契约类型化，密码仅在填写时提交。
    const updateData: UpdateUserRequest = {
      id: selectedUser.id,
      username: userForm.username.trim(),
      email: userForm.email.trim(),
      nickname: userForm.nickname.trim(),
      role: userForm.role,
      birthday: userForm.birthday
    }

    // 只有在密码不为空时才更新密码
    if (userForm.password.trim()) {
      updateData.password = userForm.password.trim()
    }

    const response = await UserAPI.updateUser(updateData)

    if (response.code === RESPONSE_CODE_SUCCESS) {
      toast.success('用户更新成功')
      isEditModalOpen = false
      resetForm()
      await loadUsers()
    } else {
      toast.error(response.message || '更新用户失败')
    }
  } catch (error) {
    console.error('Update user error:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSubmitting = false
  }
}

// 应用用户状态变更：实际的 API 调用与结果提示。
async function applyUserStatus(user: User, status: number) {
  const actionLabel = status === USER_STATUS.active ? '启用' : '禁用'

  try {
    const response = await UserAPI.updateUser({
      id: user.id,
      username: user.username,
      email: user.email,
      nickname: user.nickname || '',
      role: user.role || 'user',
      birthday: user.birthday || '',
      status
    })

    if (response.code === RESPONSE_CODE_SUCCESS) {
      toast.success(`用户${actionLabel}成功`)
      await loadUsers()
    } else {
      toast.error(response.message || `${actionLabel}用户失败`)
    }
  } catch (error) {
    console.error('Toggle user status error:', error)
    toast.error('网络错误，请稍后重试')
  }
}

// 请求切换用户状态：先经确认对话框，再执行状态变更。
function requestToggleStatus(user: User) {
  const newStatus = user.status === USER_STATUS.active ? USER_STATUS.disabled : USER_STATUS.active
  const actionLabel = newStatus === USER_STATUS.active ? '启用' : '禁用'

  requestConfirm({
    title: `${actionLabel}用户`,
    description: `确定要${actionLabel}用户 "${user.username}" 吗？该操作会立即生效。`,
    confirmText: actionLabel,
    destructive: newStatus === USER_STATUS.disabled,
    run: () => applyUserStatus(user, newStatus)
  })
}

// 删除用户
async function deleteUser(userId: number) {
  try {
    const response = await UserAPI.deleteUser(userId)

    if (response.code === RESPONSE_CODE_SUCCESS) {
      toast.success('用户删除成功')
      await loadUsers()
    } else {
      toast.error(response.message || '删除用户失败')
    }
  } catch (error) {
    console.error('Delete user error:', error)
    toast.error('网络错误，请稍后重试')
  }
}

// 请求删除用户：先经确认对话框，再执行不可恢复的删除。
function requestDeleteUser(user: User) {
  requestConfirm({
    title: '删除用户',
    description: `确定要删除用户 "${user.username}" 吗？此操作不可恢复。`,
    confirmText: '删除用户',
    destructive: true,
    run: () => deleteUser(user.id)
  })
}

// 重置表单
function resetForm() {
  userForm = {
    username: '',
    email: '',
    password: '',
    nickname: '',
    role: 'user',
    birthday: ''
  }
  formErrors = {}
}

// 打开编辑模态框
function openEditModal(user: User) {
  selectedUser = user
  userForm = {
    username: user.username,
    email: user.email,
    password: '', // 编辑时不显示密码
    nickname: user.nickname || '',
    role: user.role || 'user',
    birthday: user.birthday || ''
  }
  isEditModalOpen = true
}

// 打开创建对话框前重置表单，避免残留上次输入。
function openCreateDialog() {
  resetForm()
  isCreateModalOpen = true
}

// 批量操作函数
function toggleSelectAll() {
  if (isAllSelected) {
    selectedUserIds = []
    isAllSelected = false
  } else {
    selectedUserIds = filteredUsers.map(user => user.id)
    isAllSelected = true
  }
}

function toggleSelectUser(userId: number) {
  if (selectedUserIds.includes(userId)) {
    selectedUserIds = selectedUserIds.filter(id => id !== userId)
  } else {
    selectedUserIds = [...selectedUserIds, userId]
  }
  isAllSelected = selectedUserIds.length === filteredUsers.length
}

// 清空选择状态。
function clearSelection() {
  selectedUserIds = []
  isAllSelected = false
}

// 批量删除：逐个调用删除接口并汇总成功与失败数量。
async function batchDeleteUsers() {
  let successCount = 0
  let failCount = 0

  for (const userId of selectedUserIds) {
    try {
      const response = await UserAPI.deleteUser(userId)
      if (response.code === RESPONSE_CODE_SUCCESS) {
        successCount++
      } else {
        failCount++
      }
    } catch (error) {
      console.error('Batch delete user error:', error)
      failCount++
    }
  }

  if (successCount > 0) {
    toast.success(`成功删除 ${successCount} 个用户`)
  }
  if (failCount > 0) {
    toast.error(`删除失败 ${failCount} 个用户`)
  }

  clearSelection()
  await loadUsers()
}

// 请求批量删除：先经确认对话框，再执行批量删除。
function requestBatchDelete() {
  if (selectedUserIds.length === 0) {
    toast.error('请选择要删除的用户')
    return
  }

  requestConfirm({
    title: '批量删除用户',
    description: `确定要删除选中的 ${selectedUserIds.length} 个用户吗？此操作不可恢复。`,
    confirmText: '批量删除',
    destructive: true,
    run: batchDeleteUsers
  })
}

// 批量切换状态：逐个调用更新接口并汇总成功与失败数量。
async function batchToggleStatus(status: number) {
  const actionLabel = status === USER_STATUS.active ? '启用' : '禁用'
  let successCount = 0
  let failCount = 0

  for (const userId of selectedUserIds) {
    try {
      const user = users.find(u => u.id === userId)
      if (user) {
        const response = await UserAPI.updateUser({
          id: user.id,
          username: user.username,
          email: user.email,
          nickname: user.nickname || '',
          role: user.role || 'user',
          birthday: user.birthday || '',
          status
        })
        if (response.code === RESPONSE_CODE_SUCCESS) {
          successCount++
        } else {
          failCount++
        }
      }
    } catch (error) {
      console.error('Batch toggle user status error:', error)
      failCount++
    }
  }

  if (successCount > 0) {
    toast.success(`成功${actionLabel} ${successCount} 个用户`)
  }
  if (failCount > 0) {
    toast.error(`${actionLabel}失败 ${failCount} 个用户`)
  }

  clearSelection()
  await loadUsers()
}

// 请求批量切换状态：先经确认对话框，再执行批量启用或禁用。
function requestBatchToggle(status: number) {
  if (selectedUserIds.length === 0) {
    toast.error('请选择要操作的用户')
    return
  }

  const actionLabel = status === USER_STATUS.active ? '启用' : '禁用'
  requestConfirm({
    title: `批量${actionLabel}用户`,
    description: `确定要${actionLabel}选中的 ${selectedUserIds.length} 个用户吗？`,
    confirmText: `批量${actionLabel}`,
    destructive: status === USER_STATUS.disabled,
    run: () => batchToggleStatus(status)
  })
}

// 发起确认请求：暂存动作描述并打开对话框。
function requestConfirm(action: ConfirmAction) {
  confirmState = action
  isConfirmOpen = true
}

// 分页跳转：先更新页码再异步加载列表。
function goToPreviousPage() {
  currentPage = Math.max(1, currentPage - 1)
  void loadUsers()
}

function goToNextPage() {
  currentPage = Math.min(totalPages, currentPage + 1)
  void loadUsers()
}

// 过滤用户
let filteredUsers = $derived(
  users.filter(
    user =>
      user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (user.nickname && user.nickname.toLowerCase().includes(searchQuery.toLowerCase()))
  )
)

// 角色下拉的当前选中项文案。
const currentRoleLabel = $derived(
  roleOptions.find(option => option.role === userForm.role)?.name ?? '请选择角色'
)

// 组件挂载时检查权限和加载数据
onMount(() => {
  // 检查用户权限
  authStore.subscribe(state => {
    if (state.isAuthenticated && state.user) {
      userRole = state.user.role || 'user'
      hasPermission = ['admin', 'superadmin'].includes(userRole)

      if (hasPermission) {
        void loadUsers()
      }
    } else {
      hasPermission = false
    }
  })
})
</script>

<svelte:head>
  <title>用户管理 - MyBlog</title>
</svelte:head>

<!-- 头部导航 -->
<PageHeader crumbs={[{ label: '管理后台', href: '/manage' }, { label: '用户管理' }]} />

<!-- 主内容区域：网格底纹仅覆盖首屏，向下渐隐以保持数据区可读。 -->
<div class="relative flex-1 overflow-y-auto">
  <div class="admin-grid pointer-events-none absolute inset-0" aria-hidden="true"></div>
  <div class="relative z-10 flex flex-col gap-6 p-4 sm:p-6">
    {#if !hasPermission}
      <!-- 权限不足 -->
      <div class="flex flex-col items-center justify-center gap-3 py-24 text-center">
        <div
          class="flex size-12 items-center justify-center bg-destructive/10 text-destructive [&_svg]:size-6"
        >
          <Shield />
        </div>
        <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
          <span class="text-signal">//</span> ACCESS DENIED
        </p>
        <h2 class="text-xl font-semibold tracking-tight">权限不足</h2>
        <p class="text-sm text-muted-foreground">只有管理员和超级管理员才能访问用户管理功能。</p>
        <Button variant="outline" class="rounded-none" onclick={() => void goto('/manage')}>
          返回仪表盘
        </Button>
      </div>
    {:else}
      <!-- 页面标题与主操作：等宽眉标加 signal 主按钮。 -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div class="flex flex-col gap-2">
          <p class="font-mono text-xs tracking-[0.18em] text-muted-foreground uppercase">
            <span class="text-signal">//</span> USER MANAGEMENT
          </p>
          <h1 class="text-2xl font-bold tracking-tight sm:text-3xl">用户管理</h1>
          <p class="text-sm text-muted-foreground">
            管理系统用户，包括创建、编辑、禁用与删除操作。
          </p>
        </div>
        <Button
          class="rounded-none bg-signal text-signal-foreground hover:bg-signal/90"
          onclick={openCreateDialog}
        >
          <Plus data-icon="inline-start" />
          创建用户
        </Button>
      </div>

      <!-- 批量操作栏：signal 色浅底提示当前处于批量选择态。 -->
      {#if selectedUserIds.length > 0}
        <div
          class="flex flex-wrap items-center justify-between gap-3 border border-signal/40 bg-signal/5 px-4 py-3"
        >
          <p class="text-sm">
            已选择
            <span class="font-mono font-semibold text-signal tabular-nums">
              {selectedUserIds.length}
            </span>
            个用户
          </p>
          <div class="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              class="rounded-none"
              onclick={() => requestBatchToggle(USER_STATUS.active)}
            >
              批量启用
            </Button>
            <Button
              variant="outline"
              size="sm"
              class="rounded-none"
              onclick={() => requestBatchToggle(USER_STATUS.disabled)}
            >
              批量禁用
            </Button>
            <Button
              variant="destructive"
              size="sm"
              class="rounded-none"
              onclick={requestBatchDelete}
            >
              批量删除
            </Button>
            <Button variant="ghost" size="sm" class="rounded-none" onclick={clearSelection}>
              取消选择
            </Button>
          </div>
        </div>
      {/if}

      <!-- 用户列表：搜索工具条、表格与分页整合于同一张规格卡。 -->
      <Card.Root class="rounded-none border border-border py-0 ring-0">
        <Card.Content class="flex flex-wrap items-center gap-3 border-b border-border px-4 py-3">
          <div class="relative min-w-0 flex-1 sm:max-w-xs">
            <Search class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input.Root
              placeholder="搜索用户名、邮箱或昵称..."
              bind:value={searchQuery}
              class="rounded-none pl-9"
            />
          </div>
          <p class="ml-auto font-mono text-xs text-muted-foreground tabular-nums">
            共 {totalUsers} 个用户
          </p>
        </Card.Content>

        <Card.Content class="px-0">
          {#if isLoading}
            <!-- 加载骨架屏：与表格列结构对齐，避免加载完成后的布局跳变。 -->
            <div class="flex flex-col">
              {#each Array.from({ length: SKELETON_ROW_COUNT }) as _, index (index)}
                <div class="flex items-center gap-4 border-b border-border px-4 py-3.5">
                  <Skeleton.Root class="size-4 shrink-0 rounded-none" />
                  <div class="flex min-w-0 flex-1 items-center gap-3">
                    <Skeleton.Root class="size-8 shrink-0 rounded-none" />
                    <div class="flex flex-col gap-1.5">
                      <Skeleton.Root class="h-3 w-28 rounded-none" />
                      <Skeleton.Root class="h-2.5 w-20 rounded-none" />
                    </div>
                  </div>
                  <Skeleton.Root class="hidden h-3 w-40 rounded-none md:block" />
                  <Skeleton.Root class="h-4 w-14 rounded-none" />
                  <Skeleton.Root class="h-4 w-10 rounded-none" />
                  <Skeleton.Root class="hidden h-3.5 w-24 rounded-none lg:block" />
                  <div class="ml-auto flex gap-2">
                    <Skeleton.Root class="size-7 rounded-none" />
                    <Skeleton.Root class="size-7 rounded-none" />
                  </div>
                </div>
              {/each}
            </div>
          {:else if filteredUsers.length === 0}
            <!-- 空状态：虚线发丝框与图标暗示可填充的数据区。 -->
            <div class="p-6">
              <Empty.Root class="rounded-none border border-dashed py-12">
                <Empty.Media variant="icon" class="rounded-none">
                  <UserIcon />
                </Empty.Media>
                <Empty.Title>没有找到用户</Empty.Title>
                <Empty.Description>
                  {searchQuery
                    ? '请尝试其他搜索条件。'
                    : '系统中还没有用户，创建第一个用户开始使用。'}
                </Empty.Description>
                {#if !searchQuery}
                  <Empty.Content>
                    <Button
                      class="rounded-none bg-signal text-signal-foreground hover:bg-signal/90"
                      onclick={openCreateDialog}
                    >
                      <Plus data-icon="inline-start" />
                      创建用户
                    </Button>
                  </Empty.Content>
                {/if}
              </Empty.Root>
            </div>
          {:else}
            <Table.Root>
              <Table.Header>
                <Table.Row class="hover:bg-transparent">
                  <Table.Head class="w-12">
                    <Checkbox.Root
                      checked={isAllSelected}
                      indeterminate={selectedUserIds.length > 0 && !isAllSelected}
                      onCheckedChange={() => toggleSelectAll()}
                      aria-label="全选当前页用户"
                    />
                  </Table.Head>
                  <Table.Head class="font-mono text-xs tracking-wider uppercase">用户</Table.Head>
                  <Table.Head class="font-mono text-xs tracking-wider uppercase">邮箱</Table.Head>
                  <Table.Head class="font-mono text-xs tracking-wider uppercase">角色</Table.Head>
                  <Table.Head class="font-mono text-xs tracking-wider uppercase">状态</Table.Head>
                  <Table.Head class="font-mono text-xs tracking-wider uppercase">
                    创建时间
                  </Table.Head>
                  <Table.Head class="text-right font-mono text-xs tracking-wider uppercase">
                    操作
                  </Table.Head>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {#each filteredUsers as user (user.id)}
                  {@const isSelected = selectedUserIds.includes(user.id)}
                  {@const roleInfo = getRoleInfo(user.role || 'user')}
                  <Table.Row
                    class="hover:bg-signal/5 data-[state=selected]:bg-signal/10"
                    data-state={isSelected ? 'selected' : undefined}
                  >
                    <Table.Cell>
                      <Checkbox.Root
                        checked={isSelected}
                        onCheckedChange={() => toggleSelectUser(user.id)}
                        aria-label="选择用户 {user.username}"
                      />
                    </Table.Cell>
                    <Table.Cell>
                      <div class="flex items-center gap-3">
                        <Avatar.Root class="size-8 rounded-none [&_svg]:size-4">
                          <Avatar.Image src={user.avatar} alt={user.nickname || user.username} />
                          <Avatar.Fallback>
                            <UserIcon />
                          </Avatar.Fallback>
                        </Avatar.Root>
                        <div class="min-w-0">
                          <p class="truncate font-medium">{user.nickname || user.username}</p>
                          <p class="truncate font-mono text-xs text-muted-foreground">
                            @{user.username}
                          </p>
                        </div>
                      </div>
                    </Table.Cell>
                    <Table.Cell>
                      <span class="font-mono text-xs">{user.email}</span>
                    </Table.Cell>
                    <Table.Cell>
                      <Badge variant={roleInfo.color} class="rounded-none">
                        {roleInfo.name}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell>
                      <!-- 状态开关：状态点与文字共同指示，点击切换前先确认。 -->
                      <button
                        type="button"
                        class="inline-flex cursor-pointer items-center gap-1.5 border px-2 py-0.5 text-xs font-medium transition-colors {user.status ===
                        USER_STATUS.active
                          ? 'border-signal/40 bg-signal/5 text-signal hover:bg-signal/10'
                          : 'border-border bg-muted text-muted-foreground hover:bg-muted/70'}"
                        aria-label="切换用户 {user.username} 的状态"
                        onclick={() => requestToggleStatus(user)}
                      >
                        <span class="size-1.5 rounded-full bg-current"></span>
                        {user.status === USER_STATUS.active ? '正常' : '禁用'}
                      </button>
                    </Table.Cell>
                    <Table.Cell>
                      <span class="font-mono text-xs text-muted-foreground tabular-nums">
                        {new Date(user.createdAt).toLocaleDateString('zh-CN')}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <div class="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          class="rounded-none"
                          aria-label="编辑用户 {user.username}"
                          onclick={() => openEditModal(user)}
                        >
                          <Edit />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          class="rounded-none text-destructive hover:bg-destructive/10 hover:text-destructive"
                          aria-label="删除用户 {user.username}"
                          onclick={() => requestDeleteUser(user)}
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

        <!-- 分页：等宽字体记录页码，与表格共用发丝线分隔。 -->
        {#if totalPages > 1}
          <Card.Content
            class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-4 py-3"
          >
            <p class="font-mono text-xs text-muted-foreground tabular-nums">
              第 {currentPage} / {totalPages} 页 · 共 {totalUsers} 条记录
            </p>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                class="rounded-none"
                disabled={currentPage <= 1}
                onclick={goToPreviousPage}
              >
                <ChevronLeft data-icon="inline-start" />
                上一页
              </Button>
              <Button
                variant="outline"
                size="sm"
                class="rounded-none"
                disabled={currentPage >= totalPages}
                onclick={goToNextPage}
              >
                下一页
                <ChevronRight data-icon="inline-end" />
              </Button>
            </div>
          </Card.Content>
        {/if}
      </Card.Root>

      <!-- 创建用户模态框 -->
      <Dialog.Root bind:open={isCreateModalOpen}>
        <Dialog.Content class="rounded-none border border-border ring-0 sm:max-w-md">
          <Dialog.Header>
            <Dialog.Title>创建新用户</Dialog.Title>
            <Dialog.Description>填写以下信息来创建新的系统用户。</Dialog.Description>
          </Dialog.Header>

          <form
            class="flex flex-col gap-4"
            onsubmit={e => {
              e.preventDefault()
              void createUser()
            }}
          >
            {@render userFormFields('create', false)}
            <Dialog.Footer>
              <Button
                type="button"
                variant="outline"
                class="rounded-none"
                disabled={isSubmitting}
                onclick={() => (isCreateModalOpen = false)}
              >
                取消
              </Button>
              <Button type="submit" class="rounded-none" disabled={isSubmitting}>
                {#if isSubmitting}
                  <Spinner data-icon="inline-start" />
                {/if}
                {isSubmitting ? '创建中...' : '创建用户'}
              </Button>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog.Root>

      <!-- 编辑用户模态框 -->
      <Dialog.Root bind:open={isEditModalOpen}>
        <Dialog.Content class="rounded-none border border-border ring-0 sm:max-w-md">
          <Dialog.Header>
            <Dialog.Title>编辑用户</Dialog.Title>
            <Dialog.Description>修改用户信息。</Dialog.Description>
          </Dialog.Header>

          <form
            class="flex flex-col gap-4"
            onsubmit={e => {
              e.preventDefault()
              void updateUser()
            }}
          >
            {@render userFormFields('edit', true)}
            <Dialog.Footer>
              <Button
                type="button"
                variant="outline"
                class="rounded-none"
                disabled={isSubmitting}
                onclick={() => (isEditModalOpen = false)}
              >
                取消
              </Button>
              <Button type="submit" class="rounded-none" disabled={isSubmitting}>
                {#if isSubmitting}
                  <Spinner data-icon="inline-start" />
                {/if}
                {isSubmitting ? '更新中...' : '更新用户'}
              </Button>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog.Root>

      <!-- 危险操作确认对话框：统一承载单个与批量操作的确认流程。 -->
      <ConfirmDialog
        bind:open={isConfirmOpen}
        title={confirmState?.title ?? ''}
        description={confirmState?.description ?? ''}
        confirmText={confirmState?.confirmText}
        destructive={confirmState?.destructive}
        onConfirm={confirmState?.run}
      />
    {/if}
  </div>
</div>

<!-- 用户表单字段片段：创建与编辑对话框共用，仅密码行为随场景区分。 -->
{#snippet userFormFields(idPrefix: string, isEdit: boolean)}
  <Field.Group>
    <Field.Field data-invalid={formErrors.username ? true : undefined}>
      <Field.FieldLabel for="{idPrefix}-username">
        用户名<span class="text-destructive" aria-hidden="true"> *</span>
      </Field.FieldLabel>
      <Input.Root
        id="{idPrefix}-username"
        bind:value={userForm.username}
        placeholder="请输入用户名"
        disabled={isSubmitting}
        aria-invalid={formErrors.username ? true : undefined}
      />
      {#if formErrors.username}
        <Field.Error>{formErrors.username}</Field.Error>
      {/if}
    </Field.Field>

    <Field.Field data-invalid={formErrors.email ? true : undefined}>
      <Field.FieldLabel for="{idPrefix}-email">
        邮箱<span class="text-destructive" aria-hidden="true"> *</span>
      </Field.FieldLabel>
      <Input.Root
        id="{idPrefix}-email"
        type="email"
        bind:value={userForm.email}
        placeholder="请输入邮箱地址"
        disabled={isSubmitting}
        aria-invalid={formErrors.email ? true : undefined}
      />
      {#if formErrors.email}
        <Field.Error>{formErrors.email}</Field.Error>
      {/if}
    </Field.Field>

    <Field.Field data-invalid={formErrors.password ? true : undefined}>
      <Field.FieldLabel for="{idPrefix}-password">
        密码<span class="text-destructive" aria-hidden="true"> *</span>
      </Field.FieldLabel>
      <Input.Root
        id="{idPrefix}-password"
        type="password"
        bind:value={userForm.password}
        placeholder={isEdit ? '留空则不修改密码' : '请输入密码'}
        disabled={isSubmitting}
        aria-invalid={formErrors.password ? true : undefined}
      />
      {#if isEdit}
        <Field.Description>留空则不修改密码。</Field.Description>
      {/if}
      {#if formErrors.password}
        <Field.Error>{formErrors.password}</Field.Error>
      {/if}
    </Field.Field>

    <Field.Field>
      <Field.FieldLabel for="{idPrefix}-nickname">昵称</Field.FieldLabel>
      <Input.Root
        id="{idPrefix}-nickname"
        bind:value={userForm.nickname}
        placeholder="请输入昵称"
        disabled={isSubmitting}
      />
    </Field.Field>

    <Field.Field>
      <Field.FieldLabel for="{idPrefix}-role">角色</Field.FieldLabel>
      <Select.Root type="single" bind:value={userForm.role} disabled={isSubmitting}>
        <Select.Trigger id="{idPrefix}-role" class="w-full rounded-none">
          {currentRoleLabel}
        </Select.Trigger>
        <Select.Content class="rounded-none">
          <Select.Group>
            {#each roleOptions as option (option.role)}
              <Select.Item value={option.role} class="rounded-none">
                {option.name}
              </Select.Item>
            {/each}
          </Select.Group>
        </Select.Content>
      </Select.Root>
    </Field.Field>

    <Field.Field>
      <Field.FieldLabel for="{idPrefix}-birthday">生日</Field.FieldLabel>
      <Input.Root
        id="{idPrefix}-birthday"
        type="date"
        bind:value={userForm.birthday}
        disabled={isSubmitting}
      />
    </Field.Field>
  </Field.Group>
{/snippet}
