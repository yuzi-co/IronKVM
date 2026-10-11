import assert from 'node:assert/strict';
import { test } from 'node:test';

import { pickUpload, uiOffer } from './ui-package.ts';

const f = (name: string) => ({ name });

test('sorts a package and its signature in either order', () => {
  const pkg = f('ironkvm-ui_1.2.0.tar.gz');
  const sig = f('ironkvm-ui_1.2.0.tar.gz.minisig');
  for (const files of [
    [pkg, sig],
    [sig, pkg]
  ]) {
    const pick = pickUpload(files);
    assert.equal(pick.ok, true);
    if (pick.ok) {
      assert.equal(pick.pkg, pkg);
      assert.equal(pick.signature, sig);
    }
  }
});

test('a package alone is passed on, for the server to judge', () => {
  const pick = pickUpload([f('ironkvm-ui_1.2.0+dev.1.tar.gz')]);
  assert.equal(pick.ok, true);
  if (pick.ok) assert.equal(pick.signature, null);
});

test('refuses what is not one package and at most one signature', () => {
  assert.deepEqual(pickUpload([]), { ok: false, error: 'noPackage' });
  assert.deepEqual(pickUpload([f('a.minisig')]), { ok: false, error: 'noPackage' });
  assert.deepEqual(pickUpload([f('a'), f('b'), f('c')]), { ok: false, error: 'tooMany' });
  assert.deepEqual(pickUpload([f('ironkvm-ui_1.0.0.tar.gz'), f('ironkvm-ui_1.1.0.tar.gz')]), {
    ok: false,
    error: 'tooMany'
  });
});

test('refuses an application package or a renamed file', () => {
  assert.deepEqual(pickUpload([f('ironkvm_1.2.0.tar.gz')]), { ok: false, error: 'badName' });
  assert.deepEqual(pickUpload([f('ironkvm-ui_latest.tar.gz')]), { ok: false, error: 'badName' });
});

test('refuses a signature for another package', () => {
  assert.deepEqual(
    pickUpload([f('ironkvm-ui_1.2.0.tar.gz'), f('ironkvm-ui_1.1.0.tar.gz.minisig')]),
    { ok: false, error: 'signatureMismatch' }
  );
});

test('offers an update only when it is new and fits the server', () => {
  const api = { min: 1, max: 1 };
  assert.deepEqual(uiOffer(null), { kind: 'none' });
  assert.deepEqual(uiOffer({ version: '1.2.0', compatible: true, active: true, api }), {
    kind: 'current',
    version: '1.2.0'
  });
  assert.deepEqual(uiOffer({ version: '1.3.0', compatible: false, active: false, api }), {
    kind: 'incompatible',
    version: '1.3.0',
    api
  });
  assert.deepEqual(uiOffer({ version: '1.3.0', compatible: true, active: false, api }), {
    kind: 'available',
    version: '1.3.0'
  });
});
