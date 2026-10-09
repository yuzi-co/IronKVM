export const QualityMap = new Map([
  [1, 100],
  [2, 80],
  [3, 60],
  [4, 50]
]);

export const BitRateMap = new Map([
  [1, 5000],
  [2, 3000],
  [3, 2000],
  [4, 1000]
]);

// The bitrate, in kbit/s, of a board that has never been configured: the
// "High" step. It matches the server's DefaultBitRate (server/common/screen.go)
// and what kvm_system seeds into a missing qlty file. ironkvm-dist trial 68
// measured 5000 against it: three WebRTC viewers at 60 fps lost over 4 fps on
// the vendor slot, where 3000 lost at most 1.4.
export const DEFAULT_BIT_RATE = 3000;

// The step the menu ticks when the value it holds is none of the four: key 2,
// which is JPEG quality 80 and bitrate 3000, both the server's defaults.
export const DEFAULT_QUALITY_KEY = 2;

export function getQualityMap(videoMode: string) {
  if (videoMode === 'mjpeg') {
    return QualityMap;
  }
  if (videoMode === 'direct' || videoMode === 'h264') {
    return BitRateMap;
  }
  return null;
}

// The quality item offers four steps rather than a value, and what those steps
// mean depends on the delivery path: a JPEG quality on MJPEG, a bitrate on both
// H.264 paths. The server reports the two separately because it cannot know
// which one the viewer is asking about.
export function qualityKey(
  videoMode: string,
  settings: { quality: number; bitRate: number }
): number {
  const qualityMap = getQualityMap(videoMode);
  if (!qualityMap) return DEFAULT_QUALITY_KEY;

  const value = videoMode === 'mjpeg' ? settings.quality : settings.bitRate;
  for (const [key, mapped] of qualityMap) {
    if (mapped === value) return key;
  }

  return DEFAULT_QUALITY_KEY;
}

export function getScreenType(videoMode: string) {
  if (videoMode === 'mjpeg') {
    return 0;
  }
  if (videoMode === 'direct' || videoMode === 'h264') {
    return 1;
  }
  return null;
}
