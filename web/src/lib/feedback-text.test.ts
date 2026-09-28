import assert from 'node:assert/strict';
import { test } from 'node:test';

import { failureKind, failureText } from './feedback-text.ts';

const t = (key: string) => `<${key}>`;

test('shows the reason an API reply gives', () => {
  assert.equal(failureText({ code: -2, msg: 'port in use' }, 'fallback', t), 'port in use');
});

test('falls back when the reply has no reason', () => {
  assert.equal(failureText({ code: -1, msg: '' }, 'fallback', t), 'fallback');
  assert.equal(failureText({ code: -1 }, 'fallback', t), 'fallback');
  assert.equal(failureText(undefined, 'fallback', t), 'fallback');
  assert.equal(failureText(new Error('boom'), 'fallback', t), 'fallback');
});

test('names a timeout', () => {
  const err = { code: 'ECONNABORTED', message: 'timeout of 60000ms exceeded', request: {} };
  assert.equal(failureKind(err), 'timeout');
  assert.equal(failureText(err, 'fallback', t), '<feedback.timeout>');
});

test('names a request that got no answer', () => {
  const err = { code: 'ERR_NETWORK', message: 'Network Error', request: {} };
  assert.equal(failureKind(err), 'network');
  assert.equal(failureText(err, 'fallback', t), '<feedback.network>');
});

test('uses the reason in an HTTP error body', () => {
  const err = {
    code: 'ERR_BAD_REQUEST',
    request: {},
    response: { status: 400, data: { code: -1, msg: 'invalid parameters' } }
  };
  assert.equal(failureKind(err), 'other');
  assert.equal(failureText(err, 'fallback', t), 'invalid parameters');
});
