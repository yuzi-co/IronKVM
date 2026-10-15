import assert from 'node:assert/strict';
import { test } from 'node:test';

import { createTurnStopper, type TurnStopperOptions } from './picoclaw-stop.ts';

function fakeStopper(sent = true) {
  const calls: string[] = [];
  let fire: (() => void) | null = null;
  let armedMs = 0;
  const options: TurnStopperOptions = {
    sendCancel: () => {
      calls.push('sendCancel');
      return sent;
    },
    onSent: () => calls.push('sent'),
    onFailed: (reason) => calls.push(`failed:${reason}`),
    onPendingChange: (pending) => calls.push(`pending:${pending}`),
    timeoutMs: 5000,
    setTimer: (cb, ms) => {
      fire = cb;
      armedMs = ms;
      return () => {
        fire = null;
      };
    }
  };
  return {
    stopper: createTurnStopper(options),
    take: () => calls.splice(0, calls.length),
    fire: () => fire?.(),
    armed: () => fire !== null,
    armedMs: () => armedMs
  };
}

test('a stop the agent confirms ends without an error', () => {
  const fake = fakeStopper();
  assert.equal(fake.stopper.stop(), true);
  assert.deepEqual(fake.take(), ['sendCancel', 'sent', 'pending:true']);
  assert.equal(fake.armed(), true);
  assert.equal(fake.armedMs(), 5000);
  assert.equal(fake.stopper.isPending(), true);

  // The turn_done with stopReason "cancelled" puts the run state to idle.
  fake.stopper.settle();
  assert.deepEqual(fake.take(), ['pending:false']);
  assert.equal(fake.armed(), false, 'the timeout is cleared');
  assert.equal(fake.stopper.isPending(), false);

  fake.stopper.settle();
  assert.deepEqual(fake.take(), [], 'settling twice reports nothing');
});

test('a second press while the first stop waits sends nothing', () => {
  const fake = fakeStopper();
  fake.stopper.stop();
  fake.take();
  assert.equal(fake.stopper.stop(), false);
  assert.deepEqual(fake.take(), []);
});

test('a stop the agent never confirms fails after the timeout', () => {
  const fake = fakeStopper();
  fake.stopper.stop();
  fake.take();

  fake.fire();
  assert.deepEqual(fake.take(), ['pending:false', 'failed:timeout']);
  assert.equal(fake.stopper.isPending(), false);

  fake.stopper.settle();
  assert.deepEqual(fake.take(), [], 'the idle that follows reports nothing');

  assert.equal(fake.stopper.stop(), true, 'the user can press Stop again');
  assert.deepEqual(fake.take(), ['sendCancel', 'sent', 'pending:true']);
});

test('a cancel that cannot be sent fails at once and arms nothing', () => {
  const fake = fakeStopper(false);
  assert.equal(fake.stopper.stop(), false);
  assert.deepEqual(fake.take(), ['sendCancel', 'failed:not_sent']);
  assert.equal(fake.armed(), false);
  assert.equal(fake.stopper.isPending(), false);
});

test('the default timeout is used when none is given', () => {
  let armedMs = 0;
  const stopper = createTurnStopper({
    sendCancel: () => true,
    onSent: () => undefined,
    onFailed: () => undefined,
    setTimer: (_cb, ms) => {
      armedMs = ms;
      return () => undefined;
    }
  });
  stopper.stop();
  assert.equal(armedMs, 15000);
});
