// 脚本命令执行公共库：统一的子进程执行器，供 scripts 下各脚本复用。
// Windows 与 Unix 差异由 shell 选项吸收，调用方只面向命令与参数数组。

import { spawn, type SpawnOptions } from 'node:child_process'

// 运行命令的选项，stdio 在 SpawnOptions 基础上收窄为脚本场景常用值。
interface RunCommandOptions extends Partial<SpawnOptions> {
  stdio?: 'inherit' | 'ignore' | 'pipe'
  cwd?: string
}

// 命令执行结果：success 归一化退出状态，output 仅在 pipe 模式下返回。
interface CommandResult {
  success: boolean
  output?: string
  code: number | null
  errorMessage?: string
}

/**
 * 执行命令并归一化结果，永不 reject，spawn 失败与进程异常都收敛为 success: false。
 * pipe 模式下收集标准输出与标准错误到 output，供调用方解析。
 */
export async function tryRunCommand(
  command: string,
  args: string[],
  options: RunCommandOptions = {}
): Promise<CommandResult> {
  return new Promise(resolve => {
    const child = spawn(command, args, {
      stdio: options.stdio || 'inherit',
      shell: true,
      cwd: options.cwd || process.cwd(),
      ...options
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

    child.on('close', (code: number | null) => {
      resolve({
        success: code === 0,
        output: options.stdio === 'pipe' ? output : undefined,
        code
      })
    })

    child.on('error', (error: Error) => {
      resolve({ success: false, output, code: null, errorMessage: error.message })
    })
  })
}

/**
 * 以抛错语义执行命令，供无法继续执行的场景使用；失败消息携带命令、退出码与原因。
 */
export async function runCommand(
  command: string,
  args: string[],
  options: RunCommandOptions = {}
): Promise<void> {
  const result = await tryRunCommand(command, args, options)
  if (result.success) {
    return
  }
  if (result.errorMessage) {
    throw new Error(`无法执行命令 "${command}": ${result.errorMessage}`)
  }
  throw new Error(`命令 "${command} ${args.join(' ')}" 执行失败，退出码: ${result.code}`)
}
