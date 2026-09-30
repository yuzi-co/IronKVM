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

// sortPeers puts the online peers first, then the ones on demand, then the
// offline ones, each part by name.
export function sortPeers(peers: Peer[]): Peer[] {
  const rank = (p: Peer) => (p.online ? 0 : p.idle ? 1 : 2);
  return [...peers].sort((a, b) => {
    if (rank(a) !== rank(b)) return rank(a) - rank(b);
    return a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true });
  });
}

export function countOnline(peers: Peer[]): number {
  return peers.filter((p) => p.online).length;
}

export function countIdle(peers: Peer[]): number {
  return peers.filter((p) => !p.online && p.idle).length;
}

// hasSettled reports whether a network just switched on shows what the page
// is for: an address and a peer online, or on demand under NetBird's lazy
// connections, where no peer connects until traffic needs it. The daemon answers connect before
// that, NetBird as soon as it is Connecting and Tailscale once it is Running
// but before its peers are reachable, so until then the page asks again
// quickly. A daemon waiting for its login has nothing more coming.
export function hasSettled(state: State | undefined, ip: string, peers: Peer[] | null): boolean {
  if (state === undefined || state === 'notInstall' || state === 'notLogin') return true;
  return state === 'running' && !!ip && countOnline(peers ?? []) + countIdle(peers ?? []) > 0;
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
