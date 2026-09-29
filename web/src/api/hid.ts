import i18n from '@/i18n/index.ts';
import { http } from '@/lib/http.ts';
import { isViewOnly, VIEW_ONLY_CODE } from '@/lib/view-only.ts';

// viewOnlyRefusal answers a request that would type or press on the host
// while this tab is in view only, without sending it.
function viewOnlyRefusal() {
  return Promise.resolve({
    code: VIEW_ONLY_CODE,
    msg: i18n.t('screen.viewOnlyBlocked'),
    data: null
  });
}

export type PasteUntypeable = {
  index: number; // offset in the text, in code points
  line: number;
  column: number;
  char: string;
};

export type PasteCheck = {
  characters: number;
  keystrokes: number;
  durationMs: number;
  untypeable: PasteUntypeable[]; // the first hundred
  untypeableCount: number;
};

export type PasteStatus = {
  status: 'idle' | 'typing' | 'done' | 'canceled' | 'failed';
  typed: number;
  total: number;
  error?: 'control_busy' | 'hid_error';
};

// start typing a text on the host in the background. The server refuses a
// text with characters the layout cannot type, code -4 with a PasteCheck,
// unless skipUntypeable says to type the rest.
export function paste(content: string, layout: string, delay: number, skipUntypeable = false) {
  if (isViewOnly()) return viewOnlyRefusal();
  return http.post('/api/hid/paste', { content, layout, delay, skipUntypeable });
}

// what typing a text would take, without typing it
export function checkPaste(content: string, layout: string, delay: number) {
  return http.post('/api/hid/paste/check', { content, layout, delay });
}

// progress of the paste typing now, or of the last one
export function getPasteStatus() {
  return http.get('/api/hid/paste/status');
}

// stop the paste typing now
export function cancelPaste() {
  return http.post('/api/hid/paste/cancel');
}

// reset hid
export function reset() {
  return http.post('/api/hid/reset');
}

// get hid mode
export function getHidMode() {
  return http.get('/api/hid/mode');
}

// press and release one Consumer Control (media) or System Control (power) key
export function sendKey(page: 'consumer' | 'system', usage: number) {
  if (isViewOnly()) return viewOnlyRefusal();
  return http.post('/api/hid/key', { page, usage });
}

// get remote keyboard lock LED status
export function getKeyboardLedStatus() {
  return http.get('/api/hid/leds');
}

// set hid mode
export function setHidMode(mode: string) {
  const data = {
    mode
  };
  return http.post('/api/hid/mode', data);
}

// get shortcuts
export function getShortcuts() {
  return http.get('/api/hid/shortcuts');
}

// add shortcut
export function addShortcut(keys: any[]) {
  const data = {
    keys
  };
  return http.post('/api/hid/shortcut', data);
}

// delete shortcut
export function deleteShortcut(id: string) {
  const data = {
    id
  };
  return http.delete('/api/hid/shortcut', data);
}

// get shortcut leader key
export function getLeaderKey() {
  return http.get('/api/hid/shortcut/leader-key');
}

// set shortcut leader key
export function setLeaderKey(key: string) {
  return http.post('/api/hid/shortcut/leader-key', { key });
}

// get per-endpoint HID delivery state
export function getHidStatus() {
  return http.get('/api/hid/status');
}
