import { getDefaultStore } from 'jotai';

import { onInputTap } from '@/lib/input-tap.ts';
import { MacroRecorder, type MacroStep } from '@/lib/macro.ts';
import { macroRecordingAtom } from '@/jotai/sequence.ts';

// One recording at a time, kept outside React: the macros menu starts it and
// the sequence banner finishes it, and the menu may be closed in between.
let recorder: MacroRecorder | null = null;
let stopListening: (() => void) | null = null;

export function startRecording() {
  discardRecording();

  const store = getDefaultStore();
  const current = new MacroRecorder();
  recorder = current;
  stopListening = onInputTap((event) => {
    const now = performance.now();
    const added =
      event.kind === 'message'
        ? current.message(event.data, now)
        : current.atx(event.button, event.ms, now);
    if (added || current.full) {
      store.set(macroRecordingAtom, { count: current.count, full: current.full });
    }
  });
  store.set(macroRecordingAtom, { count: 0, full: false });
}

// finishRecording ends the recording and returns its steps, with a release
// for anything still held.
export function finishRecording(): MacroStep[] {
  const steps = recorder?.finish() ?? [];
  discardRecording();
  return steps;
}

export function discardRecording() {
  stopListening?.();
  stopListening = null;
  recorder = null;
  getDefaultStore().set(macroRecordingAtom, null);
}
