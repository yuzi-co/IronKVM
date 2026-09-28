// Waiting for the device to restart, and for it to come back.
//
// A restart used to be a fixed wait followed by a reload. Too short and the
// reload lands on a server that is still stopped; too long and the operator
// stares at a spinner for nothing. Watching the origin go down and come back
// is right in both cases.

const PROBE_INTERVAL_MS = 2000;

export type WaitDeps = {
  // probe answers whether the server is serving now.
  probe: () => Promise<boolean>;
  sleep: (ms: number) => Promise<void>;
  now: () => number;
  intervalMs: number;
  // signal stops the wait early; the wait then answers false.
  signal?: AbortSignal;
};

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// reachable answers whether this origin is still serving. redirect: 'manual'
// keeps it from following the device's own 307 into a scheme whose certificate
// the browser has not been shown yet.
export async function reachable(): Promise<boolean> {
  try {
    await fetch(`${window.location.origin}/`, { cache: 'no-store', redirect: 'manual' });
    return true;
  } catch {
    return false;
  }
}

function withDefaults(deps: Partial<WaitDeps>): WaitDeps {
  return {
    probe: deps.probe ?? reachable,
    sleep: deps.sleep ?? sleep,
    now: deps.now ?? Date.now,
    intervalMs: deps.intervalMs ?? PROBE_INTERVAL_MS,
    signal: deps.signal
  };
}

// waitUntil polls until the server is serving, or is not, and gives up rather
// than waiting for ever. It answers whether the state was reached.
export async function waitUntil(
  serving: boolean,
  budgetMs: number,
  deps: Partial<WaitDeps> = {}
): Promise<boolean> {
  const d = withDefaults(deps);
  const deadline = d.now() + budgetMs;

  for (;;) {
    if (d.signal?.aborted) return false;
    if ((await d.probe()) === serving) return true;
    if (d.now() >= deadline) return false;
    await d.sleep(d.intervalMs);
  }
}

// waitForRestart waits for the server to stop and then to serve again. A
// server that restarts between two probes is never seen down; that costs the
// down budget and nothing else. It answers whether the server came back.
export async function waitForRestart(
  downMs: number,
  upMs: number,
  deps: Partial<WaitDeps> = {}
): Promise<boolean> {
  await waitUntil(false, downMs, deps);
  if (deps.signal?.aborted) return false;
  return waitUntil(true, upMs, deps);
}

// The budgets for a reboot of the whole board and for a restart of the server
// alone. The board takes about a minute to boot, longer after an update that
// resizes the root filesystem.
export const REBOOT_DOWN_MS = 60_000;
export const REBOOT_UP_MS = 300_000;
export const SERVER_RESTART_DOWN_MS = 30_000;
export const SERVER_RESTART_UP_MS = 180_000;

// reloadAfterRestart waits for the server to go down and come back, then
// reloads the page. It reloads even when the server never answered: a page
// that cannot talk to its server is no better than a reload that fails.
// A stopped wait does not reload.
export async function reloadAfterRestart(downMs: number, upMs: number, signal?: AbortSignal) {
  await waitForRestart(downMs, upMs, { signal });
  if (signal?.aborted) return;
  window.location.reload();
}
