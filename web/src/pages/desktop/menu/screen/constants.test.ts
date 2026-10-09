import assert from 'node:assert/strict';
import { test } from 'node:test';

import { BitRateMap, DEFAULT_BIT_RATE, DEFAULT_QUALITY_KEY, qualityKey } from './constants.ts';

test('the default bitrate is 3000, the High step', () => {
  assert.equal(DEFAULT_BIT_RATE, 3000);
  assert.equal(BitRateMap.get(DEFAULT_QUALITY_KEY), DEFAULT_BIT_RATE);
});

test('the menu ticks the step the board holds', () => {
  const settings = { quality: 80, bitRate: 5000 };
  assert.equal(qualityKey('h264', settings), 1);
  assert.equal(qualityKey('direct', { ...settings, bitRate: 2000 }), 3);
  assert.equal(qualityKey('mjpeg', { ...settings, quality: 60 }), 3);
});

test('an unknown bitrate falls back to 3000', () => {
  const key = qualityKey('h264', { quality: 80, bitRate: 4000 });
  assert.equal(BitRateMap.get(key), 3000);
  assert.equal(qualityKey('', { quality: 80, bitRate: 3000 }), DEFAULT_QUALITY_KEY);
});
