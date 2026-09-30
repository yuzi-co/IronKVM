import assert from 'node:assert/strict';
import { test } from 'node:test';

import { pollWhileVisible, type PollEnv } from './visible-poll.ts';

function fakeEnv() {
  let visible = true;
  let tick: (() => void) | null = null;
  let onChange: (() => void) | null = null;
  const env: PollEnv = {
    isVisible: () => visible,
    onVisibilityChange: (cb) => {
      onChange = cb;
      return () => {
        onChange = null;
      };
    },
    setInterval: (cb) => {
      tick = cb;
      return () => {
        tick = null;
      };
    }
  };
  return {
    env,
    tick: () => tick?.(),
    setVisible: (v: boolean) => {
      visible = v;
      onChange?.();
    },
    stopped: () => tick === null && onChange === null
  };
}

test('pollWhileVisible skips hidden turns and polls at once on return', () => {
  const fake = fakeEnv();
  let calls = 0;
  pollWhileVisible(() => calls++, 1000, fake.env);

  fake.tick();
  assert.equal(calls, 1);

  fake.setVisible(false);
  assert.equal(calls, 1);
  fake.tick();
  fake.tick();
  assert.equal(calls, 1);

  fake.setVisible(true);
  assert.equal(calls, 2);
});

test('pollWhileVisible stops its timer and listener', () => {
  const fake = fakeEnv();
  let calls = 0;
  const stop = pollWhileVisible(() => calls++, 1000, fake.env);
  stop();
  assert.ok(fake.stopped());
  fake.tick();
  assert.equal(calls, 0);
});
