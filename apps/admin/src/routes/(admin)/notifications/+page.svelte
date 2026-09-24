<script lang="ts">
import {
  NOTIFICATION_PAGE_SIZE,
  NOTIFICATION_TYPE_LABELS,
  NOTIFICATION_TYPE_OPTIONS
} from '$lib/constants/notification'
import type { Notification, NotificationType } from '@myblog/api/modules/notification/types'
import { Bell, CheckCheck, Inbox, RotateCcw } from '@lucide/svelte'
import { Badge, Button, Card, Pagination, ToggleGroup } from '$ui'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { NotificationAPI } from '$lib/api'
import { onMount } from 'svelte'

let notifications = $state<Notification[]>([])
let isLoading = $state(true)
let total = $state(0)
let unreadCount = $state(0)
let currentPage = $state(1)
let typeFilter = $state<NotificationType | ''>('')

// 是否存在生效中的筛选条件，用于空态区分「无数据」与「无匹配」。
const hasActiveFilters = $derived(typeFilter !== '')

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
 * 类型筛选变更后立即回到第一页并重新加载。
 */
function handleTypeFilterChange(value: string) {
  typeFilter = value as NotificationType | ''
  currentPage = 1
  loadNotifications()
}

/**
 * 重置筛选并回到第一页。
 */
function resetAndReload() {
  typeFilter = ''
  currentPage = 1
  loadNotifications()
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
  <title>通知中心 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="通知中心" description="查看评论回复、点赞与系统消息" crumb="通知中心">
  {#snippet actions()}
    <Button variant="outline" onclick={handleMarkAllRead}>
      <CheckCheck data-icon="inline-start" />
      全部已读
    </Button>
  {/snippet}

  <Card.Root class="overflow-hidden">
    <Card.Content class="p-0">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
        <ToggleGroup.Root
          type="single"
          variant="outline"
          size="sm"
          value={typeFilter}
          onValueChange={handleTypeFilterChange}
        >
          {#each NOTIFICATION_TYPE_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <div class="flex items-center gap-3">
          <span class="text-sm text-muted-foreground tabular-nums">
            <Bell class="mr-1 inline size-3.5" />
            {unreadCount} 条未读
          </span>
          <Button variant="ghost" size="sm" onclick={resetAndReload}>
            <RotateCcw data-icon="inline-start" />
            重置
          </Button>
        </div>
      </div>

      {#if isLoading}
        <div class="flex h-48 items-center justify-center">
          <span
            class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
          ></span>
        </div>
      {:else if notifications.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <Inbox class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {hasActiveFilters ? '没有匹配的通知' : '暂无通知'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {hasActiveFilters ? '调整或重置筛选条件后再试' : '新的评论回复与点赞会在这里提醒你'}
            </p>
          </div>
          {#if hasActiveFilters}
            <Button variant="outline" size="sm" onclick={resetAndReload}>
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {/if}
        </div>
      {:else}
        <div class="divide-y">
          {#each notifications as notification (notification.id)}
            <button
              type="button"
              onclick={() => handleMarkRead(notification)}
              class="flex w-full items-start gap-4 px-4 py-3.5 text-left transition-colors hover:bg-accent/50 {notification.isRead
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
                <p class="text-xs text-muted-foreground tabular-nums">
                  {new Date(notification.createdAt).toLocaleDateString('zh-CN')}
                </p>
              </div>
            </button>
          {/each}
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3">
          <p class="text-sm text-muted-foreground">共 {total} 条通知</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={total}
            perPage={NOTIFICATION_PAGE_SIZE}
            page={currentPage}
            onPageChange={handlePageChange}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {currentPage} 页，共 {Math.max(1, Math.ceil(total / NOTIFICATION_PAGE_SIZE))} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
</PageHeader>
