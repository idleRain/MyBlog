#!/usr/bin/env -S node --import tsx

/**
 * 数据库迁移管理工具
 * 基于 golang-migrate 的数据库迁移脚本管理
 */

import { runCommand } from './lib/run-command'
import { colors } from './lib/terminal'
import { existsSync } from 'node:fs'
import path from 'node:path'

const SERVER_DIR = path.join(process.cwd(), 'server')
const MIGRATIONS_DIR = path.join(SERVER_DIR, 'migrations')

// 执行命令的辅助函数：以 server 目录为工作目录运行迁移 CLI，复用统一的命令执行器。
function executeCommand(command: string, args: string[], cwd: string = SERVER_DIR): Promise<void> {
  return runCommand(command, args, { cwd })
}

// 检查环境
function checkEnvironment() {
  if (!existsSync(SERVER_DIR)) {
    console.error(colors.red('❌ server 目录不存在'))
    process.exit(1)
  }

  if (!existsSync(MIGRATIONS_DIR)) {
    console.error(colors.red('❌ migrations 目录不存在'))
    process.exit(1)
  }
}

// 数据库连接字符串（从配置文件读取或使用默认值）
const DB_URL = process.env.DATABASE_URL || 'mysql://root:123456@tcp(localhost:3306)/blog'

// 可用命令
const commands = {
  // 创建新的迁移文件
  async create(name: string) {
    if (!name) {
      console.error(colors.red('❌ 请提供迁移文件名'))
      console.log(colors.yellow('用法: tsx scripts/migrate.ts create <migration_name>'))
      process.exit(1)
    }

    try {
      await executeCommand('migrate', [
        'create',
        '-ext',
        'sql',
        '-dir',
        './migrations',
        '-seq',
        name
      ])
      console.log(colors.green(`✅ 迁移文件创建成功: ${name}`))
    } catch (error) {
      console.error(colors.red('❌ 创建迁移文件失败:'), error)
      process.exit(1)
    }
  },

  // 运行迁移（升级到最新版本）
  async up() {
    try {
      await executeCommand('migrate', ['-path', './migrations', '-database', DB_URL, 'up'])
      console.log(colors.green('✅ 数据库迁移完成'))
    } catch (error) {
      console.error(colors.red('❌ 数据库迁移失败:'), error)
      process.exit(1)
    }
  },

  // 回滚迁移
  async down(steps?: string) {
    const stepsNum = steps ? parseInt(steps) : 1

    try {
      await executeCommand('migrate', [
        '-path',
        './migrations',
        '-database',
        DB_URL,
        'down',
        stepsNum.toString()
      ])
      console.log(colors.green(`✅ 成功回滚 ${stepsNum} 步迁移`))
    } catch (error) {
      console.error(colors.red('❌ 迁移回滚失败:'), error)
      process.exit(1)
    }
  },

  // 迁移到指定版本
  async goto(version: string) {
    if (!version) {
      console.error(colors.red('❌ 请提供目标版本号'))
      console.log(colors.yellow('用法: tsx scripts/migrate.ts goto <version>'))
      process.exit(1)
    }

    try {
      await executeCommand('migrate', [
        '-path',
        './migrations',
        '-database',
        DB_URL,
        'goto',
        version
      ])
      console.log(colors.green(`✅ 成功迁移到版本 ${version}`))
    } catch (error) {
      console.error(colors.red('❌ 迁移到指定版本失败:'), error)
      process.exit(1)
    }
  },

  // 强制设置版本（修复dirty状态）
  async force(version: string) {
    if (!version) {
      console.error(colors.red('❌ 请提供版本号'))
      console.log(colors.yellow('用法: tsx scripts/migrate.ts force <version>'))
      process.exit(1)
    }

    console.log(colors.yellow('⚠️  警告：此操作将强制设置迁移版本，可能导致数据不一致'))

    try {
      await executeCommand('migrate', [
        '-path',
        './migrations',
        '-database',
        DB_URL,
        'force',
        version
      ])
      console.log(colors.green(`✅ 成功强制设置版本为 ${version}`))
    } catch (error) {
      console.error(colors.red('❌ 强制设置版本失败:'), error)
      process.exit(1)
    }
  },

  // 查看当前版本
  async version() {
    try {
      await executeCommand('migrate', ['-path', './migrations', '-database', DB_URL, 'version'])
    } catch (error) {
      console.error(colors.red('❌ 获取版本信息失败:'), error)
      process.exit(1)
    }
  },

  // 删除所有表（慎用）
  async drop() {
    console.log(colors.red('⚠️  警告：此操作将删除所有数据库表，不可恢复！'))

    try {
      await executeCommand('migrate', ['-path', './migrations', '-database', DB_URL, 'drop'])
      console.log(colors.green('✅ 数据库表已删除'))
    } catch (error) {
      console.error(colors.red('❌ 删除数据库表失败:'), error)
      process.exit(1)
    }
  },

  // 显示帮助信息
  help() {
    console.log(colors.bold('\n📚 数据库迁移管理工具'))
    console.log('\n可用命令:')
    console.log(colors.green('  create <name>    ') + '创建新的迁移文件')
    console.log(colors.green('  up               ') + '运行所有待执行的迁移')
    console.log(colors.green('  down [steps]     ') + '回滚迁移（默认1步）')
    console.log(colors.green('  goto <version>   ') + '迁移到指定版本')
    console.log(colors.green('  force <version>  ') + '强制设置版本（修复dirty状态）')
    console.log(colors.green('  version          ') + '查看当前迁移版本')
    console.log(colors.green('  drop             ') + '删除所有数据库表（慎用）')
    console.log(colors.green('  help             ') + '显示此帮助信息')

    console.log('\n示例:')
    console.log(colors.cyan('  tsx scripts/migrate.ts create add_user_table'))
    console.log(colors.cyan('  tsx scripts/migrate.ts up'))
    console.log(colors.cyan('  tsx scripts/migrate.ts down 2'))
    console.log(colors.cyan('  tsx scripts/migrate.ts goto 1'))
    console.log(colors.cyan('  tsx scripts/migrate.ts version'))

    console.log('\n环境变量:')
    console.log(colors.yellow('  DATABASE_URL     ') + '数据库连接字符串')
    console.log(colors.yellow('                   ') + `当前: ${DB_URL}`)
  }
}

// 主函数
async function main() {
  const [command, ...args] = process.argv.slice(2)

  // 检查环境
  checkEnvironment()

  // 处理命令
  switch (command) {
    case 'create':
      await commands.create(args[0])
      break
    case 'up':
      await commands.up()
      break
    case 'down':
      await commands.down(args[0])
      break
    case 'goto':
      await commands.goto(args[0])
      break
    case 'force':
      await commands.force(args[0])
      break
    case 'version':
      await commands.version()
      break
    case 'drop':
      await commands.drop()
      break
    case 'help':
    case '--help':
    case '-h':
      commands.help()
      break
    default:
      if (!command) {
        commands.help()
      } else {
        console.error(colors.red(`❌ 未知命令: ${command}`))
        console.log(colors.yellow('使用 "help" 查看可用命令'))
        process.exit(1)
      }
  }
}

// 运行主函数
main().catch(error => {
  console.error(colors.red('❌ 执行失败:'), error)
  process.exit(1)
})
