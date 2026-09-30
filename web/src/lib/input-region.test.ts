import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  AUTO_REGION_FAST_MS,
  AUTO_REGION_SLOW_MS,
  autoRegionDelay,
  sameInputRegion
} from './input-region.ts';

const region = { frameWidth: 1920, frameHeight: 1080, left: 0, top: 60, width: 1920, height: 960 };

test('sameInputRegion matches equal numbers in different objects', () => {
  assert.equal(sameInputRegion(region, { ...region }), true);
  assert.equal(sameInputRegion(null, null), true);
  assert.equal(sameInputRegion(region, region), true);
});

test('sameInputRegion tells any changed field and null apart', () => {
  for (const key of Object.keys(region) as (keyof typeof region)[]) {
    assert.equal(sameInputRegion(region, { ...region, [key]: region[key] + 1 }), false, key);
  }
  assert.equal(sameInputRegion(region, null), false);
  assert.equal(sameInputRegion(null, region), false);
});

test('autoRegionDelay slows down once the region is stable and speeds up on change', () => {
  assert.equal(autoRegionDelay(0), AUTO_REGION_FAST_MS);
  assert.equal(autoRegionDelay(1), AUTO_REGION_FAST_MS);
  assert.equal(autoRegionDelay(2), AUTO_REGION_FAST_MS);
  assert.equal(autoRegionDelay(3), AUTO_REGION_SLOW_MS);
  assert.equal(autoRegionDelay(10), AUTO_REGION_SLOW_MS);
});
