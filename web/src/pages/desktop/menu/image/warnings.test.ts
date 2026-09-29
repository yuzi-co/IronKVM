import assert from 'node:assert/strict';
import { test } from 'node:test';

import { CD_MAX_BYTES, CD_MIN_BYTES, driveWarnings, imageWarnings } from './warnings.ts';

const GiB = 1024 ** 3;

test('the CD limit is the DVD addressing limit, about 31.6 GiB', () => {
  assert.equal(CD_MAX_BYTES, 33_957_083_136);
  assert.ok(CD_MAX_BYTES / GiB > 31.6 && CD_MAX_BYTES / GiB < 31.7);
});

test('an empty drive and the Ventoy device have nothing to warn about', () => {
  assert.deepEqual(driveWarnings({ id: 'disk', file: '', ro: false }), []);
  assert.deepEqual(
    driveWarnings({ id: 'disk', file: '/dev/mapper/ventoy', ro: false, size: 0 }),
    []
  );
});

test('a disk served read-write warns, read-only does not', () => {
  assert.deepEqual(driveWarnings({ id: 'disk', file: '/data/a.img', ro: false, size: GiB }), [
    'writable'
  ]);
  assert.deepEqual(driveWarnings({ id: 'disk', file: '/data/a.img', ro: true, size: GiB }), []);
});

test('a CD image past the limit or under 300 sectors warns', () => {
  const cd = (size: number) => driveWarnings({ id: 'cdrom', file: '/data/a.iso', ro: true, size });
  assert.deepEqual(cd(5 * GiB), []);
  assert.deepEqual(cd(CD_MAX_BYTES + 1), ['tooBigForCd']);
  assert.deepEqual(cd(CD_MIN_BYTES - 1), ['tooSmallForCd']);
});

test('a served file whose path is gone says only that', () => {
  assert.deepEqual(
    driveWarnings({ id: 'disk', file: '/data/a.img', ro: false, size: 0, missing: true }),
    ['missing']
  );
});

test('library images warn for the drive they would go into', () => {
  assert.deepEqual(imageWarnings(undefined, 'cdrom'), []);
  assert.deepEqual(imageWarnings(0, 'disk'), ['empty']);
  assert.deepEqual(imageWarnings(40 * GiB, 'disk'), []);
  assert.deepEqual(imageWarnings(40 * GiB, 'cdrom'), ['tooBigForCd']);
});
