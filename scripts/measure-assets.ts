#!/usr/bin/env -S node --import tsx

/**
 * 前台生产产物体积度量：按字体、JavaScript、样式表分类输出体积汇总，并按阈值断言。
 *
 * 该脚本解决的是「结论不可持续验证」的问题：字体 18.18 MB 这一体检结论原先完全依赖
 * build 目录恰好存在，执行一次构建清理后即无法复核。度量入口固化后，任意时刻都能
 * 复现分类体积，CI 亦在构建后自动打印数值。
 */

import {
  ASSET_KINDS,
  collectAssets,
  formatBytes,
  summarizeAssets,
  type AssetKind
} from './lib/assets'
import { colors } from './lib/terminal'
import { exit } from 'node:process'

// 前台生产构建产物目录。
const WEB_BUILD_DIRECTORY = 'apps/web/build'

// 字体总体积阈值：100 KB 覆盖少量图标字体等正常残留。
// 本地 CJK 字体若被误打回生产产物，总量会达到 MB 量级，因此该阈值能可靠拦住回流。
const FONT_BYTES_THRESHOLD = 100_000

// 单文件体积提醒阈值：超过该值的产物值得人工评估，默认 1 MB。
const LARGE_FILE_BYTES_THRESHOLD = 1_000_000

// 分类标签的中文名，用于汇总表输出。
const KIND_LABELS: Record<string, string> = {
  font: '字体',
  javascript: 'JavaScript',
  stylesheet: '样式表',
  sourcemap: '源映射',
  other: '其他'
}

// 体积提醒排除的分类：源映射体积随源码规模自然增长，提示其超限没有可执行的处置动作。
const KINDS_EXCLUDED_FROM_LARGE_FILE_NOTICE: AssetKind[] = ['sourcemap']

async function main(): Promise<void> {
  const assets = collectAssets(WEB_BUILD_DIRECTORY)
  if (assets.length === 0) {
    console.error(`❌ 未在 ${WEB_BUILD_DIRECTORY} 中找到任何产物文件`)
    console.error('   请先执行 pnpm run build:web 生成产物后再度量')
    exit(1)
  }

  const summary = summarizeAssets(assets)
  const totalBytes = assets.reduce((sum, asset) => sum + asset.bytes, 0)

  console.log(colors.bold(`📦 前台产物体积汇总（目录 ${WEB_BUILD_DIRECTORY}）`))
  for (const kind of ASSET_KINDS) {
    const label = KIND_LABELS[kind] ?? kind
    const bytes = summary.bytesByKind[kind]
    const count = summary.countByKind[kind]
    console.log(`   ${label.padEnd(12)} ${formatBytes(bytes).padStart(10)}  ${count} 个文件`)
  }
  console.log(
    `   ${'合计'.padEnd(12)} ${formatBytes(totalBytes).padStart(10)}  ${assets.length} 个文件`
  )

  // 超出提醒阈值的单文件按体积降序输出，便于定位体积来源。
  const largeFiles = assets
    .filter(asset => !KINDS_EXCLUDED_FROM_LARGE_FILE_NOTICE.includes(asset.kind))
    .filter(asset => asset.bytes >= LARGE_FILE_BYTES_THRESHOLD)
    .sort((left, right) => right.bytes - left.bytes)
  if (largeFiles.length > 0) {
    console.log(colors.yellow(`⚠️  体积超过 ${formatBytes(LARGE_FILE_BYTES_THRESHOLD)} 的单文件：`))
    for (const asset of largeFiles) {
      console.log(`   ${formatBytes(asset.bytes).padStart(10)}  ${asset.path}`)
    }
  }

  if (summary.bytesByKind.font > FONT_BYTES_THRESHOLD) {
    console.error(
      colors.red(
        `❌ 字体体积 ${formatBytes(summary.bytesByKind.font)} 超过阈值 ${formatBytes(FONT_BYTES_THRESHOLD)}，本地字体疑似回流生产产物`
      )
    )
    for (const asset of assets.filter(entry => entry.kind === 'font')) {
      console.error(`   ${formatBytes(asset.bytes).padStart(10)}  ${asset.path}`)
    }
    exit(1)
  }

  console.log(colors.green('✅ 产物体积度量完成，字体体积在阈值内'))
}

try {
  await main()
} catch (error) {
  const errorMessage = error instanceof Error ? error.message : String(error)
  console.error(`❌ 产物体积度量执行失败: ${errorMessage}`)
  exit(1)
}
