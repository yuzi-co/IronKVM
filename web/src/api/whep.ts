import { http } from '@/lib/http.ts';

// The WHEP endpoint's switch and state (ironkvm-dist#145). Path is where a
// player POSTs its offer.
export type WhepSettings = {
  enabled: boolean;
  path: string;
  sessions: number;
  maxSessions: number;
};

export function getWhepSettings() {
  return http.get('/api/stream/whep');
}

export function setWhepSettings(enabled: boolean) {
  return http.post('/api/stream/whep', { enabled });
}
