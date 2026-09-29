// The media dialog's inline warnings, as PiKVM shows them in its Drive menu,
// computed from what the drive and image lists report.

export type MediaWarning = 'missing' | 'writable' | 'tooBigForCd' | 'tooSmallForCd' | 'empty';

// The CD drive is lun.1 of the mass storage gadget, with 2048-byte sectors.
const CD_SECTOR = 2048;

// The kernel refuses a CD image under 300 sectors: the smallest track is 300
// frames (storage_common.c, fsg_lun_open).
export const CD_MIN_BYTES = 300 * CD_SECTOR;

// IronKVM's kernel carries PiKVM's CD/DVD emulation (ironkvm-dist kernel
// patch 0005): an image past the CD limit is served as a DVD. A DVD's
// physical format addresses sectors in 24 bits from a start of 0x30000, so
// no more than 0xFFFFFF - 0x30000 sectors reach the host, about 31.6 GiB.
// Past that the host reads a disc that ends early. A stock kernel without the
// patch stops at about 2.2 GiB instead; the UI cannot tell which kernel runs,
// so it warns at the limit of the one IronKVM ships.
export const CD_MAX_BYTES = (0xffffff - 0x30000) * CD_SECTOR;

export type DriveReport = {
  id: 'disk' | 'cdrom';
  file: string;
  ro: boolean;
  size?: number;
  missing?: boolean;
};

// isDevice tells a block device (the Ventoy disk) from an image file. Its
// size reads as 0 and it is read-write by design.
function isDevice(file: string) {
  return file.startsWith('/dev/');
}

// cdWarnings are the rules for an image the CD drive serves or would serve.
function cdWarnings(size: number): MediaWarning[] {
  if (size > CD_MAX_BYTES) return ['tooBigForCd'];
  if (size > 0 && size < CD_MIN_BYTES) return ['tooSmallForCd'];
  return [];
}

// driveWarnings lists what is wrong with what a drive serves now.
export function driveWarnings(drive: DriveReport): MediaWarning[] {
  if (!drive.file || isDevice(drive.file)) return [];
  if (drive.missing) return ['missing'];

  const warnings: MediaWarning[] = [];
  if (drive.id === 'disk' && !drive.ro) warnings.push('writable');
  if (drive.id === 'cdrom' && drive.size !== undefined) warnings.push(...cdWarnings(drive.size));
  return warnings;
}

// imageWarnings lists what would go wrong inserting a library image into the
// given drive. An image of unknown size (an older server) has none.
export function imageWarnings(size: number | undefined, target: 'disk' | 'cdrom'): MediaWarning[] {
  if (size === undefined) return [];
  if (size === 0) return ['empty'];
  return target === 'cdrom' ? cdWarnings(size) : [];
}
