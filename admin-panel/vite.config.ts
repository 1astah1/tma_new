import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  // Панель живёт на https://tma.happ.xin/admin/ — Caddy срезает префикс,
  // но ссылки на ассеты внутри бандла должны его содержать.
  base: '/admin/',
  plugins: [react()],
  server: {
    host: true,
    port: 5174,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
      },
      '/uploads': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
})
