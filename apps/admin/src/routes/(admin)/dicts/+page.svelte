<script lang="ts">
import type {
  CreateDictItemRequest,
  CreateDictTypeRequest,
  DictItem,
  DictType,
  UpdateDictItemRequest,
  UpdateDictTypeRequest
} from '@myblog/api/modules/dict/types'
import { SITE_NAME_ZH } from '@myblog/shared'
import { Plus } from '@lucide/svelte'
import { Button } from '$ui/button'

import DictTypeFormDialog from '$lib/components/admin/dict/dict-type-form-dialog.svelte'
import DictItemFormDialog from '$lib/components/admin/dict/dict-item-form-dialog.svelte'
import DictItemTable from '$lib/components/admin/dict/dict-item-table.svelte'
import DictTypeList from '$lib/components/admin/dict/dict-type-list.svelte'
import { dictPageState, emptyDictPageData } from './dict-page-state.svelte'
import ConfirmDialog from '$lib/components/admin/confirm-dialog.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { PageProps } from './$types'

// 删除确认的目标，种类决定分发到类型删除还是字典项删除。
type DeleteTarget = { kind: 'type' | 'item'; id: number; name: string }

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  dictPageState.initialize(data ?? emptyDictPageData())
})

// 弹窗与删除确认状态
let isTypeDialogOpen = $state(false)
let typeDialogTarget = $state<DictType | null>(null)
let isItemDialogOpen = $state(false)
let itemDialogTarget = $state<DictItem | null>(null)
let isSubmitting = $state(false)
let deleteTarget = $state<DeleteTarget | null>(null)

// 打开字典类型弹窗，目标为空表示新建。
function openTypeDialog(target: DictType | null) {
  typeDialogTarget = target
  isTypeDialogOpen = true
}

// 打开字典项弹窗，新建前必须已选中字典类型。
function openItemDialog(target: DictItem | null) {
  if (!target && !dictPageState.selectedType) return
  itemDialogTarget = target
  isItemDialogOpen = true
}

// 记录待删除目标，弹窗提交后按种类分发。
function openDeleteDialog(kind: 'type' | 'item', id: number, name: string) {
  deleteTarget = { kind, id, name }
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
    <Button onclick={() => openTypeDialog(null)}>
      <Plus data-icon="inline-start" />
      新建字典类型
    </Button>
  {/snippet}

  <div class="grid gap-4 lg:grid-cols-3">
    <DictTypeList
      types={dictPageState.types}
      selectedTypeId={dictPageState.selectedTypeId}
      isLoading={dictPageState.isLoadingTypes}
      onSelect={id => void dictPageState.selectType(id)}
      onEdit={type => openTypeDialog(type)}
      onDelete={type => openDeleteDialog('type', type.id, type.name)}
      onCreate={() => openTypeDialog(null)}
    />

    <DictItemTable
      type={dictPageState.selectedType}
      items={dictPageState.items}
      isLoading={dictPageState.isLoadingItems}
      onCreate={() => openItemDialog(null)}
      onEdit={item => openItemDialog(item)}
      onDelete={item => openDeleteDialog('item', item.id, item.label)}
    />
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
