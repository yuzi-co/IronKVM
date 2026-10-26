// Run with `pnpm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import type { Preview } from '@/api/backup.ts';

import {
  canImport,
  defaultSelection,
  downloadName,
  formatValue,
  isSelectable,
  MAX_FILE_BYTES,
  parseSettingsFile,
  toggle
} from './flow.ts';

const preview: Preview = {
  version: 1,
  source: { hostname: 'kvm' },
  sections: [
    {
      section: 'logging',
      status: 'change',
      changes: [{ field: 'level', from: 'info', to: 'debug' }]
    },
    { section: 'redfish', status: 'same' },
    { section: 'vnc', status: 'invalid', error: 'the port is used by the web server' },
    { section: 'future', status: 'unknown' },
    {
      section: 'macros',
      status: 'change',
      changes: [{ field: 'macros', from: [], to: { items: 9 } }]
    }
  ]
};

describe('parseSettingsFile', () => {
  it('accepts a settings file', () => {
    const got = parseSettingsFile('{"format":"ironkvm-settings","version":1,"sections":{}}');
    assert.equal(got.ok, true);
  });

  it('names what is wrong with a file it refuses', () => {
    assert.deepEqual(parseSettingsFile('not json'), { ok: false, reason: 'notJson' });
    assert.deepEqual(parseSettingsFile('[1]'), { ok: false, reason: 'notSettings' });
    assert.deepEqual(parseSettingsFile('{"format":"other","version":1,"sections":{}}'), {
      ok: false,
      reason: 'notSettings'
    });
    assert.deepEqual(parseSettingsFile('{"format":"ironkvm-settings","version":7,"sections":{}}'), {
      ok: false,
      reason: 'newer'
    });
    assert.deepEqual(parseSettingsFile('x'.repeat(MAX_FILE_BYTES + 1)), {
      ok: false,
      reason: 'tooLarge'
    });
  });
});

describe('selection', () => {
  it('starts with every section that changes something', () => {
    assert.deepEqual(defaultSelection(preview), ['logging', 'macros']);
  });

  it('offers only the sections that change something', () => {
    assert.deepEqual(
      preview.sections.map((s) => isSelectable(s)),
      [true, false, false, false, true]
    );
  });

  it('allows an import only with a section chosen', () => {
    assert.equal(canImport(preview, ['logging']), true);
    assert.equal(canImport(preview, []), false);
    // A section that cannot be imported never counts.
    assert.equal(canImport(preview, ['vnc']), false);
  });

  it('toggles a section in and out of the selection', () => {
    assert.deepEqual(toggle(['logging'], 'macros', true), ['logging', 'macros']);
    assert.deepEqual(toggle(['logging', 'macros'], 'logging', false), ['macros']);
    assert.deepEqual(toggle(['logging'], 'logging', true), ['logging']);
  });
});

describe('formatValue', () => {
  it('shows values the way the file holds them', () => {
    assert.equal(formatValue('debug'), 'debug');
    assert.equal(formatValue(''), '""');
    assert.equal(formatValue(5), '5');
    assert.equal(formatValue(true), 'true');
    assert.equal(formatValue(null), '-');
    assert.equal(formatValue(undefined), '-');
    assert.equal(formatValue(['a', 'b']), '["a","b"]');
  });

  it('shows a long list by its length', () => {
    assert.deepEqual(formatValue({ items: 9 }), { items: 9 });
  });

  it('cuts a long value', () => {
    const got = formatValue('x'.repeat(500));
    assert.equal(typeof got, 'string');
    assert.ok((got as string).length <= 121);
  });
});

describe('downloadName', () => {
  it('takes the name the server gives', () => {
    assert.equal(
      downloadName('attachment; filename="ironkvm-settings-kvm-20261026-090000.json"', 'x.json'),
      'ironkvm-settings-kvm-20261026-090000.json'
    );
  });

  it('falls back when there is none', () => {
    assert.equal(downloadName(null, 'x.json'), 'x.json');
    assert.equal(downloadName('attachment', 'x.json'), 'x.json');
    // A name with a path in it is not trusted.
    assert.equal(downloadName('attachment; filename="../x.json"', 'y.json'), 'y.json');
  });
});
