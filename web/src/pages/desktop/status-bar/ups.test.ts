import assert from 'node:assert/strict';
import { test } from 'node:test';

import type { NutState } from '@/api/nut.ts';

import { upsLine } from './ups.ts';

function state(over: Partial<NutState>): NutState {
  return {
    enabled: true,
    status: 'ok',
    reachable: true,
    ups: 'online',
    loggedIn: false,
    host: 'on',
    phase: 'idle',
    events: [],
    ...over
  };
}

test('shows nothing before the first answer, when off, or when not set up', () => {
  assert.equal(upsLine(null), null);
  assert.equal(upsLine(state({ enabled: false, status: 'off' })), null);
  assert.equal(upsLine(state({ status: 'unconfigured' })), null);
});

test('shows the state, the charge and the runtime in minutes', () => {
  const line = upsLine(state({ ups: 'onBattery', charge: 87, runtimeSeconds: 905 }));
  assert.deepEqual(line, { state: 'onBattery', tone: 'warning', charge: 87, runtimeMinutes: 15 });
});

test('a UPS that does not report charge or runtime shows only its state', () => {
  const line = upsLine(state({}));
  assert.deepEqual(line, { state: 'online', tone: 'ok', charge: null, runtimeMinutes: null });
});

test('low battery and a forced shutdown are errors', () => {
  assert.equal(upsLine(state({ ups: 'lowBattery' }))?.tone, 'error');
  assert.equal(upsLine(state({ ups: 'shutdown' }))?.tone, 'error');
});

test('an unreachable server shows as an unknown UPS, without the old numbers', () => {
  const line = upsLine(
    state({ status: 'unreachable', ups: 'online', charge: 90, runtimeSeconds: 600 })
  );
  assert.deepEqual(line, { state: 'unknown', tone: 'off', charge: null, runtimeMinutes: null });
});
