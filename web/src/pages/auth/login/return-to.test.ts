import assert from 'node:assert/strict';
import { test } from 'node:test';

import { loginNotice, returnTo } from './return-to.ts';

test('returns to the page the operator was sent away from', () => {
  assert.equal(
    returnTo({ from: { pathname: '/terminal', search: '?a=1', hash: '' } }),
    '/terminal?a=1'
  );
});

test('falls back to the desktop', () => {
  assert.equal(returnTo(undefined), '/');
  assert.equal(returnTo(null), '/');
  assert.equal(returnTo({}), '/');
  assert.equal(returnTo({ from: { pathname: 42 } }), '/');
});

test('never returns to an auth page or another origin', () => {
  assert.equal(returnTo({ from: { pathname: '/auth/password' } }), '/');
  assert.equal(returnTo({ from: { pathname: '/auth/login' } }), '/');
  assert.equal(returnTo({ from: { pathname: '//evil.example/' } }), '/');
  assert.equal(returnTo({ from: { pathname: 'https://evil.example/' } }), '/');
});

test('reads only the known notice', () => {
  assert.equal(loginNotice({ notice: 'passwordChanged' }), 'passwordChanged');
  assert.equal(loginNotice({ notice: 'other' }), undefined);
  assert.equal(loginNotice(null), undefined);
});
