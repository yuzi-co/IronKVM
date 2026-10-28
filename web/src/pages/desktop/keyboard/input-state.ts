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

function familyOf(code: string): Family | undefined {
  return FAMILIES.find((family) => family.codes.includes(code));
}

/**
 * Modifiers recorded as pressed that the event says are released. On a keydown
 * the family of the key behind the event is left alone, since browsers
 * disagree on whether a modifier's own keydown already carries its new state.
 * On a modifier's keyup its family is checked as well, apart from the key
 * itself, which the keyup releases anyway: a keyup that reports the family as
 * no longer held means no side of it is held. Windows sends a single Shift
 * keyup when the last of two held Shifts is released, and none for the first,
 * so this is the only point at which the first one can be let go.
 */
export function staleModifiers(
  pressed: Iterable<string>,
  event: ModifierState,
  eventCode: string,
  type: 'keydown' | 'keyup' = 'keydown'
): string[] {
  const pressedSet = new Set(pressed);
  const stale: string[] = [];

  for (const family of FAMILIES) {
    if (type === 'keydown' && family.codes.includes(eventCode)) continue;
    for (const code of family.codes) {
      if (code === eventCode) continue;
      if (pressedSet.has(code) && !family.held(event, code)) stale.push(code);
    }
  }

  return stale;
}

export interface ModifierChanges {
  release: string[];
  press: string[];
}

/**
 * The re-sync above, with a memory for the sides it releases on another side's
 * keyup. A browser that clears a modifier's flag on its own keyup whatever the
 * other side is doing would have that keyup release both sides while one is
 * still held. On Windows the two Ctrls each send their own keyup, and whether
 * the first one reports Ctrl as held has not been recorded. Such a side is
 * pressed again by the next
 * event for another key that reports its family as held while no side of the
 * family is held on the host. The memory never keeps a key held: it can only
 * restore one the browser reports as held, and it is dropped as soon as an
 * event reports the family as up, a key of the family is pressed, the side has
 * its own keyup, or every key is released.
 */
export class ModifierSync {
  private readonly implied = new Set<string>();

  /** Modifiers to release and to press again before handling a key event. */
  sync(
    pressed: Iterable<string>,
    event: ModifierState,
    type: 'keydown' | 'keyup',
    eventCode: string
  ): ModifierChanges {
    const pressedSet = new Set(pressed);
    const own = familyOf(eventCode);

    if (own && type === 'keydown') {
      for (const code of own.codes) this.implied.delete(code);
    } else {
      this.implied.delete(eventCode);
    }

    const release = staleModifiers(pressedSet, event, eventCode, type);
    for (const code of release) {
      pressedSet.delete(code);
      if (own?.codes.includes(code)) this.implied.add(code);
    }

    const press: string[] = [];
    for (const family of FAMILIES) {
      if (family === own) continue;
      for (const code of family.codes) {
        if (!this.implied.has(code)) continue;
        if (!family.held(event, code)) {
          this.implied.delete(code);
        } else if (!family.codes.some((side) => pressedSet.has(side))) {
          this.implied.delete(code);
          pressedSet.add(code);
          press.push(code);
        }
      }
    }

    return { release, press };
  }

  /** Forgets every released side, for when every key has been released. */
  reset() {
    this.implied.clear();
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
