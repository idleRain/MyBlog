#!/usr/bin/env bun

import { $ } from 'bun'
import { existsSync, writeFileSync } from 'fs'
import { join } from 'path'

console.log('🚀 开始设置 MyBlog 开发环境...\n')

// 检查系统要求
async function checkSystemRequirements() {
  console.log('📋 检查系统要求...')

  try {
    // 检查 Bun 版本
    const bunVersion = Bun.version
    console.log(`✅ Bun: ${bunVersion}`)

    // 检查 Node.js 版本（如果存在）
    try {
      const nodeVersion = await $`node --version`.text()
      console.log(`✅ Node.js: ${nodeVersion.trim()}`)
    } catch {
      console.log('ℹ️  Node.js: 未安装（使用 Bun 运行时）')
    }

    // 检查 Go 版本
    try {
      const goVersion = await $`go version`.text()
      console.log(`✅ Go: ${goVersion.trim()}`)
    } catch (error) {
      console.error('❌ Go 未安装或不在 PATH 中')
      console.error('请安装 Go 1.23.11 或更高版本')
      process.exit(1)
    }

    // 检查 MySQL（可选）
    try {
      await $`mysql --version`.quiet()
      console.log('✅ MySQL: 已安装')
    } catch {
      console.log('⚠️  MySQL: 未检测到，请确保 MySQL 服务正在运行')
    }
  } catch (error: any) {
    console.error('❌ 系统要求检查失败:', error.message)
    process.exit(1)
  }

  console.log('')
}

// 安装依赖
async function installDependencies() {
  console.log('📦 安装依赖...')

  try {
    // 安装所有依赖 (包括 workspace 中的前端依赖)
    console.log('安装项目依赖 (monorepo + workspace)...')
    await $`bun install`

    // 安装后端依赖
    console.log('安装后端依赖...')
    await $`cd server && go mod tidy`

    console.log('✅ 所有依赖安装完成\n')
  } catch (error: any) {
    console.error('❌ 依赖安装失败:', error.message)
    process.exit(1)
  }
}

// 创建环境文件
function createEnvironmentFiles() {
  console.log('🔧 创建环境配置文件...')

  // 检查并创建前端 .env 文件
  const webEnvPath = join('web', '.env')
  if (!existsSync(webEnvPath)) {
    const webEnvContent = `# 前端环境配置
VITE_API_URL=http://localhost:3000
VITE_APP_TITLE=MyBlog
`
    writeFileSync(webEnvPath, webEnvContent)
    console.log('✅ 创建 web/.env')
  } else {
    console.log('✅ web/.env 已存在')
  }

  // 检查后端配置文件
  const serverConfigPath = join('server', 'configs', 'config.yaml')
  if (existsSync(serverConfigPath)) {
    console.log('✅ server/configs/config.yaml 已存在')
  } else {
    console.log('⚠️  server/configs/config.yaml 不存在，请检查后端配置')
  }

  console.log('')
}

// 验证并修复设置
async function validateSetup() {
  console.log('🔍 验证设置...')

  const checks = [
    { name: 'package.json', path: 'package.json' },
    { name: 'web/package.json', path: 'web/package.json' },
    { name: 'server/go.mod', path: 'server/go.mod' },
    { name: 'node_modules', path: 'node_modules' }
  ]

  let needsRepair = false

  checks.forEach(check => {
    if (existsSync(check.path)) {
      console.log(`✅ ${check.name}`)
    } else {
      console.log(`❌ ${check.name} 缺失`)
      needsRepair = true
    }
  })

  // 如果有缺失，尝试修复
  if (needsRepair) {
    console.log('\n🔧 尝试修复缺失的依赖...')

    // 修复根目录依赖
    if (!existsSync('node_modules')) {
      try {
        console.log('安装根目录依赖...')
        await $`bun install`
      } catch (error) {
        console.error('❌ 根目录依赖安装失败:', error)
        process.exit(1)
      }
    }


    // 重新验证
    console.log('\n🔍 重新验证设置...')
    let allValid = true
    checks.forEach(check => {
      if (existsSync(check.path)) {
        console.log(`✅ ${check.name}`)
      } else {
        console.log(`❌ ${check.name} 仍然缺失`)
        allValid = false
      }
    })

    if (!allValid) {
      console.log('\n❌ 设置验证失败，请手动检查问题')
      process.exit(1)
    }
  }

  console.log('\n🎉 环境设置完成！')
  console.log('\n📖 下一步:')
  console.log('  bun run dev    # 启动开发服务器')
  console.log('  bun run build  # 构建项目')
  console.log('  bun run test   # 运行测试')
}

// 主函数
async function main() {
  try {
    await checkSystemRequirements()
    await installDependencies()
    createEnvironmentFiles()
    await validateSetup()
  } catch (error: any) {
    console.error('❌ 设置过程中发生错误:', error.message)
    process.exit(1)
  }
}

// 如果直接运行此脚本
if (import.meta.main) {
  await main()
}
