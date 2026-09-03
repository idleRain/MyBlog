<script lang="ts">
import { Mail, Pencil, Trash2, MoreHorizontal, Power, Users as UsersIcon } from '@lucide/svelte'
import UserStatusBadge from '$lib/components/admin/user/user-status-badge.svelte'
import { Avatar, Badge, Button, Card, Checkbox, DropdownMenu, Table } from '$ui'
import type { User } from '@myblog/api/modules/user/types'
import { getRoleInfo } from '$lib/utils/permissions'

interface Props {
  users: User[]
  selectedIds: number[]
  isAllSelected: boolean
  isLoading: boolean
  onToggleSelectAll: () => void
  onToggleSelect: (id: number) => void
  onEdit: (user: User) => void
  onToggleStatus: (user: User) => void
  onDelete: (user: User) => void
}

let {
  users,
  selectedIds,
  isAllSelected,
  isLoading,
  onToggleSelectAll,
  onToggleSelect,
  onEdit,
  onToggleStatus,
  onDelete
}: Props = $props()
</script>

<Card.Root>
  <Card.Content class="p-0">
    {#if isLoading}
      <div class="flex h-48 items-center justify-center">
        <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
        ></span>
      </div>
    {:else if users.length === 0}
      <div class="flex h-48 items-center justify-center">
        <div class="text-center">
          <UsersIcon class="mx-auto size-12 text-muted-foreground" />
          <h3 class="mt-4 text-lg font-medium">没有找到用户</h3>
          <p class="text-sm text-muted-foreground">调整搜索条件或创建新用户</p>
        </div>
      </div>
    {:else}
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head class="w-12">
              <Checkbox.Root
                checked={isAllSelected}
                onCheckedChange={onToggleSelectAll}
                aria-label="全选用户"
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
          {#each users as user (user.id)}
            {@const roleInfo = getRoleInfo(user.role || 'user')}
            <Table.Row>
              <Table.Cell>
                <Checkbox.Root
                  checked={selectedIds.includes(user.id)}
                  onCheckedChange={() => onToggleSelect(user.id)}
                  aria-label={`选择用户 ${user.username}`}
                />
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center gap-3">
                  <Avatar.Root class="size-8">
                    <Avatar.Image src={user.avatar} alt={user.nickname || user.username} />
                    <Avatar.Fallback>
                      <UsersIcon class="size-4" />
                    </Avatar.Fallback>
                  </Avatar.Root>
                  <div>
                    <p class="font-medium">{user.nickname || user.username}</p>
                    <p class="text-sm text-muted-foreground">@{user.username}</p>
                  </div>
                </div>
              </Table.Cell>
              <Table.Cell>
                <span class="flex items-center gap-2 text-sm">
                  <Mail class="size-4 text-muted-foreground" />
                  {user.email}
                </span>
              </Table.Cell>
              <Table.Cell>
                <Badge variant={roleInfo.color}>{roleInfo.name}</Badge>
              </Table.Cell>
              <Table.Cell>
                <UserStatusBadge status={user.status} />
              </Table.Cell>
              <Table.Cell>
                <span class="text-sm text-muted-foreground">
                  {new Date(user.createdAt).toLocaleDateString('zh-CN')}
                </span>
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center justify-end gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    aria-label="编辑用户"
                    onclick={() => onEdit(user)}
                  >
                    <Pencil />
                  </Button>
                  <DropdownMenu.Root>
                    <DropdownMenu.Trigger>
                      <Button variant="ghost" size="sm" aria-label="更多操作">
                        <MoreHorizontal />
                      </Button>
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content align="end">
                      {#if user.status === 1}
                        <DropdownMenu.Item onselect={() => onToggleStatus(user)}>
                          <Power data-icon="inline-start" />
                          禁用
                        </DropdownMenu.Item>
                      {:else if user.status === 0}
                        <DropdownMenu.Item onselect={() => onToggleStatus(user)}>
                          <Power data-icon="inline-start" />
                          启用
                        </DropdownMenu.Item>
                      {/if}
                      <DropdownMenu.Item variant="destructive" onselect={() => onDelete(user)}>
                        <Trash2 data-icon="inline-start" />
                        删除
                      </DropdownMenu.Item>
                    </DropdownMenu.Content>
                  </DropdownMenu.Root>
                </div>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    {/if}
  </Card.Content>
</Card.Root>
