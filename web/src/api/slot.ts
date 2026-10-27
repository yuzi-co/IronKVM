import { http } from '@/lib/http.ts';

// The A/B slots of the distribution image (server/service/slot), behind the
// "slot-install" feature. Admin only. The server runs /usr/bin/slot and every
// refusal is the tool's.

export type SlotName = 'a' | 'b' | 'recovery';

export type SlotInfo = {
  name: SlotName;
  // installed, installing, interrupted, bad, or unknown for a slot written
  // before records existed.
  state: string;
  version?: string;
  sha256?: string;
  installed?: string;
  // none, armed, booted, passed or failed.
  trial?: string;
};

export type SlotStatus = {
  running: string;
  trusted: string;
  // The slot armed for the next boot, or 'none'.
  trial?: string;
  // The slot a revert returns to, or 'none'.
  prev?: string;
  recovery?: string;
  bootloader?: string;
  kernel?: string;
  slots: SlotInfo[];
  image?: Record<string, string>;
};

export type SlotImage = {
  path: string;
  size: number;
  modified: string;
  signed: boolean;
  device?: string;
  version?: string;
};

export type InstallPhase = '' | 'verify' | 'write' | 'readback' | 'done' | 'failed';

export type InstallProgress = {
  running: boolean;
  phase: InstallPhase;
  done: number;
  total: number;
  message?: string;
  slot?: SlotName;
  source?: string;
  started?: string;
};

export type InstallRequest = {
  slot: SlotName;
  url?: string;
  file?: string;
  force?: boolean;
};

// Error codes beside -1 (server/service/slot/api.go).
export const SLOT_UNAVAILABLE = -2;
export const SLOT_BUSY = -3;

export function getSlotStatus() {
  return http.get('/api/slot');
}

export function getSlotImages() {
  return http.get('/api/slot/images');
}

export function getInstallProgress() {
  return http.get('/api/slot/install');
}

export function installSlot(req: InstallRequest) {
  return http.post('/api/slot/install', req);
}

// try can mount the slot to read its record, so it is given longer than the
// default.
export function trySlot(slot: 'a' | 'b') {
  return http.request({
    method: 'post',
    url: '/api/slot/try',
    data: { slot },
    timeout: 150 * 1000
  });
}

export function confirmSlot() {
  return http.post('/api/slot/confirm');
}

export function revertSlot() {
  return http.post('/api/slot/revert');
}
