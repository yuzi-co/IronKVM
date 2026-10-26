// Run with `npm test`. Node strips the types itself.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import {
  draftFromSettings,
  emptyPeer,
  isDefaultRoute,
  peerSnippet,
  requestFromDraft,
  splitList,
  tunnelState
} from './model.ts';

const settings = {
  publicKey: 'BOARDPUB=',
  configured: true,
  addresses: ['10.8.0.2/24', 'fd00:8::2/64'],
  dns: ['10.8.0.1'],
  listenPort: 0,
  mtu: 0,
  peers: [
    {
      publicKey: 'PEER=',
      hasPresharedKey: true,
      endpoint: 'vpn.example.com:51820',
      allowedIPs: ['10.8.0.0/24'],
      persistentKeepalive: 25
    }
  ],
  startAtBoot: false,
  defaultRoute: false
};

describe('splitList', () => {
  it('splits on commas and white space and drops blanks', () => {
    assert.deepEqual(splitList(' 10.0.0.0/24, 10.1.0.0/24\n192.168.1.0/24 ,, '), [
      '10.0.0.0/24',
      '10.1.0.0/24',
      '192.168.1.0/24'
    ]);
    assert.deepEqual(splitList(''), []);
  });
});

describe('isDefaultRoute', () => {
  it('is any /0 of either family', () => {
    assert.equal(isDefaultRoute(['10.0.0.0/24']), false);
    assert.equal(isDefaultRoute(['10.0.0.0/24', '0.0.0.0/0']), true);
    assert.equal(isDefaultRoute(['::/0']), true);
    assert.equal(isDefaultRoute(['10.0.0.0/8']), false);
  });
});

describe('draft and request', () => {
  it('round-trips the fields and never carries a key the page does not have', () => {
    const draft = draftFromSettings(settings);
    assert.equal(draft.addresses, '10.8.0.2/24, fd00:8::2/64');
    assert.equal(draft.privateKey, '');
    assert.equal(draft.peers[0].presharedKey, '');
    assert.equal(draft.peers[0].hasPresharedKey, true);

    const request = requestFromDraft(draft);
    assert.deepEqual(request.addresses, ['10.8.0.2/24', 'fd00:8::2/64']);
    assert.equal(request.privateKey, undefined);
    assert.equal(request.peers[0].presharedKey, undefined);
    assert.equal(request.peers[0].clearPresharedKey, false);
    assert.deepEqual(request.peers[0].allowedIPs, ['10.8.0.0/24']);
    assert.equal(request.peers[0].persistentKeepalive, 25);
  });

  it('sends a typed key and a cleared preshared key', () => {
    const draft = draftFromSettings(settings);
    draft.privateKey = ' PRIV= ';
    draft.peers[0].clearPresharedKey = true;
    const request = requestFromDraft(draft);
    assert.equal(request.privateKey, 'PRIV=');
    assert.equal(request.peers[0].clearPresharedKey, true);
  });

  it('starts a new peer empty', () => {
    const peer = emptyPeer();
    assert.equal(peer.publicKey, '');
    assert.equal(peer.persistentKeepalive, 25);
  });
});

describe('peerSnippet', () => {
  it('gives the peer the board key and the board addresses as hosts', () => {
    assert.equal(
      peerSnippet('BOARDPUB=', ['10.8.0.2/24', 'fd00:8::2/64']),
      '[Peer]\nPublicKey = BOARDPUB=\nAllowedIPs = 10.8.0.2/32, fd00:8::2/128'
    );
  });

  it('is empty without a key', () => {
    assert.equal(peerSnippet('', ['10.8.0.2/24']), '');
  });
});

describe('tunnelState', () => {
  const now = 1_800_000_000;
  const status = (up: boolean, handshake: number) => ({
    up,
    interface: 'wg0',
    peers: [{ publicKey: 'P', latestHandshake: handshake, rxBytes: 0, txBytes: 0, allowedIPs: [] }]
  });

  it('is off when the interface is down', () => {
    assert.equal(tunnelState(status(false, now), now), 'off');
    assert.equal(tunnelState(undefined, now), 'off');
  });

  it('is connected after a recent handshake and waiting otherwise', () => {
    assert.equal(tunnelState(status(true, now - 30), now), 'connected');
    assert.equal(tunnelState(status(true, now - 600), now), 'waiting');
    assert.equal(tunnelState(status(true, 0), now), 'waiting');
  });
});
