import assert from 'node:assert/strict';
import { test } from 'node:test';

import { DEFAULT_VIDEO_SCALE, readVideoScale } from './video-scale.ts';

function storage(value: string | null) {
  return { getItem: () => value };
}

test('a saved zoom is read back as it was saved', () => {
  assert.equal(readVideoScale(storage('2')), 2);
  assert.equal(readVideoScale(storage('1.5')), 1.5);
  assert.equal(readVideoScale(storage('0.5')), 0.5);
});

test('with nothing saved the picture starts at 100 %', () => {
  assert.equal(DEFAULT_VIDEO_SCALE, 1);
  assert.equal(readVideoScale(storage(null)), 1);
  assert.equal(readVideoScale(storage('')), 1);
});

test('a saved value that is not a usable zoom starts at 100 %', () => {
  for (const value of ['0', '-2', 'abc', 'NaN', 'Infinity', '1e9']) {
    assert.equal(readVideoScale(storage(value)), 1, value);
  }
});

test('storage that cannot be read starts at 100 %', () => {
  const blocked = {
    getItem: () => {
      throw new Error('SecurityError');
    }
  };
  assert.equal(readVideoScale(blocked), 1);
  assert.equal(readVideoScale(undefined), 1);
});
