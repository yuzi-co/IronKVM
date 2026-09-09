import { atom } from 'jotai';

import { ControlRegionMode, InputRegion, Resolution, ScreenSettings } from '@/types';

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
  codec: 1
};

// currently effective absolute mouse input region
export const inputRegionAtom = atom<InputRegion | null>(null);
export const manualInputRegionAtom = atom<InputRegion | null>(null);
export const manualRegionsAtom = atom<InputRegion[]>([]);
export const selectedManualRegionAtom = atom<string>('');
export const selectedOriginalResolutionAtom = atom<string>('');

// device-level control region mode; disabled by default
export const controlRegionModeAtom = atom<ControlRegionMode>('off');

// show the live input-region selection overlay
export const inputRegionSelectingAtom = atom(false);
