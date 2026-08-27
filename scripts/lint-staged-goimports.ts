#!/usr/bin/env bun

/**
 * lint-staged 专用的 goimports 执行入口
 * lint-staged 将匹配的文件列表作为参数传入，此脚本解析 goimports 路径后执行格式化
 */

import { ensureGoimports, runCommand } from './go-toolchain'

// 主函数
async function main(): Promise<void> {
  const goimportsPath: string = await ensureGoimports()
  await runCommand(goimportsPath, ['-w', ...process.argv.slice(2)])
}

void main()
