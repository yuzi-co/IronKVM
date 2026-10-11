import { atom } from 'jotai';

import type { SequenceStatus } from '@/api/sequence.ts';
import type { MacroStep } from '@/lib/macro.ts';

// The auto-tap or macro the server runs now, or ran last. Null until the
// first read. Whoever starts a run sets it from the reply, and the sequence
// banner polls it while it says running.
export const sequenceStatusAtom = atom<SequenceStatus | null>(null);

// The macro recording in progress: how many steps it holds, and whether it
// has reached the limit. Null when nothing is recorded.
export const macroRecordingAtom = atom<{ count: number; full: boolean } | null>(null);

// The macro the editor shows. Null when the editor is closed. A draft without
// an id is saved as a new macro.
export type MacroDraft = { id?: string; name: string; steps: MacroStep[] };
export const macroDraftAtom = atom<MacroDraft | null>(null);

// Bumped when the saved macros change, so an open list reads them again.
export const macrosVersionAtom = atom(0);
