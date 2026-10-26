import assert from 'node:assert/strict';
import { test } from 'node:test';

import { cacheHitRate, isHttpUrl, urlMountLabel, type UrlMount } from './url-media.ts';

const mount: UrlMount = {
  drive: 'cdrom',
  device: '/dev/nbd0',
  url: 'https://isos.lan:8443/linux/debian%2012.iso?',
  host: 'isos.lan:8443',
  size: 1 << 30,
  bytesRead: 0,
  cache: { hits: 0, misses: 0, prefetched: 0 },
  since: '2026-10-26T00:00:00Z'
};

test('only http and https URLs with a host are mountable', () => {
  assert.equal(isHttpUrl('https://example.com/x.iso'), true);
  assert.equal(isHttpUrl(' http://10.0.30.5:8080/x.img '), true);
  assert.equal(isHttpUrl('ftp://example.com/x.iso'), false);
  assert.equal(isHttpUrl('/data/x.iso'), false);
  assert.equal(isHttpUrl(''), false);
});

test('a URL mount is named by its file and host, anything else is not', () => {
  assert.equal(urlMountLabel('/dev/nbd0', [mount]), 'debian 12.iso (isos.lan:8443)');
  assert.equal(urlMountLabel('/dev/nbd1', [mount]), null);
  assert.equal(urlMountLabel('/data/a.iso', [mount]), null);
  assert.equal(
    urlMountLabel('/dev/nbd0', [{ ...mount, url: 'https://isos.lan:8443/?' }]),
    'isos.lan:8443'
  );
});

test('the cache hit rate is a whole percent, none before a read', () => {
  assert.equal(cacheHitRate({ hits: 0, misses: 0, prefetched: 0 }), null);
  assert.equal(cacheHitRate({ hits: 3, misses: 1, prefetched: 2 }), 75);
  assert.equal(cacheHitRate({ hits: 2, misses: 1, prefetched: 0 }), 67);
});
