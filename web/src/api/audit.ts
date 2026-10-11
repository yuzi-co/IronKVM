import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export const AUDIT_CATEGORIES = [
  'auth',
  'control',
  'power',
  'media',
  'settings',
  'macro',
  'system',
  'agent'
] as const;
export const AUDIT_OUTCOMES = ['ok', 'fail', 'denied'] as const;

export type AuditCategory = (typeof AUDIT_CATEGORIES)[number];
export type AuditOutcome = (typeof AUDIT_OUTCOMES)[number];

export type AuditEvent = {
  time: string;
  actor?: string;
  via?: string;
  key?: string;
  ip?: string;
  category: string;
  action: string;
  target?: string;
  outcome: string;
  tier?: string;
  detail?: string;
  repeat?: number;
};

export type AuditFilter = {
  category?: string;
  outcome?: string;
  actor?: string;
  ip?: string;
  q?: string;
};

export type AuditPage = { events: AuditEvent[]; total: number };

export type AuditSettings = { syslog: boolean; directory: string; maxBytes: number };

// Empty filter fields are left out of the request.
function params(filter: AuditFilter) {
  return Object.fromEntries(Object.entries(filter).filter(([, value]) => value));
}

export function getAudit(filter: AuditFilter, offset: number, limit: number) {
  return http.get('/api/audit', { ...params(filter), offset, limit });
}

// The export is a file, so it is fetched as one rather than through the JSON
// client.
export async function exportAudit(filter: AuditFilter): Promise<Blob> {
  const query = new URLSearchParams(params(filter) as Record<string, string>).toString();
  const rsp = await fetch(`${getBaseUrl('http')}/api/audit/export${query ? `?${query}` : ''}`, {
    credentials: 'include'
  });
  if (!rsp.ok) throw new Error(`export failed: ${rsp.status}`);
  return rsp.blob();
}

export function getAuditSettings() {
  return http.get('/api/audit/settings');
}

export function setAuditSyslog(syslog: boolean) {
  return http.post('/api/audit/settings', { syslog });
}

