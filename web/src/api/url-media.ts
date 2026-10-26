import { http } from '@/lib/http.ts';

// Mounting checks the server first, which the board gives up to 30 s.
const MOUNT = { timeout: 45 * 1000 };

// Whether this board can serve an image from a URL, and what it serves.
export function getUrlMediaStatus() {
  return http.get('/api/storage/url');
}

// Serve the image at url in a drive: the CD, or the disk read-only.
export function mountUrl(drive: 'cdrom' | 'disk', url: string) {
  return http.post('/api/storage/url/mount', { drive, url }, MOUNT);
}

export function ejectUrl(drive: 'cdrom' | 'disk') {
  return http.post('/api/storage/url/eject', { drive });
}
