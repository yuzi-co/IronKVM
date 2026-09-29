import assert from 'node:assert/strict';
import { test } from 'node:test';

import type { UsbNetwork, VirtualDevices } from '../api/virtual-device.ts';
import {
  applyRequest,
  boardDraft,
  changedDevices,
  draftUsed,
  fitsInDraft,
  subnetValid
} from './usb-draft.ts';

// The shipped board: HID (3) + console (2) + disk (1) + speaker (0) = 6 of 6.
function shipped(): VirtualDevices {
  return {
    console: { enabled: true, active: true, cost: 2 },
    disk: { enabled: true, active: true, cost: 1 },
    network: { enabled: false, active: false, cost: 2 },
    audio: { enabled: true, active: true, cost: 0 },
    used: 6,
    total: 6,
    fits: []
  };
}

const linkOff: UsbNetwork = {
  mode: 'off',
  subnet: '172.31.255.0/30',
  board: '172.31.255.1',
  host: '172.31.255.2',
  active: false,
  fits: false,
  refusal: ''
};

test('the board draft changes nothing', () => {
  const board = boardDraft(shipped(), linkOff);
  assert.deepEqual(changedDevices(board, board), []);
  assert.equal(draftUsed(board, shipped()), 6);
});

test('the network does not fit beside the console, and does once it is off', () => {
  const devices = shipped();
  const board = boardDraft(devices, linkOff);

  assert.equal(fitsInDraft(board, devices, 'network'), false);

  const draft = { ...board, console: false };
  assert.equal(draftUsed(draft, devices), 4);
  assert.equal(fitsInDraft(draft, devices, 'network'), true);

  const swapped = { ...draft, network: 'ncm' as const };
  assert.equal(draftUsed(swapped, devices), 6);
  assert.deepEqual(changedDevices(swapped, board), ['console', 'network']);
  // With the network on in the draft, the console no longer fits back.
  assert.equal(fitsInDraft(swapped, devices, 'console'), false);
});

test('a device on in the draft always fits, even on a board over its budget', () => {
  const devices = { ...shipped(), network: { enabled: true, active: false, cost: 2 }, used: 8 };
  const board = boardDraft(devices, { ...linkOff, mode: 'ncm' });

  assert.equal(fitsInDraft(board, devices, 'console'), true);
  assert.equal(fitsInDraft({ ...board, disk: false }, devices, 'disk'), false);
});

test('the HID cost comes from the server count', () => {
  // HID switched off on the board: the server counts only the devices.
  const devices = { ...shipped(), used: 3 };
  const draft = { ...boardDraft(devices, linkOff), network: 'ecm' as const };

  assert.equal(draftUsed(draft, devices), 5);
  assert.equal(fitsInDraft(boardDraft(devices, linkOff), devices, 'network'), true);
});

test('the subnet counts as a change only while the link stays on', () => {
  const on = boardDraft(shipped(), { ...linkOff, mode: 'ncm' });
  assert.deepEqual(changedDevices({ ...on, subnet: ' 10.9.9.0/30 ' }, on), ['network']);
  assert.deepEqual(changedDevices({ ...on, subnet: '172.31.255.0/30 ' }, on), []);

  const off = boardDraft(shipped(), linkOff);
  assert.deepEqual(changedDevices({ ...off, subnet: '10.9.9.0/30' }, off), []);
});

test('an invalid subnet gates Apply only with the link on', () => {
  const draft = boardDraft(shipped(), linkOff);
  assert.equal(subnetValid({ ...draft, subnet: 'nope' }), true);
  assert.equal(subnetValid({ ...draft, network: 'ncm', subnet: 'nope' }), false);
  assert.equal(subnetValid({ ...draft, network: 'ncm', subnet: ' 10.0.0.0/30 ' }), true);
});

test('the request carries the whole set and a subnet only with the link on', () => {
  const draft = boardDraft(shipped(), linkOff);
  assert.deepEqual(applyRequest({ ...draft, console: false, subnet: 'x' }), {
    console: false,
    disk: true,
    audio: true,
    network: { mode: 'off', subnet: '' }
  });
  assert.deepEqual(applyRequest({ ...draft, network: 'ecm', subnet: ' 10.0.0.0/30 ' }).network, {
    mode: 'ecm',
    subnet: '10.0.0.0/30'
  });
});
