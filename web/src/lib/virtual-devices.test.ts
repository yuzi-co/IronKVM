import assert from 'node:assert/strict';
import { test } from 'node:test';

import { readVirtualDevices } from './virtual-devices.ts';

const on = { enabled: true, active: true, cost: 2 };
const off = { enabled: false, active: false, cost: 1 };

test('reads server API 2, states under functions', () => {
  const devices = readVirtualDevices({
    network: false,
    disk: true,
    audio: false,
    functions: { console: off, network: off, disk: on, audio: off },
    used: 4,
    total: 6,
    fits: [['console', 'disk']]
  });
  assert.deepEqual(devices.disk, on);
  assert.deepEqual(devices.network, off);
  assert.equal(devices.used, 4);
  assert.equal(devices.total, 6);
  assert.deepEqual(devices.fits, [['console', 'disk']]);
  assert.equal('functions' in devices, false);
});

test('reads server API 1, states at the top level', () => {
  const devices = readVirtualDevices({
    console: on,
    network: off,
    disk: on,
    audio: off,
    used: 5,
    total: 6,
    fits: []
  });
  assert.deepEqual(devices.console, on);
  assert.deepEqual(devices.disk, on);
  assert.equal(devices.used, 5);
});

test('keeps the fields of an apply answer', () => {
  const usbNetwork = { mode: 'ncm' };
  const devices = readVirtualDevices({
    network: true,
    disk: false,
    audio: false,
    functions: { console: off, network: on, disk: off, audio: off },
    used: 3,
    total: 6,
    fits: [],
    usbNetwork
  });
  assert.equal(devices.usbNetwork, usbNetwork);
  assert.deepEqual(devices.network, on);
});

test('a missing state reads as off', () => {
  const devices = readVirtualDevices({ functions: {} });
  assert.deepEqual(devices.audio, { enabled: false, active: false, cost: 0 });
  assert.deepEqual(devices.fits, []);
});
