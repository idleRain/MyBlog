#!/usr/bin/env bun

/**
 * 跨平台 Go 工具脚本
 * 此脚本提供与 Makefile 相同的功能，但可在 Windows 环境中运行
 * 专为 Bun 运行时环境优化
 */

import { spawn, execSync, type SpawnOptions } from 'child_process'
import { join } from 'path'
import { existsSync, mkdirSync } from 'fs'
import { platform } from 'os'

// 支持的命令类型
type Command =
  | 'build'
  | 'test'
  | 'deps'
  | 'lint-install'
  | 'lint'
  | 'format'
  | 'fmt'
  | 'vet'
  | 'clean'
  | 'quality'
  | 'quality-check'

// 运行命令的选项接口
interface RunCommandOptions extends Partial<SpawnOptions> {
  stdio?: 'inherit' | 'ignore' | 'pipe'
  cwd?: string
}

// 命令行参数
const command: string = process.argv[2]
const serverDir: string = join(process.cwd(), 'server')

// 检查是否安装了 Go
function checkGo(): boolean {
  try {
    execSync('go version', { stdio: 'ignore' })
    return true
  } catch (error) {
    console.error('❌ 错误: 未安装 Go 或 Go 不在 PATH 中')
    console.error('请安装 Go: https://golang.org/doc/install')
    return false
  }
}

// 运行命令的辅助函数
function runCommand(
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

// 构建项目
async function buildProject(): Promise<void> {
  console.log('🔨 编译项目...')

  // 创建 bin 目录
  const binDir: string = join(serverDir, 'bin')
  if (!existsSync(binDir)) {
    mkdirSync(binDir, { recursive: true })
  }

  // 构建项目
  const outputFile: string = platform() === 'win32' ? 'bin\\myblog.exe' : 'bin/myblog'
  await runCommand('go', ['build', '-o', outputFile, './cmd/myblog'], { cwd: serverDir })

  console.log(`✅ 编译完成: ${outputFile}`)
}

// 运行测试
async function runTests(): Promise<void> {
  console.log('🧪 运行测试...')
  await runCommand('go', ['test', '-v', './...'], { cwd: serverDir })
  console.log('✅ 测试完成')
}

// 安装/更新依赖
async function updateDeps(): Promise<void> {
  console.log('📦 安装/更新依赖...')
  await runCommand('go', ['mod', 'tidy'], { cwd: serverDir })
  await runCommand('go', ['mod', 'download'], { cwd: serverDir })
  console.log('✅ 依赖更新完成')
}

// 安装 golangci-lint
async function installLint(): Promise<void> {
  console.log('📦 安装 golangci-lint 工具...')

  // 检查是否已安装
  try {
    await runCommand('golangci-lint', ['--version'], { stdio: 'ignore' })
    console.log('golangci-lint 已安装')
  } catch (error) {
    console.log('正在下载并安装 golangci-lint...')

    if (platform() === 'win32') {
      // Windows 安装方法
      console.log(
        '请访问 https://golangci-lint.run/usage/install/#windows 下载并安装 golangci-lint'
      )
      console.log('或者使用 go install 安装:')
      await runCommand('go', [
        'install',
        'github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2'
      ])
    } else {
      // Linux/macOS 安装方法
      const installScript =
        'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2'
      await runCommand(installScript, [], { shell: true })
    }
  }

  console.log('✅ golangci-lint 安装完成')
}

// 运行 lint
async function runLint(): Promise<void> {
  console.log('🔍 运行 golangci-lint 代码检查...')

  // 检查是否已安装
  try {
    await runCommand('golangci-lint', ['--version'], { stdio: 'ignore' })
  } catch (error) {
    console.log('📦 golangci-lint 未安装，正在安装...')
    await installLint()
  }

  await runCommand('golangci-lint', ['run'], { cwd: serverDir })
  console.log('✅ 代码检查完成')
}

// 格式化代码
async function formatCode(): Promise<void> {
  console.log('🎨 格式化代码...')

  // 运行 gofmt
  console.log('运行 gofmt...')
  await runCommand('go', ['fmt', './...'], { cwd: serverDir })

  // 检查并安装 goimports
  console.log('运行 goimports...')
  try {
    await runCommand('goimports', ['--help'], { stdio: 'ignore' })
  } catch (error) {
    console.log('📦 正在安装 goimports...')
    await runCommand('go', ['install', 'golang.org/x/tools/cmd/goimports@latest'])
  }

  // 运行 goimports
  await runCommand('goimports', ['-w', '.'], { cwd: serverDir })
  console.log('✅ 代码格式化完成')
}

// 代码检查
async function vetCode(): Promise<void> {
  console.log('🔍 代码检查...')
  await runCommand('go', ['vet', './...'], { cwd: serverDir })
  console.log('✅ 代码检查完成')
}

// 清理临时文件
async function cleanFiles(): Promise<void> {
  console.log('🧹 清理临时文件...')

  const dirsToClean: string[] = ['tmp', 'bin', 'logs'].map(dir => join(serverDir, dir))

  for (const dir of dirsToClean) {
    if (existsSync(dir)) {
      if (platform() === 'win32') {
        await runCommand('rmdir', ['/s', '/q', dir], { stdio: 'ignore', shell: true })
      } else {
        await runCommand('rm', ['-rf', dir])
      }
    }
  }

  console.log('✅ 清理完成')
}

// 运行完整代码质量检查
async function qualityCheck(): Promise<void> {
  await formatCode()
  await runLint()
  await vetCode()
  await runTests()
  console.log('✅ 完整代码质量检查完成')
}

// 显示帮助
function showHelp(): void {
  console.log('Go 工具脚本命令:')
  console.log('  build       - 编译项目')
  console.log('  test        - 运行测试')
  console.log('  deps        - 安装/更新依赖')
  console.log('  lint-install- 安装 golangci-lint 工具')
  console.log('  lint        - 运行代码检查 (golangci-lint)')
  console.log('  format      - 格式化代码 (gofmt + goimports)')
  console.log('  vet         - 运行 go vet 代码检查')
  console.log('  clean       - 清理临时文件')
  console.log('  quality     - 运行完整代码质量检查')
}

// 主函数
async function main(): Promise<void> {
  // 检查 Go 是否安装
  if (!checkGo()) {
    process.exit(1)
  }

  // 处理命令
  try {
    switch (command as Command) {
      case 'build':
        await buildProject()
        break
      case 'test':
        await runTests()
        break
      case 'deps':
        await updateDeps()
        break
      case 'lint-install':
        await installLint()
        break
      case 'lint':
        await runLint()
        break
      case 'format':
      case 'fmt':
        await formatCode()
        break
      case 'vet':
        await vetCode()
        break
      case 'clean':
        await cleanFiles()
        break
      case 'quality':
      case 'quality-check':
        await qualityCheck()
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

// 运行主函数
void main()
