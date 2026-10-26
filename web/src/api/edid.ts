import { presetDownloadPath } from '@/lib/edid.ts';
import { http } from '@/lib/http.ts';
import { getBaseUrl } from '@/lib/service.ts';

// The read side of EDID management (server/router/edid.go). Nothing here
// applies an EDID: writing one to the HDMI bridge is parked.

// what the board presents, or why it cannot be read
export function getCurrentEdid() {
  return http.get('/api/edid/current');
}

// the preset library
export function listEdidPresets() {
  return http.get('/api/edid/presets');
}

// one preset, as hex and decoded
export function getEdidPreset(id: string) {
  return http.get(`/api/edid/presets/${encodeURIComponent(id)}`);
}

// decode a file, binary or hex text; it is not applied
export function decodeEdid(file: Blob) {
  return http.post('/api/edid/decode', file, {
    headers: { 'Content-Type': 'application/octet-stream' }
  });
}

// the link that downloads a preset as a .bin file
export function getEdidPresetUrl(id: string) {
  return `${getBaseUrl('http')}${presetDownloadPath(id)}`;
}
