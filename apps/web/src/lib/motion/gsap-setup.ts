import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { gsap } from 'gsap'

/**
 * 全站 GSAP 入口：统一注册插件并导出实例与动效常量。
 * 组件禁止各自注册插件，避免重复初始化 ScrollTrigger。
 */
gsap.registerPlugin(ScrollTrigger)

export { gsap, ScrollTrigger }

/** 动效时长与缓动的全站基准，时间单位为秒，与 GSAP API 保持一致。 */
export const MOTION = {
  duration: {
    /** 高频交互反馈，如按压与悬停位移 */
    swift: 0.18,
    /** 常规元素的滚动进场 */
    enter: 0.6,
    /** 首屏与大型版面的编辑感进场 */
    reveal: 0.9
  },
  stagger: 0.07,
  ease: {
    /** 入场统一使用强 ease-out，避免 ease-in 起步迟滞 */
    out: 'power4.out',
    /** 屏内大位移使用强 ease-in-out，呈现自然的加减速 */
    inOut: 'power4.inOut',
    /** 滚动驱动一律线性，平滑度交给 scrub 控制 */
    linear: 'none'
  }
} as const

/** 滚动进场的默认触发锚点：元素顶部越过视口下方 15% 处。 */
export const REVEAL_START = 'top 85%'
