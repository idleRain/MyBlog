<script lang="ts">
import { LINK_STATUS_ACTIONS, LINK_STATUS_CONFIG, type LinkStatusAction } from '$lib/constants/link'
import { CircleCheck, CircleX, MoreHorizontal, Pencil, EyeOff, Trash2 } from '@lucide/svelte'
import type { FriendlyLink } from '@myblog/api/modules/friendlyLink/types'
import * as DropdownMenu from '$ui/dropdown-menu'
import type { LucideIcon } from '@lucide/svelte'
import { Button } from '$ui/button'
import * as Table from '$ui/table'
import { Badge } from '$ui/badge'

interface Props {
  links: FriendlyLink[]
  onEdit: (link: FriendlyLink) => void
  onReview: (link: FriendlyLink, key: LinkStatusAction) => void
  onDelete: (link: FriendlyLink) => void
}

let { links, onEdit, onReview, onDelete }: Props = $props()

// 审核动作图标映射，与 LINK_STATUS_ACTIONS 的动作键一一对应，供下拉菜单渲染语义图标。
// 图标只由本组件消费，故不下沉常量模块，避免页面状态模块的导入图牵连整个图标库。
const LINK_ACTION_ICONS: Record<LinkStatusAction, LucideIcon> = {
  approve: CircleCheck,
  hide: EyeOff,
  reject: CircleX
}
</script>

<!-- 友链表格：描述、回链与添加时间在窄屏折叠，站点、状态与操作保持可见 -->
<Table.Root>
  <Table.Header>
    <Table.Row>
      <Table.Head>站点</Table.Head>
      <Table.Head class="hidden md:table-cell">描述</Table.Head>
      <Table.Head class="hidden md:table-cell">回链</Table.Head>
      <Table.Head>状态</Table.Head>
      <Table.Head class="hidden md:table-cell">添加时间</Table.Head>
      <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {#each links as link (link.id)}
      <Table.Row>
        <Table.Cell>
          <div class="flex items-center gap-3">
            {#if link.logo}
              <img src={link.logo} alt={link.name} class="size-8 rounded-full object-cover" />
            {/if}
            <div class="min-w-0">
              <p class="truncate font-medium">{link.name}</p>
              <a
                href={link.url}
                target="_blank"
                rel="noreferrer"
                class="hidden truncate text-xs text-muted-foreground hover:underline md:inline"
              >
                {link.url}
              </a>
              <!-- 描述与回链在窄屏并入站点列副行，md 及以上恢复独立列 -->
              {#if link.description}
                <span class="block text-sm text-muted-foreground md:hidden">
                  {link.description}
                </span>
              {/if}
              {#if link.isReciprocal}
                <Badge variant="outline" class="md:hidden">已回链</Badge>
              {/if}
            </div>
          </div>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="line-clamp-1 max-w-48 text-sm text-muted-foreground">
            {link.description || '—'}
          </span>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          {#if link.isReciprocal}
            <Badge variant="outline">已回链</Badge>
          {:else}
            <span class="text-sm text-muted-foreground">—</span>
          {/if}
        </Table.Cell>
        <Table.Cell>
          <Badge variant={LINK_STATUS_CONFIG[link.status].variant}>
            {LINK_STATUS_CONFIG[link.status].label}
          </Badge>
        </Table.Cell>
        <Table.Cell class="hidden md:table-cell">
          <span class="text-sm text-muted-foreground tabular-nums">
            {new Date(link.createdAt).toLocaleDateString('zh-CN')}
          </span>
        </Table.Cell>
        <Table.Cell>
          <div class="flex items-center justify-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              aria-label={`编辑友链 ${link.name}`}
              onclick={() => onEdit(link)}
            >
              <Pencil />
            </Button>
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                <Button variant="ghost" size="icon" aria-label={`友链操作 ${link.name}`}>
                  <MoreHorizontal />
                </Button>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="end">
                {#each LINK_STATUS_ACTIONS[link.status] as action (action.key)}
                  {@const ActionIcon = LINK_ACTION_ICONS[action.key]}
                  <DropdownMenu.Item onSelect={() => onReview(link, action.key)}>
                    <ActionIcon data-icon="inline-start" />
                    {action.label}
                  </DropdownMenu.Item>
                {/each}
                <DropdownMenu.Item variant="destructive" onSelect={() => onDelete(link)}>
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
