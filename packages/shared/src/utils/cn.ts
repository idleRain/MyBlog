import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * 合并 Tailwind 类名，后者覆盖前者的冲突类。
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T

export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null }
