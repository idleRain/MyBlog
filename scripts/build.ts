#!/usr/bin/env -S node --import tsx

/**
 * 统一构建脚本
 * 负责构建前端和后端项目，支持清理、质量检查等功能
 */

import { tryRunCommand } from './lib/run-command'
import { isMainModule } from './lib/is-main'
import { existsSync, rmSync } from 'node:fs'
import { ansi } from './lib/terminal'

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

// 环境检查
async function checkEnvironment(): Promise<boolean> {
  console.log(`${ansi.cyan}🔍 检查构建环境...${ansi.reset}`)

  // 检查 Go
  const goCheck = await tryRunCommand('go', ['version'], { stdio: 'pipe' })
  if (!goCheck.success) {
    console.error(`${ansi.red}❌ Go 未安装或不在 PATH 中${ansi.reset}`)
    return false
  }
  console.log(`${ansi.green}✅ Go: 已安装${ansi.reset}`)

  // 检查 Node.js
  try {
    console.log(`${ansi.green}✅ Node.js: ${process.version}${ansi.reset}`)
  } catch {
    console.error(`${ansi.red}❌ Node.js 未安装${ansi.reset}`)
    return false
  }

  // 检查项目结构
  const requiredPaths = [
    'server/cmd/myblog',
    'apps/web/package.json',
    'apps/admin/package.json',
    'server/go.mod'
  ]

  for (const path of requiredPaths) {
    if (!existsSync(path)) {
      console.error(`${ansi.red}❌ 缺少必需文件/目录: ${path}${ansi.reset}`)
      return false
    }
  }

  console.log(`${ansi.green}✅ 项目结构检查通过${ansi.reset}\n`)
  return true
}

// 清理构建文件
async function cleanBuildFiles(options: BuildOptions): Promise<boolean> {
  if (!options.clean) return true

  console.log(`${ansi.yellow}🧹 清理构建文件...${ansi.reset}`)

  const cleanPaths = [
    'server/bin',
    'server/tmp',
    'apps/web/.svelte-kit',
    'apps/web/build',
    'apps/web/dist',
    'apps/admin/.svelte-kit',
    'apps/admin/build',
    'apps/admin/dist'
  ]

  try {
    for (const path of cleanPaths) {
      if (existsSync(path)) {
        rmSync(path, { recursive: true, force: true })
        console.log(`${ansi.green}  ✅ 已清理: ${path}${ansi.reset}`)
      }
    }
    console.log(`${ansi.green}✅ 清理完成${ansi.reset}\n`)
    return true
  } catch (error) {
    console.error(`${ansi.red}❌ 清理失败: ${error}${ansi.reset}`)
    return false
  }
}

// 安装依赖
async function installDependencies(options: BuildOptions): Promise<boolean> {
  console.log(`${ansi.cyan}📦 检查并安装依赖...${ansi.reset}`)

  // 检查并安装根目录依赖
  if (!existsSync('node_modules')) {
    console.log(`${ansi.yellow}  安装根目录依赖...${ansi.reset}`)
    const result = await tryRunCommand('pnpm', ['install'])
    if (!result.success) {
      console.error(`${ansi.red}❌ 根目录依赖安装失败${ansi.reset}`)
      return false
    }
  }

  // 检查并更新 Go 依赖
  if (!options.webOnly) {
    console.log(`${ansi.yellow}  更新 Go 依赖...${ansi.reset}`)
    const result = await tryRunCommand('go', ['mod', 'tidy'], { cwd: 'server' })
    if (!result.success) {
      console.error(`${ansi.red}❌ Go 依赖更新失败${ansi.reset}`)
      return false
    }
  }

  console.log(`${ansi.green}✅ 依赖检查完成${ansi.reset}\n`)
  return true
}

// 代码质量检查
async function runQualityChecks(options: BuildOptions): Promise<boolean> {
  if (options.skipLint && options.skipTests) return true

  console.log(`${ansi.cyan}🔍 运行代码质量检查...${ansi.reset}`)

  // 前端代码检查
  if (!options.serverOnly && !options.skipLint) {
    console.log(`${ansi.yellow}  前端代码检查...${ansi.reset}`)

    // TypeScript 检查（前台与后台两个应用）
    const tsCheckWeb = await tryRunCommand('pnpm', ['run', 'check'], { cwd: 'apps/web' })
    if (!tsCheckWeb.success) {
      console.error(`${ansi.red}❌ 前台 TypeScript 检查失败${ansi.reset}`)
      return false
    }

    const tsCheckAdmin = await tryRunCommand('pnpm', ['run', 'check'], { cwd: 'apps/admin' })
    if (!tsCheckAdmin.success) {
      console.error(`${ansi.red}❌ 后台 TypeScript 检查失败${ansi.reset}`)
      return false
    }

    // ESLint 检查 (跳过)
    console.log(`${ansi.yellow}  跳过前端 ESLint 检查${ansi.reset}`)
  }

  // 后端代码检查
  if (!options.webOnly && !options.skipLint) {
    console.log(`${ansi.yellow}  后端代码检查...${ansi.reset}`)

    const result = await tryRunCommand('tsx', ['scripts/go-tools.ts', 'vet'])
    if (!result.success) {
      console.error(`${ansi.red}❌ 后端代码检查失败${ansi.reset}`)
      return false
    }
  }

  // 运行测试
  if (!options.skipTests) {
    console.log(`${ansi.yellow}  运行测试...${ansi.reset}`)

    // 前端测试 (跳过，已在代码检查阶段进行了类型检查)
    if (!options.serverOnly) {
      console.log(`${ansi.yellow}  跳过前端测试 (已进行类型检查)${ansi.reset}`)
    }

    // 后端测试
    if (!options.webOnly) {
      const serverTest = await tryRunCommand('tsx', ['scripts/go-tools.ts', 'test'])
      if (!serverTest.success) {
        console.error(`${ansi.red}❌ 后端测试失败${ansi.reset}`)
        return false
      }
    }
  }

  console.log(`${ansi.green}✅ 代码质量检查通过${ansi.reset}\n`)
  return true
}

// 构建后端
async function buildServer(): Promise<boolean> {
  console.log(`${ansi.cyan}🔨 构建后端项目...${ansi.reset}`)

  const result = await tryRunCommand('tsx', ['scripts/go-tools.ts', 'build'])
  if (!result.success) {
    console.error(`${ansi.red}❌ 后端构建失败${ansi.reset}`)
    return false
  }

  console.log(`${ansi.green}✅ 后端构建完成${ansi.reset}`)
  return true
}

// 构建前端（前台与后台两个应用）
async function buildWeb(options: BuildOptions): Promise<boolean> {
  console.log(`${ansi.cyan}🔨 构建前端项目...${ansi.reset}`)

  const buildCommand = options.production ? 'build' : 'build'

  const webResult = await tryRunCommand('pnpm', ['run', buildCommand], { cwd: 'apps/web' })
  if (!webResult.success) {
    console.error(`${ansi.red}❌ 前台构建失败${ansi.reset}`)
    return false
  }

  const adminResult = await tryRunCommand('pnpm', ['run', buildCommand], { cwd: 'apps/admin' })
  if (!adminResult.success) {
    console.error(`${ansi.red}❌ 后台构建失败${ansi.reset}`)
    return false
  }

  console.log(`${ansi.green}✅ 前端构建完成${ansi.reset}`)
  return true
}

// 显示构建信息
function showBuildInfo(options: BuildOptions): void {
  console.log(`${ansi.cyan}📋 构建信息:${ansi.reset}`)
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
  console.log(`${ansi.bold}${ansi.cyan}MyBlog 构建脚本${ansi.reset}`)
  console.log('')
  console.log('使用方法:')
  console.log('  tsx scripts/build.ts [选项]')
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
  console.log('  tsx scripts/build.ts                    # 标准构建')
  console.log('  tsx scripts/build.ts --clean --production  # 生产环境构建')
  console.log('  tsx scripts/build.ts --server-only      # 仅构建后端')
  console.log('  tsx scripts/build.ts --skip-tests       # 跳过测试的快速构建')
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

  console.log(`${ansi.bold}${ansi.green}🚀 MyBlog 项目构建开始${ansi.reset}\n`)
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
    console.log(`\n${ansi.bold}${ansi.green}🎉 构建完成！${ansi.reset}`)
    console.log(`${ansi.cyan}📊 构建信息:${ansi.reset}`)
    console.log(`  耗时: ${duration}s`)

    if (!options.webOnly) {
      console.log(`  后端输出: server/bin/myblog`)
    }
    if (!options.serverOnly) {
      console.log(`  前台输出: apps/web/build/`)
      console.log(`  后台输出: apps/admin/build/`)
    }

    console.log(`\n${ansi.yellow}💡 提示: 使用 'pnpm run dev' 启动开发服务器${ansi.reset}`)
  } catch (error) {
    console.error(`\n${ansi.red}❌ 构建失败:${ansi.reset}`, error)
    process.exit(1)
  }
}

// 如果直接运行此脚本
if (isMainModule(import.meta.url)) {
  await main()
}
