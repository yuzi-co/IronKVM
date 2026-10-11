import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  DEFAULT_SHARE_DURATION,
  isShareToken,
  SHARE_DURATIONS,
  shareDurationSeconds,
  shareLinkUrl
} from './share-link.ts';

test('the token goes after the hash, where no request carries it', () => {
  const url = shareLinkUrl({ origin: 'https://kvm.lan', pathname: '/' }, 'abc-DEF_123');
  assert.equal(url, 'https://kvm.lan/#/share/abc-DEF_123');
  assert.equal(new URL(url).search, '');
});

test('the durations are ascending and inside the server limits', () => {
  const seconds = SHARE_DURATIONS.map((duration) => duration.seconds);
  assert.deepEqual(
    [...seconds].sort((a, b) => a - b),
    seconds
  );
  assert.ok(seconds[0] >= 60);
  assert.ok(seconds[seconds.length - 1] <= 30 * 24 * 60 * 60);
});

test('the default duration is one of the offered ones', () => {
  assert.ok(SHARE_DURATIONS.some((duration) => duration.key === DEFAULT_SHARE_DURATION));
  assert.equal(shareDurationSeconds('d1'), 86400);
});

test('only a token shaped like the server makes is sent', () => {
  assert.equal(isShareToken('A'.repeat(43)), true);
  assert.equal(isShareToken(undefined), false);
  assert.equal(isShareToken(''), false);
  assert.equal(isShareToken('short'), false);
  assert.equal(isShareToken('a/b'.padEnd(43, 'x')), false);
});
