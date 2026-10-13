<script lang="ts">
import { settingsPageState, emptySettingsPageData } from './settings-page-state.svelte'
import SettingField from '$lib/components/admin/setting/setting-field.svelte'
import PageHeader from '$lib/components/admin/page-header.svelte'
import { Save, Settings as SettingsIcon } from '@lucide/svelte'
import { SETTING_GROUP_LABELS } from '$lib/constants/setting'
import { SITE_NAME_ZH } from '@myblog/shared'
import { Separator } from '$ui/separator'
import type { PageProps } from './$types'
import { Spinner } from '$ui/spinner'
import { Button } from '$ui/button'
import * as Card from '$ui/card'

let { data }: PageProps = $props()

// 初始数据经 load 注入后交给页面状态模块接管，组件只做渲染与交互分发。
let initialized = false
$effect(() => {
  if (initialized) return
  initialized = true
  settingsPageState.initialize(data ?? emptySettingsPageData())
})
</script>

<svelte:head>
  <title>系统设置 - {SITE_NAME_ZH}</title>
</svelte:head>

<PageHeader
  title="系统设置"
  description="配置站点信息、内容、媒体与安全等参数；带「未生效」角标的项尚未接入业务逻辑，修改后不产生运行时效果"
  crumb="系统设置"
>
  {#snippet actions()}
    <Button
      onclick={() => settingsPageState.saveSettings()}
      disabled={settingsPageState.isSaving || settingsPageState.isLoading}
    >
      <Save data-icon="inline-start" />
      {settingsPageState.isSaving ? '保存中...' : '保存全部修改'}
    </Button>
  {/snippet}

  {#if settingsPageState.isLoading}
    <div class="flex h-48 items-center justify-center">
      <Spinner class="size-6" aria-label="正在加载设置项" />
    </div>
  {:else if settingsPageState.settings.length === 0}
    <div class="flex h-64 flex-col items-center justify-center gap-4 px-6 text-center">
      <div class="flex size-12 items-center justify-center rounded-xl border bg-muted/50">
        <SettingsIcon class="size-6 text-muted-foreground" />
      </div>
      <div class="space-y-1">
        <h3 class="text-base font-medium">暂无设置项</h3>
        <p class="text-sm text-muted-foreground">系统配置加载后展示在这里</p>
      </div>
    </div>
  {:else}
    <div class="space-y-6">
      {#each settingsPageState.groups as group, index (group.groupName)}
        <Card.Root>
          <Card.Header>
            <Card.Title>{SETTING_GROUP_LABELS[group.groupName] ?? group.groupName}</Card.Title>
            {#if group.groupName === 'security'}
              <Card.Description>
                安全参数的实际生效渠道为服务端 config.yaml 与安全中间件，本组配置尚未接入业务逻辑。
              </Card.Description>
            {/if}
          </Card.Header>
          <Card.Content>
            <div class="grid gap-6 md:grid-cols-2">
              {#each group.items as setting (setting.keyName)}
                <SettingField
                  {setting}
                  value={settingsPageState.values[setting.keyName] ?? ''}
                  onValueChange={(nextValue: string) =>
                    settingsPageState.setValue(setting.keyName, nextValue)}
                />
              {/each}
            </div>
          </Card.Content>
          {#if index < settingsPageState.groups.length - 1}
            <Separator />
          {/if}
        </Card.Root>
      {/each}
    </div>
  {/if}
</PageHeader>
