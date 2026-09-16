<script lang="ts">
import type { FriendlyLink } from '@myblog/api/modules/friendlyLink/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { ExternalLink } from '@lucide/svelte'
import { FriendlyLinkAPI } from '$lib/api'
import { toast } from 'svelte-sonner'
import { Button, Dialog } from '$ui'

// 申请字段长度上限，与后端 ApplyFriendlyLinkRequest 的 binding 规则一致。
const APPLY_NAME_MAX_LENGTH = 50
const APPLY_URL_MAX_LENGTH = 255
const APPLY_DESCRIPTION_MAX_LENGTH = 255
const APPLY_EMAIL_MAX_LENGTH = 100

// 展示中的友情链接，由根布局的全站数据提供；错误页等无数据场景降级为空列表。
interface Props {
  friendlyLinks?: FriendlyLink[]
}

let { friendlyLinks = [] }: Props = $props()

// 申请表单状态。
let applyName = $state('')
let applyUrl = $state('')
let applyDescription = $state('')
let applyEmail = $state('')
let applying = $state(false)

// 提交友链申请，成功后清空表单并提示等待审核。
async function handleApply(event: SubmitEvent) {
  event.preventDefault()
  if (applying) return

  applying = true
  try {
    // exactOptionalPropertyTypes 下可选字段仅在非空时显式传入。
    const description = applyDescription.trim()
    const response = await FriendlyLinkAPI.apply({
      name: applyName.trim(),
      url: applyUrl.trim(),
      ...(description ? { description } : {}),
      contactEmail: applyEmail.trim()
    })

    if (response.code !== RESPONSE_CODE_SUCCESS) {
      toast.error(response.message || '申请提交失败，请稍后重试')
      return
    }

    applyName = ''
    applyUrl = ''
    applyDescription = ''
    applyEmail = ''
    toast.success('申请已提交，等待站长审核')
  } catch {
    toast.error('申请提交失败，请稍后重试')
  } finally {
    applying = false
  }
}
</script>

<Dialog.Root>
  <Dialog.Trigger>
    <Button variant="ghost" size="icon" class="h-9 w-9" aria-label="友情链接">
      <ExternalLink class="h-4 w-4" />
    </Button>
  </Dialog.Trigger>
  <Dialog.Content class="max-h-[85vh] overflow-y-auto bg-popover/95 backdrop-blur-md sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>友情链接</Dialog.Title>
    </Dialog.Header>

    {#if friendlyLinks.length === 0}
      <p class="text-sm leading-relaxed text-muted-foreground">暂无友情链接，欢迎提交互换申请。</p>
    {:else}
      <div class="space-y-3">
        {#each friendlyLinks as link (link.id)}
          <a
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center justify-between bg-secondary p-3 transition-colors hover:bg-accent"
          >
            <span class="font-medium">{link.name}</span>
            <ExternalLink class="h-4 w-4 text-muted-foreground" />
          </a>
        {/each}
      </div>
    {/if}

    <!-- 申请表单：经公开 apply 端点提交，进入待审核状态。 -->
    <form onsubmit={handleApply} class="mt-6 border-t border-line pt-5">
      <h3 class="text-sm font-bold">申请友链</h3>
      <p class="mt-1 text-xs text-muted-foreground">提交后进入待审核状态，通过后即在此展示。</p>

      <div class="mt-4 space-y-3">
        <label class="block">
          <span class="mb-1.5 block text-xs font-bold">站点名称</span>
          <input
            type="text"
            bind:value={applyName}
            required
            maxlength={APPLY_NAME_MAX_LENGTH}
            placeholder="你的站点名称"
            class="w-full border border-line bg-background px-3 py-2 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-xs font-bold">站点地址</span>
          <input
            type="url"
            bind:value={applyUrl}
            required
            maxlength={APPLY_URL_MAX_LENGTH}
            placeholder="https://…"
            class="w-full border border-line bg-background px-3 py-2 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-xs font-bold">站点简介</span>
          <input
            type="text"
            bind:value={applyDescription}
            maxlength={APPLY_DESCRIPTION_MAX_LENGTH}
            placeholder="一句话介绍站点，选填"
            class="w-full border border-line bg-background px-3 py-2 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
        <label class="block">
          <span class="mb-1.5 block text-xs font-bold">联系邮箱</span>
          <input
            type="email"
            bind:value={applyEmail}
            required
            maxlength={APPLY_EMAIL_MAX_LENGTH}
            placeholder="用于接收审核结果"
            class="w-full border border-line bg-background px-3 py-2 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
          />
        </label>
      </div>

      <button
        type="submit"
        disabled={applying}
        class="mt-4 inline-flex w-full items-center justify-center bg-primary px-4 py-2.5 text-sm font-bold text-primary-foreground transition-[background-color,color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-60"
      >
        {applying ? '提交中…' : '提交申请'}
      </button>
    </form>
  </Dialog.Content>
</Dialog.Root>
