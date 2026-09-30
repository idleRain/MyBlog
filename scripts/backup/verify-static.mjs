#!/usr/bin/env node
// 备份与编排交付物的静态校验套件。
// 用途：本机未安装 Docker 且无 bash，通过静态解析对编排文件、备份脚本与操作手册做可复核检查。
// 校验项：
//   1. docker-compose.yml 的 YAML 语法正确性；
//   2. 编排内五个服务齐备，并核对资源限制、健康检查、备份服务的 profile 与卷挂载等约定项；
//   3. deploy/.env.example 的 dotenv 键值语法正确性；
//   4. 操作手册与备份脚本之间的环境变量契约一致；
//   5. 注释规范禁用词未出现在备份相关文件中。
// 脚本的外部命令依赖清单由同目录的 check-dependencies.mjs 负责核对。
// 用法：在仓库根目录执行 node scripts/backup/verify-static.mjs
// 退出码：全部通过为 0，存在失败项为 1。

import { readFileSync, existsSync } from 'node:fs'
import { dirname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parse as parseYaml } from 'yaml'

// 由本文件位置推导仓库根目录，避免硬编码任何绝对路径
const REPO_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')

// 注释与文案规范中的禁用词，出现即视为违规。
// 词表以字符拼接方式构造，避免本文件因持有词表字面量而在自检中被判定为违规。
const FORBIDDEN_WORDS = [
  ['兜', '底'].join(''),
  ['搞', '定'].join(''),
  ['干', '掉'].join(''),
  ['弄', '好'].join('')
]

// 编排内应当存在的全部服务，备份任务为一次性 job 服务
const EXPECTED_SERVICES = ['mysql', 'server', 'web', 'nginx', 'backup']

// 常驻服务，需要声明健康检查与资源限制
const RESIDENT_SERVICES = ['mysql', 'server', 'web', 'nginx']

// 需要资源限制的全部服务。
// 备份任务虽为一次性 job，仍需限制内存与 CPU，避免 mysqldump 导出大库时耗尽宿主机内存。
const SERVICES_REQUIRING_LIMITS = [...RESIDENT_SERVICES, 'backup']

// 需要健康检查的服务仅限常驻服务。
// 一次性 job 容器没有「健康」语义，且 docker compose run 不消费 healthcheck，故排除 backup。
const SERVICES_REQUIRING_HEALTHCHECK = RESIDENT_SERVICES

// 文档提及但不由备份脚本定义的变量，属跨层变量，不纳入脚本契约核对
const NON_SCRIPT_VARIABLES = new Set([
  'BACKUP_HOST_DIR', // 由 docker-compose.yml 的绑定挂载提供
  'BACKUP_RETENTION_DAYS', // 由 .env 提供，编排读取后注入为脚本的 RETENTION_DAYS
  'MAILTO' // 由 cron 自身的告警机制提供
])

const checks = []

function recordCheck(name, passed, detail) {
  checks.push({ name, passed, detail })
}

// 输出相对仓库根的路径，使校验结果不依赖执行机器的绝对路径
function toRelativePath(absolutePath) {
  return relative(REPO_ROOT, absolutePath).split('\\').join('/')
}

/**
 * 读取仓库内文件，文件缺失时返回 null 并记录失败项。
 */
function readRepoFile(relativePath) {
  const absolutePath = resolve(REPO_ROOT, relativePath)
  if (!existsSync(absolutePath)) {
    recordCheck(`文件存在 ${relativePath}`, false, '文件不存在')
    return null
  }
  return readFileSync(absolutePath, 'utf8')
}

/**
 * 校验编排文件：YAML 可解析、服务齐备、资源限制与健康检查均已声明。
 * 全部断言集中在一次解析内完成，任一项不满足即判定该项失败。
 */
function verifyComposeFile() {
  const relativePath = 'docker-compose.yml'
  const content = readRepoFile(relativePath)
  if (content === null) return

  let document
  try {
    document = parseYaml(content)
  } catch (error) {
    recordCheck(`YAML 解析 ${relativePath}`, false, `解析失败: ${error.message}`)
    return
  }
  recordCheck(`YAML 解析 ${relativePath}`, true, `卷: ${Object.keys(document?.volumes ?? {}).join(', ')}`)

  const services = document?.services ?? {}
  const missingServices = EXPECTED_SERVICES.filter(name => !Object.hasOwn(services, name))
  recordCheck(
    `${relativePath} 服务齐备`,
    missingServices.length === 0,
    `已声明: ${Object.keys(services).join(', ')}` +
      (missingServices.length > 0 ? `; 缺失: ${missingServices.join(', ')}` : '')
  )

  // 资源限制：确认 deploy.resources.limits 的 memory 与 cpus 均已给出
  const limitsMissing = SERVICES_REQUIRING_LIMITS.filter(name => {
    const limits = services[name]?.deploy?.resources?.limits
    return !limits || limits.memory === undefined || limits.cpus === undefined
  })
  recordCheck(
    `${relativePath} 资源限制`,
    limitsMissing.length === 0,
    limitsMissing.length === 0
      ? `${SERVICES_REQUIRING_LIMITS.length} 个服务均已声明 memory 与 cpus`
      : `未声明 limits 的服务: ${limitsMissing.join(', ')}`
  )

  // 健康检查：确认四个常驻服务均声明了 test
  const healthcheckMissing = SERVICES_REQUIRING_HEALTHCHECK.filter(
    name => services[name]?.healthcheck?.test === undefined
  )
  recordCheck(
    `${relativePath} 健康检查`,
    healthcheckMissing.length === 0,
    healthcheckMissing.length === 0
      ? `${SERVICES_REQUIRING_HEALTHCHECK.length} 个常驻服务均已声明 test`
      : `未声明 healthcheck 的服务: ${healthcheckMissing.join(', ')}`
  )

  // 备份为一次性 job，不应声明健康检查；此项为反向断言，防止后续误加语义不符的探针
  const backupHasHealthcheck = services.backup?.healthcheck !== undefined
  recordCheck(
    `${relativePath} 备份服务未声明健康检查`,
    !backupHasHealthcheck,
    backupHasHealthcheck
      ? '一次性 job 不应声明 healthcheck，docker compose run 不消费该配置'
      : '符合一次性 job 语义'
  )

  // 备份服务：确认以 profile 方式声明，避免被 compose up 默认拉起为常驻容器
  const backupProfiles = services.backup?.profiles ?? []
  recordCheck(
    `${relativePath} 备份服务以 profile 声明`,
    backupProfiles.length > 0,
    backupProfiles.length > 0 ? `profiles: ${backupProfiles.join(', ')}` : '未声明 profiles'
  )

  // 备份服务：确认已挂载 uploads 卷，否则无法打包媒体文件
  const backupVolumes = services.backup?.volumes ?? []
  const mountsUploads = backupVolumes.some(entry => String(entry).includes('uploads'))
  recordCheck(
    `${relativePath} 备份服务挂载 uploads`,
    mountsUploads,
    mountsUploads ? backupVolumes.join('; ') : '未挂载 uploads 卷'
  )
}

/**
 * 校验指定文件不含注释规范中的禁用词。
 */
function verifyWording(relativePaths) {
  const violations = []
  for (const relativePath of relativePaths) {
    const content = readRepoFile(relativePath)
    if (content === null) continue
    for (const word of FORBIDDEN_WORDS) {
      if (content.includes(word)) violations.push(`${relativePath} 含禁用词「${word}」`)
    }
  }
  recordCheck('注释规范禁用词检查', violations.length === 0, violations.join('; ') || '无命中')
}

/**
 * 校验操作手册引用的脚本环境变量都在脚本内确有定义，防止文档与实现漂移。
 * 仅扫描手册中标注为环境变量契约的章节，避免把示例命令里的其他全大写标识符
 * 误判为环境变量，从而产生与契约无关的误报。
 */
function verifyEnvContract(docPath, scriptPaths) {
  const docContent = readRepoFile(docPath)
  if (docContent === null) return
  const contractSection = extractContractSection(docContent)
  if (contractSection === null) {
    recordCheck('手册与脚本的环境变量契约一致性', false, '未找到环境变量契约章节')
    return
  }
  const scriptContent = scriptPaths
    .map(relativePath => readRepoFile(relativePath) ?? '')
    .join('\n')
  // 契约章节中的环境变量以反引号包裹，只取全大写下划线形式的标识符
  const documentedVariables = new Set(
    [...contractSection.matchAll(/`([A-Z][A-Z0-9_]{2,})`/g)].map(match => match[1])
  )
  const missing = [...documentedVariables].filter(
    variable => !scriptContent.includes(variable) && !NON_SCRIPT_VARIABLES.has(variable)
  )
  recordCheck(
    '手册与脚本的环境变量契约一致性',
    missing.length === 0,
    missing.length === 0
      ? `已核对 ${documentedVariables.size} 个变量`
      : `脚本中未定义: ${missing.join(', ')}`
  )
}

/**
 * 提取手册中的环境变量契约章节，范围为「脚本环境变量契约」标题到下一个同级标题之间。
 */
function extractContractSection(docContent) {
  const lines = docContent.split(/\r?\n/)
  const startIndex = lines.findIndex(line => line.includes('脚本环境变量契约'))
  if (startIndex === -1) return null
  const collected = []
  for (let index = startIndex + 1; index < lines.length; index += 1) {
    if (/^## /.test(lines[index])) break
    collected.push(lines[index])
  }
  return collected.join('\n')
}

/**
 * 校验 dotenv 文件的键值语法，并可要求必须声明指定键。
 * deploy/.env.example 采用 dotenv 格式而非 YAML，故不适用 YAML 解析。
 */
function verifyDotenv(relativePath, requiredKeys) {
  const content = readRepoFile(relativePath)
  if (content === null) return
  const malformedLines = []
  const declaredKeys = []
  for (const [index, line] of content.split(/\r?\n/).entries()) {
    const trimmed = line.trim()
    if (trimmed === '' || trimmed.startsWith('#')) continue
    const match = trimmed.match(/^([A-Z][A-Z0-9_]*)=(.*)$/)
    if (!match) {
      malformedLines.push(`第 ${index + 1} 行`)
      continue
    }
    declaredKeys.push(match[1])
  }
  const missingKeys = requiredKeys.filter(key => !declaredKeys.includes(key))
  const problems = []
  if (malformedLines.length > 0) problems.push(`格式非法行: ${malformedLines.join(', ')}`)
  if (missingKeys.length > 0) problems.push(`缺少必需键: ${missingKeys.join(', ')}`)
  recordCheck(
    `dotenv 格式 ${relativePath}`,
    problems.length === 0,
    problems.length === 0 ? `已声明键: ${declaredKeys.join(', ')}` : problems.join('; ')
  )
}

verifyComposeFile()

// BACKUP_HOST_DIR 当前未登记在模板中，由编排文件提供默认值，故此处不设为必需键。
// 待该变量补登 deploy/.env.example 后，可将其加入下方必需键列表以获得守护。
verifyDotenv('deploy/.env.example', [])

verifyWording([
  'scripts/backup/backup.sh',
  'scripts/backup/restore.sh',
  'scripts/backup/check-dependencies.mjs',
  'scripts/backup/verify-static.mjs',
  'docs/operations/backup-restore.md'
])

verifyEnvContract('docs/operations/backup-restore.md', [
  'scripts/backup/backup.sh',
  'scripts/backup/restore.sh'
])

console.log('备份与编排交付物静态校验')
console.log('='.repeat(64))
let failedCount = 0
for (const check of checks) {
  if (!check.passed) failedCount += 1
  console.log(`[${check.passed ? '通过' : '失败'}] ${check.name}`)
  if (check.detail) console.log(`        ${check.detail}`)
}
console.log(`\n合计 ${checks.length} 项，失败 ${failedCount} 项`)
process.exit(failedCount > 0 ? 1 : 0)
