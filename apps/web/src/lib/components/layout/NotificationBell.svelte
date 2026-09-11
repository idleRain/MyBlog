<script lang="ts">
import type { Notification } from '@myblog/api/modules/notification/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { formatDate } from '$lib/utils/format-date'
import { authStore } from '$lib/stores/auth'
import { NotificationAPI } from '$lib/api'
import { goto } from '$app/navigation'
import { Bell } from '@lucide/svelte'
import { DropdownMenu } from '$ui'

// 下拉面板展示的最近通知条数。
const RECENT_LIMIT = 5

// 未读数与最近通知，登录后拉取，登出时清空。
let unreadCount = $state(0)
let recent = $state<Notification[]>([])

let isAuthenticated = $state(false)
let wasAuthenticated = false
authStore.subscribe(state => {
  isAuthenticated = state.isAuthenticated
  if (isAuthenticated && !wasAuthenticated) {
    void refreshUnread()
  }
  if (!isAuthenticated) {
    unreadCount = 0
    recent = []
  }
  wasAuthenticated = isAuthenticated
})

// 拉取未读数，失败时保持当前值。
async function refreshUnread() {
  try {
    const response = await NotificationAPI.getUnreadCount()
    if (response.code === RESPONSE_CODE_SUCCESS) {
      unreadCount = response.data?.unreadCount ?? 0
    }
  } catch {
    // 未读数拉取失败保持当前值，下次打开面板时重新获取。
  }
}

// 打开面板时拉取最近通知与最新未读数。
async function loadRecent() {
  try {
    const response = await NotificationAPI.list({ page: 1, pageSize: RECENT_LIMIT })
    if (response.code === RESPONSE_CODE_SUCCESS && response.data) {
      recent = response.data.notifications
      unreadCount = response.data.unreadCount
    }
  } catch {
    // 通知列表拉取失败时保持面板为空状态。
  }
}

// 面板开关时按需加载通知列表。
function handleOpenChange(next: boolean) {
  if (next) {
    void loadRecent()
  }
}

// 标记单条通知为已读并本地同步状态。
async function markRead(id: number) {
  const response = await NotificationAPI.markRead(id)
  if (response.code !== RESPONSE_CODE_SUCCESS) return

  const target = recent.find(item => item.id === id)
  if (target && !target.isRead) {
    target.isRead = true
    unreadCount = Math.max(0, unreadCount - 1)
  }
}

// 全部标记已读并同步本地未读数。
async function markAllRead() {
  const response = await NotificationAPI.markAllRead()
  if (response.code !== RESPONSE_CODE_SUCCESS) return

  recent = recent.map(item => ({ ...item, isRead: true }))
  unreadCount = 0
}

// 点击通知：未读先标记已读，再跳转到关联页面。
function handleNotificationClick(notification: Notification) {
  if (!notification.isRead) {
    void markRead(notification.id)
  }
  if (notification.actionUrl) {
    void goto(notification.actionUrl)
  }
}
</script>

{#if isAuthenticated}
  <DropdownMenu.Root onOpenChange={handleOpenChange}>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <button
          {...props}
          type="button"
          class="relative inline-flex h-9 w-9 items-center justify-center rounded-none text-foreground transition-colors duration-200 hover:text-signal"
          aria-label="通知"
        >
          <Bell class="h-4 w-4" aria-hidden="true" />
          {#if unreadCount > 0}
            <span
              class="absolute top-0.5 right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-signal px-1 text-[10px] leading-none font-bold text-signal-foreground"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          {/if}
        </button>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="w-80">
      <div class="flex items-center justify-between px-3 py-2">
        <span class="text-sm font-bold">通知</span>
        <button
          type="button"
          onclick={markAllRead}
          class="text-xs text-signal underline-offset-4 hover:underline"
        >
          全部已读
        </button>
      </div>

      {#if recent.length === 0}
        <p class="px-3 py-6 text-center text-sm text-muted-foreground">暂无通知</p>
      {:else}
        <div class="max-h-80 overflow-y-auto">
          {#each recent as notification (notification.id)}
            <button
              type="button"
              onclick={() => handleNotificationClick(notification)}
              class="block w-full px-3 py-2.5 text-left transition-colors duration-150 hover:bg-accent"
            >
              <p
                class="flex items-center justify-between gap-2 text-sm {notification.isRead
                  ? 'text-muted-foreground'
                  : 'font-bold text-foreground'}"
              >
                {notification.title}
                {#if !notification.isRead}
                  <span class="h-2 w-2 shrink-0 rounded-full bg-signal" aria-hidden="true"></span>
                {/if}
              </p>
              <p class="mt-1 font-mono text-xs text-muted-foreground">
                {formatDate(notification.createdAt)}
              </p>
            </button>
          {/each}
        </div>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
