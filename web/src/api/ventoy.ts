import { http } from '@/lib/http.ts';

// Install downloads the Ventoy release, about 20 MB, and the request waits for it.
// Insert builds the disk head and the device table first.
const LONG = { timeout: 10 * 60 * 1000 };

// The device the Ventoy disk is served from. The drive list shows it by this path.
export const VENTOY_DEVICE = '/dev/mapper/ventoy';

export type VentoyStatus = {
  kernel: boolean;
  onData: boolean;
  installed: boolean;
  version: string;
  images: string[] | null;
  missing: string[] | null;
  inDrive: boolean;
  device: string;
  size: number;
};

export function getVentoyStatus() {
  return http.get('/api/ventoy/status');
}

export function installVentoy() {
  return http.post('/api/ventoy/install', undefined, LONG);
}

export function uninstallVentoy() {
  return http.post('/api/ventoy/uninstall');
}

// replace the set of images on the Ventoy disk
export function setVentoyImages(images: string[]) {
  return http.post('/api/ventoy/images', { images });
}

// build the Ventoy disk and insert it into the disk drive, read-only
export function insertVentoy() {
  return http.post('/api/ventoy/insert', undefined, LONG);
}

export function ejectVentoy() {
  return http.post('/api/ventoy/eject');
}
