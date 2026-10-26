import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  checkName,
  checkSize,
  defaultSize,
  GiB,
  imageFileName,
  MiB,
  PRESETS,
  toBytes
} from './new-image.ts';

const space = { min: 64 * MiB, max: 10 * GiB, free: 10 * GiB + 256 * MiB };

test('the name gets .img unless it has it', () => {
  assert.equal(imageFileName('exchange'), 'exchange.img');
  assert.equal(imageFileName('  disk.IMG '), 'disk.IMG');
  assert.equal(imageFileName(''), '');
});

test('names the server would refuse are caught before sending', () => {
  assert.equal(checkName('exchange'), '');
  assert.equal(checkName(''), 'nameRequired');
  assert.equal(checkName('a/b'), 'nameInvalid');
  assert.equal(checkName('what?'), 'nameInvalid');
  assert.equal(checkName('.hidden'), 'nameInvalid');
  assert.equal(checkName('name.'), 'nameInvalid');
  assert.equal(checkName('con'), 'nameInvalid');
  assert.equal(checkName('cd.iso'), 'nameIso');
  assert.equal(checkName('a'.repeat(300)), 'nameInvalid');
});

test('a custom size is whole MiB', () => {
  assert.equal(toBytes(1.5, 'GiB'), 1536 * MiB);
  assert.equal(toBytes(100, 'MiB'), 100 * MiB);
  assert.equal(toBytes(0.3, 'MiB'), 0);
});

test('the size fits between the minimum and the free space', () => {
  assert.equal(checkSize(1 * GiB, space), '');
  assert.equal(checkSize(10 * GiB, space), '');
  assert.equal(checkSize(32 * MiB, space), 'sizeTooSmall');
  assert.equal(checkSize(11 * GiB, space), 'sizeTooLarge');
  assert.equal(checkSize(0, space), 'sizeTooSmall');
});

test('the default is the largest preset up to 4 GiB that fits', () => {
  assert.deepEqual(
    PRESETS.map((p) => p / GiB),
    [1, 4, 8, 16, 32]
  );
  assert.equal(defaultSize(space), 4 * GiB);
  assert.equal(defaultSize({ ...space, max: 2 * GiB }), 1 * GiB);
  // Less than the smallest preset: whatever there is.
  assert.equal(defaultSize({ ...space, max: 300 * MiB }), 300 * MiB);
  assert.equal(defaultSize({ ...space, max: 0 }), 0);
});
