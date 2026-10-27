import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

// The dashcam (ironkvm-dist#119): the last minute of video, saved as an MP4
// file on /data when the host is reset, the signal goes, or a person asks.

export type DashcamTriggers = {
  // A press of the power or reset button, and one by the host watchdog.
  atx: boolean;
  // The HDMI signal going and coming back.
  signal: boolean;
  // The server starting in failsafe mode.
  failsafe: boolean;
};

export type DashcamSettings = {
  enabled: boolean;
  seconds: number;
  postSeconds: number;
  // 0 holds the stream as it is, up to the ring's fixed bound.
  maxKbps: number;
  storageMB: number;
  triggers: DashcamTriggers;
};

export type DashcamPending = {
  started: string;
  deadline: string;
  triggers: string[];
  joined: boolean;
};

export type DashcamState = {
  enabled: boolean;
  // False while the dashcam is on but the board captures no video.
  running: boolean;
  seconds: number;
  buffered: number;
  bufferedBytes: number;
  budgetBytes: number;
  pending?: DashcamPending;
  lastClip?: string[];
  lastError?: string;
};

export type DashcamClip = {
  name: string;
  size: number;
  time: string;
  triggers: string[];
};

// The save route's codes for a refusal.
export const DASHCAM_OFF = -2;
export const DASHCAM_EMPTY = -3;
export const DASHCAM_TOO_SOON = -4;

export function getDashcamState() {
  return http.get('/api/dashcam');
}

export function saveDashcamClip() {
  return http.post('/api/dashcam/save');
}

export function getDashcamClips() {
  return http.get('/api/dashcam/clips');
}

export function deleteDashcamClip(name: string) {
  return http.delete(`/api/dashcam/clips/${encodeURIComponent(name)}`);
}

// The download is a plain link, which sends the session cookie.
export function dashcamClipUrl(name: string) {
  return `${getBaseUrl('http')}/api/dashcam/clips/${encodeURIComponent(name)}`;
}

export function getDashcamSettings() {
  return http.get('/api/dashcam/settings');
}

export function setDashcamSettings(settings: DashcamSettings) {
  return http.post('/api/dashcam/settings', settings);
}
