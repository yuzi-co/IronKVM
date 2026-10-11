import assert from 'node:assert/strict';
import { test } from 'node:test';

import { HoldKeepalive, heldBy, KEEPALIVE_MESSAGE, type KeepaliveTimers } from './hold-keepalive.ts';

function fakeTimers() {
  let tick: (() => void) | null = null;
  let started = 0;
  const timers: KeepaliveTimers = {
    setInterval: (cb) => {
      started++;
      tick = cb;
      return started;
    },
    clearInterval: () => {
      tick = null;
    }
  };
  return { timers, tick: () => tick?.(), running: () => tick !== null, started: () => started };
}

const shiftDown = [1, 0x02, 0, 0, 0, 0, 0, 0, 0];
const aDown = [1, 0, 0, 0x04, 0, 0, 0, 0, 0];
const keysUp = [1, 0, 0, 0, 0, 0, 0, 0, 0];
const relativeLeftDown = [2, 1, 0, 0, 0, 0];
const relativeUp = [2, 0, 5, 0, 0, 0];
const absoluteLeftDown = [2, 1, 0x34, 0x12, 0x78, 0x56, 0];
const absoluteUp = [2, 0, 0x34, 0x12, 0x78, 0x56, 0];

test('heldBy reads the keyboard and both mouse layouts', () => {
  assert.deepEqual(heldBy(shiftDown), { kind: 'keyboard', held: true });
  assert.deepEqual(heldBy(aDown), { kind: 'keyboard', held: true });
  assert.deepEqual(heldBy(keysUp), { kind: 'keyboard', held: false });
  assert.deepEqual(heldBy(relativeLeftDown), { kind: 'relative', held: true });
  assert.deepEqual(heldBy(relativeUp), { kind: 'relative', held: false });
  assert.deepEqual(heldBy(absoluteLeftDown), { kind: 'absolute', held: true });
  assert.deepEqual(heldBy(absoluteUp), { kind: 'absolute', held: false });
  assert.equal(heldBy([0]), null);
  assert.equal(heldBy([KEEPALIVE_MESSAGE]), null);
  assert.equal(heldBy([1, 2, 3]), null);
  assert.equal(heldBy(new Uint8Array(shiftDown))?.held, true);
});

test('a held key keeps the keepalive running until it is released', () => {
  const fake = fakeTimers();
  let sent = 0;
  const keepalive = new HoldKeepalive(() => sent++, fake.timers);

  keepalive.observe(shiftDown);
  assert.equal(fake.running(), true);
  fake.tick();
  fake.tick();
  assert.equal(sent, 2);

  // A second key does not start a second interval.
  keepalive.observe([1, 0x02, 0, 0x04, 0, 0, 0, 0, 0]);
  assert.equal(fake.started(), 1);

  keepalive.observe(keysUp);
  assert.equal(fake.running(), false);
  assert.equal(keepalive.active, false);
});

test('nothing is sent while nothing is held', () => {
  const fake = fakeTimers();
  const keepalive = new HoldKeepalive(() => assert.fail('keepalive sent while idle'), fake.timers);

  keepalive.observe(keysUp);
  keepalive.observe(relativeUp);
  keepalive.observe([2, 0, 3, 250, 0, 0]);
  keepalive.observe([0]);
  assert.equal(fake.started(), 0);
});

test('the keepalive lasts while any input holds something', () => {
  const fake = fakeTimers();
  const keepalive = new HoldKeepalive(() => {}, fake.timers);

  keepalive.observe(absoluteLeftDown);
  keepalive.observe(aDown);
  keepalive.observe(keysUp);
  assert.equal(fake.running(), true, 'the absolute button is still down');
  keepalive.observe(absoluteUp);
  assert.equal(fake.running(), false);

  keepalive.observe(relativeLeftDown);
  keepalive.reset();
  assert.equal(fake.running(), false, 'reset forgets what was held');
  keepalive.observe(keysUp);
  assert.equal(fake.running(), false);
});
