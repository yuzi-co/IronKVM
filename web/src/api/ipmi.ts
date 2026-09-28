import { http } from '@/lib/http.ts';

export type IpmiUser = {
  username: string;
  role: 'admin' | 'user';
  enabled: boolean;
  hasPassword: boolean;
  nameFits: boolean;
};

export type IpmiSettings = {
  enabled: boolean;
  port: number;
  powerLed: boolean;
  users: IpmiUser[];
};

export function getIpmiSettings() {
  return http.get('/api/ipmi/settings');
}

export function setIpmiEnabled(enabled: boolean) {
  return http.post('/api/ipmi/settings', { enabled });
}

export function setIpmiPassword(username: string, password: string) {
  return http.post(`/api/ipmi/users/${encodeURIComponent(username)}/password`, { password });
}

export function clearIpmiPassword(username: string) {
  return http.delete(`/api/ipmi/users/${encodeURIComponent(username)}/password`);
}
