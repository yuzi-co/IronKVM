import { http } from '@/lib/http.ts';

// The RTSP server (ironkvm-dist#142) and the SRT listener (ironkvm-dist#143):
// their settings, and the players they serve.

export type RtspClient = {
  address: string;
  user?: string;
  transport?: 'tcp' | 'udp';
  since: string;
  playing: boolean;
  // A player on the RTSPS listener.
  secure: boolean;
};

export type RtspSettings = {
  enabled: boolean;
  port: number;
  udp: boolean;
  udpPort: number;
  // RTSPS, RTSP over TLS, on a port of its own. `enabled` is plain RTSP.
  tls: boolean;
  tlsPort: number;
  path: string;
  maxClients: number;
  // Whether the board has a certificate RTSPS can serve.
  certificate: boolean;
  listening: boolean;
  error?: string;
  clients: RtspClient[];
};

export type RtspSettingsRequest = {
  enabled: boolean;
  port: number;
  udp: boolean;
  udpPort: number;
  tls: boolean;
  tlsPort: number;
};

export function getRtspSettings() {
  return http.get('/api/stream/rtsp');
}

export function setRtspSettings(settings: RtspSettingsRequest) {
  return http.post('/api/stream/rtsp', settings);
}

export type SrtClient = {
  address: string;
  user?: string;
  since: string;
  rttMs: number;
  sendRateMbps: number;
  sent: number;
  lost: number;
  retransmitted: number;
  dropped: number;
};

export type SrtSettings = {
  enabled: boolean;
  port: number;
  latencyMs: number;
  // The server never sends the passphrase back, only whether one is set.
  passphraseSet: boolean;
  maxClients: number;
  listening: boolean;
  error?: string;
  clients: SrtClient[];
};

export type SrtSettingsRequest = {
  enabled: boolean;
  port: number;
  latencyMs: number;
  // Empty keeps the passphrase that is set.
  passphrase?: string;
};

export function getSrtSettings() {
  return http.get('/api/stream/srt');
}

export function setSrtSettings(settings: SrtSettingsRequest) {
  return http.post('/api/stream/srt', settings);
}
