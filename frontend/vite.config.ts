import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Dev-server proxy mirrors the nginx routing used in the built container, so
// the frontend can always call same-origin relative paths ('/api', '/auth',
// '/ws') whether running via `npm run dev` or inside docker compose.
export default defineConfig({
  plugins: [vue()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/auth': { target: 'http://localhost:8080', changeOrigin: true },
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
})
