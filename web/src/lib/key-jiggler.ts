// The keys the key jiggler can press, in menu order. F15 comes first: nothing
// binds it, so it is the least intrusive.
export const keyJigglerKeys = ['f15', 'shift', 'ctrl'] as const;

export type KeyJigglerKey = (typeof keyJigglerKeys)[number];

// keyJigglerKeyFrom reads the key the server reports. A missing or unknown key
// reads as F15, which is what the server presses in its place.
export function keyJigglerKeyFrom(key?: string): KeyJigglerKey {
  return (keyJigglerKeys as readonly string[]).includes(key ?? '') ? (key as KeyJigglerKey) : 'f15';
}
