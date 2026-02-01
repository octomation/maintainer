import { defineConfig } from 'vitest/config';
import preact from '@preact/preset-vite';

// The service validates Host and, on mutations, Origin: the dev proxy rewrites
// both to the service's own address, so `npm run dev` works against a running
// `maintainer service run` on the default port.
const service = process.env.MAINTAINER_SERVICE ?? 'http://maintainer.localhost:17841';

export default defineConfig({
  plugins: [preact()],
  base: '/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    reportCompressedSize: false,
  },
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': {
        target: service,
        changeOrigin: true,
        headers: { Origin: service },
      },
    },
  },
  test: {
    environment: 'happy-dom',
    // A desktop viewport: the toolbar keeps every primary filter inline; narrow tests stub matchMedia.
    environmentOptions: { happyDOM: { width: 1440, height: 900 } },
    include: ['src/**/*.test.{ts,tsx}'],
    setupFiles: ['src/test/setup.ts'],
  },
});
