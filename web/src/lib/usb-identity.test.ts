import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  identityErrors,
  normalizeUsbId,
  sameIdentity,
  usbIdValid,
  usbSerialValid,
  usbStringValid,
  type UsbIdentity
} from './usb-identity.ts';

test('usbIdValid takes what the server takes', () => {
  for (const ok of ['0x046d', '046D', '0XC31C', '1', ' ffff ']) {
    assert.equal(usbIdValid(ok), true, ok);
  }
  for (const bad of ['', '0x', '0x0000', '0', '10000', '0x1g00', '-1', '0x 12']) {
    assert.equal(usbIdValid(bad), false, bad);
  }
  assert.equal(normalizeUsbId('46D'), '0x046d');
  assert.equal(normalizeUsbId('0XC31C'), '0xc31c');
});

test('strings and serials', () => {
  assert.equal(usbStringValid('Dell KB216 Wired Keyboard'), true);
  for (const bad of ['', ' lead', 'trail ', 'tab\there', 'café', 'a'.repeat(65)]) {
    assert.equal(usbStringValid(bad), false, bad);
  }
  assert.equal(usbSerialValid(''), true);
  assert.equal(usbSerialValid('0123456789AB'), true);
  assert.equal(usbSerialValid('0123456789A'), false);
  assert.equal(usbSerialValid('0123456789AB-'), false);
});

test('a named preset checks the serial alone', () => {
  const blank: UsbIdentity = {
    vendorId: '',
    productId: '',
    manufacturer: '',
    product: '',
    serial: ''
  };
  assert.deepEqual(identityErrors('logitech', blank), {});
  assert.deepEqual(identityErrors('dell', { ...blank, serial: 'x' }), { serial: true });
  assert.deepEqual(identityErrors('custom', blank), {
    vendorId: true,
    productId: true,
    manufacturer: true,
    product: true
  });
  assert.deepEqual(
    identityErrors('custom', {
      vendorId: '1234',
      productId: '0xabcd',
      manufacturer: 'Acme',
      product: 'KVM',
      serial: ''
    }),
    {}
  );
});

test('sameIdentity compares IDs in any spelling', () => {
  const a: UsbIdentity = {
    vendorId: '0x046d',
    productId: '0xc31c',
    manufacturer: 'Logitech',
    product: 'USB Keyboard',
    serial: ''
  };
  assert.equal(sameIdentity(a, { ...a, vendorId: '46D' }), true);
  assert.equal(sameIdentity(a, { ...a, product: 'USB keyboard' }), false);
});
