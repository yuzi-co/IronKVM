// The checks the server makes on a USB identity (server/service/vm/usb-identity.go),
// made here as the operator types so Apply is only offered for what the server
// will take. The server checks again; this only saves a round trip that would
// rebuild nothing. No imports, so the node test runner can load it as it is.

export type UsbIdentity = {
  vendorId: string;
  productId: string;
  manufacturer: string;
  product: string;
  serial: string;
};

export type UsbIdentityPresetName = 'default' | 'logitech' | 'microsoft' | 'dell' | 'custom';

export const USB_STRING_MAX = 64;
export const USB_SERIAL_MIN = 12;

// usbIdValid accepts one to four hex digits, with or without 0x, other than 0.
export function usbIdValid(value: string): boolean {
  const digits = value.trim().toLowerCase().replace(/^0x/, '');
  if (!/^[0-9a-f]{1,4}$/.test(digits)) return false;
  return parseInt(digits, 16) !== 0;
}

// normalizeUsbId is the form the server saves: 0x and four lower-case digits.
export function normalizeUsbId(value: string): string {
  const digits = value.trim().toLowerCase().replace(/^0x/, '');
  return `0x${digits.padStart(4, '0')}`;
}

// usbStringValid accepts printable ASCII, not empty, with no space at either
// end, up to USB_STRING_MAX characters.
export function usbStringValid(value: string): boolean {
  if (value === '' || value.length > USB_STRING_MAX) return false;
  if (value.trim() !== value) return false;
  return /^[\x20-\x7e]+$/.test(value);
}

// usbSerialValid accepts an empty serial, the board's own, or 12 to 64 letters
// and digits.
export function usbSerialValid(value: string): boolean {
  if (value === '') return true;
  if (value.length < USB_SERIAL_MIN || value.length > USB_STRING_MAX) return false;
  return /^[0-9A-Za-z]+$/.test(value);
}

export type UsbIdentityErrors = {
  vendorId?: boolean;
  productId?: boolean;
  manufacturer?: boolean;
  product?: boolean;
  serial?: boolean;
};

// identityErrors names the fields that would be refused. A named preset uses
// its own IDs and strings, so only its serial is checked.
export function identityErrors(
  preset: UsbIdentityPresetName,
  identity: UsbIdentity
): UsbIdentityErrors {
  const errors: UsbIdentityErrors = {};
  if (!usbSerialValid(identity.serial)) errors.serial = true;
  if (preset !== 'custom') return errors;

  if (!usbIdValid(identity.vendorId)) errors.vendorId = true;
  if (!usbIdValid(identity.productId)) errors.productId = true;
  if (!usbStringValid(identity.manufacturer)) errors.manufacturer = true;
  if (!usbStringValid(identity.product)) errors.product = true;
  return errors;
}

export function sameIdentity(a: UsbIdentity, b: UsbIdentity): boolean {
  return (
    normalizeUsbId(a.vendorId) === normalizeUsbId(b.vendorId) &&
    normalizeUsbId(a.productId) === normalizeUsbId(b.productId) &&
    a.manufacturer === b.manufacturer &&
    a.product === b.product &&
    a.serial === b.serial
  );
}
