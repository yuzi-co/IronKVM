import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isValidHostname } from './hostname.ts';

test('accepts RFC 1123 host names', () => {
  for (const name of ['nanokvm', 'kvm-1', 'a', 'KVM', 'rack1.lab', '1kvm', 'a'.repeat(63)]) {
    assert.equal(isValidHostname(name), true, name);
  }
});

test('refuses anything else', () => {
  for (const name of [
    '',
    '-kvm',
    'kvm-',
    'kvm_1',
    'kvm 1',
    'kvm\n',
    'kvm/1',
    'kvm.',
    '.kvm',
    'a..b',
    'a'.repeat(64),
    `${'a'.repeat(32)}.${'a'.repeat(32)}`
  ]) {
    assert.equal(isValidHostname(name), false, JSON.stringify(name));
  }
});
