<script lang="ts">
import {
  FileText,
  Home,
  LayoutDashboard,
  LogOut,
  Settings,
  Shield,
  User as UserIcon,
  Users,
  FolderTree,
  Tags,
  MessageSquare,
  Image as ImageIcon,
  Link as LinkIcon,
  Bell,
  BarChart3,
  type LucideIcon
} from '@lucide/svelte'
import { Avatar, Badge, Button, Separator, Sidebar } from '$ui'
import { getRoleInfo } from '$lib/utils/permissions'
import { performLogout } from '$lib/utils/logout'
import type { User, UserRole } from '$lib/types'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { NotificationAPI } from '$lib/api'
import { onMount } from 'svelte'

// 侧边栏导航分组配置
interface NavGroup {
  label: string
  items: Array<{
    id: string
    title: string
    icon: LucideIcon
    url: string
    roles: UserRole[]
  }>
}

const navigationGroups: NavGroup[] = [
  {
    label: '内容管理',
    items: [
      {
        id: 'dashboard',
        title: '仪表盘',
        icon: LayoutDashboard,
        url: '/',
        roles: ['user', 'editor', 'admin', 'superadmin']
      },
      {
        id: 'articles',
        title: '文章管理',
        icon: FileText,
        url: '/posts',
        roles: ['editor', 'admin', 'superadmin']
      },
      {
        id: 'categories',
        title: '分类管理',
        icon: FolderTree,
        url: '/categories',
        roles: ['admin', 'superadmin']
      },
      { id: 'tags', title: '标签管理', icon: Tags, url: '/tags', roles: ['admin', 'superadmin'] },
      {
        id: 'comments',
        title: '评论管理',
        icon: MessageSquare,
        url: '/comments',
        roles: ['admin', 'superadmin']
      }
    ]
  },
  {
    label: '运营与系统',
    items: [
      {
        id: 'media',
        title: '媒体管理',
        icon: ImageIcon,
        url: '/media',
        roles: ['editor', 'admin', 'superadmin']
      },
      { id: 'links', title: '友情链接', icon: LinkIcon, url: '/links', roles: ['superadmin'] },
      {
        id: 'stats',
        title: '站点统计',
        icon: BarChart3,
        url: '/stats',
        roles: ['admin', 'superadmin']
      },
      {
        id: 'users',
        title: '用户管理',
        icon: Users,
        url: '/users',
        roles: ['admin', 'superadmin']
      },
      { id: 'settings', title: '系统设置', icon: Settings, url: '/settings', roles: ['superadmin'] }
    ]
  }
]

// 当前用户信息与通知未读数
let currentUser = $state<User | null>(null)
let userRole = $state<UserRole>('user')
let unreadCount = $state(0)

/**
 * 加载当前用户与通知未读数。
 */
async function loadUserState() {
  const state = authStore.getCurrentState()
  currentUser = state.user
  userRole = state.user?.role ?? 'user'

  const response = await NotificationAPI.getUnreadCount().catch(() => null)
  unreadCount = response?.data?.unreadCount ?? 0
}

/**
 * 过滤当前角色可见的导航项。
 */
const visibleGroups = $derived(
  navigationGroups
    .map(group => ({
      ...group,
      items: group.items.filter(item => item.roles.includes(userRole))
    }))
    .filter(group => group.items.length > 0)
)

async function handleLogout() {
  await performLogout({
    showToast: true,
    redirectTo: '/login'
  })
}

const roleInfo = $derived(getRoleInfo(userRole))

onMount(loadUserState)
</script>

<Sidebar.Root class="border-r">
  <Sidebar.Header class="flex h-16 items-center border-b pl-4">
    <div class="flex w-full flex-1 items-center gap-3">
      <div
        class="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
      >
        <Shield class="size-4" />
      </div>
      <div class="flex flex-col">
        <span class="text-sm font-semibold">MyBlog 管理后台</span>
        <span class="text-xs text-muted-foreground">管理控制面板</span>
      </div>
    </div>
  </Sidebar.Header>

  <Sidebar.Content class="p-2">
    <!-- 用户信息卡片 -->
    <div class="mb-4 rounded-lg border bg-card p-3">
      <div class="flex items-center gap-3">
        <Avatar.Root class="size-8">
          <Avatar.Image
            src={currentUser?.avatar}
            alt={currentUser?.nickname || currentUser?.username}
          />
          <Avatar.Fallback>
            <UserIcon class="size-4" />
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

    <!-- 快捷入口 -->
    <div class="mb-4">
      <p class="mb-2 px-2 text-xs font-medium text-muted-foreground">快捷入口</p>
      <div class="space-y-1">
        <Button
          variant="ghost"
          size="sm"
          class="h-8 w-full justify-start"
          onclick={() => goto('/notifications')}
        >
          <Bell class="mr-2 size-3.5" />
          通知中心
          {#if unreadCount > 0}
            <Badge variant="destructive" class="ml-auto px-1.5 text-xs">{unreadCount}</Badge>
          {/if}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="h-8 w-full justify-start"
          onclick={() => goto('/')}
        >
          <Home class="mr-2 size-3.5" />
          回到仪表盘
        </Button>
      </div>
    </div>

    <Separator.Root class="my-2" />

    <!-- 分组导航 -->
    <div class="space-y-4">
      {#each visibleGroups as group (group.label)}
        <div class="space-y-1">
          <p class="mb-2 px-2 text-xs font-medium text-muted-foreground">{group.label}</p>
          {#each group.items as item (item.id)}
            {@const IconComponent = item.icon}
            <Sidebar.MenuItem class="list-none">
              <Button
                variant="ghost"
                class="h-9 w-full justify-start"
                onclick={() => goto(item.url)}
              >
                <IconComponent class="mr-3 size-4" />
                {item.title}
              </Button>
            </Sidebar.MenuItem>
          {/each}
        </div>
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
      <LogOut class="mr-2 size-4" />
      退出登录
    </Button>
  </Sidebar.Footer>
</Sidebar.Root>
