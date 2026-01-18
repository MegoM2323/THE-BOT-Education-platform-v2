import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// Плагин для правильной установки favicon
const faviconPlugin = {
  name: 'fix-favicon',
  transformIndexHtml: {
    order: 'pre',
    handler(html) {
      // Заменяем все старые ссылки на favicon на новые
      html = html.replace(/<link[^>]*rel="icon"[^>]*href="\/thebot-favicon\.[^"]*"[^>]*>/g, '');
      html = html.replace(/<link[^>]*rel="apple-touch-icon"[^>]*href="\/thebot-favicon\.[^"]*"[^>]*>/g, '');

      // Добавляем правильный favicon в head
      html = html.replace(
        /<\/head>/,
        '<link rel="icon" type="image/png" href="/favicon.png">\n  </head>'
      );
      return html;
    }
  }
};

export default defineConfig({
  plugins: [
    // faviconPlugin,
    react({
      jsxRuntime: 'automatic',
      include: /\.(jsx|tsx|js)$/,
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
    extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json'],
  },
  esbuild: {
    loader: 'jsx',
    include: /\.(jsx?|tsx?)$/,
    exclude: [],
  },
  optimizeDeps: {
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
        '.jsx': 'jsx',
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => {
          // Добавляем /v1 для Go backend: /api/auth/login -> /api/v1/auth/login
          if (!path.includes('/api/v1')) {
            return path.replace('/api/', '/api/v1/');
          }
          return path;
        },
        // Явно пробрасываем X-CSRF-Token header из ответа backend'а
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes, req, res) => {
            // X-CSRF-Token уже должен быть в proxyRes.headers,
            // но убедимся что он передается клиенту
            const csrfToken = proxyRes.headers['x-csrf-token'];
            if (csrfToken) {
              res.setHeader('X-CSRF-Token', csrfToken);
            }
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
});
