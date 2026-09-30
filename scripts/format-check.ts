#!/usr/bin/env -S node --import tsx

/**
 * 前端格式只读门禁：以 prettier --check 校验应用源码，不写入任何文件。
 *
 * 覆盖范围严格对齐仓库既有规范中 lint-staged 对 apps 源码的约束，并排除
 * 非手工维护的产物，避免门禁对生成内容误报：apps/web/src/lib/paraglide 为
 * paraglide 编译产物，已被 .gitignore 排除，且被 apps/web 的 eslint 配置显式忽略。
 *
 * packages 目录不在任何既有 format 门禁覆盖内，仓库 AGENTS.md 已明确记录其存在
 * 未按规范格式化的存量文件，因此本脚本不将 packages 纳入判定范围。
 */

import { runCommand } from './lib/run-command'
import { exit } from 'node:process'

// prettier 可执行文件名，经由 npx 解析工作区根安装的版本。
const PRETTIER_COMMAND = 'prettier'

// 参与判定的源码范围，两个应用共用同一组扩展名。
const SOURCE_EXTENSIONS = '{js,ts,jsx,tsx,svelte,css}'

// 各应用需要判定的源码目录。
const APP_SOURCE_DIRECTORIES = ['apps/web/src', 'apps/admin/src']

// paraglide 生成产物目录，该目录由构建过程写出，不纳入格式门禁。
const GENERATED_EXCLUSIONS = ['!apps/web/src/lib/paraglide/**']

// 组装 prettier --check 的目标参数：每个应用的源码 glob 加上生成产物反选。
function buildCheckTargets(): string[] {
  const sourceGlobs = APP_SOURCE_DIRECTORIES.map(
    directory => `${directory}/**/*.${SOURCE_EXTENSIONS}`
  )
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
