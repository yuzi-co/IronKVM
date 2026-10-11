import assert from 'node:assert/strict';
import { test } from 'node:test';

import { parseUdpPort } from './webrtc-port.ts';

test('an empty field turns the port off', () => {
  assert.equal(parseUdpPort(''), 0);
  assert.equal(parseUdpPort('  '), 0);
});

test('accepts ports from 1024 to 65535', () => {
  for (const [text, port] of [
    ['1024', 1024],
    ['58000', 58000],
    [' 65535 ', 65535]
  ] as const) {
    assert.equal(parseUdpPort(text), port, text);
  }
});

test('refuses everything else', () => {
  for (const text of ['0', '80', '1023', '65536', '-1', '5.8', 'abc', '580000', '0x1000']) {
    assert.equal(parseUdpPort(text), null, text);
  }
});
