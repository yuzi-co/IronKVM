import { http } from '@/lib/http.ts';

export type VirtualDeviceName = 'console' | 'disk' | 'network' | 'audio';

// enabled is the /boot marker, active is the function the running gadget
// actually carries. They differ when the USB controller ran out of endpoints.
export type VirtualDeviceState = {
  enabled: boolean;
  active: boolean;
  cost: number;
};

export type VirtualDevices = {
  console: VirtualDeviceState;
  network: VirtualDeviceState;
  disk: VirtualDeviceState;
  audio: VirtualDeviceState;
  used: number;
  total: number;
  // Every largest set of optional devices that fits beside the keyboard and
  // mouse, highest priority first.
  fits: VirtualDeviceName[][];
};

// get virtual devices status
export function getVirtualDevice() {
  return http.get('/api/vm/device/virtual');
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
