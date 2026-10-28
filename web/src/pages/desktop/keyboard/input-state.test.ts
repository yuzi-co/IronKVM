import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isCompositionKey, staleModifiers } from './input-state.ts';

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

test('a composition suppresses only the keys the input method consumes', () => {
  assert.equal(isCompositionKey(false, { isComposing: true, keyCode: 65, key: 'a' }), true);
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 229, key: 'Process' }), true);
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 229, key: 'a' }), true);
  assert.equal(isCompositionKey(false, { isComposing: false, keyCode: 229, key: 'Process' }), false);
});

test('a composition that never ended does not suppress later typing', () => {
  assert.equal(isCompositionKey(true, { isComposing: false, keyCode: 65, key: 'a' }), false);
});
