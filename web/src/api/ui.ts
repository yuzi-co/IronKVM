import { http } from '@/lib/http.ts';

// Separately installed web UI packages (server/service/webui). Admin only.

export type UiRelease = {
  id: string;
  version: string;
  api: { min: number; max: number };
  bundled: boolean;
};

export type UiStatus = {
  serverApi: number;
  active: UiRelease;
  previous?: UiRelease;
  canRollback: boolean;
  // 'on', 'off' (the owner's switch file) or 'no-keys' (an empty keyring).
  signatureCheck: 'on' | 'off' | 'no-keys';
};

export type UiLatest = {
  version: string;
  api: { min: number; max: number };
  compatible: boolean;
  active: boolean;
};

export function getUiStatus() {
  return http.get('/api/ui');
}

export function getUiLatest() {
  return http.get('/api/ui/latest');
}

// Downloading and unpacking a few megabytes on the board takes a while.
export function updateUi() {
  return http.request({ method: 'post', url: '/api/ui/update', timeout: 5 * 60 * 1000 });
}

export function uploadUi(pkg: File, signature: File | null) {
  const data = new FormData();
  if (signature) data.append('signature', signature);
  data.append('package', pkg);
  return http.request({ method: 'post', url: '/api/ui/upload', data, timeout: 5 * 60 * 1000 });
}

export function rollbackUi() {
  return http.post('/api/ui/rollback');
}
