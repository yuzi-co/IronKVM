import assert from 'node:assert/strict';
import { test } from 'node:test';

import { emptyForm, isBroadcast, isMac, toRequest, validateForm } from './device-form.ts';

test('isMac accepts the three notations and rejects the rest', () => {
  for (const ok of ['aa:bb:cc:dd:ee:ff', 'AA-BB-CC-DD-EE-FF', 'aabbccddeeff', 'aabb.ccdd.eeff']) {
    assert.equal(isMac(ok), true, ok);
  }
  for (const bad of ['', 'aa:bb', 'gg:bb:cc:dd:ee:ff', 'aabbccddeeff00']) {
    assert.equal(isMac(bad), false, bad);
  }
});

test('isBroadcast takes an address or a subnet, or nothing', () => {
  for (const ok of ['', '192.168.1.255', '192.168.1.0/24']) assert.equal(isBroadcast(ok), true, ok);
  for (const bad of ['192.168.1', '192.168.1.0/31', '192.168.1.0/24/1', '300.1.1.1', 'host']) {
    assert.equal(isBroadcast(bad), false, bad);
  }
});

test('validateForm flags each bad field', () => {
  const errors = validateForm({
    ...emptyForm,
    mac: 'nope',
    broadcast: 'x',
    secureOn: '12',
    ip: '1.2.3'
  });
  assert.deepEqual(Object.keys(errors).sort(), ['broadcast', 'ip', 'mac', 'secureOn']);
  assert.deepEqual(validateForm({ ...emptyForm, mac: 'aa:bb:cc:dd:ee:ff' }), {});
});

test('toRequest trims and leaves oldMac out of an add', () => {
  const req = toRequest({ ...emptyForm, mac: ' aa:bb:cc:dd:ee:ff ', name: ' nas ' });
  assert.equal(req.oldMac, undefined);
  assert.equal(req.mac, 'aa:bb:cc:dd:ee:ff');
  assert.equal(req.name, 'nas');
});
