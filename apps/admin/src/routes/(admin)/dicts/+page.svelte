<script lang="ts">
import type {
  CreateDictItemRequest,
  CreateDictTypeRequest,
  DictItem,
  DictType,
  UpdateDictItemRequest,
  UpdateDictTypeRequest
} from '@myblog/api/modules/dict/types'
import { BookText, ListTree, Pencil, Plus, Trash2 } from '@lucide/svelte'
import { Badge, Button, Card, Table } from '$ui'
import { SITE_NAME_ZH } from '@myblog/shared'

import DictTypeFormDialog from '$lib/components/admin/dict/dict-type-form-dialog.svelte'
import DictItemFormDialog from '$lib/components/admin/dict/dict-item-form-dialog.svelte'
import { dictPageState, emptyDictPageData } from './dict-page-state.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { PageProps } from './$types'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  dictPageState.initialize(data ?? emptyDictPageData())
})

// 字典生效状态的展示配置，本页为字典体系自身的管理界面，直接使用其内置枚举。
const DICT_STATUS_CONFIG: Record<number, { label: string; variant: 'default' | 'secondary' }> = {
  1: { label: '生效', variant: 'default' },
  0: { label: '停用', variant: 'secondary' }
}

// 弹窗与删除确认状态
let isTypeDialogOpen = $state(false)
let typeDialogTarget = $state<DictType | null>(null)
let isItemDialogOpen = $state(false)
let itemDialogTarget = $state<DictItem | null>(null)
let isSubmitting = $state(false)
let deleteTarget = $state<{ kind: 'type' | 'item'; id: number; name: string } | null>(null)

// 新建字典项前必须选中一个字典类型。
function openCreateItem() {
  if (!dictPageState.selectedType) {
    return
  }
  itemDialogTarget = null
  isItemDialogOpen = true
}

// 提交字典类型表单，新建与编辑分别调用对应写操作。
// 弹窗载荷为动态键值结构，与既有表单弹窗约定一致地经类型断言收窄。
async function handleTypeConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = typeDialogTarget
    ? await dictPageState.updateType({
        id: typeDialogTarget.id,
        ...payload
      } as unknown as UpdateDictTypeRequest)
    : await dictPageState.createType(payload as unknown as CreateDictTypeRequest)
  isSubmitting = false
  if (success) isTypeDialogOpen = false
}

// 提交字典项表单，新建时补充当前选中的字典类型 ID。
async function handleItemConfirm(payload: Record<string, unknown>) {
  if (isSubmitting) return
  isSubmitting = true
  const success = itemDialogTarget
    ? await dictPageState.updateItem({
        id: itemDialogTarget.id,
        ...payload
      } as unknown as UpdateDictItemRequest)
    : await dictPageState.createItem({
        typeId: dictPageState.selectedTypeId ?? 0,
        ...payload
      } as unknown as CreateDictItemRequest)
  isSubmitting = false
  if (success) isItemDialogOpen = false
}

// 执行删除确认，按目标种类分发到类型或字典项的删除操作。
async function handleDelete() {
  if (!deleteTarget || isSubmitting) return
  isSubmitting = true
  const success =
    deleteTarget.kind === 'type'
      ? await dictPageState.removeType(deleteTarget.id)
      : await dictPageState.removeItem(deleteTarget.id)
  isSubmitting = false
  if (success) deleteTarget = null
}
</script>

<svelte:head>
  <title>字典管理 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title="字典管理"
  description="集中管理状态类枚举，支持多语言文案与生效控制"
  crumb="字典管理"
>
  {#snippet actions()}
    <Button
      onclick={() => {
        typeDialogTarget = null
        isTypeDialogOpen = true
      }}
    >
      <Plus data-icon="inline-start" />
      新建字典类型
    </Button>
  {/snippet}

  <div class="grid gap-4 lg:grid-cols-3">
    <Card.Root class="lg:col-span-1">
      <Card.Header class="pb-3">
        <Card.Title class="flex items-center gap-2 text-base">
          <ListTree class="size-4" />
          字典类型
        </Card.Title>
      </Card.Header>
      <Card.Content class="p-0">
        {#if dictPageState.isLoadingTypes}
          <div class="flex h-64 items-center justify-center">
            <span
              class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
            ></span>
          </div>
        {:else if dictPageState.types.length === 0}
          <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
            <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
              <BookText class="size-6 text-muted-foreground" />
            </div>
            <div class="space-y-1">
              <h3 class="text-base font-medium">暂无字典类型</h3>
              <p class="text-sm text-muted-foreground">创建第一个字典类型开始使用</p>
            </div>
            <Button size="sm" onclick={() => (isTypeDialogOpen = true)}>
              <Plus data-icon="inline-start" />
              新建字典类型
            </Button>
          </div>
        {:else}
          <ul class="max-h-[32rem] divide-y overflow-y-auto">
            {#each dictPageState.types as type (type.id)}
              <li>
                <button
                  type="button"
                  class="w-full px-4 py-3 text-left transition-colors hover:bg-muted/60 {type.id ===
                  dictPageState.selectedTypeId
                    ? 'bg-muted'
                    : ''}"
                  onclick={() => dictPageState.selectType(type.id)}
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-medium">{type.name}</span>
                    <Badge variant={DICT_STATUS_CONFIG[type.status]?.variant ?? 'secondary'}>
                      {DICT_STATUS_CONFIG[type.status]?.label ?? '未知'}
                    </Badge>
                  </div>
                  <div class="mt-1 flex items-center justify-between gap-2">
                    <span class="font-mono text-xs text-muted-foreground">{type.code}</span>
                    <span class="flex gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="编辑字典类型"
                        onclick={() => {
                          typeDialogTarget = type
                          isTypeDialogOpen = true
                        }}
                      >
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="删除字典类型"
                        class="text-destructive hover:text-destructive"
                        onclick={() =>
                          (deleteTarget = { kind: 'type', id: type.id, name: type.name })}
                      >
                        <Trash2 />
                      </Button>
                    </span>
                  </div>
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </Card.Content>
    </Card.Root>

    <Card.Root class="lg:col-span-2">
      <Card.Header class="flex flex-row items-center justify-between pb-3">
        <Card.Title class="flex items-center gap-2 text-base">
          字典项
          {#if dictPageState.selectedType}
            <span class="font-mono text-xs text-muted-foreground">
              {dictPageState.selectedType.code}
            </span>
          {/if}
        </Card.Title>
        <Button size="sm" onclick={openCreateItem} disabled={!dictPageState.selectedType}>
          <Plus data-icon="inline-start" />
          新建字典项
        </Button>
      </Card.Header>
      <Card.Content class="p-0">
        {#if !dictPageState.selectedType}
          <div class="flex h-64 items-center justify-center">
            <p class="text-sm text-muted-foreground">请先选择左侧的字典类型</p>
          </div>
        {:else if dictPageState.isLoadingItems}
          <div class="flex h-64 items-center justify-center">
            <span
              class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
            ></span>
          </div>
        {:else if dictPageState.items.length === 0}
          <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
            <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
              <ListTree class="size-6 text-muted-foreground" />
            </div>
            <div class="space-y-1">
              <h3 class="text-base font-medium">暂无字典项</h3>
              <p class="text-sm text-muted-foreground">为当前字典类型添加可选值</p>
            </div>
            <Button size="sm" onclick={openCreateItem}>
              <Plus data-icon="inline-start" />
              新建字典项
            </Button>
          </div>
        {:else}
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.Head>值</Table.Head>
                <Table.Head>显示名</Table.Head>
                <Table.Head>描述</Table.Head>
                <Table.Head>排序</Table.Head>
                <Table.Head>状态</Table.Head>
                <Table.Head class="w-px text-center whitespace-nowrap">操作</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {#each dictPageState.items as item (item.id)}
                <Table.Row>
                  <Table.Cell>
                    <code
                      class="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground"
                    >
                      {item.value}
                    </code>
                  </Table.Cell>
                  <Table.Cell>
                    <span class="font-medium">{item.label}</span>
                  </Table.Cell>
                  <Table.Cell>
                    <span class="text-sm text-muted-foreground">{item.description || '—'}</span>
                  </Table.Cell>
                  <Table.Cell>
                    <span class="text-sm tabular-nums">{item.sortOrder}</span>
                  </Table.Cell>
                  <Table.Cell>
                    <Badge variant={DICT_STATUS_CONFIG[item.status]?.variant ?? 'secondary'}>
                      {DICT_STATUS_CONFIG[item.status]?.label ?? '未知'}
                    </Badge>
                  </Table.Cell>
                  <Table.Cell>
                    <div class="flex items-center justify-center gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="编辑字典项"
                        onclick={() => {
                          itemDialogTarget = item
                          isItemDialogOpen = true
                        }}
                      >
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="删除字典项"
                        class="text-destructive hover:text-destructive"
                        onclick={() =>
                          (deleteTarget = { kind: 'item', id: item.id, name: item.label })}
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              {/each}
            </Table.Body>
          </Table.Root>
        {/if}
      </Card.Content>
    </Card.Root>
  </div>

  <DictTypeFormDialog
    open={isTypeDialogOpen}
    target={typeDialogTarget}
    {isSubmitting}
    onOpenChange={open => (isTypeDialogOpen = open)}
    onConfirm={handleTypeConfirm}
  />

  <DictItemFormDialog
    open={isItemDialogOpen}
    target={itemDialogTarget}
    {isSubmitting}
    onOpenChange={open => (isItemDialogOpen = open)}
    onConfirm={handleItemConfirm}
  />

  <ConfirmDialog
    title={deleteTarget?.kind === 'type' ? '删除字典类型' : '删除字典项'}
    description={deleteTarget
      ? `确定删除「${deleteTarget.name}」吗？${deleteTarget.kind === 'type' ? '其全部字典项将一并删除。' : ''}`
      : ''}
    confirmText="删除"
    destructive
    isLoading={isSubmitting}
    open={deleteTarget !== null}
    onOpenChange={open => {
      if (!open && !isSubmitting) deleteTarget = null
    }}
    onConfirm={handleDelete}
  />
</PageHeader>
