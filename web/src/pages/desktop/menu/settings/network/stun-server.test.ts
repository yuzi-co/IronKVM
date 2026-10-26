import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isStunServer, STUN_DISABLED, stunToSave } from './stun-server.ts';

test('accepts host:port with a port from 1 to 65535', () => {
  for (const text of [
    'stun.l.google.com:19302',
    ' stun.example.com:3478 ',
    'stun:1',
    '192.0.2.10:3478',
    '[2001:db8::1]:3478',
    'stun.example.com:65535'
  ]) {
    assert.equal(isStunServer(text), true, text);
  }
});

test('refuses what is not host:port', () => {
  for (const text of [
    '',
    'stun.example.com',
    'stun.example.com:',
    'stun.example.com:0',
    'stun.example.com:65536',
    'stun.example.com:+1',
    ':3478',
    'user:pass@stun.example.com:3478',
    'stun:stun.example.com:3478',
    'stun.example.com:3478/x',
    'stun example.com:3478',
    'bad_host!:3478',
    '2001:db8::1:3478',
    STUN_DISABLED
  ]) {
    assert.equal(isStunServer(text), false, text);
  }
});

test('off saves "disable", on saves the trimmed field', () => {
  assert.equal(stunToSave(false, 'stun.example.com:3478'), STUN_DISABLED);
  assert.equal(stunToSave(true, ' stun.example.com:3478 '), 'stun.example.com:3478');
});
