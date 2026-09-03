<script lang="ts">
import {
  NOTIFICATION_PAGE_SIZE,
  NOTIFICATION_TYPE_LABELS,
  NOTIFICATION_TYPE_OPTIONS
} from '$lib/constants/notification'
import type { Notification, NotificationType } from '@myblog/api/modules/notification/types'
import PageHeader from '$lib/components/admin/page-header.svelte'
import Pagination from '$lib/components/admin/pagination.svelte'
import { Bell, CheckCheck, Inbox } from '@lucide/svelte'
import { Badge, Button, Card, ToggleGroup } from '$ui'
import { NotificationAPI } from '$lib/api'
import { onMount } from 'svelte'

let notifications = $state<Notification[]>([])
let isLoading = $state(true)
let total = $state(0)
let unreadCount = $state(0)
let currentPage = $state(1)
let typeFilter = $state<NotificationType | ''>('')

/**
 * 加载通知列表并同步未读数。
 */
async function loadNotifications() {
  isLoading = true
  try {
    const response = await NotificationAPI.list({
      page: currentPage,
      pageSize: NOTIFICATION_PAGE_SIZE,
      ...(typeFilter ? { type: typeFilter } : {})
    })
    if (response.code === 200 && response.data) {
      notifications = response.data.notifications ?? []
      total = response.data.total ?? 0
      unreadCount = response.data.unreadCount ?? 0
    } else {
      toast.error(response.message || '加载通知失败')
    }
  } catch (error) {
    console.error('加载通知失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 标记单条通知为已读。
 */
async function handleMarkRead(notification: Notification) {
  if (notification.isRead) return
  try {
    const response = await NotificationAPI.markRead(notification.id)
    if (response.code === 200) {
      notification.isRead = true
      unreadCount = Math.max(0, unreadCount - 1)
    } else {
      toast.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('标记已读失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

/**
 * 标记全部通知为已读。
 */
async function handleMarkAllRead() {
  try {
    const response = await NotificationAPI.markAllRead()
    if (response.code === 200) {
      unreadCount = 0
      notifications = notifications.map(item => ({ ...item, isRead: true }))
      toast.success('全部通知已标记为已读')
    } else {
      toast.error(response.message || '操作失败')
    }
  } catch (error) {
    console.error('标记全部已读失败:', error)
    toast.error('网络错误，请稍后重试')
  }
}

function handlePageChange(page: number) {
  currentPage = page
  loadNotifications()
}

onMount(loadNotifications)
</script>

<svelte:head>
  <title>通知中心 - MyBlog</title>
</svelte:head>

<PageHeader title="通知中心" description="查看评论回复、点赞与系统消息" crumb="通知中心">
  {#snippet actions()}
    <Button variant="outline" onclick={handleMarkAllRead}>
      <CheckCheck data-icon="inline-start" />
      全部已读
    </Button>
  {/snippet}

  <Card.Root>
    <Card.Content class="p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <ToggleGroup.Root type="single" bind:value={typeFilter} variant="outline" size="sm">
          {#each NOTIFICATION_TYPE_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>
        <span class="text-sm text-muted-foreground">
          <Bell class="mr-1 inline size-3.5" />
          {unreadCount} 条未读
        </span>
      </div>
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
      {:else if notifications.length === 0}
        <div class="flex h-48 items-center justify-center">
          <div class="text-center">
            <Inbox class="mx-auto size-12 text-muted-foreground" />
            <h3 class="mt-4 text-lg font-medium">暂无通知</h3>
          </div>
        </div>
      {:else}
        <div class="divide-y">
          {#each notifications as notification (notification.id)}
            <button
              type="button"
              onclick={() => handleMarkRead(notification)}
              class="flex w-full items-start gap-4 px-6 py-4 text-left transition-colors hover:bg-accent/50 {notification.isRead
                ? ''
                : 'bg-accent/30'}"
            >
              <span
                class="mt-1.5 size-2 shrink-0 rounded-full {notification.isRead
                  ? 'bg-transparent'
                  : 'bg-primary'}"
              ></span>
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium">{notification.title}</span>
                  <Badge variant="outline">
                    {NOTIFICATION_TYPE_LABELS[notification.type] ?? notification.type}
                  </Badge>
                </div>
                {#if notification.content}
                  <p class="text-sm text-muted-foreground">{notification.content}</p>
                {/if}
                <p class="text-xs text-muted-foreground">
                  {new Date(notification.createdAt).toLocaleString('zh-CN')}
                </p>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <Pagination
    page={currentPage}
    {total}
    pageSize={NOTIFICATION_PAGE_SIZE}
    onPageChange={handlePageChange}
  />
</PageHeader>
