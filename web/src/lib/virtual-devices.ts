// GET /api/vm/device/virtual and its apply answer, read into one shape.
//
// Server API 2 moved each function's state under "functions" and gave the top
// level back to Sipeed's booleans, which clients written against Sipeed's API
// read (ironkvm-dist#124). API 1 put the states at the top level. This module
// runs under node --test, so it imports nothing.

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

const names: VirtualDeviceName[] = ['console', 'network', 'disk', 'audio'];

function readState(value: unknown): VirtualDeviceState {
  const state = (
    typeof value === 'object' && value !== null ? value : {}
  ) as Partial<VirtualDeviceState>;
  return {
    enabled: state.enabled === true,
    active: state.active === true,
    cost: typeof state.cost === 'number' ? state.cost : 0
  };
}

// readVirtualDevices takes the data field of either answer, and keeps any
// other field (an apply's usbNetwork) as it is.
export function readVirtualDevices<T extends object>(
  data: T
): VirtualDevices & Omit<T, keyof VirtualDevices> {
  const raw = data as Record<string, unknown>;
  const functions =
    typeof raw.functions === 'object' && raw.functions !== null
      ? (raw.functions as Record<string, unknown>)
      : raw;

  const devices: Record<string, unknown> = { ...raw };
  delete devices.functions;
  for (const name of names) devices[name] = readState(functions[name]);
  devices.used = typeof raw.used === 'number' ? raw.used : 0;
  devices.total = typeof raw.total === 'number' ? raw.total : 0;
  devices.fits = Array.isArray(raw.fits) ? raw.fits : [];

  return devices as VirtualDevices & Omit<T, keyof VirtualDevices>;
}
