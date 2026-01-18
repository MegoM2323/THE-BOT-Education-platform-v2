import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [
    react({
      jsxRuntime: 'automatic',
      include: /\.(jsx|tsx|js)$/,
    }),
  ],
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
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.js', './src/test/setup.js'],
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.spec.js',
        '**/*.test.js',
        '**/vite.config.js',
        '**/vitest.config.js',
      ],
      include: [
        'src/**/*.{js,jsx}',
      ],
    },
    include: [
      'src/**/*.{test,spec}.{js,jsx}',
      'test-unit/**/*.{test,spec}.{js,jsx}',
    ],
    deps: {
      inline: [/@testing-library/],
    },
    server: {
      deps: {
        inline: [/@testing-library/],
      },
    },
    transformMode: {
      web: [/\.[jt]sx?$/],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
    extensions: ['.js', '.mjs', '.jsx', '.ts', '.tsx', '.json'],
    mainFields: ['module', 'jsnext:main', 'jsnext'],
  },
});
