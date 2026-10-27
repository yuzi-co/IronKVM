import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  httpRedirectWarning,
  oidcReason,
  pendingForAccount,
  scopesFromText,
  scopesToText
} from './oidc.ts';

test('reads the reason the server sent the browser back with', () => {
  assert.equal(oidcReason('?oidc=state'), 'state');
  assert.equal(oidcReason('?x=1&oidc=account'), 'account');
  for (const reason of [
    'disabled',
    'locked',
    'provider',
    'denied',
    'token',
    'role',
    'server',
    'binding',
    'unlinked'
  ]) {
    assert.equal(oidcReason(`?oidc=${reason}`), reason);
  }
});

test('an unknown or missing reason is none, never text from the address bar', () => {
  for (const search of ['', '?', '?oidc=', '?oidc=<script>', '?other=state', '?oidc=State']) {
    assert.equal(oidcReason(search), null, search);
  }
});

test('refused sign-ins under the name of the account come first, each group newest first', () => {
  const signIns = [
    { username: 'bob', issuer: 'i', subject: 'b1', time: 300 },
    { username: 'alice', issuer: 'i', subject: 'a1', time: 100 },
    { username: 'carol', issuer: 'i', subject: 'c1', time: 200 },
    { username: 'alice', issuer: 'i', subject: 'a2', time: 150 }
  ];
  assert.deepEqual(
    pendingForAccount(signIns, 'alice').map((s) => s.subject),
    ['a2', 'a1', 'b1', 'c1']
  );
  assert.deepEqual(
    pendingForAccount(signIns, 'nobody').map((s) => s.subject),
    ['b1', 'c1', 'a2', 'a1']
  );
  // The list it was given is left as it was.
  assert.equal(signIns[0].subject, 'b1');
});

test('scopes are split on spaces and commas, without duplicates or openid', () => {
  assert.deepEqual(scopesFromText('openid profile, email  groups profile'), [
    'profile',
    'email',
    'groups'
  ]);
  assert.deepEqual(scopesFromText('  '), []);
});

test('scopes are shown without the openid the server always adds', () => {
  assert.equal(scopesToText(['openid', 'profile', 'email']), 'profile email');
  assert.equal(scopesToText(undefined), '');
});

test('warns about an http redirect URL unless it is localhost', () => {
  assert.equal(httpRedirectWarning('http://192.168.1.5/api/auth/oidc/callback'), true);
  assert.equal(httpRedirectWarning('http://kvm.lan:8080/api/auth/oidc/callback'), true);
  assert.equal(httpRedirectWarning('https://kvm.lan/api/auth/oidc/callback'), false);
  assert.equal(httpRedirectWarning('http://localhost:8080/api/auth/oidc/callback'), false);
  assert.equal(httpRedirectWarning('http://127.0.0.1/api/auth/oidc/callback'), false);
  assert.equal(httpRedirectWarning('http://[::1]/api/auth/oidc/callback'), false);
  assert.equal(httpRedirectWarning(''), false);
});
