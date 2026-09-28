import { atom } from 'jotai';

import type { PasteCheck, PasteStatus } from '@/api/hid.ts';
import * as storage from '@/lib/localstorage.ts';

// Layouts the server can type on, in the order the menu lists them.
export const pasteLayoutIds = [
  'us',
  'uk',
  'de',
  'fr',
  'es',
  'it',
  'pt-br',
  'se',
  'ru',
  'ja',
  'ko'
] as const;

// The host's layout is not the browser's, but they agree more often than not,
// so the browser's language is the first guess until the user picks one.
function guessPasteLayout() {
  const language = navigator.language.toLowerCase();
  if (language === 'en-gb' || language === 'en-ie') return 'uk';
  if (language === 'pt-br') return 'pt-br';
  if (language.startsWith('sv') || language.startsWith('fi')) return 'se';

  const primary = language.split('-')[0];
  return (pasteLayoutIds as readonly string[]).includes(primary) ? primary : 'us';
}

function storedPasteLayout() {
  const stored = storage.getPasteLayout();
  return stored && (pasteLayoutIds as readonly string[]).includes(stored)
    ? stored
    : guessPasteLayout();
}

export const defaultPasteDelay = 30;

const pasteLayoutValueAtom = atom(storedPasteLayout());
const pasteDelayValueAtom = atom(storage.getPasteDelay() ?? defaultPasteDelay);

// The layout active on the host, remembered in this browser.
export const pasteLayoutAtom = atom(
  (get) => get(pasteLayoutValueAtom),
  (_get, set, layout: string) => {
    storage.setPasteLayout(layout);
    set(pasteLayoutValueAtom, layout);
  }
);

// The pause after each key press, in milliseconds, remembered in this browser.
export const pasteDelayAtom = atom(
  (get) => get(pasteDelayValueAtom),
  (_get, set, delay: number) => {
    storage.setPasteDelay(delay);
    set(pasteDelayValueAtom, delay);
  }
);

export type PasteDialog = {
  open: boolean;
  // notice explains why the dialog opened on its own, such as a clipboard the
  // browser would not read.
  notice: string;
  // check is the server's report on a text it refused.
  check: PasteCheck | null;
};

export const pasteDialogAtom = atom<PasteDialog>({ open: false, notice: '', check: null });

// The text in the dialog. It outlives the dialog, so closing it by mistake
// does not lose a long text.
export const pasteTextAtom = atom('');

// The paste typing on the host, or the last one. Null hides the progress.
export const pasteStatusAtom = atom<PasteStatus | null>(null);
