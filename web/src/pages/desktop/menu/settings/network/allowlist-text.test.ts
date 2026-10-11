import assert from 'node:assert/strict';
import { test } from 'node:test';

import { formatNetworks, parseNetworks, sameNetworks, withClient } from './allowlist-text.ts';

test('splits on new lines, commas and spaces', () => {
  assert.deepEqual(parseNetworks(' 10.0.0.0/8\n192.168.1.7, 2001:db8::/32  \n\n'), [
    '10.0.0.0/8',
    '192.168.1.7',
    '2001:db8::/32'
  ]);
  assert.deepEqual(parseNetworks('   '), []);
});

test('formats one entry a line', () => {
  assert.equal(formatNetworks(['10.0.0.0/8', '192.168.1.7']), '10.0.0.0/8\n192.168.1.7');
  assert.equal(formatNetworks([]), '');
});

test('compares the text with the saved list', () => {
  assert.equal(sameNetworks('10.0.0.0/8\n192.168.1.7\n', ['10.0.0.0/8', '192.168.1.7']), true);
  assert.equal(sameNetworks('192.168.1.7\n10.0.0.0/8', ['10.0.0.0/8', '192.168.1.7']), false);
  assert.equal(sameNetworks('', []), true);
});

test('adds the caller once', () => {
  assert.equal(withClient('', '192.168.1.20'), '192.168.1.20');
  assert.equal(withClient('10.0.0.0/8\n', '192.168.1.20'), '10.0.0.0/8\n192.168.1.20');
  assert.equal(withClient('192.168.1.20', '192.168.1.20'), '192.168.1.20');
  assert.equal(withClient('10.0.0.0/8', ''), '10.0.0.0/8');
});
