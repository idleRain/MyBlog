#!/usr/bin/env node
// 备份与恢复脚本的外部命令依赖核对工具。
// 用途：静态提取 backup.sh 与 restore.sh 调用的全部外部命令，供容器镜像能力逐项比对。
// 背景：本仓库的备份脚本需在 backup 一次性容器内运行，而镜像内是否自带这些命令
// 只能通过实机确认，故本工具的输出是「待实机验证清单」的输入。
// 用法：node scripts/backup/check-dependencies.mjs

import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

// 脚本内使用的 shell 内建命令与关键字，不属于外部命令依赖
const SHELL_BUILTINS = new Set([
  'set', 'if', 'then', 'else', 'elif', 'fi', 'for', 'do', 'done', 'while', 'until',
  'echo', 'exit', 'cd', 'return', 'trap', 'case', 'esac', 'function', 'local',
  'export', 'unset', 'read', 'printf', 'test', 'true', 'false', 'break', 'continue'
])

// 每个外部命令对应的提供方说明，用于与目标镜像能力对照
const COMMAND_PROVIDERS = {
  mysqldump: 'MySQL 客户端，由 mysql:8.0 镜像提供',
  mysql: 'MySQL 客户端，由 mysql:8.0 镜像提供',
  date: 'GNU coreutils，由基础发行版提供',
  mkdir: 'GNU coreutils，由基础发行版提供',
  find: 'GNU findutils，由基础发行版提供',
  tar: 'GNU tar，由基础发行版提供',
  gzip: 'gzip，由基础发行版提供',
  sha256sum: 'GNU coreutils，由基础发行版提供',
  basename: 'GNU coreutils，由基础发行版提供',
  dirname: 'GNU coreutils，由基础发行版提供',
  sort: 'GNU coreutils，由基础发行版提供',
  cat: 'GNU coreutils，由基础发行版提供',
  ls: 'GNU coreutils，由基础发行版提供',
  rm: 'GNU coreutils，由基础发行版提供'
}

const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url))
const TARGET_SCRIPTS = ['backup.sh', 'restore.sh']

/**
 * 将续行符拼接为逻辑行，避免跨行的命令与管道被截断。
 */
function joinContinuationLines(source) {
  return source.replace(/\\\r?\n\s*/g, ' ')
}

/**
 * 判断片段是否为变量赋值语句。
 * 变量赋值形如 NAME=value，其中 NAME 是合法的 shell 标识符，
 * 这类片段不是外部命令调用，必须排除以免产生误报。
 */
function isVariableAssignment(segment) {
  return /^[A-Za-z_][A-Za-z0-9_]*=/.test(segment)
}

/**
 * 提取命令替换语法内部调用的命令名。
 * 形如 "$(date ...)" 或 "$(cat file)" 的片段同样是真实的外部命令调用，
 * 若只按行首与管道切分将整条命令替换表达式当作参数处理，会产生漏报。
 */
function extractSubstitutionCommands(segment) {
  const found = []
  // 逐个扫描 $( 起始的命令替换，用括号计数找到与之配对的右括号
  for (let index = 0; index < segment.length; index += 1) {
    if (segment[index] !== '$' || segment[index + 1] !== '(') continue
    let depth = 1
    let cursor = index + 2
    while (cursor < segment.length && depth > 0) {
      if (segment[cursor] === '(') depth += 1
      if (segment[cursor] === ')') depth -= 1
      cursor += 1
    }
    const inner = segment.slice(index + 2, cursor - 1)
    // 命令替换内部可能还有管道，按管道切分后逐个取命令名
    for (const innerSegment of inner.split('|')) {
      const match = innerSegment.trim().match(/^([A-Za-z_][A-Za-z0-9_.-]*)/)
      if (match) found.push(match[1])
    }
    index = cursor - 1
  }
  return found
}

/**
 * 从脚本源码中提取每个命令位置上真正调用的外部命令名。
 * 处理要点：
 * 1. 命令可能出现在行首、管道后、逻辑运算符后或条件取反符之后；
 * 2. 命令替换语法内部调用的命令同样计入依赖；
 * 3. 变量赋值、重定向与逻辑运算符不作为命令计入；
 * 4. 命令后的参数与选项一律忽略，只取命令名本身。
 */
function extractCommands(source) {
  const found = new Set()
  for (const rawLine of joinContinuationLines(source).split(/\r?\n/)) {
    const line = rawLine.trim()
    if (line === '' || line.startsWith('#')) continue
    // 先扫描命令替换再剥离括号，否则 "$(date ...)" 的括号会被提前移除而丢失命令名。
    // 本行首词可能需要跳过条件关键字，故同时保留剥离关键字后的版本用于取首词。
    const withoutKeyword = line.replace(/^(if|then|elif|while|until|do|else)\s+/i, '')
    for (const substitutionCommand of extractSubstitutionCommands(withoutKeyword)) {
      if (!SHELL_BUILTINS.has(substitutionCommand)) found.add(substitutionCommand)
    }
    // 按管道与逻辑运算符切分，得到若干个独立的命令片段
    for (const segment of withoutKeyword.split(/\|\||&&|\||;/)) {
      const withoutNegation = segment.replace(/^\s*!+\s*/, '').replace(/[()]/g, ' ').trim()
      if (withoutNegation === '' || isVariableAssignment(withoutNegation)) continue
      const match = withoutNegation.match(/^([A-Za-z_][A-Za-z0-9_.-]*)/)
      if (!match) continue
      const command = match[1]
      if (SHELL_BUILTINS.has(command) || command.startsWith('-')) continue
      found.add(command)
    }
  }
  return found
}

const report = {}
for (const scriptName of TARGET_SCRIPTS) {
  const scriptPath = resolve(SCRIPT_DIR, scriptName)
  const commands = extractCommands(readFileSync(scriptPath, 'utf8'))
  report[scriptName] = [...commands].sort()
}

console.log('备份脚本外部命令依赖清单')
console.log('='.repeat(60))
for (const [scriptName, commands] of Object.entries(report)) {
  console.log(`\n${scriptName} 共依赖 ${commands.length} 个外部命令：`)
  for (const command of commands) {
    const provider = COMMAND_PROVIDERS[command] ?? '未登记提供方，需人工确认'
    console.log(`  - ${command.padEnd(12)} ${provider}`)
  }
}

const unknown = Object.values(report)
  .flat()
  .filter(command => !Object.hasOwn(COMMAND_PROVIDERS, command))
console.log(`\n未登记提供方的命令数量: ${unknown.length}`)
process.exit(unknown.length > 0 ? 1 : 0)
