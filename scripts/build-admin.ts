#!/usr/bin/env -S node --import tsx

/**
 * 后台应用生产构建：等价于在 apps/admin 目录执行 pnpm run build。
 *
 * 独立入口的目的是让 CI 能以单条命令对单个应用做生产构建断言，而不复用根级
 * build:web。根级 build:web 走 scripts/build.ts，会附带依赖检查、类型检查与
 * 后端测试，构建错误的判定会被同批其他失败掩盖，无法定位到具体应用。
 */

import { runCommand } from './lib/run-command'

// 后台应用的 pnpm 包名，供 --filter 精确定位，避免依赖目录约定。
const ADMIN_FILTER = '@myblog/admin'

// 工作区包必须显式运行的脚本名。
const BUILD_SCRIPT = 'build'

async function main(): Promise<void> {
  console.log(`🏗️  构建后台应用 ${ADMIN_FILTER}...`)
  await runCommand('pnpm', ['--filter', ADMIN_FILTER, 'run', BUILD_SCRIPT])
  console.log('✅ 后台应用构建完成')
}

try {
  await main()
} catch (error) {
  const errorMessage = error instanceof Error ? error.message : String(error)
  console.error(`❌ 后台应用构建失败: ${errorMessage}`)
  process.exit(1)
}
