import mitt from 'mitt'

/**
 * 全局事件总线实例，供跨模块通信复用。
 */
const mitter = mitt()

export default mitter
