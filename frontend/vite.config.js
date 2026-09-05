import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { readFileSync } from 'fs'

// 单一版本源：构建期从 package.json 读取，注入 __APP_VERSION__ 供 globalConfig 使用
const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf-8'))

const proxyTarget = process.env.VITE_PROXY_TARGET || 'https://miawa.cn'

// dev/preview 代理指向线上后端联调；如需指向本地后端设置 VITE_PROXY_TARGET 即可
const proxyRoutes = ['/api/v1', '/api/v2', '/download']

// https://vite.dev/config/
export default defineConfig({
  base: '/',
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version)
  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: Object.fromEntries(
      proxyRoutes.map((route) => [route, { target: proxyTarget, changeOrigin: true }])
    ),
  },
  preview: {
    proxy: Object.fromEntries(
      proxyRoutes.map((route) => [route, { target: proxyTarget, changeOrigin: true }])
    ),
  },
  build: {
    outDir: path.resolve(__dirname, '../web/default'),
    emptyOutDir: true,
  },
})
