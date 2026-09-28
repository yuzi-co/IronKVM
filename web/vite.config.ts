import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

import { ocrAssets } from './vite-plugin-ocr';

export default defineConfig({
  plugins: [react(), ocrAssets()],
  resolve: {
    tsconfigPaths: true,
    alias: [
      // react-simple-keyboard ships an ES module at build/index.modern.esm.js
      // but advertises only "main", which points at the CommonJS build. Vite 8
      // resolves that build, finds no named exports it can bind, and gives the
      // whole module namespace the name "default". The component then arrives
      // as { KeyboardReact, default } rather than as a function, and rendering
      // it fails with "type is invalid ... but got: object".
      //
      // Point the bare specifier at the ES module the package already ships.
      // The regex anchors both ends, because a plain string also matches the
      // subpaths, and the keyboard imports its own stylesheet from one of them.
      {
        find: /^react-simple-keyboard$/,
        replacement: 'react-simple-keyboard/build/index.modern.esm.js'
      }
    ]
  },
  server: {
    port: 3001
  },
  build: {
    chunkSizeWarningLimit: 1024
  }
});
