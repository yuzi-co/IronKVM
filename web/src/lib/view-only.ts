// View only: this browser watches the host and sends it nothing. It is a
// per-tab switch, not a server setting, so it lives in memory and ends with
// the tab. Scripts, the mouse jiggler and other viewers are not affected.
//
// The flag sits in a module of its own so the websocket client and the HID
// API can read it without importing React state.

let viewOnly = false;
const listeners = new Set<(value: boolean) => void>();

export function isViewOnly() {
  return viewOnly;
}

export function setViewOnly(value: boolean) {
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
