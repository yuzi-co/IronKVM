import assert from 'node:assert/strict';
import { test } from 'node:test';

import type { PowerConfirmation } from '../../../api/power-confirm.ts';
import { pressKeys, secondsLeft, visiblePresses } from './model.ts';

const now = Date.parse('2026-10-27T12:00:00Z');

function press(over: Partial<PowerConfirmation>): PowerConfirmation {
  return {
    id: 'c1',
    sessionId: 's1',
    button: 'power',
    press: 'short',
    state: 'pending',
    createdAt: '2026-10-27T11:59:30Z',
    expiresAt: '2026-10-27T12:00:30Z',
    ...over
  };
}

test('the time left rounds up and stops at zero', () => {
  assert.equal(secondsLeft('2026-10-27T12:00:30Z', now), 30);
  assert.equal(secondsLeft('2026-10-27T12:00:00.200Z', now), 1);
  assert.equal(secondsLeft('2026-10-27T11:59:00Z', now), 0);
  assert.equal(secondsLeft('not a time', now), 0);
});

test('a press is named with the power menu strings', () => {
  assert.deepEqual(pressKeys({ button: 'power', press: 'short' }), {
    label: 'power.powerShort',
    desc: 'power.powerShortDesc'
  });
  assert.deepEqual(pressKeys({ button: 'power', press: 'long' }), {
    label: 'power.powerLong',
    desc: 'power.powerLongDesc'
  });
  assert.deepEqual(pressKeys({ button: 'reset', press: 'short' }), {
    label: 'power.reset',
    desc: 'power.resetDesc'
  });
});

test('only pending presses with time left are shown', () => {
  const list = [
    press({ id: 'live' }),
    press({ id: 'late', expiresAt: '2026-10-27T11:59:59Z' }),
    press({ id: 'pressed', state: 'pressed' })
  ];
  assert.deepEqual(
    visiblePresses(list, now).map((c) => c.id),
    ['live']
  );
});
