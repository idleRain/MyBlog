<script lang="ts">
import { onDestroy, onMount } from 'svelte'

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

let editorElement: HTMLDivElement
let editorInstance: import('@toast-ui/editor').default | null = $state(null)
let isReady = $state(false)

// 监听 documentElement 的 class 变化，暗色模式切换时同步编辑器主题。
let themeObserver: MutationObserver | null = null

/**
 * 按当前暗色模式名称返回 toast-ui 主题标识。
 */
function resolveTheme(): 'light' | 'dark' {
  return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
}

/**
 * 初始化编辑器实例，组件与样式均按需加载以控制首屏体积。
 * 后台应用关闭 SSR，onMount 仅在浏览器执行，无需区分客户端与服务端。
 */
async function initEditor() {
  if (editorInstance) return

  const { default: Editor } = await import('@toast-ui/editor')
  await import('@toast-ui/editor/dist/toastui-editor.css')
  await import('@toast-ui/editor/dist/theme/toastui-editor-dark.css')
  await import('@toast-ui/editor/dist/i18n/zh-cn')

  editorInstance = new Editor({
    el: editorElement,
    initialValue: value,
    placeholder,
    height,
    initialEditType: 'markdown',
    previewStyle: 'vertical',
    theme: resolveTheme(),
    language: 'zh-CN'
  })

  editorInstance.on('change', () => {
    value = editorInstance?.getMarkdown() ?? ''
  })
  isReady = true
}

/**
 * 启动主题监听，切换暗色模式时同步编辑器主题。
 */
function watchTheme() {
  themeObserver = new MutationObserver(() => {
    editorInstance?.setTheme(resolveTheme())
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  })
}

// 外部重置 value 时同步回编辑器，避免父组件清空表单后内容残留。
$effect(() => {
  if (!editorInstance || !isReady) return
  if (editorInstance.getMarkdown() !== value) {
    editorInstance.setMarkdown(value)
  }
})

onMount(async () => {
  await initEditor()
  watchTheme()
})

onDestroy(() => {
  themeObserver?.disconnect()
  editorInstance?.destroy()
  editorInstance = null
})
</script>

<div class={className} bind:this={editorElement}></div>
