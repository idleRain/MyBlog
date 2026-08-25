#!/usr/bin/env bun

import { $ } from 'bun'
import { spawn, type ChildProcess } from 'child_process'
import { existsSync, readFileSync } from 'fs'
import { join } from 'path'
import yaml from 'yaml'

// ---------- 常量定义 ----------

// 终端颜色样式
const COLORS = {
  blue: '\x1b[34m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  cyan: '\x1b[36m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
} as const

// 默认端口配置
const DEFAULT_SERVER_PORT = 3000
const DEFAULT_WEB_PORT = 5173

// 端口合法性范围
const PORT_RANGE_MIN = 1000
const PORT_RANGE_MAX = 65535

// 健康检查参数
const HEALTH_CHECK_MAX_RETRIES = 30
const HEALTH_CHECK_INTERVAL_MS = 1000

// 进程释放后的等待时间
const PORT_RELEASE_WAIT_MS = 1000

// 环境要求的基础文件
const REQUIRED_FILES = ['server/go.mod', 'web/package.json', 'server/configs/config.yaml']

// 无法获取进程名时的后备描述
const UNKNOWN_PROCESS_NAME = '未知进程'

// 识别服务已就绪的日志标志
const SERVER_READY_MARKERS = ['Listening and serving HTTP', '服务器启动成功']
const WEB_READY_MARKER = 'ready in'

// ---------- 接口定义 ----------

interface ServiceConfig {
  name: string
  command: string[]
  cwd?: string
  color: keyof typeof COLORS
  port?: number
  healthCheck?: () => Promise<boolean>
}

interface ServiceStatus {
  ready: boolean
  readyByOutput: boolean
  readyByHealth: boolean
  port?: number
}

// ---------- 端口读取 ----------

// 读取后端服务端口，来源为 server/configs/config.yaml。
async function readServerPort(): Promise<number> {
  const configPath = join('server', 'configs', 'config.yaml')
  if (!existsSync(configPath)) {
    console.log(
      `${COLORS.yellow}⚠️  未找到后端配置，使用默认端口 ${DEFAULT_SERVER_PORT}${COLORS.reset}`
    )
    return DEFAULT_SERVER_PORT
  }

  try {
    const configContent = readFileSync(configPath, 'utf8')
    const config = yaml.parse(configContent)
    return config.server?.port || DEFAULT_SERVER_PORT
  } catch {
    console.log(
      `${COLORS.yellow}⚠️  后端配置解析失败，使用默认端口 ${DEFAULT_SERVER_PORT}${COLORS.reset}`
    )
    return DEFAULT_SERVER_PORT
  }
}

// 读取前端 Vite 开发服务器端口，来源为 web/.env 的 VITE_SERVER_PORT。
function readWebPort(): number {
  const envPath = join('web', '.env')
  if (!existsSync(envPath)) {
    return DEFAULT_WEB_PORT
  }

  try {
    const envContent = readFileSync(envPath, 'utf8')
    const match = envContent.match(/^VITE_SERVER_PORT\s*=\s*(\d+)/m)
    return match ? parseInt(match[1], 10) : DEFAULT_WEB_PORT
  } catch {
    return DEFAULT_WEB_PORT
  }
}

// ---------- 服务定义 ----------

// 组装前后端服务配置，包含各自端口的健康检查逻辑。
async function getServices(): Promise<ServiceConfig[]> {
  const serverPort = await readServerPort()
  const webPort = readWebPort()

  return [
    {
      name: 'SERVER',
      command: ['go', 'run', 'scripts/watcher.go'],
      cwd: 'server',
      color: 'blue',
      port: serverPort,
      // 健康路由仅接受 POST 请求。
      healthCheck: async () => {
        try {
          const response = await fetch(`http://localhost:${serverPort}/api/health`, {
            method: 'POST'
          })
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
  ]
}

// ---------- 环境检查 ----------

// 检查开发环境所需的工具、文件与依赖。
async function checkEnvironment(): Promise<void> {
  console.log(`${COLORS.cyan}🔍 检查开发环境...${COLORS.reset}\n`)

  await checkCommandAvailable('Go', 'go')
  console.log(`${COLORS.green}✅ Bun: ${Bun.version}${COLORS.reset}`)

  for (const file of REQUIRED_FILES) {
    if (existsSync(file)) {
      console.log(`${COLORS.green}✅ ${file}${COLORS.reset}`)
    } else {
      console.error(`${COLORS.red}❌ 缺少文件: ${file}${COLORS.reset}`)
      process.exit(1)
    }
  }

  if (!existsSync('node_modules')) {
    console.log(`${COLORS.yellow}⚠️  根目录依赖未安装，正在安装...${COLORS.reset}`)
    await $`bun install`
  }

  console.log('')
}

// 检查指定命令行工具是否可用，不可用时退出进程。
async function checkCommandAvailable(label: string, binary: string): Promise<void> {
  try {
    await $`${binary} version`.quiet()
    console.log(`${COLORS.green}✅ ${label}: 已安装${COLORS.reset}`)
  } catch {
    console.error(`${COLORS.red}❌ ${label} 未安装或不在 PATH 中${COLORS.reset}`)
    process.exit(1)
  }
}

// ---------- 端口进程检测 ----------

// 获取正在监听指定端口的进程 PID，空闲或检测失败时返回 null。
async function getListeningPid(port: number): Promise<number | null> {
  try {
    if (process.platform === 'win32') {
      return await getListeningPidOnWindows(port)
    }
    return await getListeningPidOnUnix(port)
  } catch {
    return null
  }
}

// 在 Windows 下通过 netstat 解析监听端口的进程 PID。
async function getListeningPidOnWindows(port: number): Promise<number | null> {
  const result = await $`netstat -ano -p tcp`.text()
  for (const line of result.split('\n')) {
    if (line.includes(`:${port} `) && line.includes('LISTENING')) {
      return parseLastColumnAsInt(line)
    }
  }
  return null
}

// 在 Unix 下通过 lsof 获取监听端口的进程 PID。
async function getListeningPidOnUnix(port: number): Promise<number | null> {
  const result = await $`lsof -ti :${port}`.text()
  const pid = result.trim().split('\n')[0]
  return pid ? parseInt(pid, 10) : null
}

// 解析 netstat 输出行的最后一列并转换为整数。
function parseLastColumnAsInt(line: string): number | null {
  const parts = line.trim().split(/\s+/)
  const pid = parseInt(parts[parts.length - 1], 10)
  return Number.isNaN(pid) ? null : pid
}

// 判断端口是否被占用。
async function isPortInUse(port: number): Promise<boolean> {
  return (await getListeningPid(port)) !== null
}

// 获取进程名称，失败时返回未知进程。
async function getProcessName(pid: number): Promise<string> {
  try {
    if (process.platform === 'win32') {
      const result = await $`tasklist /FI "PID eq ${pid}" /FO CSV`.text()
      const lines = result.trim().split('\n')
      if (lines.length > 1) {
        return lines[1].split(',')[0]?.replace(/"/g, '') || UNKNOWN_PROCESS_NAME
      }
    } else {
      const result = await $`ps -p ${pid} -o comm=`.text()
      const name = result.trim()
      if (name) {
        return name
      }
    }
    return UNKNOWN_PROCESS_NAME
  } catch {
    return UNKNOWN_PROCESS_NAME
  }
}

// 强制结束指定进程，失败时返回 false。
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

// ---------- 用户交互 ----------

// 在终端中让用户选择选项，支持方向键与数字键，未启用 TTY 时回退为数字输入。
async function promptUser(message: string, options: string[]): Promise<number> {
  if (message) {
    console.log(`${COLORS.yellow}${message}${COLORS.reset}`)
  }

  let selectedIndex = 0
  let isFirstDisplay = true

  // 重绘选项列表，首次显示后向上清除旧内容。
  function displayOptions(): void {
    if (!isFirstDisplay) {
      process.stdout.write(`\x1b[${options.length}A`)
      process.stdout.write('\x1b[0J')
    }

    options.forEach((option, index) => {
      if (index === selectedIndex) {
        console.log(`${COLORS.cyan}${COLORS.bold}❯ ${option}${COLORS.reset}`)
      } else {
        console.log(`  ${option}`)
      }
    })

    isFirstDisplay = false
  }

  // 非 TTY 环境回退为数字输入。
  if (!process.stdout.isTTY || !process.stdin.isTTY) {
    console.log(`${COLORS.yellow}${message}${COLORS.reset}`)
    options.forEach((option, index) => {
      console.log(`${COLORS.cyan}${index + 1}. ${option}${COLORS.reset}`)
    })

    const input = prompt('请选择 (输入数字): ')
    const choice = parseInt(input || '1', 10) - 1
    return choice >= 0 && choice < options.length ? choice : 0
  }

  displayOptions()

  return new Promise(resolve => {
    process.stdin.setRawMode(true)
    process.stdin.resume()
    process.stdin.setEncoding('utf8')

    const onKeyPress = (key: string): void => {
      if (key === '\u001b[A' || key === 'k') {
        selectedIndex = selectedIndex === 0 ? options.length - 1 : selectedIndex - 1
      } else if (key === '\u001b[B' || key === 'j') {
        selectedIndex = selectedIndex === options.length - 1 ? 0 : selectedIndex + 1
      } else if (key === '\r' || key === '\n' || key === ' ') {
        process.stdin.setRawMode(false)
        process.stdin.pause()
        process.stdin.removeListener('data', onKeyPress)
        console.log(`${COLORS.green}✓ 已选择: ${options[selectedIndex]}${COLORS.reset}\n`)
        resolve(selectedIndex)
        return
      } else if (key === '\u0003') {
        process.stdin.setRawMode(false)
        process.stdin.pause()
        console.log(`\n${COLORS.yellow}👋 用户取消操作${COLORS.reset}`)
        process.exit(0)
        return
      } else {
        const num = parseInt(key, 10)
        if (!Number.isNaN(num) && num >= 1 && num <= options.length) {
          selectedIndex = num - 1
        }
      }

      displayOptions()
    }

    process.stdin.on('data', onKeyPress)
  })
}

// ---------- 端口冲突处理 ----------

// 检查后端端口占用，必要时结束占用进程并释放端口。
async function checkBackendPort(serverPort: number): Promise<void> {
  console.log(`${COLORS.cyan}🔌 检查后端端口 ${serverPort}...${COLORS.reset}`)

  const pid = await getListeningPid(serverPort)
  if (pid === null) {
    console.log(`${COLORS.green}✅ 端口 ${serverPort} 可用${COLORS.reset}\n`)
    return
  }

  const processName = await getProcessName(pid)
  console.log(
    `${COLORS.yellow}⚠️  端口 ${serverPort} 被进程占用：${processName} (PID: ${pid})${COLORS.reset}`
  )
  console.log(`${COLORS.cyan}使用 ↑↓ 键或 1/2 数字键选择，回车/空格确认：${COLORS.reset}\n`)

  const choice = await promptUser('', [`结束进程 ${processName} (PID: ${pid}) 并继续`, '退出'])
  if (choice !== 0) {
    console.log(`${COLORS.yellow}👋 用户选择退出${COLORS.reset}`)
    process.exit(0)
  }

  console.log(`${COLORS.cyan}🔄 正在结束进程 ${processName} (PID: ${pid})...${COLORS.reset}`)
  if (!(await killProcess(pid))) {
    console.log(`${COLORS.red}❌ 无法结束进程 ${pid}，请手动处理${COLORS.reset}`)
    process.exit(1)
  }

  console.log(`${COLORS.green}✅ 成功结束进程，端口 ${serverPort} 已释放${COLORS.reset}`)
  await new Promise(resolve => setTimeout(resolve, PORT_RELEASE_WAIT_MS))

  if (await isPortInUse(serverPort)) {
    console.log(`${COLORS.red}❌ 端口 ${serverPort} 仍被占用，请手动处理${COLORS.reset}`)
    process.exit(1)
  }

  console.log('')
}

// ---------- 服务启动与就绪 ----------

// 启动全部服务，就绪信息由输出标志与健康检查共同驱动，并接管退出信号。
async function startServices(services: ServiceConfig[]): Promise<void> {
  console.log(`${COLORS.bold}${COLORS.cyan}🚀 启动开发服务器...${COLORS.reset}\n`)

  const statusMap = new Map<string, ServiceStatus>()
  const processes: ChildProcess[] = []

  for (const service of services) {
    statusMap.set(service.name, {
      ready: false,
      readyByOutput: false,
      readyByHealth: false,
      port: service.port
    })
  }

  for (const service of services) {
    const child = spawn(service.command[0], service.command.slice(1), {
      cwd: service.cwd,
      stdio: ['inherit', 'pipe', 'pipe']
    })
    processes.push(child)

    forwardServiceOutput(service, child, statusMap)
    forwardServiceError(service, child, statusMap)
    child.on('exit', code => handleServiceExit(code, service, child, processes))
  }

  // 并发发起健康检查轮询，与输出标志共同决定就绪，失败不阻断启动。
  services.forEach(service => startHealthPolling(service, statusMap))

  setupSignalHandler(processes)

  await Promise.all(
    processes.map(child => new Promise<void>(resolve => child.on('exit', () => resolve())))
  )
}

// 轮询单个服务的健康检查，通过后标记就绪。
async function startHealthPolling(
  service: ServiceConfig,
  statusMap: Map<string, ServiceStatus>
): Promise<void> {
  if (!service.healthCheck) {
    markServiceReady(service.name, statusMap)
    return
  }

  console.log(`${COLORS.cyan}🔍 等待 ${service.name} 启动...${COLORS.reset}`)
  for (let attempt = 0; attempt < HEALTH_CHECK_MAX_RETRIES; attempt++) {
    try {
      if (await service.healthCheck()) {
        markServiceReady(service.name, statusMap)
        return
      }
    } catch {
      // 单次探测异常不终止，继续重试。
    }

    await new Promise(resolve => setTimeout(resolve, HEALTH_CHECK_INTERVAL_MS))
  }

  // 达到重试上限仍未通过时仅提示，不阻断服务继续运行。
  console.log(`${COLORS.yellow}⚠️  ${service.name} 健康检查未通过，仍会继续运行${COLORS.reset}`)
}

// 转发子进程标准输出，子进程输出统一按 UTF-8 解码。
function forwardServiceOutput(
  service: ServiceConfig,
  child: ChildProcess,
  statusMap: Map<string, ServiceStatus>
): void {
  child.stdout?.setEncoding('utf8')
  child.stdout?.on('data', (data: string) => {
    for (const line of splitOutput(data)) {
      console.log(`${COLORS[service.color]}[${service.name}]${COLORS.reset} ${line}`)
      handleServiceLine(service, line, statusMap)
    }
  })
}

// 转发子进程标准错误输出，对真实错误高亮并检测就绪标志。
function forwardServiceError(
  service: ServiceConfig,
  child: ChildProcess,
  statusMap: Map<string, ServiceStatus>
): void {
  child.stderr?.setEncoding('utf8')
  child.stderr?.on('data', (data: string) => {
    for (const line of splitOutput(data)) {
      emitServiceLine(service, line, statusMap)
    }
  })
}

// 输出单行并标记错误高亮与就绪状态。
function emitServiceLine(
  service: ServiceConfig,
  line: string,
  statusMap: Map<string, ServiceStatus>
): void {
  if (isErrorLine(line)) {
    console.log(`${COLORS.red}[${service.name}:ERROR]${COLORS.reset} ${line}`)
  } else {
    console.log(`${COLORS[service.color]}[${service.name}]${COLORS.reset} ${line}`)
  }
  handleServiceLine(service, line, statusMap)
}

// 处理单行输出，更新端口与就绪状态。
function handleServiceLine(
  service: ServiceConfig,
  line: string,
  statusMap: Map<string, ServiceStatus>
): void {
  const port = extractVitePort(line)
  if (port && statusMap.has(service.name)) {
    statusMap.get(service.name)!.port = port
  }

  if (isReadyMarker(service.name, line)) {
    markServiceReady(service.name, statusMap)
  }
}

// 判断输出是否包含服务就绪标志。
function isReadyMarker(serviceName: string, line: string): boolean {
  if (serviceName === 'SERVER') {
    return SERVER_READY_MARKERS.some(marker => line.includes(marker))
  }
  return line.includes(WEB_READY_MARKER)
}

// 将服务标记为就绪，并在全部服务就绪后展示汇总信息。
function markServiceReady(serviceName: string, statusMap: Map<string, ServiceStatus>): void {
  const status = statusMap.get(serviceName)
  if (!status || status.ready) {
    return
  }

  status.ready = true
  console.log(`${COLORS.green}✅ ${serviceName} 已就绪${COLORS.reset}`)
  checkAllServicesReady(statusMap)
}

// 当全部服务就绪时输出一次可用地址汇总。
function checkAllServicesReady(statusMap: Map<string, ServiceStatus>): void {
  const allReady = [...statusMap.values()].every(status => status.ready)
  if (!allReady) {
    return
  }

  displayServicesInfo(statusMap)
}

// 将文本切分为过滤空白的行数组。
function splitOutput(text: string): string[] {
  return text.split('\n').filter(line => line.trim())
}

// 判断行内容是否属于真实错误信息。
function isErrorLine(line: string): boolean {
  const lower = line.toLowerCase()
  return ['error', 'failed', 'panic', 'fatal'].some(word => lower.includes(word))
}

// 处理子进程退出：仅对非零退出码触发清理与终止。
function handleServiceExit(
  code: number | null,
  service: ServiceConfig,
  child: ChildProcess,
  processes: ChildProcess[]
): void {
  if (code === 0) {
    return
  }

  console.log(`${COLORS.red}❌ ${service.name} 退出，代码: ${code}${COLORS.reset}`)
  for (const other of processes) {
    if (other !== child && !other.killed) {
      other.kill()
    }
  }
  process.exit(1)
}

// 停止全部子进程。
function stopAllProcesses(processes: ChildProcess[]): void {
  for (const child of processes) {
    if (!child.killed) {
      child.kill()
    }
  }
}

// 注册 SIGINT 信号处理以停止所有服务。
function setupSignalHandler(processes: ChildProcess[]): void {
  process.on('SIGINT', () => {
    console.log(`\n${COLORS.yellow}🛑 正在停止所有服务...${COLORS.reset}`)
    stopAllProcesses(processes)
    process.exit(0)
  })
}

// 输出各服务的访问地址。
function displayServicesInfo(statusMap: Map<string, ServiceStatus>): void {
  console.log(`\n${COLORS.bold}${COLORS.green}🎉 所有服务已启动！${COLORS.reset}`)
  console.log(`${COLORS.cyan}📖 可用服务:${COLORS.reset}`)

  const serverStatus = statusMap.get('SERVER')
  if (serverStatus?.port) {
    console.log(`  ${COLORS.green}• SERVER: http://localhost:${serverStatus.port}${COLORS.reset}`)
  }

  const webPort = statusMap.get('WEB')?.port || DEFAULT_WEB_PORT
  console.log(`  ${COLORS.green}• WEB: http://localhost:${webPort}${COLORS.reset}`)

  console.log(`\n${COLORS.yellow}按 Ctrl+C 停止所有服务${COLORS.reset}\n`)
}

// ---------- 辅助函数 ----------

// 移除字符串中的 ANSI 转义序列。
function stripAnsiCodes(text: string): string {
  return text.replace(/\x1b\[[0-9;]*m/g, '')
}

// 从 Vite 输出中提取监听端口，无匹配时返回 null。
function extractVitePort(line: string): number | null {
  const cleanLine = stripAnsiCodes(line)

  const portPatterns = [
    /➜\s+Local:\s+http:\/\/localhost:(\d+)/,
    /Local:\s+http:\/\/localhost:(\d+)/,
    /http:\/\/localhost:(\d+)/,
    /localhost:(\d+)/
  ]

  for (const regex of portPatterns) {
    const match = cleanLine.match(regex)
    if (match) {
      const port = parseInt(match[1], 10)
      if (port >= PORT_RANGE_MIN && port <= PORT_RANGE_MAX) {
        return port
      }
    }
  }

  return null
}

// ---------- 主流程 ----------

async function main(): Promise<void> {
  try {
    await checkEnvironment()
    const serverPort = await readServerPort()
    await checkBackendPort(serverPort)
    const services = await getServices()
    await startServices(services)
  } catch (error) {
    console.error(`${COLORS.red}❌ 启动失败:${COLORS.reset}`, error)
    process.exit(1)
  }
}

// 直接运行时启动主流程。
if (import.meta.main) {
  await main()
}
