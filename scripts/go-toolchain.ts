#!/usr/bin/env -S node --import tsx

/**
 * Go 工具链管理脚本
 * 负责 golangci-lint 与 goimports 的检测、安装与路径解析
 * 供 go-tools.ts 导入复用，也可独立执行以确保工具就绪
 */

import { execSync, spawn, type SpawnOptions } from 'child_process'
import { existsSync } from 'fs'
import { platform } from 'os'
import { join } from 'path'
import { isMainModule } from './lib/is-main'

// golangci-lint 工具模块地址，安装时自动获取最新稳定版本
export const golangciLintModule: string =
  'github.com/golangci/golangci-lint/cmd/golangci-lint@latest'

// goimports 工具模块地址，安装时自动获取最新稳定版本
export const goimportsModule: string = 'golang.org/x/tools/cmd/goimports@latest'

// golangci-lint 单次检查的超时时间，冷启动加载依赖时耗时较长
export const golangciLintTimeout: string = '5m'

// 运行命令的选项接口
interface RunCommandOptions extends Partial<SpawnOptions> {
  stdio?: 'inherit' | 'ignore' | 'pipe'
  cwd?: string
}

// 运行命令的辅助函数，供本脚本与导入方共用
export function runCommand(
  command: string,
  args: string[],
  options: RunCommandOptions = {}
): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: options.stdio || 'inherit',
      shell: true,
      cwd: options.cwd || process.cwd(),
      ...options
    })

    child.on('close', (code: number | null) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(`命令 "${command} ${args.join(' ')}" 执行失败，退出码: ${code}`))
      }
    })

    child.on('error', (error: Error) => {
      reject(new Error(`无法执行命令 "${command}": ${error.message}`))
    })
  })
}

// 获取 Go 可执行文件安装目录，优先使用 GOBIN，未设置时退回 GOPATH/bin
function getGoBinDir(): string {
  const goBin: string = execSync('go env GOBIN').toString().trim()
  if (goBin) {
    return goBin
  }
  const goPath: string = execSync('go env GOPATH').toString().trim()
  return join(goPath, 'bin')
}

// 获取当前平台下的可执行文件后缀，Windows 平台为 .exe
function getExecutableSuffix(): string {
  return platform() === 'win32' ? '.exe' : ''
}

// 解析指定工具的可执行文件路径，优先使用 PATH 中的已安装版本
function resolveTool(toolName: string, versionCommand: string): string | null {
  try {
    execSync(versionCommand, { stdio: 'ignore' })
    return toolName
  } catch (error) {
    const goBinPath: string = join(getGoBinDir(), `${toolName}${getExecutableSuffix()}`)
    return existsSync(goBinPath) ? goBinPath : null
  }
}

// 确保 golangci-lint 可用并返回可执行文件路径，未安装时自动安装
export async function ensureGolangciLint(): Promise<string> {
  const resolvedPath: string | null = resolveTool('golangci-lint', 'golangci-lint --version')
  if (resolvedPath) {
    return resolvedPath
  }
  console.log('正在下载并安装 golangci-lint...')
  await runCommand('go', ['install', golangciLintModule])
  const installedPath: string | null = resolveTool('golangci-lint', 'golangci-lint --version')
  if (!installedPath) {
    throw new Error('golangci-lint 安装失败，请检查 Go 环境后重试')
  }
  console.log('✅ golangci-lint 安装完成')
  return installedPath
}

// 确保 goimports 可用并返回可执行文件路径，未安装时自动安装
export async function ensureGoimports(): Promise<string> {
  const resolvedPath: string | null = resolveTool('goimports', 'goimports --help')
  if (resolvedPath) {
    return resolvedPath
  }
  console.log('正在下载并安装 goimports...')
  await runCommand('go', ['install', goimportsModule])
  const installedPath: string | null = resolveTool('goimports', 'goimports --help')
  if (!installedPath) {
    throw new Error('goimports 安装失败，请检查 Go 环境后重试')
  }
  console.log('✅ goimports 安装完成')
  return installedPath
}

// 显示帮助
function showHelp(): void {
  console.log('Go 工具链脚本命令:')
  console.log('  ensure    - 确保 golangci-lint 与 goimports 均已就绪')
  console.log('  golangci  - 仅确保 golangci-lint 已就绪')
  console.log('  goimports - 仅确保 goimports 已就绪')
}

// 主函数
async function main(): Promise<void> {
  const subcommand: string = process.argv[2]
  try {
    switch (subcommand) {
      case 'ensure':
        await ensureGolangciLint()
        await ensureGoimports()
        break
      case 'golangci':
        await ensureGolangciLint()
        break
      case 'goimports':
        await ensureGoimports()
        break
      default:
        showHelp()
        break
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    console.error(`❌ 错误: ${errorMessage}`)
    process.exit(1)
  }
}

// 仅在直接运行时执行命令分发，作为模块被导入时不触发
if (isMainModule(import.meta.url)) {
  void main()
}
