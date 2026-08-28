/**
 * 防抖：连续触发时重置计时器，最后一次调用后等待 duration 再执行。
 */
export const debounce = <T extends (...args: any[]) => any>(
  fn: T,
  duration: number = 300
): ((this: ThisParameterType<T>, ...args: Parameters<T>) => void) => {
  let timer: ReturnType<typeof setTimeout> | null = null
  return function (this: ThisParameterType<T>, ...args: Parameters<T>): void {
    if (timer !== null) clearTimeout(timer)

    timer = setTimeout(() => {
      fn.apply(this, args)
      timer = null
    }, duration)
  }
}

/**
 * 节流：在间隔内最多执行一次，默认首尾各执行一次。
 */
export const throttle = <T extends (...args: any[]) => any>(
  fn: T,
  interval: number = 300,
  options: { leading?: boolean; trailing?: boolean } = {}
): ((this: ThisParameterType<T>, ...args: Parameters<T>) => void) => {
  const { leading = true, trailing = true } = options

  let lastTime = 0
  let timer: ReturnType<typeof setTimeout> | null = null
  let lastArgs: Parameters<T> | null = null
  let that: ThisParameterType<T> | null = null

  const invokeFn = () => {
    if (lastArgs && that) {
      fn.apply(that, lastArgs)
    }
    lastArgs = null
    that = null
    lastTime = Date.now()
  }

  return function (this: ThisParameterType<T>, ...args: Parameters<T>): void {
    const now = Date.now()

    if (lastTime === 0 && leading) {
      lastTime = now
      fn.apply(this, args)
      return
    }

    const remaining = interval - (now - lastTime)

    if (remaining <= 0 && !timer) {
      invokeFn()
    } else if (trailing && !timer) {
      timer = setTimeout(() => {
        if (trailing) invokeFn()
        timer = null
      }, remaining)
    }

    lastArgs = args
    that = this
  }
}
