import { useSyncExternalStore } from 'react';

import { getServerVersion } from '@/api/version.ts';
import {
  contractFromResponse,
  hasFeature,
  LEGACY_CONTRACT,
  type ServerContract,
  type ServerFeature
} from '@/lib/server-contract.ts';

// The contract is asked for once per page load: an application update restarts
// the server and the page reloads after it, and a UI update reloads the page.

let contract: ServerContract | null = null;
let pending: Promise<ServerContract> | null = null;
const listeners = new Set<() => void>();

// loadServerContract starts the request, once. main.tsx calls it before the
// first render, so the answer is usually in by the time anything asks.
export function loadServerContract(): Promise<ServerContract> {
  if (!pending) {
    pending = getServerVersion()
      .then((rsp) => (rsp.code === 0 ? contractFromResponse(rsp.data) : LEGACY_CONTRACT))
      // A 404 is an older server. Anything else (the board restarting, a
      // network error) is read the same way rather than as a server that has
      // nothing: the features then fall back to asking and coping.
      .catch(() => LEGACY_CONTRACT)
      .then((c) => {
        contract = c;
        listeners.forEach((l) => l());
        return c;
      });
  }
  return pending;
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  void loadServerContract();
  return () => {
    listeners.delete(listener);
  };
}

// currentServerContract is the server's contract, or null until it is known,
// for code outside React.
export function currentServerContract(): ServerContract | null {
  return contract;
}

// useServerContract is the server's contract, or null until it is known.
export function useServerContract(): ServerContract | null {
  return useSyncExternalStore(subscribe, () => contract);
}

// useServerFeature says whether the UI may use a server feature: null until
// the contract is known, then true or false. Gate the calls a feature makes on
// it, and declare the feature once in SERVER_FEATURES.
export function useServerFeature(name: ServerFeature): boolean | null {
  const c = useServerContract();
  return c === null ? null : hasFeature(c, name);
}
