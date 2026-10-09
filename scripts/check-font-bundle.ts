#!/usr/bin/env -S node --import tsx

/**
 * 生产构建的字体体积门禁：扫描前台 build 产物中的字体文件并断言总量低于阈值。
 * 字体来源经 vite 的 #fonts 别名按构建模式注入（dev 本地 / prod CDN）后，
 * 生产产物不应再包含本地 CJK 字体；若有人在 prod 误把本地字体改回构建，
 * 由本门禁在 CI 拦截。web 目录产物含图标等小字体文件属正常范围，按总量放宽。
 *
 * 扫描与分类口径取自 scripts/lib/assets.ts，与体积度量脚本共用同一实现。
 */

import { collectAssets, formatBytes } from './lib/assets'
import { colors } from './lib/terminal'
import { exit } from 'node:process'

// 字体产物阈值：100 KB 覆盖少量 UI 图标字体等正常残留，本地 CJK 字体总量约 18 MB 远超阈值。
const FONT_BYTES_THRESHOLD = 100_000

// 前台生产构建产物目录。
const WEB_BUILD_DIRECTORY = 'apps/web/build'

async function main(): Promise<void> {
  const fontFiles = collectAssets(WEB_BUILD_DIRECTORY).filter(asset => asset.kind === 'font')
  const totalBytes = fontFiles.reduce((sum, asset) => sum + asset.bytes, 0)

  console.log(`🌐 字体体积门禁：产物共 ${fontFiles.length} 个字体文件，合计 ${totalBytes} 字节`)

  if (totalBytes > FONT_BYTES_THRESHOLD) {
    console.error(
      colors.red(`❌ 超过 ${FONT_BYTES_THRESHOLD} 字节阈值，本地字体疑似回流生产产物：`)
    )
    for (const asset of fontFiles) {
      console.error(`   ${formatBytes(asset.bytes).padStart(10)}  ${asset.path}`)
    }
    exit(1)
  }
  console.log(colors.green('✅ 字体体积门禁通过'))
}

try {
  await main()
} catch (error) {
  const errorMessage = error instanceof Error ? error.message : String(error)
  console.error(`❌ 字体体积门禁执行失败: ${errorMessage}`)
  exit(1)
}
