import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export type WatchdogAction = 'reset' | 'power';

export type WatchdogSettings = {
  enabled: boolean;
  timeoutMinutes: number;
  action: WatchdogAction;
  cooldownMinutes: number;
  maxPerHour: number;
  pingHost: string;
};

export type WatchdogStatus =
  'off' | 'watching' | 'hostOff' | 'captureOff' | 'cooldown' | 'capped' | 'acting';

export type WatchdogState = {
  status: WatchdogStatus;
  signal: boolean;
  ledConnected: boolean;
  ledOn: boolean;
  pingHost: string;
  pingOK: boolean;
  lastSample?: string;
  lastChange?: string;
  lastPing?: string;
  lastAction?: string;
  actsInSeconds?: number;
  actionsLastHour: number;
};

export type WatchdogEntry = {
  id: string;
  time: string;
  action: WatchdogAction;
  reason: 'frozen' | 'noSignal';
  stuckSeconds: number;
  screenshot?: string;
  error?: string;
};

export function getWatchdogSettings() {
  return http.get('/api/watchdog/settings');
}

export function setWatchdogSettings(settings: WatchdogSettings) {
  return http.post('/api/watchdog/settings', settings);
}

export function getWatchdogState() {
  return http.get('/api/watchdog/state');
}

export function getWatchdogLog() {
  return http.get('/api/watchdog/log');
}

// The screenshot is an image, so the page loads it with an img element, which
// sends the session cookie as the API calls do.
export function watchdogScreenshotUrl(id: string) {
  return `${getBaseUrl('http')}/api/watchdog/log/${encodeURIComponent(id)}/screenshot`;
}
