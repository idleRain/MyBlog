#!/usr/bin/env -S node --import tsx

/**
 * 数据库种子数据管理工具
 * 用于初始化默认超级管理员账户，支持自定义用户名、密码和邮箱
 */

import { spawn } from 'child_process'
import { existsSync } from 'fs'
import path from 'path'

const SERVER_DIR = path.join(process.cwd(), 'server')
const SEED_ENTRY = 'cmd/seed/main.go'

// 颜色输出
const colors = {
  green: (text: string) => `\x1b[32m${text}\x1b[0m`,
  red: (text: string) => `\x1b[31m${text}\x1b[0m`,
  yellow: (text: string) => `\x1b[33m${text}\x1b[0m`,
  cyan: (text: string) => `\x1b[36m${text}\x1b[0m`,
  bold: (text: string) => `\x1b[1m${text}\x1b[0m`
}

// 在 server 目录下运行 Go 种子程序
function runSeed(args: string[]): Promise<void> {
  return new Promise((resolve, reject) => {
    console.log(colors.cyan(`执行: go run ${SEED_ENTRY} ${args.join(' ')}`))

    const child = spawn('go', ['run', SEED_ENTRY, ...args], {
      cwd: SERVER_DIR,
      stdio: 'inherit'
    })

    child.on('close', code => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(`种子数据初始化失败，退出码: ${code}`))
      }
    })

    child.on('error', error => {
      reject(error)
    })
  })
}

// 打印帮助信息
function showHelp(): void {
  console.log(colors.bold('\n📚 数据库种子数据工具'))
  console.log('\n用法:')
  console.log(
    colors.green('  pnpm run seed:admin [--username <用户>] [--password <密码>] [--email <邮箱>]')
  )
  console.log('\n参数:')
  console.log(colors.green('  --username    超级管理员用户名（默认 admin）'))
  console.log(colors.green('  --password    超级管理员密码（默认 Admin@123456）'))
  console.log(colors.green('  --email       超级管理员邮箱（默认 admin@myblog.local）'))
  console.log('\n示例:')
  console.log(colors.cyan('  pnpm run seed:admin'))
  console.log(colors.cyan('  pnpm run seed:admin --username root --password Root@2025'))
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
    console.log(colors.green('✅ 超级管理员初始化完成'))
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
