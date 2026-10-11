import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  ClockSampler,
  parseClockReply,
  quickestReading,
  wallClockOffset,
  type ClockSamplerOptions,
  type WallClockReading
} from './clock-offset.ts';

// A board whose clock is right, a page whose clock is skewMs ahead of it, and
// a request that takes upMs to reach the board and downMs to come back. The
// page asks at page time sentMs.
function reading(sentMs: number, skewMs: number, upMs: number, downMs: number): WallClockReading {
  const boardMs = sentMs - skewMs + upMs;
  return { sentMs, receivedMs: sentMs + upMs + downMs, boardMs };
}

for (const skewMs of [500, -500, 0]) {
  test(`a symmetric round trip finds a ${skewMs} ms skew exactly`, () => {
    const offset = wallClockOffset([reading(1_700_000_000_000, skewMs, 4, 4)]);
    assert.ok(offset);
    // board = page + offset
    assert.ok(offset.offsetMs === -skewMs, `offset ${offset.offsetMs}`);
    assert.equal(offset.errorMs, 4);
  });

  test(`an asymmetric round trip errs by half the difference, inside the bound (${skewMs} ms)`, () => {
    // 1 ms up, 9 ms down: the board read its clock early in the round trip.
    const offset = wallClockOffset([reading(1_700_000_000_000, skewMs, 1, 9)]);
    assert.ok(offset);
    const error = offset.offsetMs - -skewMs;
    assert.equal(error, -4);
    assert.ok(Math.abs(error) <= offset.errorMs);
    assert.equal(offset.errorMs, 5);
  });
}

test('the quickest reading wins, whatever its order', () => {
  const t = 1_700_000_000_000;
  const readings = [
    reading(t, 420, 30, 2), // slow and lopsided: 14 ms wrong
    reading(t + 1000, 420, 1, 2),
    reading(t + 2000, 420, 10, 10)
  ];
  const offset = wallClockOffset(readings);
  assert.ok(offset);
  assert.equal(offset.errorMs, 1.5);
  assert.ok(Math.abs(offset.offsetMs - -420) <= 1.5);
  assert.equal(quickestReading(readings), readings[1]);
});

test('a reading whose clock went backwards is skipped', () => {
  assert.equal(wallClockOffset([{ sentMs: 10, receivedMs: 5, boardMs: 1 }]), null);
  assert.equal(wallClockOffset([]), null);
});

// A sampler driven by hand: a fake clock, timers that only fire when told,
// and a board that answers when the test says.
function harness(options: Partial<ClockSamplerOptions> = {}) {
  let now = 1_700_000_000_000;
  let timer: { fn: () => void; due: number } | null = null;
  let sent = 0;
  let lastId = 0;
  let open = true;
  const sampler = new ClockSampler(
    (id) => {
      if (!open) return false;
      sent++;
      lastId = id;
      return true;
    },
    {
      now: () => now,
      setTimer: (fn, ms) => (timer = { fn, due: now + ms }),
      clearTimer: () => (timer = null),
      ...options
    }
  );
  return {
    sampler,
    sent: () => sent,
    lastId: () => lastId,
    setOpen: (value: boolean) => (open = value),
    advance(ms: number) {
      now += ms;
      while (timer && timer.due <= now) {
        const fire = timer;
        timer = null;
        fire.fn();
      }
    },
    // The board, skewMs behind the page, answers after rttMs.
    answer(rttMs: number, skewMs: number, upMs = rttMs / 2) {
      const boardMs = now - skewMs + upMs;
      now += rttMs;
      sampler.onReply(boardMs, lastId);
    }
  };
}

test('a round of five readings at start, the next half a minute later', () => {
  const h = harness();
  h.sampler.start();
  assert.equal(h.sent(), 1);
  for (const rtt of [12, 3, 8, 20, 5]) h.answer(rtt, 500);
  assert.equal(h.sent(), 5);
  const offset = h.sampler.estimate();
  assert.ok(offset);
  assert.equal(offset.errorMs, 1.5);
  assert.equal(offset.offsetMs, -500);

  // Nothing more until the next round.
  h.advance(29_000);
  assert.equal(h.sent(), 5);
  h.advance(1_000);
  assert.equal(h.sent(), 6);

  // A second start does not start a second round.
  h.sampler.start();
  assert.equal(h.sent(), 6);
});

test('a lost answer times out and the round goes on', () => {
  const h = harness();
  h.sampler.start();
  const first = h.lastId();
  h.advance(2_000);
  assert.equal(h.sent(), 2);
  // The late answer to the first request is not taken as the second's.
  h.sampler.onReply(123, first);
  assert.equal(h.sampler.estimate(), null);
  h.answer(1, 0);
  const offset = h.sampler.estimate();
  assert.equal(offset?.errorMs, 0.5);
  assert.equal(offset?.offsetMs, 0);
});

test('readings older than two rounds are forgotten, so a stepped clock is followed', () => {
  const h = harness();
  h.sampler.start();
  for (let i = 0; i < 5; i++) h.answer(2, 420);
  assert.equal(h.sampler.estimate()?.offsetMs, -420);

  // The page's clock is set right; every later round sees no skew, with a
  // worse round trip than before.
  for (let round = 0; round < 3; round++) {
    h.advance(30_000);
    for (let i = 0; i < 5; i++) h.answer(10, 0);
  }
  const offset = h.sampler.estimate();
  assert.equal(offset?.offsetMs, 0);
  assert.equal(offset?.errorMs, 5);
});

test('a closed socket is retried, and stop stops', () => {
  const h = harness();
  h.setOpen(false);
  h.sampler.start();
  assert.equal(h.sent(), 0);
  h.setOpen(true);
  h.advance(1_000);
  assert.equal(h.sent(), 1);
  h.sampler.stop();
  h.answer(2, 0);
  assert.equal(h.sampler.estimate(), null);
  h.advance(60_000);
  assert.equal(h.sent(), 1);
});

test('the signalling answer is the board clock in microseconds and the id', () => {
  assert.deepEqual(parseClockReply('1760000000123456 3'), { boardMs: 1760000000123.456, id: 3 });
  assert.equal(parseClockReply('1760000000123456'), null);
  assert.equal(parseClockReply('abc 3'), null);
  assert.equal(parseClockReply(undefined), null);
});
