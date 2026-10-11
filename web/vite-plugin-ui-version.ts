import { readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import type { Plugin } from 'vite';

// uiVersion writes dist/ui-version.json: the UI's version and the server API
// range it supports. The server reads it to name the bundled UI on the update
// page (server/service/webui), and scripts/package-ui.sh reads it for a UI
// package's manifest. The range itself is declared once, in
// src/lib/ui-api-range.json, which src/lib/server-contract.ts reads too.
//
// UI_VERSION in the environment sets the version; a release build passes the
// release's. Without it the build is "0.0.0-dev", which sorts below every
// release.
export function uiVersion(): Plugin {
  const version = process.env.UI_VERSION || '0.0.0-dev';
  let root = '';
  let outDir = '';
  return {
    name: 'ui-version',
    apply: 'build',
    configResolved(config) {
      root = config.root;
      outDir = resolve(config.root, config.build.outDir);
    },
    closeBundle() {
      const range = JSON.parse(readFileSync(resolve(root, 'src/lib/ui-api-range.json'), 'utf8'));
      const body = { version, api: { min: range.min, max: range.max } };
      writeFileSync(resolve(outDir, 'ui-version.json'), JSON.stringify(body) + '\n');
    }
  };
}
