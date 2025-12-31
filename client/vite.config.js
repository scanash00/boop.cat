// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  root: path.resolve(process.cwd(), 'client'),
  plugins: [react()],
  build: {
    outDir: path.resolve(process.cwd(), 'client/dist'),
    emptyOutDir: true,

    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-icons': ['lucide-react']
        }
      }
    },

    minify: 'esbuild',

    target: 'es2020',

    sourcemap: false,

    cssCodeSplit: true
  },

  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom', 'lucide-react']
  }
});
