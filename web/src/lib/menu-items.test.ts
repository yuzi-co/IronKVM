import assert from 'node:assert/strict';
import { test } from 'node:test';

import { migrateMenuDisabledItems } from './menu-items.ts';

test('hides Media only when both Image and Download were hidden', () => {
  assert.deepEqual(migrateMenuDisabledItems(['image', 'download']), ['media']);
  assert.deepEqual(migrateMenuDisabledItems(['wol', 'download', 'image']), ['wol', 'media']);
});

test('shows Media when only one of the old buttons was hidden', () => {
  assert.deepEqual(migrateMenuDisabledItems(['image']), []);
  assert.deepEqual(migrateMenuDisabledItems(['download', 'power']), ['power']);
});

test('keeps the Tools entries and every other id as they were', () => {
  const items = ['script', 'wol', 'picoclaw', 'power', 'speaker'];
  assert.deepEqual(migrateMenuDisabledItems(['image', ...items]), items);
});

test('returns the same list when nothing needs rewriting', () => {
  const items = ['media', 'script'];
  assert.equal(migrateMenuDisabledItems(items), items);
  const empty: string[] = [];
  assert.equal(migrateMenuDisabledItems(empty), empty);
});

test('does not add Media twice', () => {
  assert.deepEqual(migrateMenuDisabledItems(['media', 'image', 'download']), ['media']);
});
