import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  base: '/',
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // 代理 key 必须用精确 API 前缀：用 '/api' 会把前端路由 /apidocs 也劫持转发到线上
      '/api/v1': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
      '/api/v2': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
      // 下载端点同样转发线上（vite 本地无 Go 后端，/download 会回退 SPA 主界面）
      '/download': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
    },
  },
  preview: {
    proxy: {
      '/api/v1': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
      '/api/v2': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
      '/download': {
        target: 'https://miawa.cn',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: path.resolve(__dirname, '../web/default'),
    emptyOutDir: true,
  },
})
