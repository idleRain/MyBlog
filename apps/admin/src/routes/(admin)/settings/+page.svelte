<script lang="ts">
import SettingField from '$lib/components/admin/setting/setting-field.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import type { Setting } from '@myblog/api/modules/setting/types'
import { Save, Settings as SettingsIcon } from '@lucide/svelte'
import { SETTING_GROUP_LABELS } from '$lib/constants/setting'
import { Button, Card, Separator } from '$ui'
import { SettingAPI } from '$lib/api'
import { onMount } from 'svelte'

let settings = $state<Setting[]>([])
let values = $state<Record<string, string>>({})
let originalValues = $state<Record<string, string>>({})
let isLoading = $state(true)
let isSaving = $state(false)

/**
 * 按分组名聚合并保持原顺序。
 */
function groupSettings(): Array<{ groupName: string; items: Setting[] }> {
  const groups: Record<string, Setting[]> = {}
  for (const setting of settings) {
    const group = groups[setting.groupName] ?? []
    group.push(setting)
    groups[setting.groupName] = group
  }
  return Object.entries(groups).map(([groupName, items]) => ({ groupName, items }))
}

/**
 * 加载全部设置项并初始化表单值。
 */
async function loadSettings() {
  isLoading = true
  try {
    const response = await SettingAPI.list()
    if (response.code === 200 && response.data) {
      settings = response.data.settings ?? []
      const nextValues: Record<string, string> = {}
      const nextOriginals: Record<string, string> = {}
      for (const setting of settings) {
        nextValues[setting.keyName] = setting.value
        nextOriginals[setting.keyName] = setting.value
      }
      values = nextValues
      originalValues = nextOriginals
    } else {
      toast.error(response.message || '加载设置失败')
    }
  } catch (error) {
    console.error('加载设置失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isLoading = false
  }
}

/**
 * 汇总有改动的设置项，敏感掩码未改动时不会提交。
 */
function collectChangedItems(): Array<{ keyName: string; value: string }> {
  return Object.keys(values)
    .filter(keyName => originalValues[keyName] !== values[keyName])
    .map(keyName => ({ keyName, value: values[keyName] ?? '' }))
}

/**
 * 批量保存有改动的设置项。
 */
async function handleSave() {
  if (isSaving) return
  const items = collectChangedItems()
  if (items.length === 0) {
    toast.info('没有需要保存的修改')
    return
  }

  isSaving = true
  try {
    const response = await SettingAPI.update({ items })
    if (response.code === 200 && response.data) {
      toast.success(`已保存 ${items.length} 项设置`)
      await loadSettings()
    } else {
      toast.error(response.message || '保存设置失败')
    }
  } catch (error) {
    console.error('保存设置失败:', error)
    toast.error('网络错误，请稍后重试')
  } finally {
    isSaving = false
  }
}

onMount(loadSettings)
</script>

<svelte:head>
  <title>系统设置 - MyBlog</title>
</svelte:head>

<PageHeader title="系统设置" description="配置站点信息、内容、媒体与安全等参数" crumb="系统设置">
  {#snippet actions()}
    <Button onclick={handleSave} disabled={isSaving || isLoading}>
      <Save data-icon="inline-start" />
      {isSaving ? '保存中...' : '保存全部修改'}
    </Button>
  {/snippet}

  {#if isLoading}
    <div class="flex h-48 items-center justify-center">
      <span class="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
      ></span>
    </div>
  {:else if settings.length === 0}
    <div class="flex h-48 items-center justify-center">
      <div class="text-center">
        <SettingsIcon class="mx-auto size-12 text-muted-foreground" />
        <h3 class="mt-4 text-lg font-medium">暂无设置项</h3>
      </div>
    </div>
  {:else}
    <div class="space-y-6">
      {#each groupSettings() as group, index (group.groupName)}
        <Card.Root>
          <Card.Header>
            <Card.Title>{SETTING_GROUP_LABELS[group.groupName] ?? group.groupName}</Card.Title>
          </Card.Header>
          <Card.Content>
            <div class="grid gap-6 md:grid-cols-2">
              {#each group.items as setting (setting.keyName)}
                <SettingField
                  {setting}
                  value={values[setting.keyName] ?? ''}
                  onValueChange={(nextValue: string) => {
                    values[setting.keyName] = nextValue
                  }}
                />
              {/each}
            </div>
          </Card.Content>
          {#if index < groupSettings().length - 1}
            <Separator.Root />
          {/if}
        </Card.Root>
      {/each}
    </div>
  {/if}
</PageHeader>
