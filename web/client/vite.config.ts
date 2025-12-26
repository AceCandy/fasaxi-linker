import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig(async () => {
  // 强制使用localhost避免网络IP问题
  const proxyTarget = `http://localhost:9090`
  console.log(`vite proxy target: ${proxyTarget}`)
  return {
    build: {
      outDir: '../dist/web',
    },
    server: {
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './'),
      },
    },
    plugins: [
      vue(),
      vuetify({ autoImport: true }),
    ],
  }
})
