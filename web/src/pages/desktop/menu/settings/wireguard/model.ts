// The WireGuard page's rules, apart from React so they can be tested.

import type { WireguardSaveRequest, WireguardSettings, WireguardStatus } from '@/api/wireguard.ts';

export type PeerDraft = {
  publicKey: string;
  // A new preshared key; empty keeps the one the board has.
  presharedKey: string;
  hasPresharedKey: boolean;
  clearPresharedKey: boolean;
  endpoint: string;
  allowedIPs: string;
  persistentKeepalive: number;
};

export type Draft = {
  // A key to use instead of the board's; empty keeps it.
  privateKey: string;
  addresses: string;
  dns: string;
  listenPort: number;
  mtu: number;
  peers: PeerDraft[];
};

// WireGuard renews a session every two minutes while traffic flows, so a
// handshake older than three has stopped answering, or nothing was sent.
const handshakeFreshSeconds = 180;

export function splitList(text: string): string[] {
  return text
    .split(/[\s,]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

export function joinList(items: string[]): string {
  return items.join(', ');
}

// isDefaultRoute is true for every address of a family: the board's own
// traffic then goes through the tunnel.
export function isDefaultRoute(allowedIPs: string[]): boolean {
  return allowedIPs.some((ip) => /\/0$/.test(ip.trim()));
}

export function emptyPeer(): PeerDraft {
  return {
    publicKey: '',
    presharedKey: '',
    hasPresharedKey: false,
    clearPresharedKey: false,
    endpoint: '',
    allowedIPs: '',
    // The usual value behind NAT; the board usually is.
    persistentKeepalive: 25
  };
}

export function draftFromSettings(settings: WireguardSettings): Draft {
  return {
    privateKey: '',
    addresses: joinList(settings.addresses),
    dns: joinList(settings.dns),
    listenPort: settings.listenPort,
    mtu: settings.mtu,
    peers: settings.peers.map((peer) => ({
      publicKey: peer.publicKey,
      presharedKey: '',
      hasPresharedKey: peer.hasPresharedKey,
      clearPresharedKey: false,
      endpoint: peer.endpoint,
      allowedIPs: joinList(peer.allowedIPs),
      persistentKeepalive: peer.persistentKeepalive
    }))
  };
}

export function requestFromDraft(draft: Draft): WireguardSaveRequest {
  const privateKey = draft.privateKey.trim();
  return {
    privateKey: privateKey || undefined,
    addresses: splitList(draft.addresses),
    dns: splitList(draft.dns),
    listenPort: draft.listenPort || 0,
    mtu: draft.mtu || 0,
    peers: draft.peers.map((peer) => {
      const presharedKey = peer.presharedKey.trim();
      return {
        publicKey: peer.publicKey.trim(),
        presharedKey: presharedKey || undefined,
        clearPresharedKey: peer.clearPresharedKey,
        endpoint: peer.endpoint.trim(),
        allowedIPs: splitList(peer.allowedIPs),
        persistentKeepalive: peer.persistentKeepalive || 0
      };
    })
  };
}

// peerSnippet is the [Peer] section the other side needs: this board's key,
// and its tunnel addresses as host routes.
export function peerSnippet(publicKey: string, addresses: string[]): string {
  if (!publicKey) return '';
  const hosts = addresses.map((address) => {
    const ip = address.split('/')[0];
    return `${ip}/${ip.includes(':') ? 128 : 32}`;
  });
  const lines = ['[Peer]', `PublicKey = ${publicKey}`];
  if (hosts.length > 0) lines.push(`AllowedIPs = ${hosts.join(', ')}`);
  return lines.join('\n');
}

export type TunnelState = 'off' | 'waiting' | 'connected';

export function tunnelState(status: WireguardStatus | undefined, nowSeconds: number): TunnelState {
  if (!status?.up) return 'off';
  const fresh = status.peers.some(
    (peer) => peer.latestHandshake > 0 && nowSeconds - peer.latestHandshake <= handshakeFreshSeconds
  );
  return fresh ? 'connected' : 'waiting';
}
