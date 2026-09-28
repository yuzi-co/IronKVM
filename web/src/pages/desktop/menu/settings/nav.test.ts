// Run with `pnpm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { groupTabs, initialTab, readStored, writeStored } from './nav.ts';

describe('groupTabs', () => {
  it('orders groups by the sidebar and drops empty ones', () => {
    const got = groupTabs([
      { id: 'netboot', group: 'boot' },
      { id: 'about', group: 'general' },
      { id: 'account', group: 'general' }
    ]);
    assert.deepEqual(
      got.map((g) => [g.group, g.tabs.map((t) => t.id)]),
      [
        ['general', ['about', 'account']],
        ['boot', ['netboot']]
      ]
    );
  });
});

describe('initialTab', () => {
  const ids = ['about', 'appearance', 'update'];

  it('opens the update tab when an update is waiting', () => {
    assert.equal(initialTab(ids, 'appearance', true), 'update');
  });

  it('ignores a waiting update the account cannot see', () => {
    assert.equal(initialTab(['about', 'account'], 'account', true), 'account');
  });

  it('returns to the tab used last', () => {
    assert.equal(initialTab(ids, 'appearance', false), 'appearance');
  });

  it('falls back to the first tab for an unknown or missing one', () => {
    assert.equal(initialTab(ids, 'tailscale', false), 'about');
    assert.equal(initialTab(ids, null, false), 'about');
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
