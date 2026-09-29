import assert from 'node:assert/strict';
import { test } from 'node:test';

import type { VentoyStatus } from '@/api/ventoy.ts';

import {
  basename,
  formatSize,
  ventoyPresent,
  ventoyStatusText,
  ventoyUsable
} from './ventoy-status.ts';

function status(over: Partial<VentoyStatus> = {}): VentoyStatus {
  return {
    kernel: true,
    onData: true,
    installed: true,
    version: '1.1.05',
    images: [],
    missing: [],
    inDrive: false,
    device: '',
    size: 0,
    ...over
  };
}

test('Media offers Ventoy only when the kernel supports it and it is installed', () => {
  assert.equal(ventoyUsable(null), false);
  assert.equal(ventoyUsable(status({ kernel: false })), false);
  assert.equal(ventoyUsable(status({ installed: false })), false);
  assert.equal(ventoyUsable(status()), true);
});

test('the status line follows the state, the missing kernel first', () => {
  assert.deepEqual(ventoyStatusText(status({ kernel: false, installed: false })), {
    key: 'image.ventoy.statusNoKernel'
  });
  assert.deepEqual(ventoyStatusText(status({ installed: false })), {
    key: 'image.ventoy.statusNotInstalled'
  });
  assert.deepEqual(
    ventoyStatusText(status({ inDrive: true, size: 3 * 1024 ** 3, images: ['a'] })),
    {
      key: 'image.ventoy.statusInDrive',
      values: { size: '3.0 GiB' }
    }
  );
  assert.deepEqual(ventoyStatusText(status({ images: ['a', 'b'] })), {
    key: 'image.ventoy.statusSelected',
    values: { count: 2 }
  });
  assert.deepEqual(ventoyStatusText(status({ images: null })), { key: 'image.ventoy.statusReady' });
});

test('only selected images that still exist can build the disk', () => {
  assert.deepEqual(
    ventoyPresent(status({ images: ['/data/a.iso', '/data/b.iso'], missing: ['/data/b.iso'] })),
    ['/data/a.iso']
  );
  assert.deepEqual(ventoyPresent(status({ images: null, missing: null })), []);
});

test('sizes and names read as the dialog shows them', () => {
  assert.equal(formatSize(512 * 1024 * 1024), '512.0 MiB');
  assert.equal(formatSize(1536 * 1024 * 1024), '1.5 GiB');
  assert.equal(basename('/data/iso/debian.iso'), 'debian.iso');
  assert.equal(basename('C:\\images\\win.iso'), 'win.iso');
});
