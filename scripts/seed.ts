#!/usr/bin/env -S node --import tsx

/**
 * 数据库种子数据管理工具
 * 用于初始化默认超级管理员账户，支持自定义用户名、密码和邮箱；
 * 加 --demo 时同时写入可浏览的演示数据（分类、标签、文章与评论）
 */

import { runCommand } from './lib/run-command'
import { colors } from './lib/terminal'
import { existsSync } from 'node:fs'
import path from 'node:path'

const SERVER_DIR = path.join(process.cwd(), 'server')
const SEED_ENTRY = 'cmd/seed/main.go'

// 在 server 目录下运行 Go 种子程序
function runSeed(args: string[]): Promise<void> {
  console.log(colors.cyan(`执行: go run ${SEED_ENTRY} ${args.join(' ')}`))
  return runCommand('go', ['run', SEED_ENTRY, ...args], { cwd: SERVER_DIR })
}

// 打印帮助信息
function showHelp(): void {
  console.log(colors.bold('\n📚 数据库种子数据工具'))
  console.log('\n用法:')
  console.log(
    colors.green(
      '  pnpm run seed:admin [--username <用户>] [--password <密码>] [--email <邮箱>] [--demo] [--force]'
    )
  )
  console.log('\n参数:')
  console.log(colors.green('  --username    超级管理员用户名（默认 admin）'))
  console.log(colors.green('  --password    超级管理员密码（默认 Admin@123456）'))
  console.log(colors.green('  --email       超级管理员邮箱（默认 admin@myblog.local）'))
  console.log(colors.green('  --demo        同时写入演示数据：分类、标签、文章与评论'))
  console.log(colors.green('  --force       允许在 release 模式下写入演示数据'))
  console.log('\n示例:')
  console.log(colors.cyan('  pnpm run seed:admin'))
  console.log(colors.cyan('  pnpm run seed:admin --username root --password Root@2025'))
  console.log(colors.cyan('  pnpm run seed:admin --demo'))
}

// 主函数
async function main(): Promise<void> {
  const args = process.argv.slice(2)

  if (args.includes('--help') || args.includes('-h')) {
    showHelp()
    return
  }

  if (!existsSync(path.join(SERVER_DIR, SEED_ENTRY))) {
    console.error(colors.red(`❌ 未找到种子程序入口: ${SEED_ENTRY}`))
    process.exit(1)
  }

  try {
    await runSeed(args)
    console.log(
      colors.green(
        args.includes('--demo') ? '✅ 超级管理员与演示数据初始化完成' : '✅ 超级管理员初始化完成'
      )
    )
  } catch (error) {
    console.error(colors.red('❌ 初始化失败:'), error)
    process.exit(1)
  }
}

// 运行主函数
main().catch(error => {
  console.error(colors.red('❌ 执行失败:'), error)
  process.exit(1)
})
