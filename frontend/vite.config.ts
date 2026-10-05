/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: { outDir: 'dist', emptyOutDir: true },
  // The self-reading cycle's one home is beside the setup page (installer/frontend/dist), which the
  // development server must be allowed to read as the build already does.
  server: { fs: { allow: ['..'] } },
  // ribbonkit's half of the page is linked in as a package (file:../ribbonkit) and read through the
  // link, so what it imports resolves from this folder's node_modules, as tsconfig's
  // preserveSymlinks does for the type checker.
  resolve: { preserveSymlinks: true },
  test: {
    // ribbonkit's half of the page is tested here too until it leaves as a package of its own; it is
    // source, so it is transformed rather than handed to Node as a built dependency.
    include: ['src/**/*.test.{ts,tsx}', 'node_modules/@oernster/ribbonkit/web/**/*.test.{ts,tsx}'],
    exclude: ['dist/**', 'wailsjs/**'],
    server: { deps: { inline: [/@oernster\/ribbonkit/] } },
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
  },
})
