<script lang="ts">
import { onMount } from 'svelte'
import { authStore } from '$lib/stores/auth'
import {
  Sidebar,
  Breadcrumb,
  Select,
  Card,
  Button,
  Input,
  Badge,
  Avatar,
  Dialog,
  Table,
  Label
} from '$ui'
import {
  Plus,
  Search,
  Edit,
  Trash2,
  User as UserIcon,
  Mail,
  Calendar,
  Shield
} from '@lucide/svelte'
import { UserAPI } from '$lib/api'
import type { User } from '$lib/api/modules/user/types'

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

// 表单状态
let userForm = $state({
  username: '',
  email: '',
  password: '',
  nickname: '',
  role: 'user',
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

// 加载用户列表
async function loadUsers() {
  try {
    isLoading = true
    const response = await UserAPI.getUserList(currentPage, 10)

    if (response.code === 200 && response.data) {
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

    if (response.code === 200) {
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
    const updateData: any = {
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

    if (response.code === 200) {
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

// 切换用户状态
async function toggleUserStatus(user: User) {
  const newStatus = user.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'

  if (!confirm(`确定要${action}用户 "${user.username}" 吗？`)) {
    return
  }

  try {
    const response = await UserAPI.updateUser({
      id: user.id,
      username: user.username,
      email: user.email,
      nickname: user.nickname || '',
      role: user.role || 'user',
      birthday: user.birthday || '',
      status: newStatus
    })

    if (response.code === 200) {
      toast.success(`用户${action}成功`)
      await loadUsers()
    } else {
      toast.error(response.message || `${action}用户失败`)
    }
  } catch (error) {
    console.error(`Toggle user status error:`, error)
    toast.error('网络错误，请稍后重试')
  }
}

// 删除用户
async function deleteUser(userId: number) {
  if (!confirm('确定要删除此用户吗？此操作不可恢复。')) {
    return
  }

  try {
    const response = await UserAPI.deleteUser(userId)

    if (response.code === 200) {
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

async function batchDeleteUsers() {
  if (selectedUserIds.length === 0) {
    toast.error('请选择要删除的用户')
    return
  }

  if (!confirm(`确定要删除选中的 ${selectedUserIds.length} 个用户吗？此操作不可恢复。`)) {
    return
  }

  let successCount = 0
  let failCount = 0

  for (const userId of selectedUserIds) {
    try {
      const response = await UserAPI.deleteUser(userId)
      if (response.code === 200) {
        successCount++
      } else {
        failCount++
      }
    } catch (error) {
      failCount++
    }
  }

  if (successCount > 0) {
    toast.success(`成功删除 ${successCount} 个用户`)
  }
  if (failCount > 0) {
    toast.error(`删除失败 ${failCount} 个用户`)
  }

  selectedUserIds = []
  isAllSelected = false
  await loadUsers()
}

async function batchToggleStatus(status: number) {
  if (selectedUserIds.length === 0) {
    toast.error('请选择要操作的用户')
    return
  }

  const action = status === 1 ? '启用' : '禁用'
  if (!confirm(`确定要${action}选中的 ${selectedUserIds.length} 个用户吗？`)) {
    return
  }

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
          status: status
        })
        if (response.code === 200) {
          successCount++
        } else {
          failCount++
        }
      }
    } catch (error) {
      failCount++
    }
  }

  if (successCount > 0) {
    toast.success(`成功${action} ${successCount} 个用户`)
  }
  if (failCount > 0) {
    toast.error(`${action}失败 ${failCount} 个用户`)
  }

  selectedUserIds = []
  isAllSelected = false
  await loadUsers()
}

// 获取角色信息
function getRoleInfo(role: string) {
  switch (role) {
    case 'superadmin':
      return { name: '超级管理员', variant: 'destructive' }
    case 'admin':
      return { name: '管理员', variant: 'default' }
    case 'editor':
      return { name: '编辑者', variant: 'secondary' }
    default:
      return { name: '用户', variant: 'outline' }
  }
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

// 组件挂载时检查权限和加载数据
onMount(() => {
  // 检查用户权限
  authStore.subscribe(state => {
    if (state.isAuthenticated && state.user) {
      userRole = state.user.role || 'user'
      hasPermission = ['admin', 'superadmin'].includes(userRole)

      if (hasPermission) {
        loadUsers()
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
<header class="flex h-16 shrink-0 items-center gap-2 border-b px-6">
  <Sidebar.Trigger />
  <Breadcrumb.Root>
    <Breadcrumb.List>
      <Breadcrumb.Item>
        <Breadcrumb.Link href="/manage">管理后台</Breadcrumb.Link>
      </Breadcrumb.Item>
      <Breadcrumb.Separator />
      <Breadcrumb.Item>
        <Breadcrumb.Page>用户管理</Breadcrumb.Page>
      </Breadcrumb.Item>
    </Breadcrumb.List>
  </Breadcrumb.Root>
</header>

<!-- 主内容区域 -->
<main class="flex-1 space-y-6 p-6">
  {#if !hasPermission}
    <div class="flex h-96 items-center justify-center">
      <div class="text-center">
        <Shield class="mx-auto h-16 w-16 text-muted-foreground" />
        <h2 class="mt-4 text-xl font-semibold">权限不足</h2>
        <p class="mt-2 text-sm text-muted-foreground">只有管理员和超级管理员才能访问用户管理功能</p>
        <Button class="mt-4" onclick={() => goto('/manage')}>返回仪表盘</Button>
      </div>
    </div>
  {:else}
    <!-- 页面标题和操作 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">用户管理</h1>
        <p class="text-muted-foreground">管理系统用户，包括创建、编辑和删除操作</p>
      </div>

      <Dialog.Root bind:open={isCreateModalOpen}>
        <Dialog.Trigger>
          <Button
            onclick={() => {
              resetForm()
              isCreateModalOpen = true
            }}
          >
            <Plus class="mr-2 h-4 w-4" />
            创建用户
          </Button>
        </Dialog.Trigger>
        <Dialog.Content class="sm:max-w-md">
          <Dialog.Header>
            <Dialog.Title>创建新用户</Dialog.Title>
            <Dialog.Description>填写以下信息来创建新的系统用户</Dialog.Description>
          </Dialog.Header>

          <div class="space-y-4">
            <div class="space-y-2">
              <Label.Root for="username">用户名 *</Label.Root>
              <Input.Root
                id="username"
                bind:value={userForm.username}
                placeholder="请输入用户名"
                disabled={isSubmitting}
              />
              {#if formErrors.username}
                <p class="text-sm text-destructive">{formErrors.username}</p>
              {/if}
            </div>

            <div class="space-y-2">
              <Label.Root for="email">邮箱 *</Label.Root>
              <Input.Root
                id="email"
                type="email"
                bind:value={userForm.email}
                placeholder="请输入邮箱地址"
                disabled={isSubmitting}
              />
              {#if formErrors.email}
                <p class="text-sm text-destructive">{formErrors.email}</p>
              {/if}
            </div>

            <div class="space-y-2">
              <Label.Root for="password">密码 *</Label.Root>
              <Input.Root
                id="password"
                type="password"
                bind:value={userForm.password}
                placeholder="请输入密码"
                disabled={isSubmitting}
              />
              {#if formErrors.password}
                <p class="text-sm text-destructive">{formErrors.password}</p>
              {/if}
            </div>

            <div class="space-y-2">
              <Label.Root for="nickname">昵称</Label.Root>
              <Input.Root
                id="nickname"
                bind:value={userForm.nickname}
                placeholder="请输入昵称"
                disabled={isSubmitting}
              />
            </div>

            <div class="space-y-2">
              <Select.Root
                bind:value={userForm.role}
                class="bg-surface flex h-10 w-full rounded-md border border-input px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isSubmitting}
              >
                <Label.Root for="role">角色</Label.Root>
                <Select.Trigger class="w-full">请选择角色</Select.Trigger>
                <Select.Content>
                  <Select.Item value="user">用户</Select.Item>
                  <Select.Item value="editor">编辑者</Select.Item>
                  <Select.Item value="admin">管理员</Select.Item>
                  <Select.Item value="superadmin">超级管理员</Select.Item>
                </Select.Content>
              </Select.Root>
            </div>

            <div class="space-y-2">
              <Label.Root for="birthday">生日</Label.Root>
              <Input.Root
                id="birthday"
                type="date"
                bind:value={userForm.birthday}
                disabled={isSubmitting}
              />
            </div>
          </div>

          <Dialog.Footer>
            <Button variant="outline" onclick={() => (isCreateModalOpen = false)}>取消</Button>
            <Button onclick={createUser} disabled={isSubmitting}>
              {#if isSubmitting}
                <div
                  class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                ></div>
                创建中...
              {:else}
                创建用户
              {/if}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog.Root>
    </div>

    <!-- 搜索和筛选 -->
    <Card.Root>
      <Card.Content class="p-6">
        <div class="flex items-center space-x-4">
          <div class="relative flex-1">
            <Search
              class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input.Root
              placeholder="搜索用户名、邮箱或昵称..."
              bind:value={searchQuery}
              class="pl-9"
            />
          </div>
          <div class="text-sm text-muted-foreground">
            共 {totalUsers} 个用户
          </div>
        </div>
      </Card.Content>
    </Card.Root>

    <!-- 批量操作栏 -->
    {#if selectedUserIds.length > 0}
      <Card.Root>
        <Card.Content class="p-4">
          <div class="flex items-center justify-between">
            <div class="text-sm text-muted-foreground">
              已选择 {selectedUserIds.length} 个用户
            </div>
            <div class="flex items-center space-x-2">
              <Button variant="outline" size="sm" onclick={() => batchToggleStatus(1)}>
                批量启用
              </Button>
              <Button variant="outline" size="sm" onclick={() => batchToggleStatus(0)}>
                批量禁用
              </Button>
              <Button variant="destructive" size="sm" onclick={batchDeleteUsers}>批量删除</Button>
              <Button
                variant="ghost"
                size="sm"
                onclick={() => {
                  selectedUserIds = []
                  isAllSelected = false
                }}
              >
                取消选择
              </Button>
            </div>
          </div>
        </Card.Content>
      </Card.Root>
    {/if}

    <!-- 用户列表 -->
    <Card.Root>
      <Card.Content class="p-0">
        {#if isLoading}
          <div class="flex h-48 items-center justify-center">
            <div
              class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
            ></div>
          </div>
        {:else if filteredUsers.length === 0}
          <div class="flex h-48 items-center justify-center">
            <div class="text-center">
              <UserIcon class="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 class="mt-4 text-lg font-medium">没有找到用户</h3>
              <p class="text-sm text-muted-foreground">
                {searchQuery ? '请尝试其他搜索条件' : '开始创建第一个用户'}
              </p>
            </div>
          </div>
        {:else}
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.Head class="w-12">
                  <input
                    type="checkbox"
                    checked={isAllSelected}
                    onchange={toggleSelectAll}
                    class="rounded border-gray-300"
                  />
                </Table.Head>
                <Table.Head>用户</Table.Head>
                <Table.Head>邮箱</Table.Head>
                <Table.Head>角色</Table.Head>
                <Table.Head>状态</Table.Head>
                <Table.Head>创建时间</Table.Head>
                <Table.Head class="text-right">操作</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {#each filteredUsers as user (user.id)}
                <Table.Row>
                  <Table.Cell>
                    <input
                      type="checkbox"
                      checked={selectedUserIds.includes(user.id)}
                      onchange={() => toggleSelectUser(user.id)}
                      class="rounded border-gray-300"
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <div class="flex items-center space-x-3">
                      <Avatar.Root class="h-8 w-8">
                        <Avatar.Image src={user.avatar} alt={user.nickname || user.username} />
                        <Avatar.Fallback>
                          <UserIcon class="h-4 w-4" />
                        </Avatar.Fallback>
                      </Avatar.Root>
                      <div>
                        <p class="font-medium">{user.nickname || user.username}</p>
                        <p class="text-sm text-muted-foreground">@{user.username}</p>
                      </div>
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    <div class="flex items-center space-x-2">
                      <Mail class="h-4 w-4 text-muted-foreground" />
                      <span>{user.email}</span>
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    {@const roleInfo = getRoleInfo(user.role || 'user')}
                    <Badge variant={roleInfo.variant}>
                      {roleInfo.name}
                    </Badge>
                  </Table.Cell>
                  <Table.Cell>
                    <button
                      onclick={() => toggleUserStatus(user)}
                      class="inline-flex cursor-pointer items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {user.status ===
                      1
                        ? 'bg-green-100 text-green-800 hover:bg-green-200'
                        : 'bg-gray-100 text-gray-800 hover:bg-gray-200'}"
                    >
                      {user.status === 1 ? '正常' : '禁用'}
                    </button>
                  </Table.Cell>
                  <Table.Cell>
                    <div class="flex items-center space-x-2">
                      <Calendar class="h-4 w-4 text-muted-foreground" />
                      <span class="text-sm">
                        {new Date(user.createdAt).toLocaleDateString('zh-CN')}
                      </span>
                    </div>
                  </Table.Cell>
                  <Table.Cell class="text-right">
                    <div class="flex items-center justify-end space-x-2">
                      <Button variant="ghost" size="sm" onclick={() => openEditModal(user)}>
                        <Edit class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onclick={() => deleteUser(user.id)}
                        class="text-destructive hover:text-destructive"
                      >
                        <Trash2 class="h-4 w-4" />
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

    <!-- 分页 -->
    {#if totalPages > 1}
      <div class="flex items-center justify-center space-x-2">
        <Button
          variant="outline"
          size="sm"
          disabled={currentPage <= 1}
          onclick={() => {
            currentPage = Math.max(1, currentPage - 1)
            loadUsers()
          }}
        >
          上一页
        </Button>

        <span class="text-sm text-muted-foreground">
          第 {currentPage} 页，共 {totalPages} 页
        </span>

        <Button
          variant="outline"
          size="sm"
          disabled={currentPage >= totalPages}
          onclick={() => {
            currentPage = Math.min(totalPages, currentPage + 1)
            loadUsers()
          }}
        >
          下一页
        </Button>
      </div>
    {/if}

    <!-- 编辑用户模态框 -->
    <Dialog.Root bind:open={isEditModalOpen}>
      <Dialog.Content class="sm:max-w-md">
        <Dialog.Header>
          <Dialog.Title>编辑用户</Dialog.Title>
          <Dialog.Description>修改用户信息</Dialog.Description>
        </Dialog.Header>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label.Root for="edit-username">用户名 *</Label.Root>
            <Input.Root
              id="edit-username"
              bind:value={userForm.username}
              placeholder="请输入用户名"
              disabled={isSubmitting}
            />
            {#if formErrors.username}
              <p class="text-sm text-destructive">{formErrors.username}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label.Root for="edit-email">邮箱 *</Label.Root>
            <Input.Root
              id="edit-email"
              type="email"
              bind:value={userForm.email}
              placeholder="请输入邮箱地址"
              disabled={isSubmitting}
            />
            {#if formErrors.email}
              <p class="text-sm text-destructive">{formErrors.email}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label.Root for="edit-password"
              >密码 <span class="text-xs text-muted-foreground">(留空则不修改)</span></Label.Root
            >
            <Input.Root
              id="edit-password"
              type="password"
              bind:value={userForm.password}
              placeholder="留空则不修改密码"
              disabled={isSubmitting}
            />
            {#if formErrors.password}
              <p class="text-sm text-destructive">{formErrors.password}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label.Root for="edit-nickname">昵称</Label.Root>
            <Input.Root
              id="edit-nickname"
              bind:value={userForm.nickname}
              placeholder="请输入昵称"
              disabled={isSubmitting}
            />
          </div>

          <div class="space-y-2">
            <Select.Root
              bind:value={userForm.role}
              class="bg-surface flex h-10 w-full rounded-md border border-input px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              disabled={isSubmitting}
            >
              <Label.Root for="role">角色</Label.Root>
              <Select.Trigger class="w-full">请选择角色</Select.Trigger>
              <Select.Content>
                <Select.Item value="user">用户</Select.Item>
                <Select.Item value="editor">编辑者</Select.Item>
                <Select.Item value="admin">管理员</Select.Item>
                <Select.Item value="superadmin">超级管理员</Select.Item>
              </Select.Content>
            </Select.Root>
            <Label.Root for="edit-role">角色</Label.Root>
          </div>

          <div class="space-y-2">
            <Label.Root for="edit-birthday">生日</Label.Root>
            <Input.Root
              id="edit-birthday"
              type="date"
              bind:value={userForm.birthday}
              disabled={isSubmitting}
            />
          </div>
        </div>

        <Dialog.Footer>
          <Button variant="outline" onclick={() => (isEditModalOpen = false)}>取消</Button>
          <Button onclick={updateUser} disabled={isSubmitting}>
            {#if isSubmitting}
              <div
                class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
              ></div>
              更新中...
            {:else}
              更新用户
            {/if}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog.Root>
  {/if}
</main>
