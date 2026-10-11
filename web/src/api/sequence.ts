import i18n from '@/i18n/index.ts';
import { http } from '@/lib/http.ts';
import type { AtxButton, Macro, MacroStep } from '@/lib/macro.ts';
import { isViewOnly, VIEW_ONLY_CODE } from '@/lib/view-only.ts';

// The key auto-tap and the macros. The server plays both as timed input
// sequences, one at a time, and keeps playing when this tab closes.

export type SequenceReason =
  | ''
  | 'finished'
  | 'stopped'
  | 'timeout'
  | 'input'
  | 'video'
  | 'shutdown'
  | 'busy'
  | 'hid_error'
  | 'atx_error';

export type SequenceStatus = {
  running: boolean;
  kind: '' | 'autotap' | 'macro';
  name: string;
  step: number;
  total: number;
  taps: number;
  elapsedMs: number;
  timeoutMs: number;
  reason: SequenceReason;
};

export type AutoTapRequest = {
  label: string;
  modifiers: number;
  keys: number[];
  intervalMs: number;
  timeoutS: number;
  stopOnInput: boolean;
  stopOnVideo: boolean;
  atx: '' | AtxButton;
};

function viewOnlyRefusal() {
  return Promise.resolve({
    code: VIEW_ONLY_CODE,
    msg: i18n.t('screen.viewOnlyBlocked'),
    data: null
  });
}

// start pressing a key over and over on the host
export function startAutoTap(req: AutoTapRequest) {
  if (isViewOnly()) return viewOnlyRefusal();
  return http.post('/api/hid/autotap', req);
}

// play a saved macro by id, or steps that are not saved
export function playMacro(macro: { id: string } | { name: string; steps: MacroStep[] }) {
  if (isViewOnly()) return viewOnlyRefusal();
  return http.post('/api/hid/macro/play', macro);
}

// the sequence running now, or the last one
export function getSequenceStatus() {
  return http.get('/api/hid/sequence');
}

// stop the running sequence
export function stopSequence() {
  return http.post('/api/hid/sequence/stop');
}

// the macros saved on the device
export function getMacros() {
  return http.get('/api/hid/macros');
}

// save a macro: a new one without an id, or the one the id names
export function saveMacro(macro: Partial<Macro> & { name: string; steps: MacroStep[] }) {
  return http.post('/api/hid/macro', macro);
}

// delete a saved macro
export function deleteMacro(id: string) {
  return http.delete('/api/hid/macro', { id });
}
