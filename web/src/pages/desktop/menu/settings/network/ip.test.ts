import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isValidIPv4 } from './ip.ts';

test('accepts dotted quads', () => {
  for (const value of ['0.0.0.0', '10.0.0.1', '192.168.1.254', '255.255.255.255']) {
    assert.equal(isValidIPv4(value), true, value);
  }
});

test('refuses anything else', () => {
  for (const value of [
    '',
    '10.0.0',
    '10.0.0.1.1',
    '256.0.0.1',
    '010.0.0.1',
    '1.2.3.a',
    ' 1.2.3.4'
  ]) {
    assert.equal(isValidIPv4(value), false, value);
  }
});
