// Writes a gzipped copy next to every compressible asset in dist.
//
// The device serves these from a 1GHz core, so compressing per request would
// cost more CPU than it has to spare. The server picks the .gz up when the
// client advertises gzip support. For a client that does not, the server
// decompresses the .gz itself (server/router/static.go), so the original is
// deleted once its copy exists: keeping both doubled the flash the web files
// take. Two kinds of file stay as they are. index.html and the assets it names
// are read as plain files by the server's UI package check (an older server
// reads them that way too), and they are what every client fetches first.
//
// Uses only Node built-ins on purpose: no extra dependency for a build step.

import { gzipSync, constants } from 'node:zlib';
import { readdir, readFile, rm, stat, writeFile } from 'node:fs/promises';
import { join, extname } from 'node:path';
import { fileURLToPath } from 'node:url';

// fileURLToPath, not URL.pathname: the latter yields "/D:/..." on Windows.
const DIST = fileURLToPath(new URL('../dist/', import.meta.url));

const COMPRESSIBLE = new Set([
  '.css',
  '.html',
  '.js',
  '.json',
  '.map',
  '.mjs',
  '.svg',
  '.txt',
  '.wasm',
  '.xml'
]);

// Below this the gzip header costs more than it saves.
const MIN_BYTES = 1024;

async function* walk(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* walk(full);
    } else if (entry.isFile()) {
      yield full;
    }
  }
}

// The pattern the server's package check uses to find what index.html loads
// (server/service/webui/package.go).
const ASSET_REFERENCE = /(?:src|href)="\/?(assets\/[^"?#]+)"/g;

async function entryFiles() {
  const index = await readFile(join(DIST, 'index.html'), 'utf8');
  const keep = new Set([join(DIST, 'index.html')]);
  for (const match of index.matchAll(ASSET_REFERENCE)) {
    keep.add(join(DIST, ...match[1].split('/')));
  }
  return keep;
}

async function main() {
  const keep = await entryFiles();
  let compressed = 0;
  let saved = 0;
  let removed = 0;

  for await (const file of walk(DIST)) {
    if (file.endsWith('.gz') || !COMPRESSIBLE.has(extname(file))) continue;

    const { size } = await stat(file);
    if (size < MIN_BYTES) continue;

    const source = await readFile(file);
    const gzipped = gzipSync(source, { level: constants.Z_BEST_COMPRESSION });

    // A copy that is not smaller would only waste flash.
    if (gzipped.length >= source.length) continue;

    await writeFile(`${file}.gz`, gzipped);
    compressed += 1;
    saved += source.length - gzipped.length;

    if (!keep.has(file)) {
      await rm(file);
      removed += 1;
    }
  }

  console.log(
    `precompressed ${compressed} files, saving ${(saved / 1024).toFixed(0)} KB over the wire; ` +
      `removed ${removed} originals`
  );
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
