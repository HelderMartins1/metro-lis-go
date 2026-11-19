import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    svelte(),
    VitePWA({
      registerType: 'autoUpdate',
      // Our own service worker (src/sw.js), to receive push notifications.
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.js',
      injectManifest: {
        globPatterns: ['**/*.{js,css,html,svg,png,woff2}'],
      },
      includeAssets: ['icon.svg', 'apple-touch-icon.png'],
      manifest: {
        name: 'Cais',
        short_name: 'Cais',
        description: 'Tempos de espera do metro de Lisboa, estação a estação. App não oficial.',
        lang: 'pt-PT',
        theme_color: '#f6f6f3',
        background_color: '#f6f6f3',
        display: 'standalone',
        icons: [
          { src: 'icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: 'icon-512.png', sizes: '512x512', type: 'image/png' },
          { src: 'icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  build: {
    // The Go server embeds this folder (see web/web.go).
    outDir: '../web/dist',
    emptyOutDir: true,
  },
  server: {
    // src/lib/store.svelte.js imports ../models/network.json, outside this folder.
    fs: { allow: ['..'] },
    proxy: {
      '/api': { target: process.env.API_TARGET || 'http://localhost:8080', changeOrigin: true },
    },
  },
});
