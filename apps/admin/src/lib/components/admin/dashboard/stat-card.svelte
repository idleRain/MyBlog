<script lang="ts">
import type { Component } from 'svelte'
import { Card } from '$ui'

interface Props {
  label: string
  value: number | string
  icon: Component
  hint?: string
}

let { label, value, icon: IconComponent, hint }: Props = $props()
</script>

<Card.Root>
  <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
    <Card.Title class="text-sm font-medium">{label}</Card.Title>
    <IconComponent class="size-4 text-muted-foreground" />
  </Card.Header>
  <Card.Content>
    <div class="text-2xl font-bold">{value}</div>
    {#if hint}
      <p class="text-xs text-muted-foreground">{hint}</p>
    {/if}
  </Card.Content>
</Card.Root>
