/**
 * 代码块客户端交互：为渲染管线输出的 figure.code-block 提供复制按钮行为。
 * 交互以事件委托绑定在正文容器上，容器内 HTML 经 {@html} 注入后无需逐块绑定。
 */

import type { Action } from 'svelte/action'

// 复制结果反馈的展示时长，超时后按钮恢复初始文案与图标。
const COPY_FEEDBACK_MS = 2000

// 外壳结构与按钮的标识选择器，与渲染管线 render.ts 的输出约定一一对应。
const CODE_BLOCK_SELECTOR = '.code-block'
const COPY_BUTTON_SELECTOR = '[data-code-copy]'
const COPY_TEXT_SELECTOR = '.code-block-copy-text'

// 按钮三种状态下的文案：初始态、复制成功态与复制失败态。
const COPY_LABEL = '复制'
const COPIED_LABEL = '已复制'
const COPY_FAILED_LABEL = '复制失败'

/**
 * 将代码文本写入剪贴板，优先使用异步剪贴板 API，
 * 非安全上下文降级为隐藏文本域触发 execCommand 命令，两条路径都失败时向上抛错。
 */
async function writeClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text)
    return
  }
  if (!copyViaHiddenTextarea(text)) {
    throw new Error('剪贴板降级写入失败')
  }
}

/** 降级复制：挂载临时只读文本域并触发 execCommand 命令，返回命令是否执行成功。 */
function copyViaHiddenTextarea(text: string): boolean {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    return document.execCommand('copy')
  } finally {
    textarea.remove()
  }
}

/**
 * 代码块交互增强动作，绑定容器后拦截内部复制按钮的点击并给出状态反馈。
 * 返回的销毁钩子负责清理监听与未完成的反馈定时器，避免组件卸载后残留副作用。
 */
export const enhanceCodeBlocks: Action<HTMLElement> = container => {
  // 每个按钮的反馈复原定时器，重复点击时重置计时，卸载时统一清理。
  const revertTimers = new Map<HTMLButtonElement, ReturnType<typeof setTimeout>>()

  /** 从按钮所在的外壳中读取待复制的源码文本，去除结尾换行以对齐高亮块的复制结果。 */
  function readCodeText(button: HTMLButtonElement): string | null {
    const block = button.closest<HTMLElement>(CODE_BLOCK_SELECTOR)
    const text = block?.querySelector('pre code')?.textContent
    if (text === undefined || text === null) return null
    return text.replace(/\n$/, '')
  }

  /** 将按钮切换到指定反馈态，并在展示时限后恢复初始文案。 */
  function showFeedback(button: HTMLButtonElement, label: string, copied: boolean): void {
    const previousTimer = revertTimers.get(button)
    if (previousTimer) clearTimeout(previousTimer)
    button.dataset.copied = copied ? 'true' : 'false'
    const textNode = button.querySelector(COPY_TEXT_SELECTOR)
    if (textNode) textNode.textContent = label
    revertTimers.set(
      button,
      setTimeout(() => {
        revertTimers.delete(button)
        button.dataset.copied = 'false'
        if (textNode) textNode.textContent = COPY_LABEL
      }, COPY_FEEDBACK_MS)
    )
  }

  /** 处理复制点击：读取源码文本写入剪贴板，并按结果切换按钮反馈态。 */
  async function handleCopyClick(button: HTMLButtonElement): Promise<void> {
    const text = readCodeText(button)
    if (text === null) return
    try {
      await writeClipboard(text)
      showFeedback(button, COPIED_LABEL, true)
    } catch (error) {
      console.error('复制代码失败:', error)
      showFeedback(button, COPY_FAILED_LABEL, false)
    }
  }

  /** 容器级点击委托：仅响应来自复制按钮的点击事件。 */
  function handleContainerClick(event: MouseEvent): void {
    const target = event.target instanceof Element ? event.target : null
    const button = target?.closest<HTMLButtonElement>(COPY_BUTTON_SELECTOR)
    if (button) void handleCopyClick(button)
  }

  container.addEventListener('click', handleContainerClick)

  return {
    destroy() {
      for (const timer of revertTimers.values()) clearTimeout(timer)
      revertTimers.clear()
      container.removeEventListener('click', handleContainerClick)
    }
  }
}
