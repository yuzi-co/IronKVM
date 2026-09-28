import { http } from '@/lib/http.ts';

export type VncSettings = {
  enabled: boolean;
  port: number;
  maxFps: number;
  vncAuth: boolean;
  // The server never sends the password back, only whether one is set.
  passwordSet: boolean;
};

export type VncSettingsRequest = {
  enabled: boolean;
  port: number;
  maxFps: number;
  vncAuth: boolean;
  // Empty keeps the password that is set.
  password?: string;
};

export type VncSession = {
  client: string;
  user?: string;
  method: 'vencrypt' | 'vnc';
  since: string;
  width: number;
  height: number;
  framesSent: number;
  bytesSent: number;
};

export type VncState = {
  listening: boolean;
  port?: number;
  error?: string;
  lastError?: string;
  session?: VncSession;
};

export function getVncSettings() {
  return http.get('/api/vnc/settings');
}

export function setVncSettings(settings: VncSettingsRequest) {
  return http.post('/api/vnc/settings', settings);
}

export function getVncState() {
  return http.get('/api/vnc/state');
}

export function disconnectVnc() {
  return http.post('/api/vnc/disconnect');
}
