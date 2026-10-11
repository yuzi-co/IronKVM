// Writes a UI package's manifest.json for a built web/dist to stdout.
//
//   node scripts/ui-manifest.mjs <dist dir>
//
// The version and the server API range come from <dist>/ui-version.json, which
// the build writes (vite-plugin-ui-version.ts). Every regular file under dist is
// listed with its SHA-256; anything that is not a regular file or a directory
// stops the script, because the server refuses such a package anyway.
// server/service/webui/package.go is the reader, and the format is described
// there and in docs/specs/2026-10-11-ui-packages.md.

import { createHash } from 'node:crypto';
import { lstatSync, readdirSync, readFileSync } from 'node:fs';
import { join, relative, sep } from 'node:path';

const dist = process.argv[2];
if (!dist) {
  console.error('usage: node scripts/ui-manifest.mjs <dist dir>');
  process.exit(2);
}

const stamp = JSON.parse(readFileSync(join(dist, 'ui-version.json'), 'utf8'));
const files = {};

function walk(dir) {
  for (const name of readdirSync(dir).sort()) {
    const path = join(dir, name);
    const st = lstatSync(path);
    if (st.isDirectory()) {
      walk(path);
    } else if (st.isFile()) {
      const rel = relative(dist, path).split(sep).join('/');
      files[rel] = createHash('sha256').update(readFileSync(path)).digest('hex');
    } else {
      console.error(`ui-manifest: ${path} is neither a file nor a directory`);
      process.exit(1);
    }
  }
}

walk(dist);
if (!files['index.html']) {
  console.error('ui-manifest: no index.html in ' + dist);
  process.exit(1);
}

const manifest = {
  format: 1,
  name: 'ironkvm-ui',
  version: stamp.version,
  api: { min: stamp.api.min, max: stamp.api.max },
  files
};
process.stdout.write(JSON.stringify(manifest, null, 2) + '\n');
