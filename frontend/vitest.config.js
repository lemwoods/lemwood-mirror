import { defineConfig } from 'vitest/config'
import path from 'path'
import { readFileSync } from 'fs'

// 与 vite.config.js 对齐：node 测试环境也需提供 __APP_VERSION__ 替换
const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf-8'))

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version)
  },
  test: {
    environment: 'node',
    include: ['tests/**/*.test.js']
  }
})
