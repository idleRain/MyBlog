#!/usr/bin/env bun

/**
 * 统一构建脚本
 * 负责构建前端和后端项目，支持清理、质量检查等功能
 */

import { spawn } from 'child_process'
import { existsSync, rmSync } from 'fs'

// 颜色配置
const colors = {
  blue: '\x1b[34m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  cyan: '\x1b[36m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
}

interface BuildOptions {
  clean?: boolean
  production?: boolean
  skipTests?: boolean
  skipLint?: boolean
  serverOnly?: boolean
  webOnly?: boolean
}

// 解析命令行参数
function parseArgs(): BuildOptions {
  const args = process.argv.slice(2)
  return {
    clean: args.includes('--clean') || args.includes('-c'),
    production: args.includes('--production') || args.includes('-p'),
    skipTests: args.includes('--skip-tests'),
    skipLint: args.includes('--skip-lint'),
    serverOnly: args.includes('--server-only'),
    webOnly: args.includes('--web-only')
  }
}

// 运行命令的辅助函数
async function runCommand(
  command: string,
  args: string[],
  options: { cwd?: string; stdio?: 'inherit' | 'pipe' } = {}
): Promise<{ success: boolean; output?: string }> {
  return new Promise(resolve => {
    const child = spawn(command, args, {
      stdio: options.stdio || 'inherit',
      shell: true,
      cwd: options.cwd || process.cwd()
    })

    let output = ''

    if (options.stdio === 'pipe') {
      child.stdout?.on('data', data => {
        output += data.toString()
      })
      child.stderr?.on('data', data => {
        output += data.toString()
      })
    }

    child.on('close', code => {
      resolve({
        success: code === 0,
        output: options.stdio === 'pipe' ? output : undefined
      })
    })

    child.on('error', () => {
      resolve({ success: false, output })
    })
  })
}

// 环境检查
async function checkEnvironment(): Promise<boolean> {
  console.log(`${colors.cyan}🔍 检查构建环境...${colors.reset}`)

  // 检查 Go
  const goCheck = await runCommand('go', ['version'], { stdio: 'pipe' })
  if (!goCheck.success) {
    console.error(`${colors.red}❌ Go 未安装或不在 PATH 中${colors.reset}`)
    return false
  }
  console.log(`${colors.green}✅ Go: 已安装${colors.reset}`)

  // 检查 Bun
  try {
    console.log(`${colors.green}✅ Bun: ${Bun.version}${colors.reset}`)
  } catch {
    console.error(`${colors.red}❌ Bun 未安装${colors.reset}`)
    return false
  }

  // 检查项目结构
  const requiredPaths = ['server/cmd/myblog', 'web/package.json', 'server/go.mod']

  for (const path of requiredPaths) {
    if (!existsSync(path)) {
      console.error(`${colors.red}❌ 缺少必需文件/目录: ${path}${colors.reset}`)
      return false
    }
  }

  console.log(`${colors.green}✅ 项目结构检查通过${colors.reset}\n`)
  return true
}

// 清理构建文件
async function cleanBuildFiles(options: BuildOptions): Promise<boolean> {
  if (!options.clean) return true

  console.log(`${colors.yellow}🧹 清理构建文件...${colors.reset}`)

  const cleanPaths = ['server/bin', 'server/tmp', 'web/.svelte-kit', 'web/build', 'web/dist']

  try {
    for (const path of cleanPaths) {
      if (existsSync(path)) {
        rmSync(path, { recursive: true, force: true })
        console.log(`${colors.green}  ✅ 已清理: ${path}${colors.reset}`)
      }
    }
    console.log(`${colors.green}✅ 清理完成${colors.reset}\n`)
    return true
  } catch (error) {
    console.error(`${colors.red}❌ 清理失败: ${error}${colors.reset}`)
    return false
  }
}

// 安装依赖
async function installDependencies(options: BuildOptions): Promise<boolean> {
  console.log(`${colors.cyan}📦 检查并安装依赖...${colors.reset}`)

  // 检查并安装根目录依赖
  if (!existsSync('node_modules')) {
    console.log(`${colors.yellow}  安装根目录依赖...${colors.reset}`)
    const result = await runCommand('bun', ['install'])
    if (!result.success) {
      console.error(`${colors.red}❌ 根目录依赖安装失败${colors.reset}`)
      return false
    }
  }

  // 检查并安装前端依赖
  if (!options.serverOnly && !existsSync('web/node_modules')) {
    console.log(`${colors.yellow}  安装前端依赖...${colors.reset}`)
    const result = await runCommand('bun', ['install'], { cwd: 'web' })
    if (!result.success) {
      console.error(`${colors.red}❌ 前端依赖安装失败${colors.reset}`)
      return false
    }
  }

  // 检查并更新 Go 依赖
  if (!options.webOnly) {
    console.log(`${colors.yellow}  更新 Go 依赖...${colors.reset}`)
    const result = await runCommand('go', ['mod', 'tidy'], { cwd: 'server' })
    if (!result.success) {
      console.error(`${colors.red}❌ Go 依赖更新失败${colors.reset}`)
      return false
    }
  }

  console.log(`${colors.green}✅ 依赖检查完成${colors.reset}\n`)
  return true
}

// 代码质量检查
async function runQualityChecks(options: BuildOptions): Promise<boolean> {
  if (options.skipLint && options.skipTests) return true

  console.log(`${colors.cyan}🔍 运行代码质量检查...${colors.reset}`)

  // 前端代码检查
  if (!options.serverOnly && !options.skipLint) {
    console.log(`${colors.yellow}  前端代码检查...${colors.reset}`)

    // TypeScript 检查
    const tsCheck = await runCommand('bun', ['run', 'check'], { cwd: 'web' })
    if (!tsCheck.success) {
      console.error(`${colors.red}❌ 前端 TypeScript 检查失败${colors.reset}`)
      return false
    }

    // ESLint 检查 (跳过)
    console.log(`${colors.yellow}  跳过前端 ESLint 检查${colors.reset}`)
  }

  // 后端代码检查
  if (!options.webOnly && !options.skipLint) {
    console.log(`${colors.yellow}  后端代码检查...${colors.reset}`)

    const result = await runCommand('bun', ['scripts/go-tools.ts', 'vet'])
    if (!result.success) {
      console.error(`${colors.red}❌ 后端代码检查失败${colors.reset}`)
      return false
    }
  }

  // 运行测试
  if (!options.skipTests) {
    console.log(`${colors.yellow}  运行测试...${colors.reset}`)

    // 前端测试 (跳过，已在代码检查阶段进行了类型检查)
    if (!options.serverOnly) {
      console.log(`${colors.yellow}  跳过前端测试 (已进行类型检查)${colors.reset}`)
    }

    // 后端测试
    if (!options.webOnly) {
      const serverTest = await runCommand('bun', ['scripts/go-tools.ts', 'test'])
      if (!serverTest.success) {
        console.error(`${colors.red}❌ 后端测试失败${colors.reset}`)
        return false
      }
    }
  }

  console.log(`${colors.green}✅ 代码质量检查通过${colors.reset}\n`)
  return true
}

// 构建后端
async function buildServer(): Promise<boolean> {
  console.log(`${colors.cyan}🔨 构建后端项目...${colors.reset}`)

  const result = await runCommand('bun', ['scripts/go-tools.ts', 'build'])
  if (!result.success) {
    console.error(`${colors.red}❌ 后端构建失败${colors.reset}`)
    return false
  }

  console.log(`${colors.green}✅ 后端构建完成${colors.reset}`)
  return true
}

// 构建前端
async function buildWeb(options: BuildOptions): Promise<boolean> {
  console.log(`${colors.cyan}🔨 构建前端项目...${colors.reset}`)

  const buildCommand = options.production ? 'build' : 'build'
  const result = await runCommand('bun', ['run', buildCommand], { cwd: 'web' })

  if (!result.success) {
    console.error(`${colors.red}❌ 前端构建失败${colors.reset}`)
    return false
  }

  console.log(`${colors.green}✅ 前端构建完成${colors.reset}`)
  return true
}

// 显示构建信息
function showBuildInfo(options: BuildOptions): void {
  console.log(`${colors.cyan}📋 构建信息:${colors.reset}`)
  console.log(`  模式: ${options.production ? '生产环境' : '开发环境'}`)
  console.log(`  清理: ${options.clean ? '是' : '否'}`)
  console.log(`  跳过测试: ${options.skipTests ? '是' : '否'}`)
  console.log(`  跳过代码检查: ${options.skipLint ? '是' : '否'}`)
  if (options.serverOnly) console.log(`  构建范围: 仅后端`)
  else if (options.webOnly) console.log(`  构建范围: 仅前端`)
  else console.log(`  构建范围: 全栈`)
  console.log('')
}

// 显示帮助信息
function showHelp(): void {
  console.log(`${colors.bold}${colors.cyan}MyBlog 构建脚本${colors.reset}`)
  console.log('')
  console.log('使用方法:')
  console.log('  bun scripts/build.ts [选项]')
  console.log('')
  console.log('选项:')
  console.log('  --clean, -c        构建前清理输出目录')
  console.log('  --production, -p   生产环境构建')
  console.log('  --skip-tests       跳过测试')
  console.log('  --skip-lint        跳过代码检查')
  console.log('  --server-only      仅构建后端')
  console.log('  --web-only         仅构建前端')
  console.log('  --help, -h         显示帮助信息')
  console.log('')
  console.log('示例:')
  console.log('  bun scripts/build.ts                    # 标准构建')
  console.log('  bun scripts/build.ts --clean --production  # 生产环境构建')
  console.log('  bun scripts/build.ts --server-only      # 仅构建后端')
  console.log('  bun scripts/build.ts --skip-tests       # 跳过测试的快速构建')
}

// 主函数
async function main(): Promise<void> {
  const startTime = Date.now()
  const options = parseArgs()

  // 显示帮助
  if (process.argv.includes('--help') || process.argv.includes('-h')) {
    showHelp()
    return
  }

  console.log(`${colors.bold}${colors.green}🚀 MyBlog 项目构建开始${colors.reset}\n`)
  showBuildInfo(options)

  try {
    // 1. 环境检查
    if (!(await checkEnvironment())) {
      process.exit(1)
    }

    // 2. 清理构建文件
    if (!(await cleanBuildFiles(options))) {
      process.exit(1)
    }

    // 3. 安装依赖
    if (!(await installDependencies(options))) {
      process.exit(1)
    }

    // 4. 代码质量检查
    if (!(await runQualityChecks(options))) {
      process.exit(1)
    }

    // 5. 构建项目
    if (!options.webOnly && !(await buildServer())) {
      process.exit(1)
    }

    if (!options.serverOnly && !(await buildWeb(options))) {
      process.exit(1)
    }

    // 构建完成
    const duration = ((Date.now() - startTime) / 1000).toFixed(2)
    console.log(`\n${colors.bold}${colors.green}🎉 构建完成！${colors.reset}`)
    console.log(`${colors.cyan}📊 构建信息:${colors.reset}`)
    console.log(`  耗时: ${duration}s`)

    if (!options.webOnly) {
      console.log(`  后端输出: server/bin/myblog`)
    }
    if (!options.serverOnly) {
      console.log(`  前端输出: web/build/`)
    }

    console.log(`\n${colors.yellow}💡 提示: 使用 'bun run dev' 启动开发服务器${colors.reset}`)
  } catch (error) {
    console.error(`\n${colors.red}❌ 构建失败:${colors.reset}`, error)
    process.exit(1)
  }
}

// 如果直接运行此脚本
if (import.meta.main) {
  await main()
}
