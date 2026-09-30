import { rmSync } from 'node:fs';
import { resolve } from 'node:path';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import type { Plugin } from 'vite';

import { ocrAssets } from './vite-plugin-ocr';

// The mock service worker in public/ serves `npm run mocked`, a dev-server
// mode. Vite copies all of public/ into every build, so a production build
// would ship it to the board for nothing; drop it there unless the build
// itself is a mocked one.
function dropMockWorker(): Plugin {
  let outDir = '';
  let keep = false;
  return {
    name: 'drop-mock-worker',
    apply: 'build',
    configResolved(config) {
      outDir = resolve(config.root, config.build.outDir);
      keep = config.mode === 'mocked';
    },
    closeBundle() {
      if (!keep) rmSync(resolve(outDir, 'mockServiceWorker.js'), { force: true });
    }
  };
}

export default defineConfig({
  plugins: [react(), ocrAssets(), dropMockWorker()],
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
