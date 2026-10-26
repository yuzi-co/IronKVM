import assert from 'node:assert/strict';
import { test } from 'node:test';

import { groupLicenseFiles, licenseFileUrl } from './licenses.ts';

test('the summary files come first, in reading order, and the texts after', () => {
  const groups = groupLicenseFiles([
    { path: 'MANIFEST', size: 3 },
    { path: 'texts/MIT', size: 1 },
    { path: 'COMPONENTS', size: 2 },
    { path: 'texts/Apache-2.0', size: 1 },
    { path: 'NOTICE', size: 1 },
    { path: 'README', size: 1 }
  ]);

  assert.deepEqual(
    groups.summary.map((file) => file.path),
    ['NOTICE', 'COMPONENTS', 'MANIFEST', 'README']
  );
  assert.deepEqual(groups.texts, [
    { path: 'texts/Apache-2.0', name: 'Apache-2.0', size: 1 },
    { path: 'texts/MIT', name: 'MIT', size: 1 }
  ]);
});

test('a manifest without some summary files lists what it has', () => {
  const groups = groupLicenseFiles([{ path: 'texts/MIT', size: 1 }]);
  assert.deepEqual(groups.summary, []);
  assert.equal(groups.texts.length, 1);
});

test('nothing at all is an empty manifest', () => {
  assert.deepEqual(groupLicenseFiles(undefined), { summary: [], texts: [] });
});

test('a file link names the file in the query, encoded', () => {
  assert.equal(
    licenseFileUrl('https://kvm.local', 'texts/GPL-2.0-or-later WITH Linux-syscall-note'),
    'https://kvm.local/api/licenses/file?path=texts%2FGPL-2.0-or-later%20WITH%20Linux-syscall-note'
  );
});
