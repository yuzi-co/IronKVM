import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export type HostLogEntry = {
  seq: number;
  time: string;
  source: string;
  proto: 'syslog' | 'netconsole';
  // -1 when the sender gave none.
  facility: number;
  severity: number;
  host?: string;
  tag?: string;
  msg: string;
};

export type HostLogListener = { enabled: boolean; port: number };

export type HostLogSettings = {
  syslog: HostLogListener;
  netconsole: HostLogListener;
  bind: '' | 'usb';
  sources: string[];
  memoryKb: number;
  disk: boolean;
  diskKb: number;
  patterns: string[];
};

export type HostLogListenerState = {
  state: 'off' | 'listening' | 'waiting' | 'error';
  addr?: string;
  error?: string;
};

export type HostLogState = {
  syslog: HostLogListenerState;
  netconsole: HostLogListenerState;
  sources: string[];
  received: number;
  filtered: number;
  rateLimited: number;
  buffered: number;
  bufferBytes: number;
  diskBytes: number;
};

// The live tail sends the newest lines once, then batches of new ones.
export type HostLogFrame = { type: 'backlog' | 'entries'; entries: HostLogEntry[] };

export type HostLogFilter = { q: string; severity: number | null };

export function getHostLogSettings() {
  return http.get('/api/hostlog/settings');
}

export function setHostLogSettings(settings: HostLogSettings) {
  return http.post('/api/hostlog/settings', settings);
}

export function getHostLogState() {
  return http.get('/api/hostlog/state');
}

export function clearHostLog() {
  return http.delete('/api/hostlog/lines');
}

function filterQuery(filter: HostLogFilter) {
  const params = new URLSearchParams();
  if (filter.q.trim()) params.set('q', filter.q.trim());
  if (filter.severity !== null) params.set('severity', String(filter.severity));
  return params;
}

// The download is a plain link, which sends the session cookie.
export function hostLogDownloadUrl(filter: HostLogFilter) {
  const query = filterQuery(filter).toString();
  return `${getBaseUrl('http')}/api/hostlog/download${query ? `?${query}` : ''}`;
}

export function hostLogStreamUrl(filter: HostLogFilter) {
  const query = filterQuery(filter).toString();
  return `${getBaseUrl('ws')}/api/hostlog/stream${query ? `?${query}` : ''}`;
}
