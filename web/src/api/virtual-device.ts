import { http } from '@/lib/http.ts';
import type { UsbIdentity, UsbIdentityPresetName } from '@/lib/usb-identity.ts';
import {
  readVirtualDevices,
  type VirtualDeviceName,
  type VirtualDevices
} from '@/lib/virtual-devices.ts';

export type {
  VirtualDeviceName,
  VirtualDeviceState,
  VirtualDevices
} from '@/lib/virtual-devices.ts';

// get virtual devices status, in one shape whichever API version the server
// speaks
export async function getVirtualDevice() {
  const rsp = await http.get('/api/vm/device/virtual');
  if (rsp.code === 0 && rsp.data) rsp.data = readVirtualDevices(rsp.data);
  return rsp;
}

// mount/unmount virtual device
export function updateVirtualDevice(device: VirtualDeviceName) {
  const data = {
    device
  };

  return http.post('/api/vm/device/virtual', data);
}

// The USB network link to the host. rndis is only ever read back, from a board
// an older server set up; the settings page offers off, ncm and ecm.
export type UsbNetworkMode = 'off' | 'ncm' | 'ecm' | 'rndis';

export type UsbNetwork = {
  mode: UsbNetworkMode;
  subnet: string;
  board: string;
  host: string;
  active: boolean;
  fits: boolean;
  refusal: string;
};

export function getUsbNetwork() {
  return http.get('/api/vm/device/usb-network');
}

// An empty subnet keeps the one in use.
export function setUsbNetwork(mode: Exclude<UsbNetworkMode, 'rndis'>, subnet: string) {
  return http.post('/api/vm/device/usb-network', { mode, subnet });
}

// What an apply answers: the virtual devices as GET reports them, and the USB
// network link beside them.
export type UsbApplied = VirtualDevices & { usbNetwork: UsbNetwork };

// Apply every USB device at once. The board checks the whole set against its
// endpoint budget and rebuilds the gadget once. An empty subnet keeps the one
// in use; rndis is accepted only on a board that already runs it.
export type UsbApplyRequest = {
  console: boolean;
  disk: boolean;
  audio: boolean;
  network: { mode: UsbNetworkMode; subnet: string };
};

export async function applyUsbDevices(request: UsbApplyRequest) {
  const rsp = await http.post('/api/vm/device/virtual/apply', request);
  if (rsp.code === 0 && rsp.data) rsp.data = readVirtualDevices(rsp.data);
  return rsp;
}

// The IDs and strings the gadget shows the host. preset is default, a named
// preset or custom; live is what the running gadget carries; bootOverride is
// true while /boot/usb.vid or /boot/usb.pid exist, which apply only when no
// identity is saved.
export type UsbIdentityState = {
  preset: UsbIdentityPresetName;
  identity: UsbIdentity;
  live: UsbIdentity;
  presets: (UsbIdentity & { name: UsbIdentityPresetName })[];
  bootOverride: boolean;
};

export function getUsbIdentity() {
  return http.get('/api/vm/device/usb-identity');
}

// Saving rebuilds the gadget, so the host enumerates it again. A named preset
// takes only the serial from identity.
export function setUsbIdentity(preset: UsbIdentityPresetName, identity: UsbIdentity) {
  return http.post('/api/vm/device/usb-identity', { preset, ...identity });
}
