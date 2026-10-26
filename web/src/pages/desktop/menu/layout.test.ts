import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  NARROW_ITEM_SIZE,
  NARROW_MENU_MAX_WIDTH,
  narrowMenuPriority,
  narrowMenuSlots,
  splitNarrowMenu
} from './layout.ts';

test('the narrow bar holds more entries as the phone gets wider', () => {
  assert.equal(narrowMenuSlots(320), 4);
  assert.equal(narrowMenuSlots(360), 5);
  assert.equal(narrowMenuSlots(412), 6);
  assert.equal(narrowMenuSlots(NARROW_MENU_MAX_WIDTH), 12);
});

test('the bar plus both indicators fits the viewport it was sized for', () => {
  for (const width of [280, 320, 360, 375, 390, 412, 480, 600, NARROW_MENU_MAX_WIDTH]) {
    const entries = narrowMenuSlots(width) + 2; // Settings and More
    const bar = 12 + 17 + entries * NARROW_ITEM_SIZE + 30 + 34;
    assert.ok(bar + 8 <= width, `${width} px: bar ${bar} px`);
  }
});

test('a width too small or not a number leaves only Settings and More', () => {
  assert.equal(narrowMenuSlots(100), 0);
  assert.equal(narrowMenuSlots(Number.NaN), 0);
});

test('a 320 px phone keeps the keyboard, mouse and power on the bar', () => {
  const { bar, more } = splitNarrowMenu(narrowMenuPriority, narrowMenuSlots(320));
  assert.deepEqual(bar, ['virtualKeyboard', 'keyboard', 'mouse', 'power']);
  assert.equal(more[0], 'media');
  assert.equal(bar.length + more.length, narrowMenuPriority.length);
});

test('entries the account cannot use leave their slot to the next one', () => {
  // A view-only account has no keyboard, mouse or power.
  const { bar, more } = splitNarrowMenu(['fullscreen', 'screen', 'speaker'], 2);
  assert.deepEqual(bar, ['screen', 'speaker']);
  assert.deepEqual(more, ['fullscreen']);
});

test('negative or fractional slots are read as whole slots from zero', () => {
  assert.deepEqual(splitNarrowMenu(['mouse'], -1), { bar: [], more: ['mouse'] });
  assert.deepEqual(splitNarrowMenu(['mouse', 'power'], 1.9), { bar: ['mouse'], more: ['power'] });
});
