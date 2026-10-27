import { http } from '@/lib/http.ts';

// A press of the host's power or reset button that the on-board agent asked
// for. The server does not press until a signed-in user confirms it here, and
// drops it unconfirmed after a minute (server/service/picoclaw/power_confirm.go).
export type PowerConfirmation = {
  id: string;
  agent?: string;
  sessionId: string;
  button: string;
  press: string;
  state: string;
  message?: string;
  createdAt: string;
  expiresAt: string;
  decidedBy?: string;
};

const base = '/api/agent/power/confirmations';

// listPowerConfirmations returns the presses that wait for a person.
export function listPowerConfirmations() {
  return http.get(base);
}

export function confirmPowerConfirmation(id: string) {
  return http.post(`${base}/${encodeURIComponent(id)}/confirm`);
}

export function denyPowerConfirmation(id: string) {
  return http.post(`${base}/${encodeURIComponent(id)}/deny`);
}
