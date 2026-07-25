import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';
import { VitePWA } from 'vite-plugin-pwa';

/** Proxy API through the Vite origin so HTTPS (ngrok) has no mixed content. */
const apiProxy = {
  '/v1': {
    target: 'http://127.0.0.1:8080',
    changeOrigin: true,
  },
  '/healthz': {
    target: 'http://127.0.0.1:8080',
    changeOrigin: true,
  },
} as const;

export default defineConfig({
  plugins: [
    preact(),
    VitePWA({
      // Registration is wired in main.tsx (step 3).
      injectRegister: false,
      registerType: 'autoUpdate',
      // Keep existing public/manifest.webmanifest linked from index.html.
      manifest: false,
      includeAssets: [
        'favicon.svg',
        'apple-touch-icon.png',
        'icons/icon-192.png',
        'icons/icon-512.png',
        'icons/icon-192.svg',
        'icons/icon-512.svg',
        'icons.svg',
      ],
      workbox: {
        navigateFallback: '/index.html',
        // Do not treat missing static files / API paths as SPA routes.
        navigateFallbackDenylist: [/^\/api/, /^\/v1/, /^\/healthz/, /\.[^/]+$/],
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest,woff2}'],
        // Must hit the network so free ngrok can show Visit Site in the PWA.
        globIgnores: ['**/tunnel-unlock.html'],
      },
    }),
  ],
  // Same origin for `dev` and `preview` so the iPhone Home Screen icon stays valid.
  // Allow ngrok Host headers (URL changes each session; leading dot = any subdomain).
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    allowedHosts: ['.ngrok-free.app', '.ngrok-free.dev', '.ngrok.io'],
    proxy: { ...apiProxy },
  },
  preview: {
    host: true,
    port: 5173,
    strictPort: true,
    allowedHosts: ['.ngrok-free.app', '.ngrok-free.dev', '.ngrok.io'],
    proxy: { ...apiProxy },
  },
});
