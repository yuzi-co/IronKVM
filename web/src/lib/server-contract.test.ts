import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  apiSupported,
  contractFromResponse,
  formatRange,
  hasFeature,
  LEGACY_CONTRACT,
  SERVER_FEATURES,
  UI_API_RANGE
} from './server-contract.ts';

test('reads a server that answers the contract', () => {
  const c = contractFromResponse({ version: '2.4.3', api: 1, features: ['status-bar', 7] });
  assert.equal(c.known, true);
  assert.equal(c.version, '2.4.3');
  assert.equal(c.api, 1);
  assert.deepEqual([...c.features], ['status-bar']);
});

test('a malformed answer reads as a legacy server', () => {
  for (const data of [null, undefined, 'x', {}, { api: '1' }, { api: 0 }, { api: 1.5 }]) {
    assert.equal(contractFromResponse(data).known, false, JSON.stringify(data));
  }
});

test('a declared contract is taken at its word', () => {
  const c = contractFromResponse({ version: '', api: 1, features: ['status-bar'] });
  assert.equal(hasFeature(c, 'status-bar'), true);
  assert.equal(hasFeature(c, 'direct-clock'), false);
});

test('a legacy server is asked for the older features, so the callers keep their 404 handling', () => {
  assert.equal(hasFeature(LEGACY_CONTRACT, 'status-bar'), true);
  assert.equal(hasFeature(LEGACY_CONTRACT, 'direct-clock'), true);
});

test('a legacy server never has a feature that came with the contract', () => {
  assert.equal(hasFeature(LEGACY_CONTRACT, 'ui-update'), false);
});

test('a legacy server speaks the first API version, which this UI supports', () => {
  assert.equal(apiSupported(LEGACY_CONTRACT), true);
});

test('the API range is checked at both ends', () => {
  const range = { min: 2, max: 3 };
  const at = (api: number) => contractFromResponse({ api, features: [] });
  assert.equal(apiSupported(at(1), range), false);
  assert.equal(apiSupported(at(2), range), true);
  assert.equal(apiSupported(at(3), range), true);
  assert.equal(apiSupported(at(4), range), false);
});

test('the range reads as one number or a span', () => {
  assert.equal(formatRange({ min: 1, max: 1 }), '1');
  assert.equal(formatRange({ min: 1, max: 3 }), '1–3');
  assert.ok(UI_API_RANGE.min <= UI_API_RANGE.max);
});

test('the feature table is sorted and has no duplicates', () => {
  const sorted = [...SERVER_FEATURES].sort();
  assert.deepEqual([...SERVER_FEATURES], sorted);
  assert.equal(new Set(SERVER_FEATURES).size, SERVER_FEATURES.length);
});
