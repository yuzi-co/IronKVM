// What the prompt for an agent's press shows, decided without the UI so a test
// can check it.

import type { PowerConfirmation } from '@/api/power-confirm.ts';

// secondsLeft is how long a press still waits, rounded up, never below 0.
export function secondsLeft(expiresAt: string, now: number): number {
  const at = Date.parse(expiresAt);
  if (Number.isNaN(at)) return 0;
  return Math.max(0, Math.ceil((at - now) / 1000));
}

// pressKeys are the translation keys of a press's name and of what it does
// to the host. They are the power menu's own.
export function pressKeys(c: Pick<PowerConfirmation, 'button' | 'press'>): {
  label: string;
  desc: string;
} {
  if (c.button === 'reset') {
    return { label: 'power.reset', desc: 'power.resetDesc' };
  }
  if (c.press === 'long') {
    return { label: 'power.powerLong', desc: 'power.powerLongDesc' };
  }
  return { label: 'power.powerShort', desc: 'power.powerShortDesc' };
}

// visiblePresses drops the presses whose time is up, which the server ends at
// its next look; the poll may still list them for a moment.
export function visiblePresses(list: PowerConfirmation[], now: number): PowerConfirmation[] {
  return list.filter((c) => c.state === 'pending' && secondsLeft(c.expiresAt, now) > 0);
}
