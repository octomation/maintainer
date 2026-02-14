import { defineConfig } from 'vitest/config';
import preact from '@preact/preset-vite';

// The service validates Host and, on mutations, Origin: the dev proxy rewrites
// both to the service's own address. The data is served only at the address of
// the current run: set MAINTAINER_SERVICE to what `maintainer service open
// --print` prints; the plain default answers the API with 401.
const service = (process.env.MAINTAINER_SERVICE ?? 'http://maintainer.localhost:17841').replace(/\/$/, '');

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
