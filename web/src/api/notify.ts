import { http } from '@/lib/http.ts';

// The events the owner can switch on and off, by the name the server uses.
export const NOTIFY_EVENTS = [
  'signal.lost',
  'signal.restored',
  'watchdog.action',
  'login.failures',
  'atx.press',
  'failsafe.entered',
  'hostlog.match'
] as const;

export type NotifyEventType = (typeof NOTIFY_EVENTS)[number];

export type NotifyChannel = 'webhook' | 'ntfy' | 'telegram';

export type NotifyRule = { enabled: boolean; cooldownSeconds: number };

// A secret is never sent to the page: it only learns whether one is set.
export type NotifySettings = {
  enabled: boolean;
  snapshot: boolean;
  events: Record<NotifyEventType, NotifyRule>;
  loginThreshold: number;
  loginWindowSeconds: number;
  webhook: { enabled: boolean; urlSet: boolean; urlHint?: string; includeSnapshot: boolean };
  ntfy: { enabled: boolean; server: string; topic: string; tokenSet: boolean };
  telegram: { enabled: boolean; botTokenSet: boolean; chatId: string };
};

// What a save sends. A secret left undefined is kept, an empty string clears it.
export type NotifySave = {
  enabled: boolean;
  snapshot: boolean;
  events: Record<string, NotifyRule>;
  loginThreshold: number;
  loginWindowSeconds: number;
  webhook: { enabled: boolean; url?: string; includeSnapshot: boolean };
  ntfy: { enabled: boolean; server: string; topic: string; token?: string };
  telegram: { enabled: boolean; botToken?: string; chatId: string };
};

export type NotifyChannelResult = { time: string; ok: boolean; error?: string };

export type NotifyEntry = {
  time: string;
  type: string;
  title: string;
  suppressed?: boolean;
  channels?: Record<string, string>;
};

export type NotifyState = {
  channels: Partial<Record<NotifyChannel, NotifyChannelResult>>;
  recent: NotifyEntry[];
};

export function getNotifySettings() {
  return http.get('/api/notify/settings');
}

export function setNotifySettings(settings: NotifySave) {
  return http.post('/api/notify/settings', settings);
}

export function getNotifyState() {
  return http.get('/api/notify/state');
}

// channel '' tests every channel that is set up.
export function sendNotifyTest(channel: NotifyChannel | '') {
  return http.post('/api/notify/test', { channel });
}
