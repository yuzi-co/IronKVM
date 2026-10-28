import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isCompositionKey, ModifierSync, staleModifiers } from './input-state.ts';

function event(held: {
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  meta?: boolean;
  altGraph?: boolean;
}) {
  return {
    ctrlKey: held.ctrl ?? false,
    shiftKey: held.shift ?? false,
    altKey: held.alt ?? false,
    metaKey: held.meta ?? false,
    getModifierState: (key: string) => key === 'AltGraph' && (held.altGraph ?? false)
  };
}

test('a modifier whose keyup was lost is released by the next key event', () => {
  const pressed = ['AltLeft', 'ControlLeft', 'ShiftRight'];
  assert.deepEqual(staleModifiers(pressed, event({ shift: true }), 'KeyA'), [
    'ControlLeft',
    'AltLeft'
  ]);
});

test('both sides of a modifier are released when the event reports neither', () => {
  const pressed = ['ControlLeft', 'ControlRight', 'MetaLeft', 'MetaRight'];
  assert.deepEqual(staleModifiers(pressed, event({}), 'KeyB'), [
    'ControlLeft',
    'ControlRight',
    'MetaLeft',
    'MetaRight'
  ]);
});

test('held modifiers and ordinary keys are kept', () => {
  const pressed = ['ControlLeft', 'MetaRight', 'KeyC'];
  assert.deepEqual(staleModifiers(pressed, event({ ctrl: true, meta: true }), 'KeyV'), []);
});

test("a modifier's own event does not re-sync its family", () => {
  // Browsers disagree on whether a modifier's keydown and keyup carry its new
  // state, so ShiftLeft's keyup may report shiftKey either way.
  assert.deepEqual(staleModifiers(['ControlLeft'], event({}), 'ControlRight'), []);
  assert.deepEqual(staleModifiers(['AltLeft'], event({}), 'AltLeft'), []);
  assert.deepEqual(staleModifiers(['ShiftLeft', 'ShiftRight'], event({}), 'ShiftRight'), []);
});

test('the event for one modifier still re-syncs the others', () => {
  assert.deepEqual(staleModifiers(['ControlLeft', 'ShiftLeft'], event({}), 'ShiftLeft'), [
    'ControlLeft'
  ]);
});

test('AltGraph keeps the right Alt key held, not the left', () => {
  assert.deepEqual(staleModifiers(['AltRight'], event({ altGraph: true }), 'KeyQ'), []);
  assert.deepEqual(staleModifiers(['AltRight'], event({}), 'KeyQ'), ['AltRight']);
  assert.deepEqual(
    staleModifiers(['AltLeft', 'AltRight'], event({ altGraph: true }), 'KeyQ'),
    ['AltLeft']
  );
});

test('Windows AltGr, reported as Ctrl+Alt, keeps AltRight held', () => {
  assert.deepEqual(staleModifiers(['AltRight'], event({ ctrl: true, alt: true }), 'KeyQ'), []);
});

// Chromium on Windows reports a left Ctrl and a left Alt held together as
// AltGraph as well, with ctrlKey and altKey still set. The AltGraph rule above
// is about the right Alt only, so a real Ctrl+Alt chord must keep both keys.
test('Windows Ctrl+Alt, also reported as AltGraph, keeps the left Ctrl held', () => {
  const ctrlAlt = event({ ctrl: true, alt: true, altGraph: true });
  assert.deepEqual(staleModifiers(['ControlLeft'], ctrlAlt, 'AltLeft'), []);
  assert.deepEqual(staleModifiers(['ControlLeft', 'AltLeft'], ctrlAlt, 'Delete'), []);
  assert.deepEqual(staleModifiers(['ControlLeft', 'AltLeft'], ctrlAlt, 'KeyT'), []);
});

test('a Ctrl let go while Alt is held is released by the next event', () => {
  assert.deepEqual(
    staleModifiers(['ControlLeft', 'AltLeft'], event({ alt: true }), 'Tab'),
    ['ControlLeft']
  );
});

test("a modifier's keyup re-syncs its own family, apart from itself", () => {
  assert.deepEqual(staleModifiers(['ShiftLeft', 'ShiftRight'], event({}), 'ShiftRight', 'keyup'), [
    'ShiftLeft'
  ]);
  assert.deepEqual(
    staleModifiers(['ShiftLeft', 'ShiftRight'], event({ shift: true }), 'ShiftRight', 'keyup'),
    []
  );
  // Linux reports a right Alt used as AltGr as AltGraph, not as Alt.
  assert.deepEqual(
    staleModifiers(['AltLeft', 'AltRight'], event({ altGraph: true }), 'AltLeft', 'keyup'),
    []
  );
});

type Held = Parameters<typeof event>[0];
type Step = ['down' | 'up', string, Held, 'repeat'?] | ['blur'];

// Replays browser events the way the keyboard component handles them and
// returns what the host is sent. Each key event first has the modifiers
// re-synced, then is passed on. Like the component, the replay does not look at
// the repeat flag: a keydown is dropped only for a key already held, so a
// first keydown that Windows flags as a repeat still presses its key.
function replay(steps: Step[]): string[] {
  const sync = new ModifierSync();
  const pressed = new Set<string>();
  const host: string[] = [];
  const release = (code: string) => {
    pressed.delete(code);
    host.push(`${code} up`);
  };
  const press = (code: string) => {
    pressed.add(code);
    host.push(`${code} down`);
  };

  for (const step of steps) {
    if (step[0] === 'blur') {
      pressed.forEach((code) => host.push(`${code} up`));
      pressed.clear();
      sync.reset();
      continue;
    }
    const [type, code, held] = step;
    const changes = sync.sync(pressed, event(held), type === 'down' ? 'keydown' : 'keyup', code);
    changes.release.forEach(release);
    changes.press.forEach(press);
    if (type === 'down') {
      if (!pressed.has(code)) press(code);
    } else if (pressed.has(code)) {
      release(code);
    }
  }
  return host;
}

// Brave on Windows, recorded with the W3C key-event-viewer: hold the left
// Shift, press the right Shift, release the left Shift, tap A, release the
// right Shift. Releasing the first Shift of a held pair sends no event, Shift
// is reported as held while either side is, the right Shift's first keydown is
// already flagged as a repeat, and the last release sends one keyup, for the
// right Shift only.
const windowsShiftPair: Step[] = [
  ['down', 'ShiftLeft', { shift: true }],
  ['down', 'ShiftLeft', { shift: true }, 'repeat'],
  ['down', 'ShiftLeft', { shift: true }, 'repeat'],
  ['down', 'ShiftLeft', { shift: true }, 'repeat'],
  ['down', 'ShiftRight', { shift: true }, 'repeat'],
  ['down', 'ShiftRight', { shift: true }, 'repeat'],
  ['down', 'ShiftRight', { shift: true }, 'repeat'],
  ['down', 'ShiftRight', { shift: true }, 'repeat'],
  ['down', 'KeyA', { shift: true }],
  ['up', 'KeyA', { shift: true }],
  ['up', 'ShiftRight', {}]
];

test('the Windows Shift pair trace shifts A and leaves no Shift held', () => {
  assert.deepEqual(replay([...windowsShiftPair, ['down', 'KeyB', {}], ['up', 'KeyB', {}]]), [
    'ShiftLeft down',
    'ShiftRight down',
    'KeyA down',
    'KeyA up',
    'ShiftLeft up',
    'ShiftRight up',
    'KeyB down',
    'KeyB up'
  ]);
});

// The same hands on the target before the re-sync had a memory, where the host
// saw the left Shift released as it was let go and the right Shift released
// just before A. The re-sync of that build skipped the family of a modifier's
// own event and did nothing else, so the left Shift can only have gone with a
// ShiftLeft keyup, and the right Shift, still held, with an A keydown that
// reported Shift as up. The browser of that run did not behave like the one in
// the trace above.
const windowsShiftPairRunA: Step[] = [
  ['down', 'ShiftLeft', { shift: true }],
  ['down', 'ShiftRight', { shift: true }, 'repeat'],
  ['up', 'ShiftLeft', {}],
  ['down', 'KeyA', {}],
  ['up', 'KeyA', {}],
  ['up', 'ShiftRight', {}]
];

test('the earlier re-sync explains what the host saw for the Shift pair', () => {
  const pressed = new Set<string>();
  const host: string[] = [];
  for (const step of windowsShiftPairRunA) {
    if (step[0] === 'blur') continue;
    const [type, code, held] = step;
    for (const stale of staleModifiers(pressed, event(held), code)) {
      pressed.delete(stale);
      host.push(`${stale} up`);
    }
    if (type === 'down' && !pressed.has(code)) {
      pressed.add(code);
      host.push(`${code} down`);
    } else if (type === 'up' && pressed.delete(code)) {
      host.push(`${code} up`);
    }
  }
  assert.deepEqual(host, [
    'ShiftLeft down',
    'ShiftRight down',
    'ShiftLeft up',
    'ShiftRight up',
    'KeyA down',
    'KeyA up'
  ]);
});

test('a Shift keyup that reports Shift as up releases both Shifts', () => {
  // The browser reports A as unshifted, so it reaches the host unshifted.
  assert.deepEqual(replay(windowsShiftPairRunA), [
    'ShiftLeft down',
    'ShiftRight down',
    'ShiftRight up',
    'ShiftLeft up',
    'KeyA down',
    'KeyA up'
  ]);
});

// On Windows each Ctrl sends its own keyup, and the host saw the left Ctrl
// released as it was let go and C typed with the right Ctrl held. Whether the
// left Ctrl's keyup reports Ctrl as held was not recorded, so both are replayed.
for (const ctrlOnKeyUp of [true, false]) {
  test(`the Ctrl pair keeps the right Ctrl for C, ctrlKey ${ctrlOnKeyUp} on the left keyup`, () => {
    const leftRelease = ctrlOnKeyUp
      ? ['ControlLeft up']
      : ['ControlRight up', 'ControlLeft up', 'ControlRight down'];
    assert.deepEqual(
      replay([
        ['down', 'ControlLeft', { ctrl: true }],
        ['down', 'ControlRight', { ctrl: true }, 'repeat'],
        ['up', 'ControlLeft', { ctrl: ctrlOnKeyUp }],
        ['down', 'KeyC', { ctrl: true }],
        ['up', 'KeyC', { ctrl: true }],
        ['up', 'ControlRight', {}],
        ['down', 'KeyD', {}]
      ]),
      [
        'ControlLeft down',
        'ControlRight down',
        ...leftRelease,
        'KeyC down',
        'KeyC up',
        'ControlRight up',
        'KeyD down'
      ]
    );
  });
}

test('a side released with its pair is not pressed again once reported up', () => {
  assert.deepEqual(
    replay([
      ['down', 'ControlLeft', { ctrl: true }],
      ['down', 'ControlRight', { ctrl: true }],
      ['up', 'ControlLeft', {}],
      ['down', 'KeyC', {}],
      ['up', 'KeyC', { ctrl: true }]
    ]),
    [
      'ControlLeft down',
      'ControlRight down',
      'ControlRight up',
      'ControlLeft up',
      'KeyC down',
      'KeyC up'
    ]
  );
});

test('both Shifts held over a key, released right then left', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['down', 'KeyA', { shift: true }],
      ['up', 'KeyA', { shift: true }],
      ['up', 'ShiftRight', { shift: true }],
      ['down', 'KeyB', { shift: true }],
      ['up', 'KeyB', { shift: true }],
      ['up', 'ShiftLeft', {}]
    ]),
    [
      'ShiftLeft down',
      'ShiftRight down',
      'KeyA down',
      'KeyA up',
      'ShiftRight up',
      'KeyB down',
      'KeyB up',
      'ShiftLeft up'
    ]
  );
});

test('a lone Shift whose keyup was lost is still released by the next event', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftRight', { shift: true }],
      ['down', 'KeyA', {}],
      ['up', 'KeyA', {}]
    ]),
    ['ShiftRight down', 'ShiftRight up', 'KeyA down', 'KeyA up']
  );
});

test('a Shift pair whose keyups were lost is released by the next event', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['down', 'KeyA', {}],
      ['up', 'KeyA', {}]
    ]),
    ['ShiftLeft down', 'ShiftRight down', 'ShiftLeft up', 'ShiftRight up', 'KeyA down', 'KeyA up']
  );
});

test('Windows Ctrl+Alt+Del keeps the left Ctrl through the replay', () => {
  const ctrlAlt = { ctrl: true, alt: true, altGraph: true };
  assert.deepEqual(
    replay([
      ['down', 'ControlLeft', { ctrl: true }],
      ['down', 'AltLeft', ctrlAlt],
      ['down', 'Delete', ctrlAlt],
      ['up', 'Delete', ctrlAlt]
    ]),
    ['ControlLeft down', 'AltLeft down', 'Delete down', 'Delete up']
  );
});

test('blur releases every key, and a side released with its pair stays released', () => {
  assert.deepEqual(
    replay([
      ['down', 'ControlLeft', { ctrl: true }],
      ['down', 'ControlRight', { ctrl: true }],
      ['up', 'ControlLeft', {}],
      ['down', 'KeyA', { ctrl: true }],
      ['blur'],
      ['down', 'KeyB', { ctrl: true }]
    ]),
    [
      'ControlLeft down',
      'ControlRight down',
      'ControlRight up',
      'ControlLeft up',
      'ControlRight down',
      'KeyA down',
      'ControlRight up',
      'KeyA up',
      'KeyB down'
    ]
  );
});

test('a composition suppresses only the keys the input method consumes', () => {
  assert.equal(isCompositionKey(false, { isComposing: true, keyCode: 65, key: 'a' }), true);
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 229, key: 'Process' }), true);
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 229, key: 'a' }), true);
  assert.equal(isCompositionKey(false, { isComposing: false, keyCode: 229, key: 'Process' }), false);
});

test('a composition that never ended does not suppress later typing', () => {
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 65, key: 'a' }), false);
});
