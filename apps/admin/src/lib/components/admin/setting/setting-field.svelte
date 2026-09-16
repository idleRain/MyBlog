<script lang="ts">
import type { Setting } from '@myblog/api/modules/setting/types'
import { isSettingEffective } from '$lib/constants/setting'
import { Input, Label, Switch, Textarea } from '$ui'

interface Props {
  setting: Setting
  value: string
  onValueChange: (value: string) => void
}

let { setting, value, onValueChange }: Props = $props()

const isReadonly = $derived(setting.isReadonly)
// 未被业务消费的设置项以角标提示，避免管理员误以为修改即生效。
const isEffective = $derived(isSettingEffective(setting.keyName))
</script>

<div class="space-y-2">
  <div class="flex items-center justify-between gap-2">
    <Label.Root for={`setting-${setting.keyName}`}>
      {setting.label || setting.keyName}
      {#if isReadonly}
        <span class="text-xs text-muted-foreground">（只读）</span>
      {/if}
      {#if !isEffective}
        <span
          class="ml-1.5 inline-flex items-center rounded-none border border-amber-600/40 px-1.5 py-0.5 text-[10px] leading-none font-bold text-amber-600 dark:border-amber-400/40 dark:text-amber-400"
          title="该设置项尚未接入业务逻辑，修改后不会产生任何效果"
        >
          未生效
        </span>
      {/if}
    </Label.Root>
  </div>

  {#if setting.type === 'boolean'}
    <Switch.Switch
      id={`setting-${setting.keyName}`}
      checked={value === 'true' || value === '1'}
      onCheckedChange={(checked: boolean) => onValueChange(String(checked))}
      disabled={isReadonly}
    />
  {:else if setting.type === 'json' || setting.type === 'array'}
    <Textarea.Textarea
      id={`setting-${setting.keyName}`}
      rows={3}
      {value}
      oninput={(event: Event) => onValueChange((event.target as HTMLTextAreaElement).value)}
      disabled={isReadonly}
      class="resize-y font-mono text-xs"
    />
  {:else}
    <Input.Root
      id={`setting-${setting.keyName}`}
      type={setting.type === 'number' ? 'number' : 'text'}
      {value}
      oninput={(event: Event) => onValueChange((event.target as HTMLInputElement).value)}
      disabled={isReadonly}
    />
  {/if}

  {#if setting.description}
    <p class="text-xs text-muted-foreground">{setting.description}</p>
  {/if}
</div>
