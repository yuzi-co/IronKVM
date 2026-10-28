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

type Held = Parameters<typeof event>[0];
type Step = ['down' | 'up', string, Held] | ['blur'];

// Replays browser events the way the keyboard component handles them and
// returns what the host is sent: each key event first releases the stale
// modifiers, then is passed on.
function replay(steps: Step[]): string[] {
  const sync = new ModifierSync();
  const pressed = new Set<string>();
  const host: string[] = [];
  const release = (code: string) => {
    pressed.delete(code);
    host.push(`${code} up`);
  };

  for (const step of steps) {
    if (step[0] === 'blur') {
      pressed.forEach((code) => host.push(`${code} up`));
      pressed.clear();
      sync.reset();
      continue;
    }
    const [type, code, held] = step;
    for (const stale of sync.stale(pressed, event(held), code)) release(stale);
    if (type === 'down') {
      if (pressed.has(code)) continue;
      sync.keyDown(code);
      pressed.add(code);
      host.push(`${code} down`);
    } else {
      sync.keyUp(code, pressed);
      if (pressed.has(code)) release(code);
    }
  }
  return host;
}

// Windows keeps one Shift state for both keys, so after the left Shift is
// released the browser reports Shift as up while the right one is still held.
test('a Shift held after the other Shift is released keeps shifting', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['up', 'ShiftLeft', {}],
      ['down', 'KeyA', {}],
      ['up', 'KeyA', {}],
      ['up', 'ShiftRight', {}]
    ]),
    ['ShiftLeft down', 'ShiftRight down', 'ShiftLeft up', 'KeyA down', 'KeyA up', 'ShiftRight up']
  );
});

test('both Shifts held over a key, released right then left', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['down', 'KeyA', { shift: true }],
      ['up', 'KeyA', { shift: true }],
      ['up', 'ShiftRight', {}],
      ['up', 'ShiftLeft', {}]
    ]),
    ['ShiftLeft down', 'ShiftRight down', 'KeyA down', 'KeyA up', 'ShiftRight up', 'ShiftLeft up']
  );
});

for (const [article, name, left, right, flag] of [
  ['a', 'Ctrl', 'ControlLeft', 'ControlRight', 'ctrl'],
  ['an', 'Alt', 'AltLeft', 'AltRight', 'alt']
] as const) {
  test(`${article} ${name} held after the other ${name} is released stays held`, () => {
    assert.deepEqual(
      replay([
        ['down', left, { [flag]: true }],
        ['down', right, { [flag]: true }],
        ['up', left, {}],
        ['down', 'KeyC', {}],
        ['up', 'KeyC', {}],
        ['up', right, {}]
      ]),
      [`${left} down`, `${right} down`, `${left} up`, 'KeyC down', 'KeyC up', `${right} up`]
    );
  });

  test(`both ${name}s held over a key, released right then left`, () => {
    assert.deepEqual(
      replay([
        ['down', left, { [flag]: true }],
        ['down', right, { [flag]: true }],
        ['down', 'KeyC', { [flag]: true }],
        ['up', 'KeyC', { [flag]: true }],
        ['up', right, {}],
        ['up', left, {}]
      ]),
      [`${left} down`, `${right} down`, 'KeyC down', 'KeyC up', `${right} up`, `${left} up`]
    );
  });
}

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

test('an unconfirmed Shift is re-synced again once the browser reports it held', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['up', 'ShiftLeft', {}],
      ['down', 'KeyA', { shift: true }],
      ['up', 'KeyA', { shift: true }],
      ['down', 'KeyB', {}]
    ]),
    [
      'ShiftLeft down',
      'ShiftRight down',
      'ShiftLeft up',
      'KeyA down',
      'KeyA up',
      'ShiftRight up',
      'KeyB down'
    ]
  );
});

test('blur releases every key, an unconfirmed Shift included', () => {
  assert.deepEqual(
    replay([
      ['down', 'ShiftLeft', { shift: true }],
      ['down', 'ShiftRight', { shift: true }],
      ['up', 'ShiftLeft', {}],
      ['down', 'KeyA', {}],
      ['blur'],
      ['down', 'ShiftRight', { shift: true }],
      ['down', 'KeyB', {}]
    ]),
    [
      'ShiftLeft down',
      'ShiftRight down',
      'ShiftLeft up',
      'KeyA down',
      'ShiftRight up',
      'KeyA up',
      'ShiftRight down',
      'ShiftRight up',
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
