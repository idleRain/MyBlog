#!/usr/bin/env -S node --import tsx

/**
 * 生产构建的字体体积门禁：扫描前台 build 产物中的字体文件并断言总量低于阈值。
 * 字体来源经 vite 的 #fonts 别名按构建模式注入（dev 本地 / prod CDN）后，
 * 生产产物不应再包含本地 CJK 字体；若有人在 prod 误把本地字体改回构建，
 * 由本门禁在 CI 拦截。web 目录产物含图标等小字体文件属正常范围，按总量放宽。
 */

import { readdirSync, statSync } from 'node:fs'
import { exit } from 'node:process'
import { join } from 'node:path'

// 字体产物阈值：100 KB 覆盖少量 UI 图标字体等正常残留，本地 CJK 字体总量约 18 MB 远超阈值。
const FONT_BYTES_THRESHOLD = 100_000

// 前台生产构建产物目录。
const WEB_BUILD_DIRECTORY = 'apps/web/build'

// 参与统计的字体扩展名，woff2 为现代浏览器主格式，woff 为旧格式兜底。
const FONT_EXTENSIONS = ['.woff', '.woff2']

// 递归收集目录下的字体文件路径。
function collectFontFiles(directory: string): string[] {
  const fonts: string[] = []
  try {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const entryPath = join(directory, entry.name)
      if (entry.isDirectory()) {
        fonts.push(...collectFontFiles(entryPath))
        continue
      }
      if (FONT_EXTENSIONS.some(extension => entry.name.toLowerCase().endsWith(extension))) {
        fonts.push(entryPath)
      }
    }
  } catch (error) {
    if (
      !(error instanceof Error) ||
      !String((error as NodeJS.ErrnoException).code).includes('ENOENT')
    ) {
      throw error
    }
  }
  return fonts
}

async function main(): Promise<void> {
  const fontFiles = collectFontFiles(WEB_BUILD_DIRECTORY)
  const totalBytes = fontFiles.reduce((sum, path) => sum + statSync(path).size, 0)

  console.log(`🌐 字体体积门禁：产物共 ${fontFiles.length} 个字体文件，合计 ${totalBytes} 字节`)

  if (totalBytes > FONT_BYTES_THRESHOLD) {
    console.error(`❌ 超过 ${FONT_BYTES_THRESHOLD} 字节阈值，本地字体疑似回流生产产物：`)
    for (const path of fontFiles) {
      console.error(`   ${path}`)
    }
    exit(1)
  }
  console.log('✅ 字体体积门禁通过')
}

try {
  await main()
} catch (error) {
  const errorMessage = error instanceof Error ? error.message : String(error)
  console.error(`❌ 字体体积门禁执行失败: ${errorMessage}`)
  exit(1)
}
