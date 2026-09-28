import { http } from '@/lib/http.ts';

// Download image
export function downloadImage(file?: string, sha256sum?: string) {
  const data = {
    file: file ?? '',
    sha256sum: sha256sum ?? ''
  };
  return http.post('/api/download/image', data);
}

export function cancelDownloadImage() {
  return http.post('/api/download/image/cancel');
}

export function statusImage() {
  return http.get('/api/download/image/status');
}

export function imageEnabled() {
  return http.get('/api/download/image/enabled');
}

// Download the netboot.xyz boot menu ISO. The server holds its URL and checksum.
export function downloadBootMenu() {
  return http.post('/api/download/image/netboot');
}

// Upload an ISO from this computer. It goes through the shared client so the
// base URL and an expired session are handled like any other request, and it
// has no time limit, since a large image over a slow link takes a while.
export function uploadImageFile(
  file: File,
  sha256sum: string,
  onProgress: (percent: number) => void
) {
  const formData = new FormData();
  formData.append('file', file);

  return http.post('/api/download/file', formData, {
    headers: { 'X-SHA256-Sum': sha256sum },
    timeout: 0,
    onUploadProgress: (event) => {
      if (event.total) onProgress(Math.round((event.loaded / event.total) * 100));
    }
  });
}
