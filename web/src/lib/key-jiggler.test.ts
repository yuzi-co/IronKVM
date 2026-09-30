import assert from 'node:assert/strict';
import { test } from 'node:test';

import { keyJigglerKeyFrom, keyJigglerKeys } from './key-jiggler.ts';

test('every known key reads as itself', () => {
  for (const key of keyJigglerKeys) {
    assert.equal(keyJigglerKeyFrom(key), key);
  }
});

test('a missing or unknown key reads as F15', () => {
  assert.equal(keyJigglerKeyFrom(undefined), 'f15');
  assert.equal(keyJigglerKeyFrom(''), 'f15');
  assert.equal(keyJigglerKeyFrom('alt'), 'f15');
  assert.equal(keyJigglerKeyFrom('F15'), 'f15');
});
