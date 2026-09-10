import { gsap, MOTION, REVEAL_START } from './gsap-setup'

export type ScrollRevealOptions = {
  /** 垂直位移距离，单位像素，默认上浮 28px */
  y?: number
  /** 水平位移距离，单位像素，用于侧栏卡片等横向进场 */
  x?: number
  /** 动画时长，单位秒 */
  duration?: number
  /** 相对延迟，单位秒，用于同屏元素的节奏编排 */
  delay?: number
}

/**
 * 滚动进场动作：元素进入视口后位移并淡入，仅播放一次。
 * 动画只作用于 transform 与 opacity，滚动过程中不触发布局与重绘。
 * 减少动态偏好下不注册任何动画，元素保持静态可见。
 */
export function scrollReveal(node: HTMLElement, options: ScrollRevealOptions = {}) {
  const media = gsap.matchMedia()

  media.add('(prefers-reduced-motion: no-preference)', () => {
    gsap.from(node, {
      y: options.y ?? 28,
      x: options.x ?? 0,
      autoAlpha: 0,
      duration: options.duration ?? MOTION.duration.reveal,
      delay: options.delay ?? 0,
      ease: MOTION.ease.out,
      scrollTrigger: { trigger: node, start: REVEAL_START, once: true }
    })
  })

  return {
    destroy() {
      media.revert()
    }
  }
}

export type ScrollStaggerOptions = {
  /** 组内子元素选择器，缺省时取全部直接子元素 */
  selector?: string
  /** 垂直位移距离，单位像素 */
  y?: number
  /** 相邻元素的时间间隔，单位秒 */
  each?: number
  /** 动画时长，单位秒 */
  duration?: number
}

/** 滚动进场动作的组内错峰版本：子元素按固定间隔依次浮现。 */
export function scrollStagger(node: HTMLElement, options: ScrollStaggerOptions = {}) {
  const targets = options.selector ? node.querySelectorAll(options.selector) : node.children

  const media = gsap.matchMedia()

  media.add('(prefers-reduced-motion: no-preference)', () => {
    gsap.from(targets, {
      y: options.y ?? 24,
      autoAlpha: 0,
      duration: options.duration ?? MOTION.duration.enter,
      stagger: options.each ?? MOTION.stagger,
      ease: MOTION.ease.out,
      scrollTrigger: { trigger: node, start: REVEAL_START, once: true }
    })
  })

  return {
    destroy() {
      media.revert()
    }
  }
}
