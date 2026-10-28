// Browser key events are not a reliable stream of matched pairs. A keyup can be
// lost to an OS shortcut, an Alt menu or a dialog that takes focus inside the
// page, and an input-method composition can end without its compositionend
// reaching the document. Every key event carries the browser's own view of the
// modifiers, though, so these helpers derive the true state from each new event
// rather than waiting for the next blur to clean up (ironkvm-dist#154).

export interface ModifierState {
  readonly ctrlKey: boolean;
  readonly shiftKey: boolean;
  readonly altKey: boolean;
  readonly metaKey: boolean;
  getModifierState(key: string): boolean;
}

type Family = {
  codes: readonly string[];
  held: (event: ModifierState, code: string) => boolean;
};

const FAMILIES: readonly Family[] = [
  { codes: ['ControlLeft', 'ControlRight'], held: (event) => event.ctrlKey },
  { codes: ['ShiftLeft', 'ShiftRight'], held: (event) => event.shiftKey },
  {
    codes: ['AltLeft', 'AltRight'],
    // Linux reports a right Alt that acts as AltGr as AltGraph, not as Alt.
    // Windows reports AltGr as Ctrl+Alt, which altKey covers.
    held: (event, code) =>
      event.altKey || (code === 'AltRight' && event.getModifierState('AltGraph'))
  },
  { codes: ['MetaLeft', 'MetaRight'], held: (event) => event.metaKey }
];

/**
 * Modifiers recorded as pressed that the event says are released. The family of
 * the key behind the event is left alone: browsers disagree on whether a
 * modifier's own keydown and keyup already carry its new state, and with both
 * sides of one held, the flag cannot say which side was let go.
 */
export function staleModifiers(
  pressed: Iterable<string>,
  event: ModifierState,
  eventCode: string
): string[] {
  const pressedSet = new Set(pressed);
  const stale: string[] = [];

  for (const family of FAMILIES) {
    if (family.codes.includes(eventCode)) continue;
    for (const code of family.codes) {
      if (pressedSet.has(code) && !family.held(event, code)) stale.push(code);
    }
  }

  return stale;
}

/**
 * The re-sync above with a memory for one case it gets wrong. Windows keeps one
 * combined state per modifier family, so once either Shift is released the
 * browser reports Shift as up, although the other Shift is still physically
 * held. Ctrl and Alt are tracked the same way, so they are given the same
 * treatment, and so is Meta, at no cost. A side that is still pressed when its
 * family's other side has its keyup is therefore marked as unconfirmed: the
 * flag can no longer say whether it is held, so a flag that reads as released
 * does not release it. It is kept until its own keyup, a blur or visibility
 * change that releases every key, or an event that reports the family as held
 * again. That last event shows the browser's state agrees with ours once more,
 * so the mark is dropped and the re-sync covers the key from then on. The cost
 * is that a side whose keyup is lost while it is unconfirmed stays held until
 * the next blur, which is the behaviour from before the re-sync existed.
 */
export class ModifierSync {
  private readonly unconfirmed = new Set<string>();

  /** Modifiers to release before handling an event for eventCode. */
  stale(pressed: Iterable<string>, event: ModifierState, eventCode: string): string[] {
    const pressedSet = new Set(pressed);

    for (const family of FAMILIES) {
      if (family.codes.includes(eventCode)) continue;
      for (const code of family.codes) {
        if (this.unconfirmed.has(code) && family.held(event, code)) {
          this.unconfirmed.delete(code);
        }
      }
    }

    return staleModifiers(pressedSet, event, eventCode).filter(
      (code) => !this.unconfirmed.has(code)
    );
  }

  /** Records a key's own keydown. A fresh press starts with nothing in doubt. */
  keyDown(code: string) {
    this.unconfirmed.delete(code);
  }

  /** Records a key's own keyup, with the keys still pressed when it came. */
  keyUp(code: string, pressed: Iterable<string>) {
    this.unconfirmed.delete(code);
    const family = FAMILIES.find((f) => f.codes.includes(code));
    if (!family) return;
    for (const other of pressed) {
      if (other !== code && family.codes.includes(other)) this.unconfirmed.add(other);
    }
  }

  /** Forgets every mark, for when every key has been released. */
  reset() {
    this.unconfirmed.clear();
  }
}

export interface CompositionKeyEvent {
  readonly isComposing: boolean;
  readonly keyCode: number;
  readonly key: string;
}

// keyCode 229, key "Process", marks a key the input method is consuming.
const IME_KEY_CODE = 229;

/**
 * Whether a keydown belongs to an input-method composition. Once one has
 * started, a keydown that is neither flagged as composing nor addressed to the
 * input method shows it is over, even if its compositionend never arrived.
 */
export function isCompositionKey(started: boolean, event: CompositionKeyEvent): boolean {
  if (event.isComposing) return true;
  if (!started) return false;
  return event.keyCode === IME_KEY_CODE || event.key === 'Process';
}
