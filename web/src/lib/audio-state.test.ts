import assert from 'node:assert/strict';
import { test } from 'node:test';

import { audioStateFromByte, audioStateFromName } from './audio-state.ts';

test('a direct state byte maps to the server state it names', () => {
  assert.equal(audioStateFromByte(0), 'unknown');
  assert.equal(audioStateFromByte(1), 'playing');
  assert.equal(audioStateFromByte(2), 'idle');
  assert.equal(audioStateFromByte(3), 'failing');
});

test('a state byte from a newer server is unknown, not a crash', () => {
  assert.equal(audioStateFromByte(4), 'unknown');
  assert.equal(audioStateFromByte(255), 'unknown');
});

test('a WebRTC state name maps to itself, and anything else to unknown', () => {
  assert.equal(audioStateFromName('idle'), 'idle');
  assert.equal(audioStateFromName('playing'), 'playing');
  assert.equal(audioStateFromName('failing'), 'failing');
  assert.equal(audioStateFromName('muted'), 'unknown');
  assert.equal(audioStateFromName(undefined), 'unknown');
  assert.equal(audioStateFromName(2), 'unknown');
});
