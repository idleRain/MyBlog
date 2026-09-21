<script lang="ts">
import { replaceAll } from '@milkdown/kit/utils'
import '@milkdown/crepe/theme/common/style.css'
import { onDestroy, onMount } from 'svelte'
import '@milkdown/crepe/theme/frame.css'
import { Crepe } from '@milkdown/crepe'

interface Props {
  value?: string
  height?: string
  placeholder?: string
  class?: string
}

// value 为双向绑定，编辑器内容变化时同步到父组件。
let {
  value = $bindable(''),
  height = '480px',
  placeholder = '请输入 Markdown 内容...',
  class: className
}: Props = $props()

let host: HTMLDivElement
let crepe: Crepe | null = null
let isReady = $state(false)

// 外部重置内容回写编辑器期间抑制变更同步，防止自渲染内容再次回写父组件形成回环。
let isSyncingFromExternal = false

/**
 * 创建 Crepe 编辑器实例并订阅 markdown 更新事件。
 * admin 应用关闭 SSR，初始化仅发生在浏览器环境。
 */
async function initEditor() {
  const instance = new Crepe({
    root: host,
    defaultValue: value,
    features: {
      // LaTeX 功能依赖额外体积且当前写作场景不使用，保持关闭。
      [Crepe.Feature.Latex]: false
    },
    featureConfigs: {
      [Crepe.Feature.Placeholder]: { text: placeholder, mode: 'block' }
    }
  })

  instance.on(listener => {
    listener.markdownUpdated((_ctx, markdown, prevMarkdown) => {
      if (markdown !== prevMarkdown && !isSyncingFromExternal) {
        value = markdown
      }
    })
  })

  await instance.create()
  crepe = instance
  isReady = true
}

// 外部重置 value 时同步回编辑器，父组件清空表单或加载文章后内容不残留。
$effect(() => {
  if (!isReady || !crepe) return
  if (crepe.getMarkdown() === value) return
  isSyncingFromExternal = true
  crepe.editor.action(replaceAll(value))
  isSyncingFromExternal = false
})

onMount(() => {
  void initEditor()
})

onDestroy(() => {
  void crepe?.destroy()
  crepe = null
})
</script>

<div bind:this={host} class={className} style="height: {height}; overflow-y: auto;"></div>
