import assert from 'node:assert/strict';
import { test } from 'node:test';
import semver from 'semver';

import { compareVersions, isValidVersion, versionGt, versionGte } from './version.ts';

test('orders release versions by major, minor and patch', () => {
  assert.equal(compareVersions('1.2.3', '1.2.4'), -1);
  assert.equal(compareVersions('1.10.0', '1.9.9'), 1);
  assert.equal(compareVersions('2.0.0', '2.0.0'), 0);
});

test('accepts a leading v and surrounding whitespace', () => {
  assert.equal(compareVersions('v1.2.3', '1.2.3'), 0);
  assert.equal(versionGt(' v2.4.4 ', '2.4.3'), true);
});

test('sorts a prerelease below its release', () => {
  assert.equal(versionGt('1.2.3', '1.2.3-rc1'), true);
  assert.equal(versionGte('1.2.3-rc1', '1.2.3'), false);
  assert.equal(compareVersions('1.2.3-rc.2', '1.2.3-rc.10'), -1);
  assert.equal(compareVersions('1.2.3-rc2', '1.2.3-rc10'), 1);
  assert.equal(compareVersions('1.2.3-1', '1.2.3-alpha'), -1);
  assert.equal(compareVersions('1.2.3-alpha', '1.2.3-alpha.1'), -1);
});

test('ignores the build stamp the server appends as metadata', () => {
  assert.equal(compareVersions('2.4.3+dev.20260930.1023.7ff5405f', '2.4.3'), 0);
  assert.equal(versionGte('2.4.3+dev.20260930.7ff5405f', '2.4.3'), true);
  assert.equal(versionGt('2.4.4', '2.4.3+iron.5.dev.20260930.7ff5405f'), true);
});

test('refuses what is not a version, as semver does', () => {
  for (const bad of ['', 'dev.20260930.7ff5405f', '1.2', '01.2.3', '1.2.3-01', 'latest']) {
    assert.equal(isValidVersion(bad), false, bad);
    assert.throws(() => compareVersions(bad, '1.0.0'), TypeError, bad);
    assert.throws(() => compareVersions('1.0.0', bad), TypeError, bad);
  }
  assert.equal(isValidVersion('1.2.3-rc1+build.7'), true);
});

// Every pair from this list is compared by both, including the invalid ones,
// where both must throw.
const samples = [
  '0.0.0',
  '1.0.0',
  'v1.0.0',
  '=1.0.0',
  ' 1.0.0 ',
  '1.0.0-0',
  '1.0.0-1',
  '1.0.0-01',
  '1.0.0-alpha',
  '1.0.0-alpha.1',
  '1.0.0-alpha.beta',
  '1.0.0-beta',
  '1.0.0-beta.2',
  '1.0.0-beta.11',
  '1.0.0-rc1',
  '1.0.0-rc10',
  '1.0.0-rc.1',
  '1.0.0-1a',
  '1.0.0-a-b',
  '1.0.0+build',
  '1.0.0+dev.20260930.7ff5405f',
  '1.0.0-rc1+dev.20260930.1023.07ff5405',
  '1.0.1',
  '1.1.0',
  '2.4.3',
  '2.4.3+iron.5',
  '2.4.10',
  '10.0.0',
  '9007199254740991.0.0',
  '9007199254740992.0.0',
  '1.0.0-9007199254740993',
  '1.0.0-9007199254740992',
  '1.2',
  '1.2.3.4',
  'dev.20260930.7ff5405f',
  '',
  'x.y.z',
  '1.2.3-',
  '1.2.3+',
  `1.2.3-${'a'.repeat(260)}`
];

test('agrees with semver on every pair of samples', () => {
  const outcome = (fn: () => unknown) => {
    try {
      return fn();
    } catch (e) {
      return e instanceof TypeError ? 'TypeError' : e;
    }
  };

  for (const a of samples) {
    assert.equal(isValidVersion(a), semver.valid(a) !== null, `valid ${a}`);
    for (const b of samples) {
      assert.equal(
        outcome(() => versionGt(a, b)),
        outcome(() => semver.gt(a, b)),
        `gt ${a} ${b}`
      );
      assert.equal(
        outcome(() => versionGte(a, b)),
        outcome(() => semver.gte(a, b)),
        `gte ${a} ${b}`
      );
    }
  }
});
