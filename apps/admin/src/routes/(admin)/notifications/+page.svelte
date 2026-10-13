<script lang="ts">
import {
  NOTIFICATION_PAGE_SIZE,
  NOTIFICATION_TYPE_LABELS,
  NOTIFICATION_TYPE_OPTIONS
} from '$lib/constants/notification'
import {
  notificationsPageState,
  emptyNotificationsPageData
} from './notifications-page-state.svelte'
import { Bell, CheckCheck, Inbox, RotateCcw } from '@lucide/svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import * as ToggleGroup from '$ui/toggle-group'
import { SITE_NAME_ZH } from '@myblog/shared'
import * as Pagination from '$ui/pagination'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import { Badge } from '$ui/badge'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  notificationsPageState.initialize(data ?? emptyNotificationsPageData())
})
</script>

<svelte:head>
  <title>通知中心 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader title="通知中心" description="查看评论回复、点赞与系统消息" crumb="通知中心">
  {#snippet actions()}
    <Button variant="outline" onclick={() => notificationsPageState.markAllRead()}>
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
          value={notificationsPageState.typeFilter}
          onValueChange={value => notificationsPageState.setTypeFilter(value)}
        >
          {#each NOTIFICATION_TYPE_OPTIONS as option (option.label)}
            <ToggleGroup.Item value={option.value}>{option.label}</ToggleGroup.Item>
          {/each}
        </ToggleGroup.Root>

        <div class="flex items-center gap-3">
          <span class="text-sm text-muted-foreground tabular-nums">
            <Bell class="mr-1 inline size-3.5" />
            {notificationsPageState.unreadCount} 条未读
          </span>
          <Button variant="ghost" size="sm" onclick={() => notificationsPageState.resetFilters()}>
            <RotateCcw data-icon="inline-start" />
            重置
          </Button>
        </div>
      </div>

      {#if notificationsPageState.isLoading}
        <div class="flex h-48 items-center justify-center">
          <Spinner class="size-6" aria-label="正在加载通知" />
        </div>
      {:else if notificationsPageState.notifications.length === 0}
        <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
          <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
            <Inbox class="size-6 text-muted-foreground" />
          </div>
          <div class="space-y-1">
            <h3 class="text-base font-medium">
              {notificationsPageState.hasActiveFilters ? '没有匹配的通知' : '暂无通知'}
            </h3>
            <p class="text-sm text-muted-foreground">
              {notificationsPageState.hasActiveFilters
                ? '调整或重置筛选条件后再试'
                : '新的评论回复与点赞会在这里提醒你'}
            </p>
          </div>
          {#if notificationsPageState.hasActiveFilters}
            <Button
              variant="outline"
              size="sm"
              onclick={() => notificationsPageState.resetFilters()}
            >
              <RotateCcw data-icon="inline-start" />
              重置筛选
            </Button>
          {/if}
        </div>
      {:else}
        <div class="divide-y">
          {#each notificationsPageState.notifications as notification (notification.id)}
            <button
              type="button"
              onclick={() => notificationsPageState.markRead(notification)}
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
          <p class="text-sm text-muted-foreground">共 {notificationsPageState.total} 条通知</p>
          <Pagination.Root
            class="mx-0 w-auto"
            count={notificationsPageState.total}
            perPage={NOTIFICATION_PAGE_SIZE}
            page={notificationsPageState.currentPage}
            onPageChange={page => notificationsPageState.changePage(page)}
          >
            <Pagination.Content>
              <Pagination.PrevButton />
              <span class="px-2 text-sm text-muted-foreground tabular-nums">
                第 {notificationsPageState.currentPage} 页，共 {notificationsPageState.totalPages} 页
              </span>
              <Pagination.NextButton />
            </Pagination.Content>
          </Pagination.Root>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>
</PageHeader>
