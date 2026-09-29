import assert from 'node:assert/strict';
import { test } from 'node:test';

import { compareVersions, updateView } from './upstream-update.ts';

test('compares versions number by number', () => {
  assert.equal(compareVersions('3.0.3', '3.0.3'), 0);
  assert.equal(compareVersions('3.0.10', '3.0.9'), 1);
  assert.equal(compareVersions('v1.1.17', '1.1.18'), -1);
  assert.equal(compareVersions('2.0', '2.0.0'), 0);
  assert.equal(compareVersions('1.10', '1.9.9'), 1);
});

const base = {
  installed: true,
  version: '3.0.3',
  latest: '',
  checkedAt: null,
  checkError: '',
  updateAvailable: false,
  unverifiable: '',
  job: { state: 'idle' }
};

test('names what the update row shows', () => {
  assert.equal(updateView({ ...base, installed: false }), 'notInstalled');
  assert.equal(updateView(base), 'unchecked');
  const checked = { ...base, checkedAt: '2026-09-29T10:00:00Z' };
  assert.equal(updateView({ ...checked, checkError: 'offline' }), 'checkFailed');
  assert.equal(updateView({ ...checked, latest: '3.0.3' }), 'upToDate');
  assert.equal(updateView({ ...checked, latest: '3.0.2' }), 'upToDate');
  assert.equal(updateView({ ...checked, latest: '3.1.0', updateAvailable: true }), 'available');
  assert.equal(
    updateView({ ...checked, latest: '3.1.0', unverifiable: 'no checksum file' }),
    'unverifiable'
  );
  assert.equal(updateView({ ...checked, job: { state: 'running' } }), 'running');
});
