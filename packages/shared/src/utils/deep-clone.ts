/**
 * 深拷贝对象，支持日期、正则、数组缓冲、Map 与 Set。
 * visited 记录已访问对象，防止循环引用导致无限递归。
 */
export const deepClone = <T>(source: T, visited = new WeakMap()): T => {
  // 基础类型直接返回。
  if (typeof source !== 'object' || source === null) return source
  if (visited.has(source)) return visited.get(source)
  // 处理内置对象。
  if (source instanceof Date) return new Date(source.getTime()) as T
  if (source instanceof RegExp) return new RegExp(source) as T
  if (source instanceof ArrayBuffer) return source.slice(0) as T
  if (ArrayBuffer.isView(source))
    return new (source as any).constructor(source.buffer.slice(0)) as T
  // Map 与 Set 递归处理。
  if (source instanceof Map) {
    const cloned = new Map()
    visited.set(source, cloned)
    for (const [key, value] of source.entries()) {
      cloned.set(deepClone(key, visited), deepClone(value, visited))
    }
    return cloned as T
  }
  if (source instanceof Set) {
    const cloned = new Set()
    visited.set(source, cloned)
    for (const value of source) {
      cloned.add(deepClone(value, visited))
    }
    return cloned as T
  }
  // 数组或普通对象递归克隆所有可枚举属性。
  const cloned: Array<any> | Record<any, any> = Array.isArray(source) ? [] : {}
  visited.set(source, cloned)
  for (const key in source) {
    if (Object.prototype.hasOwnProperty.call(source, key)) {
      cloned[key] = deepClone(source[key], visited)
    }
  }
  return cloned as T
}
