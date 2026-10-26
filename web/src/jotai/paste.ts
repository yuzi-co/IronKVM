import { atom } from 'jotai';

import type { PasteCheck, PasteStatus } from '@/api/hid.ts';
import * as storage from '@/lib/localstorage.ts';
import { guessPasteLayout, initialPasteLayout, type HostLayout } from '@/lib/paste-layout.ts';

export { pasteLayoutIds } from '@/lib/paste-layout.ts';

function storedPasteLayout() {
  return initialPasteLayout(storage.getPasteLayout(), null, guessPasteLayout(navigator.language));
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

// The layout the owner set for the host on the device. The dialog starts on
// it while this browser has not picked a layout of its own; it is not stored.
export const hostPasteLayoutAtom = atom(null, (_get, set, host: HostLayout) => {
  set(
    pasteLayoutValueAtom,
    initialPasteLayout(storage.getPasteLayout(), host, guessPasteLayout(navigator.language))
  );
});

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
