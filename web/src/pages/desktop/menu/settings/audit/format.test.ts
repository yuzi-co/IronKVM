import assert from 'node:assert/strict';
import { test } from 'node:test';

import { auditFileName, subject, via, who } from './format.ts';

test('who names the key beside the account', () => {
  assert.equal(who({ actor: 'alice' }), 'alice');
  assert.equal(who({ actor: 'alice', key: 'ci' }), 'alice (ci)');
  assert.equal(who({}), '-');
});

test('via shows the start of the agent session', () => {
  assert.equal(via({}), '');
  assert.equal(via({ via: 'session' }), 'session');
  assert.equal(via({ via: 'agent', session: '3f2a9c1e-77aa-4b1e-9d1c-0a1b2c3d4e5f' }), 'agent · 3f2a9c1e');
});

test('subject joins target and detail and shows a repeat count', () => {
  assert.equal(subject({}), '');
  assert.equal(subject({ target: 'a.iso' }), 'a.iso');
  assert.equal(subject({ target: 'power', detail: '800ms' }), 'power - 800ms');
  assert.equal(subject({ detail: 'invalid password', repeat: 12 }), 'invalid password - x12');
  assert.equal(subject({ target: 'x', repeat: 1 }), 'x');
});

test('auditFileName carries the host and the day', () => {
  assert.equal(
    auditFileName('kvm.lan', new Date('2026-10-12T08:00:00Z')),
    'ironkvm-audit-kvm.lan-2026-10-12.jsonl'
  );
});
