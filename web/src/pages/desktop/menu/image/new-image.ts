// The rules of a new exchange image, mirrored from the server so the dialog
// can say what is wrong before anything is sent. The server checks them all
// again.

export const MiB = 1024 * 1024;
export const GiB = 1024 * MiB;

// The sizes offered as buttons; anything else is a custom size.
export const PRESETS = [1 * GiB, 4 * GiB, 8 * GiB, 16 * GiB, 32 * GiB];

// What the server says an image may be: min and max in bytes, and the free
// space max was derived from.
export type ImageSpace = { min: number; max: number; free: number };

export type SizeUnit = 'MiB' | 'GiB';

const FORBIDDEN = /[<>:"/\\|?*]/;
const RESERVED = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])$/i;

function hasControl(s: string): boolean {
  return [...s].some((c) => c.charCodeAt(0) < 0x20 || c === '\x7f');
}

// imageFileName is the file the server will create for name.
export function imageFileName(name: string): string {
  const trimmed = name.trim();
  if (trimmed === '') return '';
  return trimmed.toLowerCase().endsWith('.img') ? trimmed : `${trimmed}.img`;
}

// checkName returns the locale key of what is wrong with name, or ''.
export function checkName(name: string): string {
  const trimmed = name.trim();
  if (trimmed === '') return 'nameRequired';
  if (trimmed.toLowerCase().endsWith('.iso')) return 'nameIso';
  const file = imageFileName(trimmed);
  if (
    FORBIDDEN.test(trimmed) ||
    hasControl(trimmed) ||
    trimmed.startsWith('.') ||
    /[. ]$/.test(trimmed) ||
    RESERVED.test(trimmed.split('.')[0]) ||
    [...file].length > 255
  ) {
    return 'nameInvalid';
  }
  return '';
}

// toBytes turns a custom size into bytes, in whole MiB.
export function toBytes(value: number, unit: SizeUnit): number {
  const mib = Math.floor(value * (unit === 'GiB' ? 1024 : 1));
  return Math.max(0, mib) * MiB;
}

// checkSize returns the locale key of what is wrong with size, or ''.
export function checkSize(size: number, space: ImageSpace): string {
  if (size < space.min) return 'sizeTooSmall';
  if (size > space.max) return 'sizeTooLarge';
  return '';
}

// defaultSize is what the dialog starts with: 4 GiB, or the largest preset
// below it that fits, or all the room there is when no preset does.
export function defaultSize(space: ImageSpace): number {
  const fitting = PRESETS.filter((p) => p <= 4 * GiB && p <= space.max);
  if (fitting.length > 0) return fitting[fitting.length - 1];
  return Math.max(0, Math.floor(space.max / MiB) * MiB);
}
