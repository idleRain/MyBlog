<script lang="ts">
import type { Setting } from '@myblog/api/modules/setting/types'
import { Input, Label, Switch, Textarea } from '$ui'

interface Props {
  setting: Setting
  value: string
  onValueChange: (value: string) => void
}

let { setting, value, onValueChange }: Props = $props()

const isReadonly = $derived(setting.isReadonly)
</script>

<div class="space-y-2">
  <div class="flex items-center justify-between gap-2">
    <Label.Root for={`setting-${setting.keyName}`}>
      {setting.label || setting.keyName}
      {#if isReadonly}
        <span class="text-xs text-muted-foreground">（只读）</span>
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
