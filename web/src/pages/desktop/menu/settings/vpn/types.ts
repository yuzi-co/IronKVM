import type { ReactNode } from 'react';

export type State = 'notInstall' | 'notRunning' | 'notLogin' | 'stopped' | 'running';

export type Peer = { name: string; ip: string; online: boolean };

// Bytes. groupHigh and groupMax are 0 when the group sets no limit.
export type Memory = {
  daemonRss: number;
  groupCurrent: number;
  groupHigh: number;
  groupMax: number;
};

export type Status = {
  state: State;
  version: string;
  ip: string;
  name: string;
  account: string;
  control: boolean;
  peers: Peer[] | null;
  uptimeSec: number;
  bootEnabled: boolean;
  memory: Memory;
  blockedBy: string;
};

export type UpdateInfo = { current: string; latest: string };

// The server's reply, as lib/http.ts returns it.
export type Rsp = { code: number; msg: string; data: any };

// What the shared page needs from an add-on's API module. Login differs
// between the two and is each add-on's own form.
export type VpnApi = {
  install: () => Promise<Rsp>;
  uninstall: () => Promise<Rsp>;
  getStatus: () => Promise<Rsp>;
  start: () => Promise<Rsp>;
  stop: () => Promise<Rsp>;
  restart: () => Promise<Rsp>;
  up: () => Promise<Rsp>;
  down: () => Promise<Rsp>;
  logout: () => Promise<Rsp>;
  setBoot: (enabled: boolean) => Promise<Rsp>;
  getUpdate: () => Promise<Rsp>;
  update: () => Promise<Rsp>;
};

export type VpnInfo = {
  id: 'tailscale' | 'netbird';
  title: string;
  api: VpnApi;
  logoutLabel: string;
  logoutWarning: string;
  renderLogin: (onSuccess: () => void) => ReactNode;
  installHelp?: ReactNode;
};
