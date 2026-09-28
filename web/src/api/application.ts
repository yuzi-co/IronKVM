import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

export type UpdateServerConfig = {
  enabled: boolean;
  url: string;
};

// get application version
export function getVersion() {
  return http.get('/api/application/version');
}

// update application to latest version
export function update() {
  return http.request({
    method: 'post',
    url: '/api/application/update',
    timeout: 15 * 60 * 1000
  });
}

export type OfflineUpdateResult = {
  status: number;
  ok: boolean;
  // The parsed JSON reply, or null when the body was not JSON.
  body: { code: number; msg?: string } | null;
};

// offlineUpdate uploads an update package. It uses XMLHttpRequest rather than
// fetch because fetch reports no upload progress, and a package is large
// enough on a slow link that a bar is the only sign the upload is moving.
export function offlineUpdate(
  data: FormData,
  sha256Checksum = '',
  onProgress?: (percent: number) => void
): Promise<OfflineUpdateResult> {
  const baseUrl = getBaseUrl('http');
  const url = `${baseUrl}/api/application/update/offline`;

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', url);
    if (sha256Checksum) xhr.setRequestHeader('X-SHA256-Checksum', sha256Checksum);
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && onProgress) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    };
    xhr.onload = () => {
      let body: OfflineUpdateResult['body'];
      try {
        body = JSON.parse(xhr.responseText);
      } catch {
        body = null;
      }
      resolve({ status: xhr.status, ok: xhr.status >= 200 && xhr.status < 300, body });
    };
    // Shaped like an axios network error, so the page words it the same way.
    xhr.onerror = () => reject(Object.assign(new Error('Network Error'), { code: 'ERR_NETWORK' }));
    xhr.send(data);
  });
}

// enable/disable preview updates
export function setPreviewUpdates(enable: boolean) {
  const data = {
    enable
  };
  return http.post('/api/application/preview', data);
}

// get preview updates state
export function getPreviewUpdates() {
  return http.get('/api/application/preview');
}

// get custom update server configuration
export function getUpdateServer() {
  return http.get('/api/application/update-server');
}

// enable/disable custom update server
export function setUpdateServer(config: UpdateServerConfig) {
  return http.post('/api/application/update-server', config);
}
