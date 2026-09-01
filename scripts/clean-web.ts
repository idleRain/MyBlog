#!/usr/bin/env -S node --import tsx

import { existsSync, rmSync } from 'fs'
import { join } from 'path'
import { isMainModule } from './lib/is-main'

// 前端各应用待清理的构建产物目录，与 build.ts 的清理范围保持一致。
const FRONTEND_CLEAN_PATHS = [
  ['apps/web', '.svelte-kit'],
  ['apps/web', 'build'],
  ['apps/web', 'dist'],
  ['apps/admin', '.svelte-kit'],
  ['apps/admin', 'build'],
  ['apps/admin', 'dist']
]

// 清理全部前端应用的构建产物目录，不存在的目录自动跳过。
function cleanFrontendBuilds(): void {
  let cleanedCount = 0
  for (const [appDir, target] of FRONTEND_CLEAN_PATHS) {
    const targetPath = join(appDir, target)
    if (existsSync(targetPath)) {
      rmSync(targetPath, { recursive: true, force: true })
      console.log(`✅ 已清理: ${targetPath}`)
      cleanedCount += 1
    }
  }
  console.log(cleanedCount > 0 ? '✅ 前端产物清理完成' : '✅ 无前端产物需要清理')
}

// 直接运行时启动清理流程。
if (isMainModule(import.meta.url)) {
  cleanFrontendBuilds()
}
