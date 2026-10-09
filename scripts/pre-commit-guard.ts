#!/usr/bin/env -S node --import tsx

/**
 * pre-commit 门禁入口，承载 husky 钩子中的测试编排。
 * 钩子脚本保持简短，受影响范围的判定与检查调度全部集中在此脚本内。
 *
 * 门禁按暂存区实际触及的范围条件触发：仅改动文档时零测试开销，
 * 改动 packages 时只跑对应工作区包，改动 server 时只跑受影响的 Go 包。
 */

import { tryRunCommand } from './lib/run-command'
import { join, posix, sep } from 'node:path'
import { existsSync } from 'node:fs'

// 仓库根目录，husky 钩子始终从仓库根触发。
const repositoryRoot: string = process.cwd()

// server 模块目录名，Go 检查在此目录下执行。
const serverDirectoryName = 'server'

// packages 目录前缀，用于判定提交是否触及前端工作区包。
const packagesDirectoryPrefix = 'packages'

// apps 目录前缀，用于判定提交是否触及两个 SvelteKit 应用。
const appsDirectoryPrefix = 'apps'

// 纳入 vitest 门禁的工作区包，与根 test:packages 的口径保持一致。
const gatedWorkspacePackages: string[] = ['packages/api', 'packages/auth', 'packages/http']

// 纳入 vitest 门禁的应用，与根 test:apps 的口径保持一致。
// 两应用各自持有页面状态模块与渲染管线的单测，提交前必须与 packages 同级把关。
const gatedApplicationPackages: string[] = ['apps/web', 'apps/admin']

// vitest 入口相对工作区包根目录的路径，pnpm 隔离布局下每个包各自持有。
const vitestEntryRelativePath = join('node_modules', 'vitest', 'vitest.mjs')

// Go 测试的空值关闭参数，用于规避实现变更时的陈旧缓存命中。
const goTestNoCacheFlag = '-count=1'

/**
 * 判定提交是否触及指定目录前缀下的文件。
 * @param stagedFiles 暂存区文件路径列表，统一使用正斜杠分隔。
 * @param pathPrefix 待匹配的目录前缀。
 */
function touchesDirectory(stagedFiles: string[], pathPrefix: string): boolean {
  const normalizedPrefix = `${pathPrefix}/`
  return stagedFiles.some(filePath => filePath.startsWith(normalizedPrefix))
}

/**
 * 收集暂存区文件列表，使用零终止符分隔以规避非 ASCII 路径的转义歧义。
 * 命令失败时返回空列表，使钩子在无法判定范围时不误报测试失败。
 */
async function collectStagedFiles(): Promise<string[]> {
  const result = await tryRunCommand('git', ['diff', '--cached', '--name-only', '-z'], {
    stdio: 'pipe',
    shell: false
  })
  if (!result.success || !result.output) {
    return []
  }
  return result.output
    .split('\0')
    .filter(filePath => filePath.length > 0)
    .map(filePath => filePath.split(sep).join(posix.sep))
}

/**
 * 由暂存区中的 Go 文件推导受影响的包导入路径。
 * 同一目录下的多个文件只产生一个待测包，避免重复执行同一包的测试。
 * @param stagedFiles 暂存区文件路径列表。
 */
function collectAffectedGoPackages(stagedFiles: string[]): string[] {
  const packagePaths = new Set<string>()
  for (const filePath of stagedFiles) {
    if (!filePath.startsWith(`${serverDirectoryName}/`) || !filePath.endsWith('.go')) {
      continue
    }
    const directoryPath = posix.dirname(filePath)
    packagePaths.add(`./${directoryPath.slice(serverDirectoryName.length + 1)}`)
  }
  return [...packagePaths]
}

/**
 * 逐包执行 vitest，任一包失败即终止。
 * 每个包在其自身目录下执行，使 vitest 解析到该包内的依赖与配置。
 * @param workspacePaths 待测试的工作区包相对路径列表。
 * @param label 失败提示中使用的工作区类别名称。
 */
async function runVitestInPackages(workspacePaths: string[], label: string): Promise<boolean> {
  for (const workspacePath of workspacePaths) {
    const packageDirectory = join(repositoryRoot, workspacePath)
    const vitestEntry = join(packageDirectory, vitestEntryRelativePath)
    if (!existsSync(vitestEntry)) {
      console.error(`❌ 未找到 ${workspacePath} 的 vitest 入口，请先执行 pnpm install`)
      return false
    }
    console.log(`🧪 运行 ${workspacePath} 单元测试`)
    const result = await tryRunCommand('node', [vitestEntry, 'run'], { cwd: packageDirectory })
    if (!result.success) {
      console.error(`❌ ${label} 单元测试未通过，提交已被阻止`)
      return false
    }
  }
  return true
}

/**
 * 在 server 目录下执行受影响包的 Go 测试。
 * @param goPackages 待测试的包导入路径列表。
 */
async function runAffectedGoTests(goPackages: string[]): Promise<boolean> {
  if (goPackages.length === 0) {
    return true
  }
  console.log(`🧪 运行受影响的 Go 包测试: ${goPackages.join(' ')}`)
  const result = await tryRunCommand('go', ['test', goTestNoCacheFlag, ...goPackages], {
    cwd: join(repositoryRoot, serverDirectoryName)
  })
  return result.success
}

/**
 * 按暂存区触及范围调度各项门禁，任一项失败即中断提交流程。
 */
async function runGuard(): Promise<void> {
  const stagedFiles = await collectStagedFiles()
  if (stagedFiles.length === 0) {
    return
  }

  if (touchesDirectory(stagedFiles, packagesDirectoryPrefix)) {
    const packageTestsPassed = await runVitestInPackages(gatedWorkspacePackages, 'packages')
    if (!packageTestsPassed) {
      process.exit(1)
    }
  }

  if (touchesDirectory(stagedFiles, appsDirectoryPrefix)) {
    const appTestsPassed = await runVitestInPackages(gatedApplicationPackages, 'apps')
    if (!appTestsPassed) {
      process.exit(1)
    }
  }

  const goTestsPassed = await runAffectedGoTests(collectAffectedGoPackages(stagedFiles))
  if (!goTestsPassed) {
    console.error('❌ 受影响的 Go 包测试未通过，提交已被阻止')
    process.exit(1)
  }
}

void runGuard()
