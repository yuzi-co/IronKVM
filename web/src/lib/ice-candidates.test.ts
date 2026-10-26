import assert from 'node:assert/strict';
import { test } from 'node:test';

import { CandidateInbox, CandidateOutbox, iceFailure } from './ice-candidates.ts';

const candidate = (n: number) => ({
  candidate: `candidate:${n} 1 udp 1 192.0.2.${n} 5000 typ host`
});

test('a candidate gathered before the offer is sent waits for it', () => {
  const sent: string[] = [];
  const outbox = new CandidateOutbox((c) => {
    sent.push(c.candidate ?? '');
    return true;
  });

  assert.equal(outbox.add(candidate(1)), true);
  assert.equal(outbox.add(candidate(2)), true);
  assert.deepEqual(sent, []);

  assert.equal(outbox.offerSent(), true);
  assert.deepEqual(sent, [candidate(1).candidate, candidate(2).candidate]);

  outbox.add(candidate(3));
  assert.deepEqual(sent, [candidate(1).candidate, candidate(2).candidate, candidate(3).candidate]);
  assert.equal(outbox.count, 3);
});

test('the empty end-of-candidates marker goes on but is not counted', () => {
  const sent: string[] = [];
  const outbox = new CandidateOutbox((c) => {
    sent.push(c.candidate ?? '');
    return true;
  });

  outbox.offerSent();
  outbox.add({ candidate: '', sdpMid: '0' });
  assert.deepEqual(sent, ['']);
  assert.equal(outbox.count, 0);
});

test('a candidate that cannot be sent reports the failure', () => {
  const outbox = new CandidateOutbox(() => false);

  assert.equal(outbox.add(candidate(1)), true, 'held, not sent yet');
  assert.equal(outbox.offerSent(), false);
  assert.equal(outbox.add(candidate(2)), false);
});

test('a candidate from the board waits for the answer', () => {
  const added: number[] = [];
  const inbox = new CandidateInbox<number>((c) => added.push(c));

  inbox.receive(1, false);
  inbox.receive(2, false);
  assert.deepEqual(added, []);

  inbox.answerApplied();
  assert.deepEqual(added, [1, 2]);

  inbox.receive(3, true);
  assert.deepEqual(added, [1, 2, 3]);

  inbox.answerApplied();
  assert.deepEqual(added, [1, 2, 3], 'a held candidate is added once');
});

test('gathering that ends with no candidate is told apart', () => {
  assert.equal(
    iceFailure({
      gatheringState: 'complete',
      localCandidates: 0,
      answerApplied: true,
      connectionState: 'new'
    }),
    'no-candidates'
  );
  assert.equal(
    iceFailure({
      gatheringState: 'gathering',
      localCandidates: 0,
      answerApplied: false,
      connectionState: 'new'
    }),
    null
  );
});

test('checks that never reach the board mean no route', () => {
  for (const connectionState of ['new', 'connecting', 'failed'] as const) {
    assert.equal(
      iceFailure({
        gatheringState: 'complete',
        localCandidates: 3,
        answerApplied: true,
        connectionState
      }),
      'no-route',
      connectionState
    );
  }
});

test('a failure before the answer, or after a connection, is not a routing one', () => {
  assert.equal(
    iceFailure({
      gatheringState: 'complete',
      localCandidates: 3,
      answerApplied: false,
      connectionState: 'new'
    }),
    null
  );
  assert.equal(
    iceFailure({
      gatheringState: 'complete',
      localCandidates: 3,
      answerApplied: true,
      connectionState: 'disconnected'
    }),
    null
  );
});
