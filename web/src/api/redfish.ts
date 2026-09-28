import { http } from '@/lib/http.ts';

export type RedfishSettings = {
  enabled: boolean;
  https: boolean;
  serviceRoot: string;
  resetTypes: string[];
};

export type RedfishSession = {
  id: string;
  user: string;
  createdAt: string;
  lastUsed: string;
};

export function getRedfishSettings() {
  return http.get('/api/redfish/settings');
}

export function setRedfishEnabled(enabled: boolean) {
  return http.post('/api/redfish/settings', { enabled });
}

export function getRedfishSessions() {
  return http.get('/api/redfish/sessions');
}

export function endRedfishSession(id: string) {
  return http.delete(`/api/redfish/sessions/${encodeURIComponent(id)}`);
}
