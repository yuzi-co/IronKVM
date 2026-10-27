import assert from 'node:assert/strict';
import { test } from 'node:test';

import { hostPart, rtspUrl, srtUrl } from './urls.ts';

test('an IPv6 address is bracketed and a name is not', () => {
  assert.equal(hostPart('kvm.local'), 'kvm.local');
  assert.equal(hostPart('10.0.30.222'), '10.0.30.222');
  assert.equal(hostPart('fe80::1'), '[fe80::1]');
  assert.equal(hostPart('[fe80::1]'), '[fe80::1]');
});

test('the RTSP address carries the token as the password', () => {
  assert.equal(
    rtspUrl('kvm.local', 8554, '/stream'),
    'rtsp://viewer:<token>@kvm.local:8554/stream'
  );
  assert.equal(
    rtspUrl('10.0.30.222', 8555, 'stream', 'nkvm_a/b'),
    'rtsp://viewer:nkvm_a%2Fb@10.0.30.222:8555/stream'
  );
});

test('the RTSPS address differs only in its scheme and port', () => {
  assert.equal(
    rtspUrl('kvm.local', 8322, '/stream', undefined, true),
    'rtsps://viewer:<token>@kvm.local:8322/stream'
  );
  assert.equal(
    rtspUrl('fe80::1', 8322, '/stream', 'nkvm_abc', true),
    'rtsps://viewer:nkvm_abc@[fe80::1]:8322/stream'
  );
});

test('the SRT address calls the board with the passphrase and the token', () => {
  assert.equal(
    srtUrl('kvm.local', 8890),
    'srt://kvm.local:8890?mode=caller&passphrase=<passphrase>&streamid=<token>'
  );
  assert.equal(
    srtUrl('fe80::1', 9000, 'two words&more', 'nkvm_abc'),
    'srt://[fe80::1]:9000?mode=caller&passphrase=two%20words%26more&streamid=nkvm_abc'
  );
});
