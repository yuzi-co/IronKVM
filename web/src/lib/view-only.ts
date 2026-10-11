// View only: this browser watches the host and sends it nothing. It is a
// per-tab switch, not a server setting, so it lives in memory and ends with
// the tab. Scripts, the mouse jiggler and other viewers are not affected.
//
// The flag sits in a module of its own so the websocket client and the HID
// API can read it without importing React state.

let viewOnly = false;
// A view-only account (or a share link) cannot turn view only off: the server
// refuses it every input route anyway, and a switch that does nothing would
// only invite pressing it.
let locked = false;
const listeners = new Set<(value: boolean) => void>();

export function isViewOnly() {
  return viewOnly;
}

export function isViewOnlyLocked() {
  return locked;
}

export function lockViewOnly() {
  locked = true;
  setViewOnly(true);
}

export function setViewOnly(value: boolean) {
  if (locked && !value) return;
  if (value === viewOnly) return;
  viewOnly = value;
  listeners.forEach((listener) => listener(value));
}

export function onViewOnlyChange(listener: (value: boolean) => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

// The answer the HID API gives in place of a request view only held back, in
// the server's envelope so callers handle it as any refusal.
export const VIEW_ONLY_CODE = -100;
