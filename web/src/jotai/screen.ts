import { atom } from 'jotai';

import { ControlRegionMode, InputRegion, Resolution, ScreenSettings } from '@/types';
import type { CaptureReport } from '@/lib/health.ts';
import { sameInputRegion } from '@/lib/input-region.ts';
import { KeyboardReport } from '@/lib/keyboard.ts';
import { getPauseWhenHidden, setPauseWhenHidden } from '@/lib/localstorage.ts';
import { setViewOnly } from '@/lib/view-only.ts';
import { client, MessageEvent } from '@/lib/websocket.ts';

export const isHdmiEnabledAtom = atom(true);

// video mode
// direct: stream H.264 over HTTP
// h264: stream H.264 over WebRTC
// mjpeg: stream JPEG over HTTP
export const videoModeAtom = atom('');

export const videoScaleAtom = atom<number>(1.0);

// capture resolution, as the board reports it
export const resolutionAtom = atom<Resolution | null>(null);

// What the server holds for the capture pipeline, read once when the desktop
// loads. There is one encoder, so these are the board's settings and not this
// browser's: the menu draws itself from this rather than from localStorage.
// Null until the read answers, or for good if it fails.
export const screenSettingsAtom = atom<ScreenSettings | null>(null);

// What the menu shows while the read has not answered. These match the
// server's own defaults, so an unconfigured board and an unreachable one look
// the same, which is the truth: neither has told us anything.
export const defaultScreenSettings: ScreenSettings = {
  width: 0,
  height: 0,
  quality: 80,
  bitRate: 3000,
  fps: 30,
  gop: 30,
  codec: 1,
  aspect: 0,
  aspectSupported: false
};

// currently effective absolute mouse input region. Timers recompute it, so a
// write that brings the same numbers keeps the current object: readers do not
// re-render and the mouse handlers are left alone.
const baseInputRegionAtom = atom<InputRegion | null>(null);
export const inputRegionAtom = atom(
  (get) => get(baseInputRegionAtom),
  (get, set, next: InputRegion | null) => {
    if (!sameInputRegion(get(baseInputRegionAtom), next)) set(baseInputRegionAtom, next);
  }
);
export const manualInputRegionAtom = atom<InputRegion | null>(null);
export const manualRegionsAtom = atom<InputRegion[]>([]);
export const selectedManualRegionAtom = atom<string>('');
export const selectedOriginalResolutionAtom = atom<string>('');

// device-level control region mode; disabled by default
export const controlRegionModeAtom = atom<ControlRegionMode>('off');

// show the live input-region selection overlay
export const inputRegionSelectingAtom = atom(false);

// The capture status the board last reported for this viewer's video mode,
// or null while the stream is fine. The desktop keeps it; the toolbar's
// Screen dot and alert icon read it.
export const captureStatusAtom = atom<CaptureReport>(null);

// True while the stream is stopped because the tab is hidden.
export const streamPausedAtom = atom(false);

const pauseWhenHiddenBaseAtom = atom(getPauseWhenHidden());

// Whether to stop the stream while the tab is hidden, remembered per browser.
export const pauseWhenHiddenAtom = atom(
  (get) => get(pauseWhenHiddenBaseAtom),
  (_get, set, enabled: boolean) => {
    set(pauseWhenHiddenBaseAtom, enabled);
    setPauseWhenHidden(enabled);
  }
);

const viewOnlyBaseAtom = atom(false);

// View only: this tab sends the host no keyboard or mouse input. Turning it
// on first releases every key, so a key held at that moment is not left down
// on the host with nothing to let it go.
export const viewOnlyAtom = atom(
  (get) => get(viewOnlyBaseAtom),
  (_get, set, enabled: boolean) => {
    if (enabled) {
      client.send(new Uint8Array([MessageEvent.Keyboard, ...new KeyboardReport().reset()]));
    }
    setViewOnly(enabled);
    set(viewOnlyBaseAtom, enabled);
  }
);
