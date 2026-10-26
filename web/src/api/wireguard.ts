import { http } from '@/lib/http.ts';

// The built-in WireGuard tunnel (server/service/wireguard). The server never
// sends the private key or a preshared key back: only the board's public key
// and whether a peer has a preshared key.

export type WireguardPeer = {
  publicKey: string;
  hasPresharedKey: boolean;
  endpoint: string;
  allowedIPs: string[];
  persistentKeepalive: number;
};

export type WireguardSettings = {
  // The board's public key, for the peer's side; empty before a key exists.
  publicKey: string;
  configured: boolean;
  addresses: string[];
  dns: string[];
  listenPort: number;
  mtu: number;
  peers: WireguardPeer[];
  startAtBoot: boolean;
  defaultRoute: boolean;
};

export type WireguardPeerStatus = {
  publicKey: string;
  endpoint?: string;
  // Unix seconds, 0 for never.
  latestHandshake: number;
  rxBytes: number;
  txBytes: number;
  allowedIPs: string[];
};

export type WireguardStatus = {
  up: boolean;
  interface: string;
  publicKey?: string;
  listenPort?: number;
  peers: WireguardPeerStatus[];
  warnings?: string[];
  error?: string;
};

export type WireguardState = {
  settings: WireguardSettings;
  status: WireguardStatus;
  // What the last change ignored or could not do.
  warnings?: string[];
};

export type WireguardPeerRequest = {
  publicKey: string;
  // Empty keeps the stored one, unless clearPresharedKey.
  presharedKey?: string;
  clearPresharedKey?: boolean;
  endpoint: string;
  allowedIPs: string[];
  persistentKeepalive: number;
};

export type WireguardSaveRequest = {
  // Empty keeps the board's key; a board without one makes one.
  privateKey?: string;
  addresses: string[];
  dns: string[];
  listenPort: number;
  mtu: number;
  peers: WireguardPeerRequest[];
};

const base = '/api/network/wireguard';

export function getState() {
  return http.get(base);
}

export function save(request: WireguardSaveRequest) {
  return http.post(base, request);
}

export function importConfig(config: string) {
  return http.post(`${base}/import`, { config });
}

export function up() {
  return http.post(`${base}/up`);
}

export function down() {
  return http.post(`${base}/down`);
}

export function setBoot(enabled: boolean) {
  return http.post(`${base}/boot`, { enabled });
}

export function regenerateKey() {
  return http.post(`${base}/key`);
}

export function remove() {
  return http.delete(base);
}
