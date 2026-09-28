import assert from 'node:assert/strict';
import { test } from 'node:test';

import { wifiCredentialsError } from './wifi.ts';

test('an open network needs no password', () => {
  assert.equal(wifiCredentialsError('cafe', ''), '');
});

test('a passphrase is 8 to 63 characters', () => {
  assert.equal(wifiCredentialsError('home', '12345678'), '');
  assert.equal(wifiCredentialsError('home', 'a'.repeat(63)), '');
  assert.equal(wifiCredentialsError('home', 'short'), 'password');
  assert.equal(wifiCredentialsError('home', 'a'.repeat(64)), 'password');
});

test('the network name is required and at most 32 characters', () => {
  assert.equal(wifiCredentialsError('', '12345678'), 'ssid');
  assert.equal(wifiCredentialsError('s'.repeat(33), ''), 'ssid');
});
