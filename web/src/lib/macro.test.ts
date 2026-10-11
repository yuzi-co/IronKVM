import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  AUTO_TAP_KEYS,
  AUTO_TAP_PRESETS,
  autoTapSettingsFrom,
  comboLabel,
  DEFAULT_AUTO_TAP,
  durationMs,
  keyLabel,
  MacroRecorder,
  MAX_DELAY_MS,
  stepFromMessage,
  usageOf
} from './macro.ts';

test('a keyboard message becomes the keyboard state', () => {
  assert.deepEqual(stepFromMessage([1, 0x05, 0, 0x4c, 0, 0, 0, 0, 0]), {
    type: 'key',
    modifiers: 0x05,
    keys: [0x4c]
  });
  assert.deepEqual(stepFromMessage([1, 0, 0, 0, 0, 0, 0, 0, 0]), { type: 'key' });
});

test('pointer messages become relative and absolute steps', () => {
  assert.deepEqual(stepFromMessage([2, 1, 0xff, 2, 0xfe, 0]), {
    type: 'mouse',
    buttons: 1,
    x: -1,
    y: 2,
    wheel: -2
  });
  assert.deepEqual(stepFromMessage([2, 0, 0x34, 0x12, 0x00, 0x00, 0]), {
    type: 'mouse',
    absolute: true,
    x: 0x1234,
    y: 0
  });
});

test('heartbeats, keepalives and touch frames are not steps', () => {
  assert.equal(stepFromMessage([0]), null);
  assert.equal(stepFromMessage([4]), null);
  assert.equal(stepFromMessage([3, 1, 2, 3]), null);
});

test('the recorder puts each gap between steps as a pause', () => {
  const r = new MacroRecorder();
  r.message([1, 0, 0, 0x04, 0, 0, 0, 0, 0], 1000);
  r.message([1, 0, 0, 0, 0, 0, 0, 0, 0], 1080);
  r.atx('power', 500, 1300);
  assert.deepEqual(r.finish(), [
    { type: 'key', keys: [0x04] },
    { type: 'delay', ms: 80 },
    { type: 'key' },
    { type: 'delay', ms: 220 },
    { type: 'atx', button: 'power', ms: 500 }
  ]);
});

test('the gap after an ATX press is counted from the release', () => {
  const r = new MacroRecorder();
  r.atx('reset', 800, 0);
  r.message([1, 0, 0, 0x3b, 0, 0, 0, 0, 0], 1000);
  r.message([1, 0, 0, 0, 0, 0, 0, 0, 0], 1100);
  r.atx('power', 800, 1200);
  r.message([1, 0, 0, 0x3b, 0, 0, 0, 0, 0], 1500);
  assert.deepEqual(r.finish().slice(0, 2), [
    { type: 'atx', button: 'reset', ms: 800 },
    { type: 'delay', ms: 200 }
  ]);
  // The key came while the button was still down, so no pause precedes it.
  assert.equal(r.finish()[6].type, 'atx');
  assert.deepEqual(r.finish()[7], { type: 'key', keys: [0x3b] });
});

test('a long gap is split into pauses the server takes', () => {
  const r = new MacroRecorder();
  r.message([1, 0, 0, 0x04, 0, 0, 0, 0, 0], 0);
  r.message([1, 0, 0, 0, 0, 0, 0, 0, 0], MAX_DELAY_MS + 5);
  const steps = r.finish();
  assert.deepEqual(steps[1], { type: 'delay', ms: MAX_DELAY_MS });
  assert.deepEqual(steps[2], { type: 'delay', ms: 5 });
});

test('pointer moves close together are merged, clicks are not', () => {
  const r = new MacroRecorder();
  r.message([2, 0, 10, 0, 10, 0, 0], 0);
  r.message([2, 0, 20, 0, 20, 0, 0], 16);
  r.message([2, 0, 30, 0, 30, 0, 0], 32);
  r.message([2, 1, 30, 0, 30, 0, 0], 40);
  r.message([2, 0, 30, 0, 30, 0, 0], 100);
  assert.deepEqual(r.finish(), [
    { type: 'mouse', absolute: true, x: 30, y: 30 },
    { type: 'delay', ms: 40 },
    { type: 'mouse', absolute: true, buttons: 1, x: 30, y: 30 },
    { type: 'delay', ms: 60 },
    { type: 'mouse', absolute: true, x: 30, y: 30 }
  ]);
});

test('relative moves are summed while they fit in a report', () => {
  const r = new MacroRecorder();
  r.message([2, 0, 100, 0, 0, 0], 0);
  r.message([2, 0, 20, 5, 0, 0], 10);
  r.message([2, 0, 20, 0, 0, 0], 20);
  assert.deepEqual(r.finish(), [
    { type: 'mouse', x: 120, y: 5 },
    { type: 'delay', ms: 20 },
    { type: 'mouse', x: 20 }
  ]);
});

test('finish releases what is still held and drops the trailing pause', () => {
  const r = new MacroRecorder();
  r.message([1, 0x01, 0, 0, 0, 0, 0, 0, 0], 0);
  r.message([2, 1, 0x10, 0, 0x20, 0, 0], 10);
  r.message([2, 1, 0, 0, 0, 0], 20);
  assert.deepEqual(r.finish(), [
    { type: 'key', modifiers: 0x01 },
    { type: 'delay', ms: 10 },
    { type: 'mouse', absolute: true, buttons: 1, x: 0x10, y: 0x20 },
    { type: 'delay', ms: 10 },
    { type: 'mouse', buttons: 1 },
    { type: 'key' },
    { type: 'mouse' },
    { type: 'mouse', absolute: true, x: 0x10, y: 0x20 }
  ]);
});

test('the recorder stops at the step limit and keeps room for the releases', () => {
  const r = new MacroRecorder(10);
  let accepted = 0;
  for (let i = 0; i < 20; i++) {
    if (r.message([1, 0, 0, i % 2 ? 0 : 0x04, 0, 0, 0, 0, 0], i * 100)) accepted++;
  }
  assert.equal(r.full, true);
  assert.ok(r.finish().length <= 10);
  assert.ok(accepted < 20);
});

test('durationMs counts pauses and ATX presses', () => {
  assert.equal(
    durationMs([{ type: 'delay', ms: 100 }, { type: 'atx', button: 'reset' }, { type: 'key' }]),
    900
  );
});

test('labels name keys as their caps do', () => {
  assert.equal(keyLabel(0x04), 'A');
  assert.equal(keyLabel(0x1e), '1');
  assert.equal(keyLabel(0x4c), 'Del');
  assert.equal(keyLabel(0x3b), 'F2');
  assert.equal(keyLabel(0x03), '0x03');
  assert.equal(comboLabel(0x05, [0x4c]), 'Ctrl+Alt+Del');
  assert.equal(comboLabel(), '');
  assert.equal(usageOf('F12'), 0x45);
});

test('stored auto-tap settings keep only what the server takes', () => {
  assert.deepEqual(autoTapSettingsFrom(null), DEFAULT_AUTO_TAP);
  assert.deepEqual(
    autoTapSettingsFrom({
      code: 'Delete',
      interval: 10,
      timeout: 9999,
      firstPress: 'reset',
      stopOnInput: false,
      stopOnVideo: 'yes'
    }),
    {
      code: 'Delete',
      interval: 50,
      timeout: 600,
      firstPress: 'reset',
      stopOnInput: false,
      stopOnVideo: false
    }
  );
  assert.equal(autoTapSettingsFrom({ code: 'NoSuchKey' }).code, 'F2');
  for (const code of AUTO_TAP_PRESETS) assert.ok(AUTO_TAP_KEYS.includes(code), code);
});
