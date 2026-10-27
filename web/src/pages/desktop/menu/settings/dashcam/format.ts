// What the dashcam page shows that can be tested without a browser.

// The limits of the settings, as server/service/dashcam/settings.go keeps
// them.
export const LIMITS = {
  minSeconds: 30,
  maxSeconds: 120,
  maxPostSeconds: 30,
  minKbps: 250,
  maxKbps: 20000,
  minStorageMB: 20,
  maxStorageMB: 100000
} as const;

// The triggers as the server names them in a clip, and the key of each one's
// label under settings.dashcam.triggerNames.
const TRIGGER_KEYS: Record<string, string> = {
  power: 'power',
  reset: 'reset',
  watchdog: 'watchdog',
  'signal-lost': 'signalLost',
  'signal-restored': 'signalRestored',
  failsafe: 'failsafe',
  manual: 'manual',
  api: 'api'
};

// triggerKey is the label key of a trigger, or null for one this UI does not
// know, which is then shown as the server spells it.
export function triggerKey(trigger: string): string | null {
  return TRIGGER_KEYS[trigger] ?? null;
}

// clipSize shows a clip's size in kB or MB.
export function clipSize(bytes: number): string {
  if (bytes < 1000 * 1000) return `${Math.max(1, Math.round(bytes / 1000))} kB`;
  return `${(bytes / (1000 * 1000)).toFixed(1)} MB`;
}

// settingsProblem names the first setting out of its limits, or null.
export function settingsProblem(s: {
  seconds: number;
  postSeconds: number;
  maxKbps: number;
  storageMB: number;
}): 'seconds' | 'postSeconds' | 'maxKbps' | 'storageMB' | null {
  const within = (v: number, lo: number, hi: number) => Number.isInteger(v) && v >= lo && v <= hi;
  if (!within(s.seconds, LIMITS.minSeconds, LIMITS.maxSeconds)) return 'seconds';
  if (!within(s.postSeconds, 0, LIMITS.maxPostSeconds)) return 'postSeconds';
  if (s.maxKbps !== 0 && !within(s.maxKbps, LIMITS.minKbps, LIMITS.maxKbps)) return 'maxKbps';
  if (!within(s.storageMB, LIMITS.minStorageMB, LIMITS.maxStorageMB)) return 'storageMB';
  return null;
}
