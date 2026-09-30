import type { InputRegion } from '@/types';

// sameInputRegion says whether two regions cover the same pixels. The region
// is recomputed on timers, and a new object with the same numbers would
// otherwise re-render everything that reads it and re-register the mouse.
export function sameInputRegion(a: InputRegion | null, b: InputRegion | null): boolean {
  if (a === b) return true;
  if (!a || !b) return false;
  return (
    a.frameWidth === b.frameWidth &&
    a.frameHeight === b.frameHeight &&
    a.left === b.left &&
    a.top === b.top &&
    a.width === b.width &&
    a.height === b.height
  );
}

// Auto region detection reads a frame back from the GPU on every check, so it
// looks often only while the picture is changing. Once the same region has
// been seen AUTO_REGION_STABLE_CHECKS times in a row it slows down, and the
// first different answer brings it back to the fast rate.
export const AUTO_REGION_FAST_MS = 1000;
export const AUTO_REGION_SLOW_MS = 5000;
export const AUTO_REGION_STABLE_CHECKS = 3;

export function autoRegionDelay(confirmations: number): number {
  return confirmations >= AUTO_REGION_STABLE_CHECKS ? AUTO_REGION_SLOW_MS : AUTO_REGION_FAST_MS;
}
