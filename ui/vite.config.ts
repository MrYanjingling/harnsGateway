import { fileURLToPath } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { mockApiPlugin } from './mock'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const useMock = env.VITE_USE_MOCK === 'true'
  const gatewayTarget = env.VITE_GATEWAY_TARGET || 'http://127.0.0.1:32200'
  const dataTarget = env.VITE_DATA_TARGET || 'http://127.0.0.1:32201'

  return {
    plugins: [react(), tailwindcss(), ...(useMock ? [mockApiPlugin()] : [])],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      // 两套服务共用 /api/v1 前缀,按路径分流;网关路径必须写在前面。
      proxy: useMock
        ? undefined
        : {
            '/api/v1/devices': { target: gatewayTarget, changeOrigin: true },
            '/api/v1/gatewayMeta': { target: gatewayTarget, changeOrigin: true },
            '/api/v1/gatewayCpu': { target: gatewayTarget, changeOrigin: true },
            '/api/v1/gatewayMem': { target: gatewayTarget, changeOrigin: true },
            '/api/v1/gatewayDisk': { target: gatewayTarget, changeOrigin: true },
            '/api/v1/metrics': { target: dataTarget, changeOrigin: true },
            '/api/v1/kpi': { target: dataTarget, changeOrigin: true },
            '/api/v1/kpis': { target: dataTarget, changeOrigin: true },
            '/api/v1/tags': { target: dataTarget, changeOrigin: true },
            '/api/v1/fmcs': { target: dataTarget, changeOrigin: true },
            '/api/v1/escada': { target: dataTarget, changeOrigin: true },
          },
    },
    build: {
      outDir: 'dist',
    },
  }
})
