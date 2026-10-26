// The zoom the Scale menu saves for the picture, per browser.

export const VIDEO_SCALE_KEY = 'nano-kvm-video-scale';

export const DEFAULT_VIDEO_SCALE = 1;

// The Scale menu offers 50 % to 200 %. A little room above that lets a
// later menu offer more without old browsers losing their choice; anything
// beyond it is not a zoom anyone picked.
const MAX_VIDEO_SCALE = 4;

type ReadableStorage = Pick<Storage, 'getItem'>;

// readVideoScale is the saved zoom, or 100 % when nothing usable is saved or
// the storage cannot be read (a private window, blocked site data). It is
// read when the app starts, so the picture opens at the saved zoom before
// the Scale menu has ever been opened.
export function readVideoScale(storage: ReadableStorage | undefined): number {
  let raw: string | null;
  try {
    raw = storage?.getItem(VIDEO_SCALE_KEY) ?? null;
  } catch {
    return DEFAULT_VIDEO_SCALE;
  }
  if (!raw) return DEFAULT_VIDEO_SCALE;

  const scale = Number(raw);
  if (!Number.isFinite(scale) || scale <= 0 || scale > MAX_VIDEO_SCALE) {
    return DEFAULT_VIDEO_SCALE;
  }
  return scale;
}
