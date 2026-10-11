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

test('accepts IPv6 addresses in every notation', async () => {
  const { isValidIPv6 } = await import('./ip.ts');
  for (const value of ['2001:db8::1', 'fd00:0:0:1::50', '::1', 'fe80::1', '::ffff:192.168.1.1']) {
    assert.equal(isValidIPv6(value), true, value);
  }
});

test('refuses what is not a bare IPv6 address', async () => {
  const { isValidIPv6 } = await import('./ip.ts');
  for (const value of [
    '',
    '192.168.1.1',
    'fe80::1%eth0',
    '[2001:db8::1]',
    '2001:db8::1/64',
    '2001:db8::g',
    '2001:db8:::1',
    ' 2001:db8::1',
    'example.com'
  ]) {
    assert.equal(isValidIPv6(value), false, value);
  }
});
