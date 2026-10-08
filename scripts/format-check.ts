#!/usr/bin/env -S node --import tsx

/**
 * 前端格式只读门禁：以 prettier --check 校验应用源码，不写入任何文件。
 *
 * 覆盖范围严格对齐仓库既有规范：apps 源码对齐 lint-staged 的约束，公共包
 * 源码在存量格式化完成后一并纳入判定。生成内容通过排除避免误报且已被
 * .gitignore 排除：apps/web/src/lib/paraglide 为 paraglide 编译产物，
 * apps/web 的 eslint 配置同样显式忽略；packages/ui 与根目录生成产物
 * 由根 .prettierignore 统一排除，脚本内无需重复反选。
 */

import { runCommand } from './lib/run-command'
import { exit } from 'node:process'

// prettier 可执行文件名，经由 npx 解析工作区根安装的版本。
const PRETTIER_COMMAND = 'prettier'

// 参与判定的源码范围，所有源码目录共用同一组扩展名。
const SOURCE_EXTENSIONS = '{js,ts,jsx,tsx,svelte,css,json}'

// 各应用需要判定的源码目录。
const APP_SOURCE_DIRECTORIES = ['apps/web/src', 'apps/admin/src']

// 已纳入判定的公共包目录；packages/ui 属生成代码目录，经 .prettierignore 排除。
const PACKAGE_SOURCE_DIRECTORIES = [
  'packages/shared',
  'packages/http',
  'packages/api',
  'packages/auth'
]

// paraglide 生成产物目录，该目录由构建过程写出，不纳入格式门禁。
const GENERATED_EXCLUSIONS = ['!apps/web/src/lib/paraglide/**']

// 组装 prettier --check 的目标参数：应用与公共包的源码 glob 加上生成产物反选。
function buildCheckTargets(): string[] {
  const sourceDirectories = [...APP_SOURCE_DIRECTORIES, ...PACKAGE_SOURCE_DIRECTORIES]
  const sourceGlobs = sourceDirectories.map(directory => `${directory}/**/*.${SOURCE_EXTENSIONS}`)
  return [...sourceGlobs, ...GENERATED_EXCLUSIONS]
}

async function main(): Promise<void> {
  const targets = buildCheckTargets()
  console.log('🎨 运行前端格式只读检查...')
  console.log(`检查范围: ${targets.join(' ')}`)

  await runCommand(PRETTIER_COMMAND, ['--check', ...targets])
  console.log('✅ 前端格式检查通过')
}

try {
  await main()
} catch (error) {
  const errorMessage = error instanceof Error ? error.message : String(error)
  console.error(`❌ 前端格式检查失败: ${errorMessage}`)
  exit(1)
}
