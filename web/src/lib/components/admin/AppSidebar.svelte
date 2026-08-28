<script lang="ts">
import { Button, Badge, Sidebar, Avatar, Separator } from '$ui'
import { authStore } from '$lib/stores/auth'
import { getRoleInfo } from '$lib/utils/permissions'
import { performLogout } from '$lib/utils/logout'
import type { User, UserRole, SidebarMenuItem } from '$lib/types'
import {
  LayoutDashboard,
  Users,
  FileText,
  Settings,
  LogOut,
  User as UserIcon,
  Shield,
  Home
} from '@lucide/svelte'

// 基础导航菜单配置
const baseNavigation: SidebarMenuItem[] = [
  {
    id: 'dashboard',
    title: '仪表盘',
    icon: LayoutDashboard,
    url: '/manage',
    roles: ['user', 'editor', 'admin', 'superadmin']
  },
  {
    id: 'articles',
    title: '文章管理',
    icon: FileText,
    url: '/manage/posts',
    roles: ['editor', 'admin', 'superadmin']
  }
]

// 管理员导航菜单
const adminNavigation: SidebarMenuItem[] = [
  {
    id: 'users',
    title: '用户管理',
    icon: Users,
    url: '/manage/users',
    roles: ['admin', 'superadmin']
  },
  {
    id: 'settings',
    title: '系统设置',
    icon: Settings,
    url: '/manage/settings',
    roles: ['superadmin']
  }
]

// 获取当前用户信息
let currentUser = $state<User | null>(null)
let userRole = $state<UserRole>('user')

$effect(() => {
  authStore.subscribe(state => {
    if (state.isAuthenticated && state.user) {
      currentUser = state.user
      userRole = state.user.role || 'user'
    }
  })
})

// 退出登录
async function handleLogout() {
  await performLogout({
    showToast: true,
    redirectTo: '/login'
  })
}

// 导入的 getRoleInfo 函数已经提供了角色信息

// 根据用户角色获取角色信息和过滤导航菜单
let roleInfo = $derived(getRoleInfo(userRole))
let filteredNavigation = $derived(
  [...baseNavigation, ...adminNavigation].filter(item => item.roles.includes(userRole))
)
</script>

<Sidebar.Root class="border-r">
  <Sidebar.Header class="p-4">
    <div class="flex items-center gap-3">
      <div
        class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
      >
        <Shield class="h-4 w-4" />
      </div>
      <div class="flex flex-col">
        <span class="text-sm font-semibold">MyBlog 管理后台</span>
        <span class="text-xs text-muted-foreground">管理控制面板</span>
      </div>
    </div>
  </Sidebar.Header>

  <Separator.Root />

  <Sidebar.Content class="p-2">
    <!-- 用户信息卡片 -->
    <div class="mb-4 rounded-lg border bg-card p-3">
      <div class="flex items-center gap-3">
        <Avatar.Root class="h-8 w-8">
          <Avatar.Image
            src={currentUser?.avatar}
            alt={currentUser?.nickname || currentUser?.username}
          />
          <Avatar.Fallback>
            <UserIcon class="h-4 w-4" />
          </Avatar.Fallback>
        </Avatar.Root>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium">
            {currentUser?.nickname || currentUser?.username || '未知用户'}
          </p>
          <div class="mt-1 flex items-center gap-1">
            <Badge variant={roleInfo.color} class="px-1.5 py-0.5 text-xs">
              {roleInfo.name}
            </Badge>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="mb-4">
      <p class="mb-2 px-2 text-xs font-medium text-muted-foreground">快速操作</p>
      <div class="space-y-1">
        <Button
          variant="ghost"
          size="sm"
          class="h-8 w-full justify-start"
          onclick={() => goto('/')}
        >
          <Home class="mr-2 h-3 w-3" />
          回到首页
        </Button>
      </div>
    </div>

    <Separator.Root class="my-2" />

    <!-- 主导航菜单 -->
    <div class="space-y-1">
      <p class="mb-2 px-2 text-xs font-medium text-muted-foreground">管理功能</p>
      {#each filteredNavigation as item, index (index)}
        {@const IconComponent = item.icon}
        <Sidebar.MenuItem class="list-none">
          <Button variant="ghost" class="h-9 w-full justify-start" onclick={() => goto(item.url)}>
            <IconComponent class="mr-3 h-4 w-4" />
            {item.title}
          </Button>
        </Sidebar.MenuItem>
      {/each}
    </div>
  </Sidebar.Content>

  <Separator.Root />

  <Sidebar.Footer class="p-4">
    <Button
      variant="ghost"
      size="sm"
      class="w-full justify-start text-destructive hover:bg-destructive/10 hover:text-destructive"
      onclick={handleLogout}
    >
      <LogOut class="mr-2 h-4 w-4" />
      退出登录
    </Button>
  </Sidebar.Footer>
</Sidebar.Root>
