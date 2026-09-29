import { http } from '@/lib/http.ts';

// The opt-in updates of a downloaded third-party component (netboot.xyz, Ventoy).
// Each component answers under its own path: GET for the status, POST /check to
// ask GitHub, POST to start the update, which runs in the background.

export type UpdateJob = {
  state: 'idle' | 'running' | 'done' | 'failed';
  version: string;
  progress: number;
  error: string;
};

export type UpdateStatus = {
  installed: boolean;
  version: string;
  pinned: string;
  latest: string;
  checkedAt: string | null;
  checkError: string;
  updateAvailable: boolean;
  unverifiable: string;
  inUse: string;
  job: UpdateJob;
};

export const UPDATE_PATHS = {
  netbootxyz: '/api/netboot/update',
  ventoy: '/api/ventoy/update',
  bootMenu: '/api/download/image/netboot/update'
} as const;

export function getUpdate(path: string) {
  return http.get(path);
}

// force asks GitHub even when the last answer is fresh.
export function checkUpdate(path: string, force = true) {
  return http.post(`${path}/check`, { force });
}

export function startUpdate(path: string) {
  return http.post(path);
}
