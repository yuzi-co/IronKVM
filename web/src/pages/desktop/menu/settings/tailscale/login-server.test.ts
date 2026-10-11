// Run with `npm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { isLoginServer, sameLoginServer } from './login-server.ts';

describe('isLoginServer', () => {
  it('takes empty for Tailscale and an https URL for Headscale', () => {
    for (const ok of [
      '',
      '  ',
      'https://headscale.example.com',
      'https://headscale.example.com/',
      ' https://hs.example.com:8443 ',
      'https://10.0.0.5',
      'https://[fd00::1]:443',
      'https://example.com/headscale'
    ]) {
      assert.equal(isLoginServer(ok), true, ok);
    }
  });

  it('refuses anything else', () => {
    for (const bad of [
      'http://headscale.example.com',
      'headscale.example.com',
      'https://',
      'https://user:pass@hs.example.com',
      'https://hs.example.com?x=1',
      'https://hs.example.com#top',
      'https://hs .example.com',
      '--auth-key=tskey-abc'
    ]) {
      assert.equal(isLoginServer(bad), false, bad);
    }
  });
});

describe('sameLoginServer', () => {
  it('ignores a trailing slash and spaces', () => {
    assert.equal(sameLoginServer('https://hs.example.com/', ' https://hs.example.com'), true);
    assert.equal(sameLoginServer('', ''), true);
    assert.equal(sameLoginServer('https://a.example.com', 'https://b.example.com'), false);
  });
});
