import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import type { Plugin } from 'vite';

// The files tesseract.js loads at run time, served from the board rather than
// from its default CDN: the board is often on a network with no way out, and
// the web UI must not depend on a third party to read its own screen.
//
// Only the SIMD build of the LSTM engine is shipped. WebAssembly SIMD is in
// every current browser (Chrome 91, Firefox 89, Safari 16.4), the web UI says
// so when a browser lacks it, and each further build (plain, relaxed SIMD, or
// with the legacy engine) would add about 4MB to the board's web files.
// The .wasm is shipped beside its loader rather than the .wasm.js that embeds
// it, because base64 makes the embedded copy a third larger.
//
// eng.traineddata is the 4.0.0_best_int model: integer LSTM weights, the one
// tesseract.js picks by default, and a quarter of the size of the float one.
const require = createRequire(import.meta.url);
const tesseractRequire = createRequire(require.resolve('tesseract.js'));

const sources: Record<string, string> = {
  'worker.min.js': require.resolve('tesseract.js/dist/worker.min.js'),
  'tesseract-core-simd-lstm.js': tesseractRequire.resolve(
    'tesseract.js-core/tesseract-core-simd-lstm.js'
  ),
  'tesseract-core-simd-lstm.wasm': tesseractRequire.resolve(
    'tesseract.js-core/tesseract-core-simd-lstm.wasm'
  ),
  'eng.traineddata.gz': require.resolve('@tesseract.js-data/eng/4.0.0_best_int/eng.traineddata.gz')
};

const contentTypes: Record<string, string> = {
  '.js': 'text/javascript',
  '.wasm': 'application/wasm',
  '.gz': 'application/octet-stream'
};

function readSources() {
  return Object.entries(sources).map(([name, path]) => ({ name, source: readFileSync(path) }));
}

// ocrAssets copies the files into the build and tells the app where they are.
//
// They go under assets/, where the server lets the browser cache a file for a
// year, in a directory named by a hash of their content. The worker finds the
// core and the model by fixed file names, so the files cannot carry a hash of
// their own the way Vite names the rest; naming the directory instead keeps
// the rule that a cached file under assets/ can never be the wrong one.
export function ocrAssets(): Plugin {
  const files = readSources();
  const hash = createHash('sha256');
  for (const { name, source } of files) {
    hash.update(name).update(source);
  }
  const dir = `assets/ocr-${hash.digest('hex').slice(0, 10)}`;

  return {
    name: 'ocr-assets',

    config() {
      return { define: { __OCR_ASSET_DIR__: JSON.stringify(dir) } };
    },

    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const path = req.url?.split('?')[0] ?? '';
        const file = files.find(({ name }) => path === `/${dir}/${name}`);
        if (!file) {
          next();
          return;
        }
        const extension = file.name.slice(file.name.lastIndexOf('.'));
        res.setHeader('Content-Type', contentTypes[extension] ?? 'application/octet-stream');
        res.end(file.source);
      });
    },

    generateBundle() {
      for (const { name, source } of files) {
        this.emitFile({ type: 'asset', fileName: `${dir}/${name}`, source });
      }
    }
  };
}
