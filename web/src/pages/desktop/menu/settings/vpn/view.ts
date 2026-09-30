import type { Memory, Peer, State } from './types.ts';

// Tag is what the badge beside the title says: connected, waiting for a
// login, off (the daemon stopped, or running with the network down), or error
// when the status could not be read.
export type Tag = 'connected' | 'needsLogin' | 'off' | 'error';

export function statusTag(state: State | undefined, failed: boolean): Tag | undefined {
  if (failed) return 'error';
  switch (state) {
    case undefined:
    case 'notInstall':
      return undefined;
    case 'running':
      return 'connected';
    case 'notLogin':
      return 'needsLogin';
    default:
      return 'off';
  }
}

// isSwitchedOn is the Connected switch. A daemon waiting for its login counts
// as on: the operator turned it on and the login below is the next step, and
// turning it off stops the daemon. A daemon running with the network down is
// off, and turning it on brings the network up.
export function isSwitchedOn(state: State | undefined): boolean {
  return state === 'running' || state === 'notLogin';
}

// hasDaemon reports whether the daemon runs, which is when it can be
// restarted or logged out.
export function hasDaemon(state: State | undefined): boolean {
  return state === 'notLogin' || state === 'stopped' || state === 'running';
}

// sortPeers puts the online peers first, then each part by name.
export function sortPeers(peers: Peer[]): Peer[] {
  return [...peers].sort((a, b) => {
    if (a.online !== b.online) return a.online ? -1 : 1;
    return a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true });
  });
}

export function countOnline(peers: Peer[]): number {
  return peers.filter((p) => p.online).length;
}

// memoryUse is the daemon's resident memory against the addons group's limit:
// memory.high, where the kernel starts to throttle, or memory.max without it.
// pressed is set once the whole group is within a tenth of memory.high.
export function memoryUse(memory: Memory): { used: number; limit: number; pressed: boolean } {
  return {
    used: memory.daemonRss,
    limit: memory.groupHigh || memory.groupMax,
    pressed: memory.groupHigh > 0 && memory.groupCurrent >= memory.groupHigh * 0.9
  };
}
