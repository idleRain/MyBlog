#!/usr/bin/env bun

import { $ } from 'bun'
import { spawn } from 'child_process'
import { existsSync, readFileSync } from 'fs'
import { join } from 'path'
import yaml from 'yaml'

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

interface ServiceConfig {
  name: string
  command: string[]
  cwd?: string
  color: string
  healthCheck?: () => Promise<boolean>
  port?: number
}

// 读取后端端口配置
async function getServerPort(): Promise<number> {
  let serverPort = 3000

  try {
    const serverConfigPath = join('server', 'configs', 'config.yaml')
    if (existsSync(serverConfigPath)) {
      const configContent = readFileSync(serverConfigPath, 'utf8')
      const config = yaml.parse(configContent)
      serverPort = config.server?.port || 3000
    }
  } catch (error) {
    console.log(`${colors.yellow}⚠️  无法读取后端配置，使用默认端口 3000${colors.reset}`)
  }

  return serverPort
}

// 服务配置
async function getServices() {
  const serverPort = await getServerPort()

  return [
    {
      name: 'SERVER',
      command: ['go', 'run', 'scripts/watcher.go'],
      cwd: 'server',
      color: 'blue',
      port: serverPort,
      healthCheck: async () => {
        try {
          const response = await fetch(`http://localhost:${serverPort}/api/health`)
          return response.ok
        } catch {
          return false
        }
      }
    },
    {
      name: 'WEB',
      command: ['bun', 'run', 'dev'],
      cwd: 'web',
      color: 'green',
      // 前端不需要端口检测，Vite 会自动处理
      healthCheck: async () => {
        try {
          const response = await fetch(`http://localhost:5173/`)
          return response.ok
        } catch {
          return false
        }
      }
    }
  ] as ServiceConfig[]
}

// 环境检查
async function checkEnvironment() {
  console.log(`${colors.cyan}🔍 检查开发环境...${colors.reset}\n`)

  // 检查 Go
  try {
    await $`go version`.quiet()
    console.log(`${colors.green}✅ Go: 已安装${colors.reset}`)
  } catch {
    console.error(`${colors.red}❌ Go 未安装或不在 PATH 中${colors.reset}`)
    process.exit(1)
  }

  // 检查 Bun
  try {
    console.log(`${colors.green}✅ Bun: ${Bun.version}${colors.reset}`)
  } catch {
    console.error(`${colors.red}❌ Bun 未安装${colors.reset}`)
    process.exit(1)
  }

  // 检查项目文件
  const requiredFiles = ['server/go.mod', 'web/package.json', 'server/configs/config.yaml']

  for (const file of requiredFiles) {
    if (existsSync(file)) {
      console.log(`${colors.green}✅ ${file}${colors.reset}`)
    } else {
      console.error(`${colors.red}❌ 缺少文件: ${file}${colors.reset}`)
      process.exit(1)
    }
  }

  // 检查依赖
  if (!existsSync('node_modules')) {
    console.log(`${colors.yellow}⚠️  根目录依赖未安装，正在安装...${colors.reset}`)
    await $`bun install`
  }

  console.log('')
}

// 检查端口是否被占用
async function isPortInUse(port: number): Promise<boolean> {
  try {
    if (process.platform === 'win32') {
      const result = await $`netstat -ano -p tcp`.text()
      const lines = result.split('\n')
      for (const line of lines) {
        if (line.includes(`:${port} `) && line.includes('LISTENING')) {
          return true
        }
      }
      return false
    } else {
      try {
        await $`lsof -i :${port}`.quiet()
        return true
      } catch {
        return false
      }
    }
  } catch {
    return false
  }
}

// 获取占用端口的进程ID
async function getProcessUsingPort(port: number): Promise<number | null> {
  try {
    if (process.platform === 'win32') {
      const result = await $`netstat -ano -p tcp`.text()
      const lines = result.split('\n')
      for (const line of lines) {
        if (line.includes(`:${port} `) && line.includes('LISTENING')) {
          const parts = line.trim().split(/\s+/)
          const pid = parts[parts.length - 1]
          return parseInt(pid, 10)
        }
      }
    } else {
      const result = await $`lsof -ti :${port}`.text()
      const pid = result.trim().split('\n')[0]
      return pid ? parseInt(pid, 10) : null
    }
    return null
  } catch {
    return null
  }
}

// 获取进程名称
async function getProcessName(pid: number): Promise<string> {
  try {
    if (process.platform === 'win32') {
      const result = await $`tasklist /FI "PID eq ${pid}" /FO CSV`.text()
      const lines = result.trim().split('\n')
      if (lines.length > 1) {
        const parts = lines[1].split(',')
        return parts[0]?.replace(/"/g, '') || '未知进程'
      }
    } else {
      const result = await $`ps -p ${pid} -o comm=`.text()
      return result.trim() || '未知进程'
    }
    return '未知进程'
  } catch {
    return '未知进程'
  }
}

// 杀死进程
async function killProcess(pid: number): Promise<boolean> {
  try {
    if (process.platform === 'win32') {
      await $`taskkill /F /PID ${pid}`.quiet()
    } else {
      await $`kill -9 ${pid}`.quiet()
    }
    return true
  } catch {
    return false
  }
}

// 用户选择处理（支持上下键切换）
async function promptUser(message: string, options: string[]): Promise<number> {
  if (message) {
    console.log(`${colors.yellow}${message}${colors.reset}`)
  }

  let selectedIndex = 0 // 默认选中第一个
  let isFirstDisplay = true

  // 显示选项
  function displayOptions() {
    // 清除之前的显示（向上移动光标）
    if (!isFirstDisplay && process.stdout.isTTY) {
      process.stdout.write(`\x1b[${options.length}A`) // 向上移动对应行数
      process.stdout.write('\x1b[0J') // 清除光标到屏幕末尾
    }

    options.forEach((option, index) => {
      if (index === selectedIndex) {
        // 选中状态：高亮显示，带箭头
        console.log(`${colors.cyan}${colors.bold}❯ ${option}${colors.reset}`)
      } else {
        // 未选中状态：普通显示，留空间对齐
        console.log(`  ${option}`)
      }
    })

    isFirstDisplay = false
  }

  // 如果不是TTY环境，回退到简单选择
  if (!process.stdout.isTTY || !process.stdin.isTTY) {
    console.log(`${colors.yellow}${message}${colors.reset}`)
    options.forEach((option, index) => {
      console.log(`${colors.cyan}${index + 1}. ${option}${colors.reset}`)
    })

    const input = prompt('请选择 (输入数字): ')
    const choice = parseInt(input || '1', 10) - 1

    if (choice >= 0 && choice < options.length) {
      return choice
    }
    return 0 // 默认第一个选项
  }

  // 初始显示
  displayOptions()

  return new Promise(resolve => {
    // 设置 stdin 为 raw 模式以捕获按键
    process.stdin.setRawMode(true)
    process.stdin.resume()
    process.stdin.setEncoding('utf8')

    const onKeyPress = (key: string) => {
      switch (key) {
        case '\u001b[A': // 上箭头键
        case 'k': // vim 风格向上
          selectedIndex = selectedIndex > 0 ? selectedIndex - 1 : options.length - 1
          displayOptions()
          break

        case '\u001b[B': // 下箭头键
        case 'j': // vim 风格向下
          selectedIndex = selectedIndex < options.length - 1 ? selectedIndex + 1 : 0
          displayOptions()
          break

        case '\r': // 回车键
        case '\n':
        case ' ': // 空格键也可以确认
          // 恢复 stdin 设置
          process.stdin.setRawMode(false)
          process.stdin.pause()
          process.stdin.removeListener('data', onKeyPress)
          console.log(`${colors.green}✓ 已选择: ${options[selectedIndex]}${colors.reset}\n`)
          resolve(selectedIndex)
          break

        case '\u0003': // Ctrl+C
          // 恢复 stdin 设置
          process.stdin.setRawMode(false)
          process.stdin.pause()
          console.log(`\n${colors.yellow}👋 用户取消操作${colors.reset}`)
          process.exit(0)
          break

        case '1': // 数字键快捷选择
          if (options.length >= 1) {
            selectedIndex = 0
            displayOptions()
          }
          break

        case '2': // 数字键快捷选择
          if (options.length >= 2) {
            selectedIndex = 1
            displayOptions()
          }
          break

        default:
          // 忽略其他按键
          break
      }
    }

    process.stdin.on('data', onKeyPress)
  })
}

// 检查并处理后端端口占用
async function checkBackendPort(serverPort: number) {
  console.log(`${colors.cyan}🔌 检查后端端口 ${serverPort}...${colors.reset}`)

  const isInUse = await isPortInUse(serverPort)

  if (!isInUse) {
    console.log(`${colors.green}✅ 端口 ${serverPort} 可用${colors.reset}\n`)
    return
  }

  // 端口被占用，获取进程信息
  const pid = await getProcessUsingPort(serverPort)

  if (!pid) {
    console.log(`${colors.yellow}⚠️  端口 ${serverPort} 被占用，但无法获取进程信息${colors.reset}`)
    console.log(`${colors.red}❌ 请手动停止占用端口的进程后重试${colors.reset}`)
    process.exit(1)
  }

  const processName = await getProcessName(pid)

  console.log(
    `${colors.yellow}⚠️  端口 ${serverPort} 被进程占用：${processName} (PID: ${pid})${colors.reset}`
  )
  console.log(`${colors.cyan}使用 ↑↓ 键或 1/2 数字键选择，回车/空格确认：${colors.reset}\n`)

  // 提供用户选择
  const choice = await promptUser('', [`结束进程 ${processName} (PID: ${pid}) 并继续`, '退出'])

  if (choice === 0) {
    // 选择结束进程
    console.log(`${colors.cyan}🔄 正在结束进程 ${processName} (PID: ${pid})...${colors.reset}`)
    const killed = await killProcess(pid)

    if (killed) {
      console.log(`${colors.green}✅ 成功结束进程，端口 ${serverPort} 已释放${colors.reset}`)

      // 等待端口完全释放
      await new Promise(resolve => setTimeout(resolve, 1000))

      // 再次检查
      const stillInUse = await isPortInUse(serverPort)
      if (stillInUse) {
        console.log(`${colors.red}❌ 端口 ${serverPort} 仍被占用，请手动处理${colors.reset}`)
        process.exit(1)
      }
    } else {
      console.log(`${colors.red}❌ 无法结束进程 ${pid}，请手动处理${colors.reset}`)
      process.exit(1)
    }
  } else {
    // 选择退出
    console.log(`${colors.yellow}👋 用户选择退出${colors.reset}`)
    process.exit(0)
  }

  console.log('')
}

// 健康检查
async function healthCheck(service: ServiceConfig, maxRetries = 30) {
  if (!service.healthCheck) return true

  console.log(`${colors.cyan}🔍 等待 ${service.name} 启动...${colors.reset}`)

  for (let i = 0; i < maxRetries; i++) {
    try {
      const isHealthy = await service.healthCheck()
      if (isHealthy) {
        console.log(`${colors.green}✅ ${service.name} 健康检查通过${colors.reset}`)
        return true
      }
    } catch (error) {
      // 继续重试
    }

    await new Promise(resolve => setTimeout(resolve, 1000))
  }

  console.log(`${colors.red}❌ ${service.name} 健康检查失败${colors.reset}`)
  return false
}

// 移除ANSI颜色代码
function stripAnsiCodes(text: string): string {
  return text.replace(/\x1b\[[0-9;]*m/g, '')
}

// 从Vite输出中提取端口号
function extractVitePort(line: string): number | null {
  // 先移除ANSI颜色代码
  const cleanLine = stripAnsiCodes(line)

  // 匹配 Vite 输出中的端口信息，根据实际输出格式
  const portMatches = [
    /➜\s+Local:\s+http:\/\/localhost:(\d+)/, // ➜ Local: http://localhost:8900/
    /Local:\s+http:\/\/localhost:(\d+)/, // Local: http://localhost:8900/
    /http:\/\/localhost:(\d+)/, // 通用 http://localhost:端口
    /localhost:(\d+)/ // 通用 localhost:端口
  ]

  for (const regex of portMatches) {
    const match = cleanLine.match(regex)
    if (match) {
      const port = parseInt(match[1], 10)
      if (port >= 1000 && port <= 65535) {
        return port
      }
    }
  }
  return null
}

// 启动服务
async function startServices(services: ServiceConfig[]) {
  console.log(`${colors.bold}${colors.cyan}🚀 启动开发服务器...${colors.reset}\n`)

  const processes: any[] = []
  const serviceStatus: { [key: string]: { ready: boolean; port?: number } } = {}

  // 初始化状态
  services.forEach(service => {
    serviceStatus[service.name] = { ready: false, port: service.port }
  })

  // 启动所有服务
  for (const service of services) {
    console.log(`${colors.cyan}启动 ${service.name}...${colors.reset}`)

    const child = spawn(service.command[0], service.command.slice(1), {
      cwd: service.cwd,
      stdio: ['inherit', 'pipe', 'pipe'],
      shell: true
    })

    // 设置颜色输出并监控服务状态
    child.stdout?.on('data', data => {
      const output = data.toString()
      const lines = output.split('\n').filter((line: string) => line.trim())

      lines.forEach((line: string) => {
        console.log(
          `${colors[service.color as keyof typeof colors]}[${service.name}]${colors.reset} ${line}`
        )

        // 检测服务启动状态
        if (service.name === 'SERVER') {
          // 后端服务启动标志
          if (line.includes('Listening and serving HTTP') || line.includes('服务器启动成功')) {
            serviceStatus[service.name].ready = true
            checkAllServicesReady()
          }
        } else if (service.name === 'WEB') {
          // 前端服务启动标志
          if (line.includes('ready in')) {
            serviceStatus[service.name].ready = true

            // 延迟一点检查，等待端口信息输出
            setTimeout(() => checkAllServicesReady(), 100)
          }

          // 提取Vite实际使用的端口（独立于ready状态）
          const port = extractVitePort(line)
          if (port) {
            serviceStatus[service.name].port = port
            // 如果服务已经ready，更新显示的信息
            if (serviceStatus[service.name].ready && allServicesDisplayed) {
              updateWebPortDisplay(port)
            }
          }
        }
      })
    })

    child.stderr?.on('data', data => {
      const output = data.toString()
      const lines = output.split('\n').filter((line: string) => line.trim())

      lines.forEach((line: string) => {
        // 判断是否为真正的错误信息
        const isActualError =
          line.toLowerCase().includes('error') ||
          line.toLowerCase().includes('failed') ||
          line.toLowerCase().includes('panic') ||
          line.toLowerCase().includes('fatal')

        if (isActualError) {
          console.log(`${colors.red}[${service.name}:ERROR]${colors.reset} ${line}`)
        } else {
          // 对于非错误的 stderr 输出，使用正常颜色显示
          console.log(
            `${colors[service.color as keyof typeof colors]}[${service.name}]${colors.reset} ${line}`
          )

          // 也要检查 stderr 中的启动信息（某些框架可能输出到stderr）
          if (service.name === 'WEB') {
            const port = extractVitePort(output)
            if (port) {
              serviceStatus[service.name].port = port
            }
          }
        }
      })
    })

    child.on('exit', code => {
      if (code !== 0) {
        console.log(`${colors.red}❌ ${service.name} 退出，代码: ${code}${colors.reset}`)
        // 清理其他进程
        processes.forEach(p => {
          if (p !== child && !p.killed) {
            p.kill()
          }
        })
        process.exit(1)
      }
    })

    processes.push(child)
  }

  let allServicesDisplayed = false

  // 检查所有服务是否都已启动
  function checkAllServicesReady() {
    if (allServicesDisplayed) return

    const allReady = Object.values(serviceStatus).every(status => status.ready)

    if (allReady) {
      allServicesDisplayed = true
      // 延迟显示，等待端口信息完全获取
      setTimeout(() => displayServicesInfo(), 500)
    }
  }

  // 显示服务信息
  function displayServicesInfo() {
    console.log(`\n${colors.bold}${colors.green}🎉 所有服务已启动！${colors.reset}`)
    console.log(`${colors.cyan}📖 可用服务:${colors.reset}`)

    // 显示后端服务
    const serverPort = serviceStatus['SERVER'].port
    if (serverPort) {
      console.log(`  ${colors.green}• SERVER: http://localhost:${serverPort}${colors.reset}`)
    }

    // 显示前端服务（使用动态获取的端口）
    const webPort = serviceStatus['WEB'].port || 5173
    console.log(`  ${colors.green}• WEB: http://localhost:${webPort}${colors.reset}`)

    console.log(`\n${colors.yellow}按 Ctrl+C 停止所有服务${colors.reset}\n`)
  }

  // 更新前端端口显示（当端口在服务启动后才确定时）
  function updateWebPortDisplay(port: number) {
    // 清除之前的显示并重新输出
    process.stdout.write('\x1b[4A') // 向上移动4行
    process.stdout.write('\x1b[0J') // 清除光标到屏幕末尾

    console.log(`${colors.cyan}📖 可用服务:${colors.reset}`)

    const serverPort = serviceStatus['SERVER'].port
    if (serverPort) {
      console.log(`  ${colors.green}• SERVER: http://localhost:${serverPort}${colors.reset}`)
    }

    console.log(`  ${colors.green}• WEB: http://localhost:${port}${colors.reset}`)
    console.log(`\n${colors.yellow}按 Ctrl+C 停止所有服务${colors.reset}\n`)
  }

  // 设置信号处理
  process.on('SIGINT', () => {
    console.log(`\n${colors.yellow}🛑 正在停止所有服务...${colors.reset}`)
    processes.forEach(child => {
      if (!child.killed) {
        child.kill('SIGTERM')
      }
    })
    process.exit(0)
  })

  // 等待所有进程
  await Promise.all(
    processes.map(
      child =>
        new Promise(resolve => {
          child.on('exit', resolve)
        })
    )
  )
}

// 主函数
async function main() {
  try {
    await checkEnvironment()

    const serverPort = await getServerPort()
    await checkBackendPort(serverPort)

    const services = await getServices()
    await startServices(services)
  } catch (error) {
    console.error(`${colors.red}❌ 启动失败:${colors.reset}`, error)
    process.exit(1)
  }
}

// 如果直接运行此脚本
if (import.meta.main) {
  await main()
}
