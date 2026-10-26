import assert from 'node:assert/strict';
import { test } from 'node:test';

import { breadcrumbs, browseState, fileSize, joinPath, parentPath } from './file-paths.ts';

test('a path joins onto the root without a leading slash', () => {
  assert.equal(joinPath('/', 'a.txt'), 'a.txt');
  assert.equal(joinPath('docs', 'a.txt'), 'docs/a.txt');
  assert.equal(joinPath('docs/sub', 'a.txt'), 'docs/sub/a.txt');
});

test('the parent of a top-level directory is the root', () => {
  assert.equal(parentPath('docs'), '/');
  assert.equal(parentPath('docs/sub'), 'docs');
  assert.equal(parentPath('/'), '/');
});

test('breadcrumbs lead from the root to the directory', () => {
  assert.deepEqual(breadcrumbs('/'), [{ name: '', path: '/' }]);
  assert.deepEqual(breadcrumbs('EFI/BOOT'), [
    { name: '', path: '/' },
    { name: 'EFI', path: 'EFI' },
    { name: 'BOOT', path: 'EFI/BOOT' }
  ]);
});

test('an image is free, read by the host, or written by the host', () => {
  const drives = [
    { id: 'disk' as const, type: 'disk' as const, file: '/data/rw.img', ro: false },
    { id: 'cdrom' as const, type: 'cdrom' as const, file: '/data/cd.iso', ro: true }
  ];
  assert.equal(browseState('/data/free.img', drives, []), 'free');
  assert.equal(browseState('/data/cd.iso', drives, []), 'hostReads');
  assert.equal(browseState('/data/rw.img', drives, []), 'hostWrites');
  assert.equal(browseState('/data/v.iso', drives, ['/data/v.iso']), 'hostReads');
});

test('a size is named in the unit that keeps it short', () => {
  assert.equal(fileSize(0), '0 B');
  assert.equal(fileSize(1023), '1023 B');
  assert.equal(fileSize(1536), '1.5 KiB');
  assert.equal(fileSize(200 * 1024 * 1024), '200 MiB');
  assert.equal(fileSize(4.7 * 1024 ** 3), '4.7 GiB');
});
