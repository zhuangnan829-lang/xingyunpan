/**
 * Bundle Size Analysis Script
 * 
 * This script analyzes the production build output to ensure
 * optimal bundle sizes and code splitting.
 */

import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const distPath = path.resolve(__dirname, '../dist/assets')

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

function analyzeBundle() {
  console.log('\n📦 Bundle Size Analysis\n')
  console.log('=' .repeat(60))

  if (!fs.existsSync(distPath)) {
    console.error('❌ Build directory not found. Please run "npm run build" first.')
    process.exit(1)
  }

  const files = fs.readdirSync(distPath)
  const jsFiles = files.filter(f => f.endsWith('.js'))
  const cssFiles = files.filter(f => f.endsWith('.css'))

  let totalJsSize = 0
  let totalCssSize = 0

  console.log('\n📄 JavaScript Files:')
  console.log('-'.repeat(60))
  
  jsFiles.forEach(file => {
    const filePath = path.join(distPath, file)
    const stats = fs.statSync(filePath)
    totalJsSize += stats.size
    
    const sizeStr = formatBytes(stats.size)
    const status = stats.size > 500 * 1024 ? '⚠️ ' : '✅'
    console.log(`${status} ${file.padEnd(40)} ${sizeStr}`)
  })

  console.log('\n🎨 CSS Files:')
  console.log('-'.repeat(60))
  
  cssFiles.forEach(file => {
    const filePath = path.join(distPath, file)
    const stats = fs.statSync(filePath)
    totalCssSize += stats.size
    
    const sizeStr = formatBytes(stats.size)
    const status = stats.size > 200 * 1024 ? '⚠️ ' : '✅'
    console.log(`${status} ${file.padEnd(40)} ${sizeStr}`)
  })

  console.log('\n📊 Summary:')
  console.log('-'.repeat(60))
  console.log(`Total JavaScript: ${formatBytes(totalJsSize)}`)
  console.log(`Total CSS: ${formatBytes(totalCssSize)}`)
  console.log(`Total Assets: ${formatBytes(totalJsSize + totalCssSize)}`)

  console.log('\n🎯 Performance Targets:')
  console.log('-'.repeat(60))
  
  const totalSize = totalJsSize + totalCssSize
  const targetSize = 1 * 1024 * 1024 // 1MB target for initial load
  
  if (totalSize < targetSize) {
    console.log(`✅ Total size (${formatBytes(totalSize)}) is under target (${formatBytes(targetSize)})`)
  } else {
    console.log(`⚠️  Total size (${formatBytes(totalSize)}) exceeds target (${formatBytes(targetSize)})`)
    console.log('   Consider further code splitting or lazy loading.')
  }

  // Check for large chunks
  const largeChunks = jsFiles.filter(file => {
    const filePath = path.join(distPath, file)
    const stats = fs.statSync(filePath)
    return stats.size > 500 * 1024 // 500KB
  })

  if (largeChunks.length > 0) {
    console.log('\n⚠️  Large Chunks Detected:')
    console.log('-'.repeat(60))
    largeChunks.forEach(file => {
      const filePath = path.join(distPath, file)
      const stats = fs.statSync(filePath)
      console.log(`   ${file}: ${formatBytes(stats.size)}`)
    })
    console.log('   Consider splitting these chunks further.')
  }

  console.log('\n' + '='.repeat(60) + '\n')
}

analyzeBundle()
