// The USB section of Settings > Device as a draft: the switches and the network
// controls change nothing on the board until Apply sends the whole set in one
// request, and the board rebuilds its gadget once.
//
// The numbers come from the server (each device's cost, the used count and the
// total), and the server checks the applied set again, so these rules only
// decide what the page offers.

import type {
  UsbApplyRequest,
  UsbNetwork,
  UsbNetworkMode,
  VirtualDeviceName,
  VirtualDevices
} from '../api/virtual-device.ts';

export type UsbDraft = {
  console: boolean;
  disk: boolean;
  audio: boolean;
  network: UsbNetworkMode;
  subnet: string;
};

export const USB_DEVICES: VirtualDeviceName[] = ['console', 'disk', 'network', 'audio'];

// The shape only. The server holds the real rules (private, /24 to /30, the
// network address) and says which one a subnet breaks.
const SUBNET_SHAPE = /^\d{1,3}(\.\d{1,3}){3}\/\d{1,2}$/;

// boardDraft is the draft that changes nothing: what the board carries now.
export function boardDraft(devices: VirtualDevices, network: UsbNetwork): UsbDraft {
  return {
    console: devices.console.enabled,
    disk: devices.disk.enabled,
    audio: devices.audio.enabled,
    network: network.mode,
    subnet: network.subnet
  };
}

export function isOn(draft: UsbDraft, device: VirtualDeviceName): boolean {
  return device === 'network' ? draft.network !== 'off' : draft[device];
}

// changedDevices lists the rows the draft changes. The subnet counts only while
// the link stays on, because its field is hidden with the link off and the
// draft then sends no subnet at all.
export function changedDevices(draft: UsbDraft, board: UsbDraft): VirtualDeviceName[] {
  return USB_DEVICES.filter((device) => {
    if (device !== 'network') return draft[device] !== board[device];

    if (draft.network !== board.network) return true;
    return draft.network !== 'off' && draft.subnet.trim() !== board.subnet;
  });
}

// draftUsed counts the inbound endpoints the draft would use. The keyboard and
// mouse are whatever the server's count holds beyond the devices it reports
// enabled, so switching HID off on the board is accounted for without knowing
// how.
export function draftUsed(draft: UsbDraft, devices: VirtualDevices): number {
  let used = devices.used;

  for (const device of USB_DEVICES) {
    if (devices[device].enabled) used -= devices[device].cost;
    if (isOn(draft, device)) used += devices[device].cost;
  }

  return used;
}

// fitsInDraft answers whether a row can be switched on beside the rest of the
// draft. A row already on in the draft always can stay on.
export function fitsInDraft(
  draft: UsbDraft,
  devices: VirtualDevices,
  device: VirtualDeviceName
): boolean {
  if (isOn(draft, device)) return true;

  return devices[device].cost <= devices.total - draftUsed(draft, devices);
}

export function subnetShapeValid(subnet: string): boolean {
  return SUBNET_SHAPE.test(subnet.trim());
}

// subnetValid gates Apply. With the link off the subnet is not sent.
export function subnetValid(draft: UsbDraft): boolean {
  return draft.network === 'off' || subnetShapeValid(draft.subnet);
}

export function applyRequest(draft: UsbDraft): UsbApplyRequest {
  return {
    console: draft.console,
    disk: draft.disk,
    audio: draft.audio,
    network: {
      mode: draft.network,
      subnet: draft.network === 'off' ? '' : draft.subnet.trim()
    }
  };
}
