// Run with `pnpm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { filterTabs, groupTabs, initialTab, readStored, writeStored } from './nav.ts';

describe('groupTabs', () => {
  it('orders groups by the sidebar and drops empty ones', () => {
    const got = groupTabs([
      { id: 'preferences', group: 'browser' },
      { id: 'netboot', group: 'boot' },
      { id: 'account', group: 'access' },
      { id: 'tls', group: 'access' },
      { id: 'device', group: 'system' }
    ]);
    assert.deepEqual(
      got.map((g) => [g.group, g.tabs.map((t) => t.id)]),
      [
        ['system', ['device']],
        ['access', ['account', 'tls']],
        ['boot', ['netboot']],
        ['browser', ['preferences']]
      ]
    );
  });

  it('leaves footer tabs out of the groups', () => {
    const got = groupTabs([
      { id: 'about', group: 'footer' },
      { id: 'account', group: 'access' }
    ]);
    assert.deepEqual(
      got.map((g) => g.group),
      ['access']
    );
  });
});

describe('initialTab', () => {
  const ids = ['about', 'preferences', 'update'];

  it('opens the update tab when an update is waiting', () => {
    assert.equal(initialTab(ids, 'preferences', true), 'update');
  });

  it('ignores a waiting update the account cannot see', () => {
    assert.equal(initialTab(['about', 'account'], 'account', true), 'account');
  });

  it('returns to the tab used last', () => {
    assert.equal(initialTab(ids, 'preferences', false), 'preferences');
  });

  it('maps a remembered tab that was renamed', () => {
    assert.equal(initialTab(ids, 'appearance', false), 'preferences');
  });

  it('falls back to the first tab for an unknown or missing one', () => {
    assert.equal(initialTab(ids, 'tailscale', false), 'about');
    assert.equal(initialTab(ids, null, false), 'about');
  });
});

describe('filterTabs', () => {
  const tabs = [
    { id: 'device', label: 'Device' },
    { id: 'performance', label: 'Performance' },
    { id: 'network', label: 'Réseau' },
    { id: 'tls', label: 'TLS' }
  ];
  const ids = (query: string) => filterTabs(tabs, query).map((t) => t.id);

  it('matches everything on an empty query', () => {
    assert.deepEqual(ids('  '), ['device', 'performance', 'network', 'tls']);
  });

  it('matches the tab name, ignoring case', () => {
    assert.deepEqual(ids('PERF'), ['performance']);
    assert.deepEqual(ids('rés'), ['network']);
  });

  it('matches search words', () => {
    assert.deepEqual(ids('oled'), ['device']);
    assert.deepEqual(ids('swap'), ['performance']);
    assert.deepEqual(ids('https'), ['tls']);
    assert.deepEqual(ids('mdns'), ['network']);
  });

  it('needs every word of the query', () => {
    assert.deepEqual(ids('power led'), ['device']);
    assert.deepEqual(ids('oled swap'), []);
  });

  it('matches nothing for an unknown word', () => {
    assert.deepEqual(ids('bluetooth'), []);
  });
});

describe('stored values', () => {
  const throwing = {
    getItem(): string | null {
      throw new Error('blocked');
    },
    setItem() {
      throw new Error('blocked');
    }
  };

  it('treats storage that throws as empty', () => {
    assert.equal(readStored(throwing, 'k'), null);
    assert.doesNotThrow(() => writeStored(throwing, 'k', 'v'));
  });

  it('treats missing storage as empty', () => {
    assert.equal(readStored(undefined, 'k'), null);
  });

  it('round-trips a value', () => {
    const map = new Map<string, string>();
    const storage = {
      getItem: (k: string) => map.get(k) ?? null,
      setItem: (k: string, v: string) => void map.set(k, v)
    };
    writeStored(storage, 'k', 'vnc');
    assert.equal(readStored(storage, 'k'), 'vnc');
  });
});
