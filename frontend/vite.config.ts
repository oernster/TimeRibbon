/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: { outDir: 'dist', emptyOutDir: true },
  // The self-reading cycle's one home is beside the setup page (installer/frontend/dist), which the
  // development server must be allowed to read as the build already does.
  server: { fs: { allow: ['..'] } },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
  },
})
