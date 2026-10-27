import assert from 'node:assert/strict';
import { test } from 'node:test';

import type { InstallProgress, SlotStatus } from '@/api/slot.ts';

import {
  armedTrial,
  installTargets,
  isHttpsUrl,
  onTrial,
  progressPercent,
  revertTarget,
  tryTargets
} from './slot-install.ts';

const progress = (p: Partial<InstallProgress>): InstallProgress => ({
  running: true,
  phase: 'write',
  done: 0,
  total: 0,
  ...p
});

const status = (p: Partial<SlotStatus>): SlotStatus => ({
  running: 'a',
  trusted: 'a',
  trial: 'none',
  prev: 'none',
  slots: [
    { name: 'a', state: 'unknown' },
    { name: 'b', state: 'installed' },
    { name: 'recovery', state: 'installed' }
  ],
  ...p
});

test('the bar fills over the write and again over the read back', () => {
  assert.equal(progressPercent(progress({ phase: 'verify' })), 0);
  assert.equal(progressPercent(progress({ phase: 'write', done: 50, total: 100 })), 25);
  assert.equal(progressPercent(progress({ phase: 'write', done: 100, total: 100 })), 50);
  assert.equal(progressPercent(progress({ phase: 'readback', done: 0, total: 100 })), 50);
  assert.equal(progressPercent(progress({ phase: 'readback', done: 100, total: 100 })), 100);
  assert.equal(progressPercent(progress({ phase: 'done', done: 0, total: 0 })), 100);
  assert.equal(progressPercent(progress({ phase: 'write', done: 200, total: 100 })), 50);
});

test('an image is never written to the running slot or to one being written', () => {
  assert.deepEqual(installTargets(status({})), ['b', 'recovery']);
  const busy = status({
    slots: [
      { name: 'a', state: 'installed' },
      { name: 'b', state: 'installing' }
    ]
  });
  assert.deepEqual(installTargets(busy), []);
});

test('a trial is offered for a whole image on the other slot, once', () => {
  assert.deepEqual(tryTargets(status({})), ['b']);
  assert.deepEqual(tryTargets(status({ trial: 'b' })), []);
  for (const state of ['bad', 'interrupted', 'installing']) {
    const st = status({
      slots: [
        { name: 'a', state: 'installed' },
        { name: 'b', state }
      ]
    });
    assert.deepEqual(tryTargets(st), [], state);
  }
  // Running b, the trial is of a, which has no record and is tried as before.
  assert.deepEqual(tryTargets(status({ running: 'b', trusted: 'b' })), ['a']);
});

test('the armed trial comes from the backend, or from the record without one', () => {
  assert.equal(armedTrial(status({ trial: 'b' })), 'b');
  assert.equal(armedTrial(status({})), '');
  const recorded = status({
    trial: undefined,
    slots: [{ name: 'b', state: 'installed', trial: 'armed' }]
  });
  assert.equal(armedTrial(recorded), 'b');
});

test('confirm is offered on a trial boot and revert while a previous slot is kept', () => {
  assert.equal(onTrial(status({})), false);
  assert.equal(onTrial(status({ running: 'b' })), true);
  assert.equal(onTrial(status({ running: 'recovery' })), false);
  assert.equal(revertTarget(status({})), '');
  assert.equal(revertTarget(status({ trusted: 'b', prev: 'a' })), 'a');
  assert.equal(revertTarget(status({ trusted: 'a', prev: 'a' })), '');
});

test('only an https URL is taken', () => {
  assert.equal(isHttpsUrl('https://example.com/slot-b.img.zst'), true);
  assert.equal(isHttpsUrl('http://example.com/slot-b.img.zst'), false);
  assert.equal(isHttpsUrl('https://example.com/a b'), false);
  assert.equal(isHttpsUrl('/data/slot-b.img.zst'), false);
  assert.equal(isHttpsUrl(''), false);
});
