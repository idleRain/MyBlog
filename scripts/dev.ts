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

// 读取配置文件
async function loadConfigs() {
  let serverPort = 3000
  let webPort = 5173

  // 读取后端配置
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

  // 读取前端配置
  try {
    const webEnvPath = join('web', '.env')
    if (existsSync(webEnvPath)) {
      const envContent = readFileSync(webEnvPath, 'utf8')
      const envMatch = envContent.match(/VITE_SERVER_PORT=(\d+)/)
      if (envMatch) {
        webPort = parseInt(envMatch[1], 10)
      }
    }
  } catch (error) {
    console.log(`${colors.yellow}⚠️  无法读取前端配置，使用默认端口 5173${colors.reset}`)
  }

  return { serverPort, webPort }
}

// 服务配置（动态生成）
async function getServices() {
  const { serverPort, webPort } = await loadConfigs()

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
      port: webPort,
      healthCheck: async () => {
        try {
          const response = await fetch(`http://localhost:${webPort}/`)
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

  if (!existsSync('web/node_modules')) {
    console.log(`${colors.yellow}⚠️  前端依赖未安装，正在安装...${colors.reset}`)
    await $`cd web && bun install`
  }

  console.log('')
}

// 端口检查
async function checkPorts(services: ServiceConfig[]) {
  console.log(`${colors.cyan}🔌 检查端口占用...${colors.reset}`)

  for (const service of services) {
    if (service.port) {
      try {
        const response = await fetch(`http://localhost:${service.port}`)
        if (response.ok) {
          console.log(
            `${colors.yellow}⚠️  端口 ${service.port} 已被占用（${service.name}）${colors.reset}`
          )
        }
      } catch {
        console.log(`${colors.green}✅ 端口 ${service.port} 可用（${service.name}）${colors.reset}`)
      }
    }
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

// 启动服务
async function startServices(services: ServiceConfig[]) {
  console.log(`${colors.bold}${colors.cyan}🚀 启动开发服务器...${colors.reset}\n`)

  const processes: any[] = []

  // 启动所有服务
  for (const service of services) {
    console.log(`${colors.cyan}启动 ${service.name}...${colors.reset}`)

    const child = spawn(service.command[0], service.command.slice(1), {
      cwd: service.cwd,
      stdio: ['inherit', 'pipe', 'pipe'],
      shell: true
    })

    // 设置颜色输出
    child.stdout?.on('data', data => {
      const lines = data
        .toString()
        .split('\n')
        .filter((line: string) => line.trim())
      lines.forEach((line: string) => {
        console.log(
          `${colors[service.color as keyof typeof colors]}[${service.name}]${colors.reset} ${line}`
        )
      })
    })

    child.stderr?.on('data', data => {
      const lines = data
        .toString()
        .split('\n')
        .filter((line: string) => line.trim())
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

    // 等待服务启动
    await new Promise(resolve => setTimeout(resolve, 2000))
  }

  console.log(`\n${colors.bold}${colors.green}🎉 所有服务已启动！${colors.reset}`)
  console.log(`${colors.cyan}📖 可用服务:${colors.reset}`)
  services.forEach(service => {
    console.log(
      `  ${colors.green}• ${service.name}: http://localhost:${service.port}${colors.reset}`
    )
  })
  console.log(`\n${colors.yellow}按 Ctrl+C 停止所有服务${colors.reset}\n`)

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
    const services = await getServices()
    await checkPorts(services)
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
