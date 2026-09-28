import assert from 'node:assert/strict';
import { test } from 'node:test';

import { isValidSerialPort, validatePicocomParameters } from './validater.ts';

test('accepts ordinary serial devices', () => {
  for (const port of ['/dev/ttyS1', '/dev/ttyUSB0', '/dev/ttyACM0', '/dev/tty.usbserial-1_0']) {
    assert.equal(isValidSerialPort(port), true, port);
  }
});

test('accepts the stable names under /dev/serial', () => {
  assert.equal(isValidSerialPort('/dev/serial/by-id/usb-FTDI_FT232R-if00-port0'), true);
  assert.equal(
    isValidSerialPort('/dev/serial/by-path/platform-xhci-hcd.0.auto-usb-0:1:1.0-port0'),
    true
  );
});

// '.-_' in a character class is the range from '.' to '_', which let these
// through before the hyphen was moved.
test('rejects characters the old range let through', () => {
  for (const port of ['/dev/tty[0]', '/dev/tty^S', '/dev/tty@1', '/dev/tty=1', '/dev/tty;1']) {
    assert.equal(isValidSerialPort(port), false, port);
  }
});

test('rejects paths outside /dev and traversal', () => {
  for (const port of ['', 'ttyS1', '/tmp/x', '/dev/../etc/passwd', '/dev/tty S1']) {
    assert.equal(isValidSerialPort(port), false, port);
  }
});

test('checks the line settings', () => {
  const ok = {
    port: '/dev/ttyS1',
    baud: '115200',
    parity: 'none',
    flowControl: 'none',
    dataBits: '8',
    stopBits: '1'
  };
  assert.equal(validatePicocomParameters(ok), true);
  assert.equal(validatePicocomParameters({ ...ok, baud: '12345' }), false);
  assert.equal(validatePicocomParameters({ ...ok, dataBits: '9' }), false);
});
