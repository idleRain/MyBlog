<script lang="ts">
import { Mail, Pencil, Trash2, MoreHorizontal, Power, Users as UsersIcon } from '@lucide/svelte'
import UserStatusBadge from '$lib/components/admin/user/user-status-badge.svelte'
import { Avatar, Badge, Button, Checkbox, DropdownMenu, Table } from '$ui'
import type { User } from '@myblog/api/modules/user/types'
import { getRoleInfo } from '$lib/utils/permissions'

interface Props {
  users: User[]
  selectedIds: number[]
  isAllSelected: boolean
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
  onToggleSelectAll,
  onToggleSelect,
  onEdit,
  onToggleStatus,
  onDelete
}: Props = $props()
</script>

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
      <Table.Head class="hidden md:table-cell">创建时间</Table.Head>
      <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
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
              <!-- 邮箱在窄屏并入用户列副行，md 及以上保持独立列展示完整信息 -->
              <p class="truncate text-sm text-muted-foreground md:hidden">{user.email}</p>
            </div>
          </div>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="flex items-center gap-2 text-sm">
            <Mail class="size-4 shrink-0 text-muted-foreground" />
            <span class="truncate">{user.email}</span>
          </span>
        </Table.Cell>
        <Table.Cell>
          <Badge variant={roleInfo.color}>{roleInfo.name}</Badge>
        </Table.Cell>
        <Table.Cell>
          <UserStatusBadge status={user.status} />
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="text-sm text-muted-foreground tabular-nums">
            {new Date(user.createdAt).toLocaleDateString('zh-CN')}
          </span>
        </Table.Cell>
        <Table.Cell>
          <div class="flex items-center justify-center gap-1">
            <Button variant="ghost" size="sm" aria-label="编辑用户" onclick={() => onEdit(user)}>
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
                  <DropdownMenu.Item onSelect={() => onToggleStatus(user)}>
                    <Power data-icon="inline-start" />
                    禁用
                  </DropdownMenu.Item>
                {:else if user.status === 0}
                  <DropdownMenu.Item onSelect={() => onToggleStatus(user)}>
                    <Power data-icon="inline-start" />
                    启用
                  </DropdownMenu.Item>
                {/if}
                <DropdownMenu.Item variant="destructive" onSelect={() => onDelete(user)}>
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
