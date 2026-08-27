<script lang="ts">
import { page } from '$app/state'
import { authStore } from '$lib/stores/auth'
import type { SidebarMenuItem, User, UserRole } from '$lib/types'
import { performLogout } from '$lib/utils/logout'
import { getRoleInfo } from '$lib/utils/permissions'
import { Avatar, Badge, Button, Sidebar } from '$ui'
import {
  FileText,
  Home,
  LayoutDashboard,
  LogOut,
  Settings,
  Shield,
  User as UserIcon,
  Users
} from '@lucide/svelte'

// 导航菜单配置：roles 决定各角色可见的菜单项，顺序即渲染顺序。
const navigation: SidebarMenuItem[] = [
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
  },
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

let roleInfo = $derived(getRoleInfo(userRole))
let filteredNavigation = $derived(navigation.filter(item => item.roles.includes(userRole)))

// 当前激活菜单项 id，先精确匹配路径，再按最长前缀匹配以兼容子路由。
const activeItemId = $derived.by(() => {
  const pathname = page.url.pathname
  const exactMatched = filteredNavigation.find(item => pathname === item.url)
  if (exactMatched) return exactMatched.id
  const prefixMatched = filteredNavigation
    .filter(item => pathname.startsWith(`${item.url}/`))
    .sort((a, b) => b.url.length - a.url.length)
  return prefixMatched[0]?.id
})
</script>

<Sidebar.Root class="border-r">
  <!-- 品牌区：signal 色块加双行标题，文字在图标收起模式下隐藏。 -->
  <Sidebar.Header class="border-b px-3 py-3">
    <a
      href="/manage"
      class="flex h-8 items-center gap-3 outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <div
        class="flex size-8 shrink-0 items-center justify-center bg-signal text-signal-foreground [&_svg]:size-4"
      >
        <Shield />
      </div>
      <div class="flex min-w-0 flex-col group-data-[collapsible=icon]:hidden">
        <span class="truncate text-sm leading-none font-semibold">MyBlog 管理后台</span>
        <span class="mt-1 font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">
          Admin Console
        </span>
      </div>
    </a>
  </Sidebar.Header>

  <Sidebar.Content class="flex flex-col gap-1">
    <!-- 主导航菜单：激活项以 signal 左侧刻线与文字变色标记，对应图纸选中态。 -->
    <Sidebar.Group class="p-2">
      <Sidebar.GroupLabel
        class="h-6 rounded-none px-2 font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase"
      >
        管理
      </Sidebar.GroupLabel>
      <Sidebar.GroupContent>
        <Sidebar.Menu class="gap-1">
          {#each filteredNavigation as item (item.id)}
            {@const IconComponent = item.icon}
            <Sidebar.MenuItem class="list-none">
              <Sidebar.MenuButton
                isActive={activeItemId === item.id}
                tooltipContent={item.title}
                class="rounded-none data-active:bg-signal/10 data-active:text-signal data-active:shadow-[inset_2px_0_0_0_var(--signal)]"
                onclick={() => goto(item.url)}
              >
                <IconComponent />
                <span>{item.title}</span>
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>

    <!-- 站内导航 -->
    <Sidebar.Group class="p-2">
      <Sidebar.GroupLabel
        class="h-6 rounded-none px-2 font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase"
      >
        通用
      </Sidebar.GroupLabel>
      <Sidebar.GroupContent>
        <Sidebar.Menu class="gap-1">
          <Sidebar.MenuItem class="list-none">
            <Sidebar.MenuButton
              tooltipContent="回到首页"
              class="rounded-none"
              onclick={() => goto('/')}
            >
              <Home />
              <span>回到首页</span>
            </Sidebar.MenuButton>
          </Sidebar.MenuItem>
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
  </Sidebar.Content>

  <!-- 底部用户区：头像加角色徽标，承载身份与退出动作。 -->
  <Sidebar.Footer class="border-t p-3">
    <div class="flex items-center gap-3">
      <Avatar.Root class="size-8 shrink-0 rounded-none [&_svg]:size-4">
        <Avatar.Image
          src={currentUser?.avatar}
          alt={currentUser?.nickname || currentUser?.username || '用户头像'}
        />
        <Avatar.Fallback>
          <UserIcon />
        </Avatar.Fallback>
      </Avatar.Root>
      <div class="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
        <p class="truncate text-sm leading-none font-medium">
          {currentUser?.nickname || currentUser?.username || '未知用户'}
        </p>
        <div class="mt-1.5">
          <Badge variant={roleInfo.color} class="rounded-none px-1.5 py-0 text-[10px]">
            {roleInfo.name}
          </Badge>
        </div>
      </div>
    </div>
    <Button
      variant="ghost"
      size="sm"
      class="mt-3 w-full justify-start rounded-none text-destructive hover:bg-destructive/10 hover:text-destructive"
      onclick={() => void handleLogout()}
    >
      <LogOut data-icon="inline-start" />
      退出登录
    </Button>
  </Sidebar.Footer>
</Sidebar.Root>
