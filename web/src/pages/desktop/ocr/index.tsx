import { useCallback, useRef, useState } from 'react';
import { useAtom } from 'jotai';

import { getScreenshot } from '@/api/stream.ts';
import { ocrSelectingAtom } from '@/jotai/ocr.ts';

import { selectionToFrame, upscaleFactor } from './crop.ts';
import { OcrDialog, type OcrState } from './dialog.tsx';
import { isOcrSupported, recognizeText, type OcrLanguage } from './engine.ts';
import { OcrSelection, type OcrSelectionResult } from './selection.tsx';

function errorDetail(error: unknown) {
  return error instanceof Error ? error.message : String(error);
}

// cropFrame cuts the selection out of the frame the board sent, enlarged when
// that helps recognition. It returns null when the selection holds none of the
// picture.
async function cropFrame(jpeg: Blob, selected: OcrSelectionResult): Promise<Blob | null> {
  const bitmap = await createImageBitmap(jpeg);
  try {
    const crop = selectionToFrame(selected.selection, selected.picture, {
      width: bitmap.width,
      height: bitmap.height
    });
    if (!crop) {
      return null;
    }

    const scale = upscaleFactor(crop);
    const canvas = document.createElement('canvas');
    canvas.width = crop.width * scale;
    canvas.height = crop.height * scale;
    const context = canvas.getContext('2d');
    if (!context) {
      throw new Error('the browser gave no 2D canvas');
    }
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = 'high';
    context.drawImage(
      bitmap,
      crop.left,
      crop.top,
      crop.width,
      crop.height,
      0,
      0,
      canvas.width,
      canvas.height
    );

    // PNG, because a second round of JPEG would blur the text again.
    return await new Promise<Blob>((resolve, reject) => {
      canvas.toBlob(
        (blob) => (blob ? resolve(blob) : reject(new Error('the crop could not be encoded'))),
        'image/png'
      );
    });
  } finally {
    bitmap.close();
  }
}

// Ocr reads text off the host screen. The user drags a rectangle over the
// video; the board sends a full frame; the browser crops it to the rectangle
// and reads the text with tesseract.js. Nothing of the recognition runs on the
// board, which has no CPU to spare for it.
//
// The frame is taken when the drag ends, so the text read is the text the user
// saw when letting go.
export const Ocr = () => {
  const [selecting, setSelecting] = useAtom(ocrSelectingAtom);

  const [open, setOpen] = useState(false);
  const [state, setState] = useState<OcrState>({ phase: 'capturing' });
  const [image, setImage] = useState<Blob | null>(null);
  const [language, setLanguage] = useState<OcrLanguage>('eng');

  // Each capture or recognition takes a number, and only the latest one may
  // change the dialog. A slow one that the user abandoned then ends unseen.
  const runRef = useRef(0);

  const read = useCallback(async (crop: Blob, lang: OcrLanguage) => {
    const run = ++runRef.current;
    const isCurrent = () => run === runRef.current;

    setState({ phase: 'loading', progress: 0 });
    try {
      const text = await recognizeText(crop, lang, ({ phase, progress }) => {
        if (isCurrent()) setState({ phase, progress });
      });
      if (isCurrent()) setState({ phase: 'done', text: text.trimEnd() });
    } catch (error) {
      if (isCurrent()) setState({ phase: 'error', error: 'recognize', detail: errorDetail(error) });
    }
  }, []);

  async function handleSelect(selected: OcrSelectionResult) {
    const run = ++runRef.current;
    const isCurrent = () => run === runRef.current;

    setSelecting(false);
    setImage(null);
    setOpen(true);

    if (!isOcrSupported()) {
      setState({ phase: 'error', error: 'unsupported', detail: '' });
      return;
    }

    setState({ phase: 'capturing' });
    let crop: Blob | null;
    try {
      crop = await cropFrame(await getScreenshot(), selected);
    } catch (error) {
      if (isCurrent()) setState({ phase: 'error', error: 'capture', detail: errorDetail(error) });
      return;
    }
    if (!isCurrent()) return;
    if (!crop) {
      setState({ phase: 'error', error: 'outside', detail: '' });
      return;
    }

    setImage(crop);
    await read(crop, language);
  }

  const cancelSelection = useCallback(() => setSelecting(false), [setSelecting]);

  function handleLanguageChange(next: OcrLanguage) {
    setLanguage(next);
    if (image) {
      void read(image, next);
    }
  }

  function close() {
    runRef.current++;
    setOpen(false);
  }

  function selectAgain() {
    close();
    setSelecting(true);
  }

  return (
    <>
      {selecting && (
        <OcrSelection
          onSelect={(selected) => void handleSelect(selected)}
          onCancel={cancelSelection}
        />
      )}
      <OcrDialog
        open={open}
        state={state}
        image={image}
        language={language}
        onLanguageChange={handleLanguageChange}
        onTextChange={(text) => setState({ phase: 'done', text })}
        onSelectAgain={selectAgain}
        onClose={close}
      />
    </>
  );
};
