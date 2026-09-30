// Run with `npm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import {
  countOnline,
  hasDaemon,
  hasSettled,
  isSwitchedOn,
  memoryUse,
  sortPeers,
  statusTag
} from './view.ts';

describe('statusTag', () => {
  it('maps each state to its badge', () => {
    assert.equal(statusTag('running', false), 'connected');
    assert.equal(statusTag('notLogin', false), 'needsLogin');
    assert.equal(statusTag('stopped', false), 'off');
    assert.equal(statusTag('notRunning', false), 'off');
  });

  it('shows no badge before the status or without an install', () => {
    assert.equal(statusTag(undefined, false), undefined);
    assert.equal(statusTag('notInstall', false), undefined);
  });

  it('says error when the status could not be read', () => {
    assert.equal(statusTag('running', true), 'error');
    assert.equal(statusTag(undefined, true), 'error');
  });
});

describe('isSwitchedOn and hasDaemon', () => {
  it('counts a daemon waiting for its login as on', () => {
    assert.equal(isSwitchedOn('running'), true);
    assert.equal(isSwitchedOn('notLogin'), true);
    assert.equal(isSwitchedOn('stopped'), false);
    assert.equal(isSwitchedOn('notRunning'), false);
    assert.equal(isSwitchedOn(undefined), false);
  });

  it('knows when the daemon runs', () => {
    assert.equal(hasDaemon('stopped'), true);
    assert.equal(hasDaemon('notLogin'), true);
    assert.equal(hasDaemon('notRunning'), false);
    assert.equal(hasDaemon('notInstall'), false);
  });
});

describe('sortPeers', () => {
  const peers = [
    { name: 'nas', ip: '100.64.0.3', online: false },
    { name: 'laptop10', ip: '100.64.0.4', online: true },
    { name: 'Desktop', ip: '100.64.0.1', online: true },
    { name: 'laptop9', ip: '100.64.0.5', online: true },
    { name: 'backup', ip: '100.64.0.2', online: false }
  ];

  it('puts online peers first, each part by name', () => {
    assert.deepEqual(
      sortPeers(peers).map((p) => p.name),
      ['Desktop', 'laptop9', 'laptop10', 'backup', 'nas']
    );
  });

  it('leaves its argument alone', () => {
    sortPeers(peers);
    assert.equal(peers[0].name, 'nas');
  });

  it('counts the online ones', () => {
    assert.equal(countOnline(peers), 3);
    assert.equal(countOnline([]), 0);
  });
});

describe('hasSettled', () => {
  const online = [{ name: 'nas', ip: '100.64.0.2', online: true }];
  const offline = [{ name: 'nas', ip: '100.64.0.2', online: false }];

  it('waits for an address and a peer online', () => {
    assert.equal(hasSettled('running', '', []), false);
    assert.equal(hasSettled('running', '', null), false);
    assert.equal(hasSettled('running', '100.64.0.1', []), false);
    assert.equal(hasSettled('running', '100.64.0.1', offline), false);
    assert.equal(hasSettled('running', '', online), false);
    assert.equal(hasSettled('running', '100.64.0.1', online), true);
  });

  it('waits while the daemon or the network is still coming up', () => {
    assert.equal(hasSettled('notRunning', '', []), false);
    assert.equal(hasSettled('stopped', '', []), false);
  });

  it('has nothing to wait for while a login is needed', () => {
    assert.equal(hasSettled('notLogin', '', []), true);
    assert.equal(hasSettled('notInstall', '', []), true);
    assert.equal(hasSettled(undefined, '', null), true);
  });
});

describe('memoryUse', () => {
  it('measures the daemon against memory.high', () => {
    assert.deepEqual(
      memoryUse({ daemonRss: 30e6, groupCurrent: 40e6, groupHigh: 64e6, groupMax: 96e6 }),
      { used: 30e6, limit: 64e6, pressed: false }
    );
  });

  it('falls back to memory.max and flags a group near its limit', () => {
    assert.deepEqual(
      memoryUse({ daemonRss: 30e6, groupCurrent: 60e6, groupHigh: 0, groupMax: 96e6 }),
      { used: 30e6, limit: 96e6, pressed: false }
    );
    assert.equal(
      memoryUse({ daemonRss: 30e6, groupCurrent: 60e6, groupHigh: 64e6, groupMax: 0 }).pressed,
      true
    );
  });

  it('has no limit without a group', () => {
    assert.equal(
      memoryUse({ daemonRss: 30e6, groupCurrent: 0, groupHigh: 0, groupMax: 0 }).limit,
      0
    );
  });
});
