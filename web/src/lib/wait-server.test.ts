import assert from 'node:assert/strict';
import { test } from 'node:test';

import { waitForRestart, waitUntil } from './wait-server.ts';

// fakeServer answers probes from a script of states, one per probe, and keeps
// the clock moving by the interval on every sleep.
function fakeServer(states: boolean[]) {
  let clock = 0;
  let probes = 0;
  return {
    deps: {
      probe: async () => states[Math.min(probes++, states.length - 1)],
      sleep: async (ms: number) => {
        clock += ms;
      },
      now: () => clock,
      intervalMs: 1000
    },
    probes: () => probes,
    clock: () => clock
  };
}

test('waitUntil returns as soon as the state is reached', async () => {
  const s = fakeServer([true, true, false]);
  assert.equal(await waitUntil(false, 10_000, s.deps), true);
  assert.equal(s.probes(), 3);
});

test('waitUntil gives up after its budget', async () => {
  const s = fakeServer([true]);
  assert.equal(await waitUntil(false, 5000, s.deps), false);
  assert.equal(s.clock(), 5000);
});

test('waitForRestart sees the server go down and come back', async () => {
  const s = fakeServer([true, false, false, true]);
  assert.equal(await waitForRestart(10_000, 10_000, s.deps), true);
  assert.equal(s.probes(), 4);
});

test('waitForRestart reports a server that never comes back', async () => {
  const s = fakeServer([true, false]);
  assert.equal(await waitForRestart(10_000, 3000, s.deps), false);
});

test('a restart missed between two probes costs only the down budget', async () => {
  const s = fakeServer([true]);
  assert.equal(await waitForRestart(4000, 10_000, s.deps), true);
  assert.equal(s.clock(), 4000);
});

test('an aborted wait stops probing and answers false', async () => {
  const s = fakeServer([true]);
  const abort = new AbortController();
  let probes = 0;
  const deps = {
    ...s.deps,
    signal: abort.signal,
    probe: async () => {
      if (++probes === 2) abort.abort();
      return true;
    }
  };
  assert.equal(await waitForRestart(60_000, 60_000, deps), false);
  assert.equal(probes, 2);
});
