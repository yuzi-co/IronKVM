// The text recognition engine. tesseract.js is imported here and only here,
// with a dynamic import, so that its code is fetched the first time someone
// reads text and not with the desktop.

// The languages the build carries a model for. vite-plugin-ocr.ts copies the
// models, and a language added there needs an entry here and a name in the
// locales.
export const ocrLanguages = ['eng'] as const;

export type OcrLanguage = (typeof ocrLanguages)[number];

export type OcrPhase = 'loading' | 'recognizing';

export type OcrProgress = {
  phase: OcrPhase;
  // From 0 to 1 within the phase.
  progress: number;
};

// The smallest WebAssembly module that uses a SIMD instruction, as
// wasm-feature-detect checks for it. The core build that ships needs SIMD, and
// a browser without it would fail inside the worker with an error that names
// nothing the user can act on.
const simdProbe = new Uint8Array([
  0, 97, 115, 109, 1, 0, 0, 0, 1, 5, 1, 96, 0, 1, 123, 3, 2, 1, 0, 10, 10, 1, 8, 0, 65, 0, 253, 15,
  253, 98, 11
]);

export function isOcrSupported() {
  try {
    return typeof WebAssembly === 'object' && WebAssembly.validate(simdProbe);
  } catch {
    return false;
  }
}

// The worker has to be a real URL on this origin rather than the Blob URL
// tesseract.js makes by default: the emscripten loader in the core finds the
// .wasm beside the script that runs it, and a Blob URL has no directory.
function assetUrl(name = '') {
  return new URL(`${import.meta.env.BASE_URL}${__OCR_ASSET_DIR__}/${name}`, window.location.href)
    .href;
}

// recognizeText reads the text in an image. Each call starts a worker and ends
// it, so the memory the engine takes, tens of megabytes, is held only while
// it reads.
export async function recognizeText(
  image: Blob,
  language: OcrLanguage,
  onProgress: (progress: OcrProgress) => void
): Promise<string> {
  const tesseract = await import('tesseract.js');
  // tesseract.js is CommonJS. The bundler puts its exports on default, and a
  // module that was ES would have them at the top.
  const { createWorker, OEM } = (
    'default' in tesseract ? tesseract.default : tesseract
  ) as typeof tesseract;

  const worker = await createWorker(language, OEM.LSTM_ONLY, {
    workerPath: assetUrl('worker.min.js'),
    // A path that ends in .js is loaded as it is. A directory would make the
    // worker pick a build by the browser's features, and only one ships.
    corePath: assetUrl('tesseract-core-simd-lstm.js'),
    langPath: assetUrl().replace(/\/$/, ''),
    workerBlobURL: false,
    // The browser caches the model by its URL, which changes with its content.
    // tesseract.js would also keep a copy in IndexedDB under the language code
    // alone, and serve that copy after an update changed the model.
    cacheMethod: 'none',
    logger: (message) => {
      onProgress({
        phase: message.status === 'recognizing text' ? 'recognizing' : 'loading',
        progress: message.progress
      });
    }
  });

  try {
    // Screens lay out columns and tables with runs of spaces, and a copy that
    // keeps them is easier to read.
    await worker.setParameters({ preserve_interword_spaces: '1' });
    const { data } = await worker.recognize(image);
    return data.text;
  } finally {
    await worker.terminate();
  }
}
