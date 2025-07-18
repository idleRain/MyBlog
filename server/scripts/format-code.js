#!/usr/bin/env bun

/**
 * 跨平台 Go 代码格式化脚本
 * 此脚本自动检测操作系统并运行相应的格式化脚本
 * 专为 Bun 运行时环境优化
 */

import { execSync, spawn } from 'child_process'

// 检查是否在正确的目录
const goModPath = path.join(process.cwd(), 'go.mod')
if (!fs.existsSync(goModPath)) {
  console.error('❌ 错误: 请在 Go 项目根目录运行此脚本')
  process.exit(1)
}

console.log('🎨 开始格式化 Go 代码...')

// 格式化函数
async function formatCode() {
  try {
    // 运行 gofmt
    console.log('📝 运行 gofmt 格式化代码...')
    await runCommand('go', ['fmt', './...'])
    console.log('✅ gofmt 格式化完成')

    // 检查并安装 goimports
    try {
      await runCommand('goimports', ['--help'], { stdio: 'ignore' })
    } catch (error) {
      console.log('📦 goimports 未安装，正在安装...')
      await runCommand('go', ['install', 'golang.org/x/tools/cmd/goimports@latest'])
      console.log('✅ goimports 安装完成')
    }

    // 运行 goimports
    console.log('📝 运行 goimports 整理导入...')
    await runCommand('goimports', ['-w', '.'])
    console.log('✅ goimports 整理完成')

    console.log('🎉 Go 代码格式化全部完成！')

    // 显示统计信息
    const goFiles = await countGoFiles()
    console.log('')
    console.log('📊 格式化统计:')
    console.log(`   - 已处理的 Go 文件数量: ${goFiles}`)
    console.log('   - 跳过的目录: vendor/, tmp/, .git/')

  } catch (error) {
    console.error('❌ 格式化过程中出现错误:', error.message)
    process.exit(1)
  }
}

// 运行命令的辅助函数
function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: options.stdio || 'inherit',
      shell: true,
      ...options
    })

    child.on('close', (code) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(`命令 "${command} ${args.join(' ')}" 执行失败，退出码: ${code}`))
      }
    })

    child.on('error', (error) => {
      reject(new Error(`无法执行命令 "${command}": ${error.message}`))
    })
  })
}

// 计算 Go 文件数量
async function countGoFiles() {
  try {
    if (process.platform === 'win32') {
      // Windows
      const output = execSync('dir /s /b *.go 2>nul | find /c /v ""', { encoding: 'utf8' })
      return parseInt(output.trim()) || 0
    } else {
      // Unix/Linux/macOS
      const output = execSync('find . -name "*.go" -not -path "./vendor/*" -not -path "./tmp/*" -not -path "./.git/*" | wc -l', { encoding: 'utf8' })
      return parseInt(output.trim()) || 0
    }
  } catch (error) {
    return '未知'
  }
}

// 运行格式化
formatCode()
