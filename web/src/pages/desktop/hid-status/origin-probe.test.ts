import assert from 'node:assert/strict';
import { test } from 'node:test';

import { originMismatch, probeFromError } from './origin-probe.ts';

test('a 403 naming both hosts is an origin mismatch', () => {
  const probe = probeFromError({
    response: {
      status: 403,
      data: {
        code: -1,
        msg: 'forbidden origin',
        data: { origin: 'kvm.lan.yuzi.co', host: 'ironkvm.lan' }
      }
    }
  });
  assert.deepEqual(originMismatch(probe), { origin: 'kvm.lan.yuzi.co', host: 'ironkvm.lan' });
});

test('a network error is not an origin mismatch', () => {
  assert.equal(originMismatch(probeFromError(new Error('Network Error'))), null);
});

test('a 401 or a 403 without the names keeps the current choice', () => {
  assert.equal(originMismatch({ status: 401 }), null);
  assert.equal(originMismatch({ status: 403 }), null);
  assert.equal(originMismatch({ status: 403, data: { origin: 'a' } }), null);
  assert.equal(originMismatch({ status: 403, data: { origin: '', host: 'b' } }), null);
  assert.equal(originMismatch(null), null);
});

test('a success is not an origin mismatch', () => {
  assert.equal(originMismatch({ status: 200 }), null);
});
