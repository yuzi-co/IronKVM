import { atom } from 'jotai';

import * as api from '@/api/power-confirm.ts';
import type { PowerConfirmation } from '@/api/power-confirm.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

// How often the presses the agent asked for are read. A press waits a minute
// for a person, so a short delay matters; the answer is a short list from
// memory.
const POLL_MS = 2000;

const basePendingAtom = atom<PowerConfirmation[]>([]);

// read is the mounted atom's reader, for refreshPowerConfirmations.
let read: (() => void) | null = null;

// pendingPowerConfirmationsAtom is the one reader of the pending presses. The
// prompt over the desktop and the one in the chat panel both use it, so the
// board answers one poll however many are on screen, and none while the tab
// is hidden or nobody who may answer is signed in.
export const pendingPowerConfirmationsAtom = atom((get) => get(basePendingAtom));

basePendingAtom.onMount = (set) => {
  let active = true;

  const reader = () => {
    api
      .listPowerConfirmations()
      .then((rsp) => {
        if (!active || rsp.code !== 0) return;
        set(Array.isArray(rsp.data) ? rsp.data : []);
      })
      // A failed poll keeps the last list; the next one tries again.
      .catch(() => {});
  };

  read = reader;
  reader();
  const stop = pollWhileVisible(reader, POLL_MS);
  return () => {
    active = false;
    if (read === reader) read = null;
    stop();
  };
};

// refreshPowerConfirmations reads the list now, as when the chat socket says
// a press changed. It does nothing while no prompt is on screen.
export function refreshPowerConfirmations() {
  read?.();
}

// dropPowerConfirmationAtom removes an answered press at once, before the
// next poll.
export const dropPowerConfirmationAtom = atom(null, (get, set, id: string) => {
  set(
    basePendingAtom,
    get(basePendingAtom).filter((c) => c.id !== id)
  );
});
