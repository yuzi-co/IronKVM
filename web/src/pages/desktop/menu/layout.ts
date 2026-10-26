// The menu bar on a narrow screen. Below the sm breakpoint the grouped bar
// is wider than a phone, so the bar keeps as many entries as the width
// allows, in the order a phone needs them, and the rest move into More.
// Settings and More always stay on the bar.

// The widest viewport, in CSS pixels, that gets the narrow bar. It matches
// Tailwind's sm breakpoint, which the rest of the page follows.
export const NARROW_MENU_MAX_WIDTH = 639;

// A bar entry on a narrow screen. Larger than the 30 px of the wide bar, for
// a finger.
export const NARROW_ITEM_SIZE = 36;

// The room the bar needs around its entries: the screen gutter on each side,
// the bar's own padding, the separator before Settings, and the two
// indicators that show only when something needs attention (the alerts
// triangle and the room microphone). Those are kept free even when they are
// not showing, so the bar does not overflow when one appears.
const GUTTER = 4;
const BAR_PADDING = 12;
const SEPARATOR = 17;
const INDICATORS = 30 + 34;

// The entries a narrow bar may hold, most needed first. The virtual keyboard
// comes first: a phone has no other keyboard to type on the host with.
export const narrowMenuPriority = [
  'virtualKeyboard',
  'keyboard',
  'mouse',
  'power',
  'media',
  'screen',
  'text',
  'terminal',
  'tools',
  'speaker',
  'roomMic',
  'fullscreen'
] as const;

export type NarrowMenuEntry = (typeof narrowMenuPriority)[number];

// narrowMenuSlots is how many entries, besides Settings and More, fit on the
// bar at this viewport width.
export function narrowMenuSlots(viewportWidth: number): number {
  const fixed = 2 * GUTTER + BAR_PADDING + SEPARATOR + INDICATORS + 2 * NARROW_ITEM_SIZE;
  if (!Number.isFinite(viewportWidth)) return 0;
  return Math.max(0, Math.floor((viewportWidth - fixed) / NARROW_ITEM_SIZE));
}

// splitNarrowMenu puts the available entries in priority order and fills the
// bar's slots with the first of them. The rest go into More, in the same
// order.
export function splitNarrowMenu(
  available: readonly NarrowMenuEntry[],
  slots: number
): { bar: NarrowMenuEntry[]; more: NarrowMenuEntry[] } {
  const present = new Set(available);
  const ordered = narrowMenuPriority.filter((entry) => present.has(entry));
  const count = Math.max(0, Math.floor(slots));
  return { bar: ordered.slice(0, count), more: ordered.slice(count) };
}
