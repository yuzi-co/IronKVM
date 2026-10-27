import type { InstallProgress, SlotName, SlotStatus } from '@/api/slot.ts';

// What Settings > Update offers for the A/B slots, decided from `slot status`
// as the server reads it. The tool refuses anything these let through that it
// would not do, so they only keep the page from offering a button that is
// certain to fail.

// progressPercent is how far the install is, for the bar. The verify phase
// has no byte count, and a read back is the second half of the work after
// the write, so the bar fills over the write and again over the read back
// rather than jumping back to zero between the two.
export function progressPercent(p: InstallProgress): number {
  if (p.phase === 'done') return 100;
  if (p.total <= 0) return 0;
  const part = Math.min(1, Math.max(0, p.done / p.total));
  if (p.phase === 'write') return Math.floor(part * 50);
  if (p.phase === 'readback') return 50 + Math.floor(part * 50);
  return 0;
}

// installTargets are the slots an image may be written to: every slot the
// board has except the one running, and except one already being written.
export function installTargets(st: SlotStatus): SlotName[] {
  return st.slots
    .filter((s) => s.name !== st.running && s.state !== 'installing')
    .map((s) => s.name);
}

// tryTargets are the slots a trial may be armed for: A or B, not the one
// running, holding a whole image, and not armed already. A slot with no
// record ("unknown") was written before records existed and is tried as the
// tool always tried it.
export function tryTargets(st: SlotStatus): ('a' | 'b')[] {
  const armed = armedTrial(st);
  return st.slots
    .filter((s) => s.name === 'a' || s.name === 'b')
    .filter((s) => s.name !== st.running && s.name !== armed)
    .filter((s) => s.state === 'installed' || s.state === 'unknown')
    .map((s) => s.name as 'a' | 'b');
}

// armedTrial is the slot the next boot will try, or ''.
export function armedTrial(st: SlotStatus): string {
  if (st.trial && st.trial !== 'none') return st.trial;
  return st.slots.find((s) => s.trial === 'armed')?.name ?? '';
}

// onTrial says the board is running a slot it does not yet trust: the boot
// after a trial was armed. Confirming makes it the trusted slot.
export function onTrial(st: SlotStatus): boolean {
  return (st.running === 'a' || st.running === 'b') && st.running !== st.trusted;
}

// revertTarget is the slot a revert makes trusted again, or ''.
export function revertTarget(st: SlotStatus): string {
  if (!st.prev || st.prev === 'none' || st.prev === st.trusted) return '';
  return st.prev;
}

// isHttpsUrl is the check the server makes before it hands a URL to the tool.
export function isHttpsUrl(value: string): boolean {
  if (!value || /\s/.test(value)) return false;
  try {
    const u = new URL(value);
    return u.protocol === 'https:' && u.hostname !== '';
  } catch {
    return false;
  }
}
