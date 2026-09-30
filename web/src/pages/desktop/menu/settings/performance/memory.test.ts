// Run with `npm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { groupUse, ramUse } from './memory.ts';

const MiB = 1024 * 1024;

describe('ramUse', () => {
  it('counts what is not available as used', () => {
    assert.deepEqual(ramUse({ total: 200 * MiB, available: 80 * MiB }), {
      used: 120 * MiB,
      percent: 60,
      low: false
    });
  });

  it('flags less than 15% available', () => {
    assert.equal(ramUse({ total: 200 * MiB, available: 30 * MiB }).low, false);
    assert.equal(ramUse({ total: 200 * MiB, available: 29 * MiB }).low, true);
  });

  it('says nothing without a reading', () => {
    assert.deepEqual(ramUse({ total: 0, available: 0 }), { used: 0, percent: 0, low: false });
  });

  it('keeps a torn reading inside the bar', () => {
    assert.equal(ramUse({ total: 100 * MiB, available: 120 * MiB }).percent, 0);
    assert.equal(ramUse({ total: 100 * MiB, available: -1 }).percent, 100);
  });
});

describe('groupUse', () => {
  it('measures against memory.high, else memory.max', () => {
    assert.deepEqual(groupUse({ current: 40 * MiB, high: 64 * MiB, max: 96 * MiB }), {
      limit: 64 * MiB,
      pressed: false
    });
    assert.equal(groupUse({ current: 40 * MiB, high: 0, max: 96 * MiB }).limit, 96 * MiB);
    assert.equal(groupUse({ current: 40 * MiB, high: 0, max: 0 }).limit, 0);
  });

  it('flags a group within a tenth of memory.high', () => {
    assert.equal(groupUse({ current: 58 * MiB, high: 64 * MiB, max: 0 }).pressed, true);
    assert.equal(groupUse({ current: 57 * MiB, high: 64 * MiB, max: 0 }).pressed, false);
  });
});
