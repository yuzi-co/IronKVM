import { atom } from 'jotai';

import * as api from '@/api/vm';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

// The host's front panel as the board reads it: the power LED, whether its
// header is wired, and the HDD LED where the board has one.
export type Gpio = {
  pwr: boolean;
  ledConnected: boolean;
  hasHdd: boolean;
  hdd: boolean;
};

const GPIO_POLL_MS = 5000;

// Bumped by every local change, so a poll that was already in flight when the
// operator flipped a switch does not put the old value back.
let writeVersion = 0;

const baseGpioAtom = atom<Gpio | null>(null);

// gpioAtom is the one reader of /api/vm/gpio. The power menu and the settings
// pages that show the LED all use it, so the board answers one poll however
// many of them are on screen, and none while the tab is hidden. It is null
// until the first answer.
export const gpioAtom = atom((get) => get(baseGpioAtom));

baseGpioAtom.onMount = (set) => {
  let active = true;

  function read() {
    const version = writeVersion;
    api
      .getGpio()
      .then((rsp) => {
        if (!active || rsp.code !== 0 || version !== writeVersion) return;
        set({
          pwr: !!rsp.data.pwr,
          ledConnected: !!rsp.data.ledConnected,
          hasHdd: !!rsp.data.hasHdd,
          hdd: !!rsp.data.hdd
        });
      })
      // A failed poll keeps the last state; the next one tries again.
      .catch(() => {});
  }

  read();
  const stop = pollWhileVisible(read, GPIO_POLL_MS);
  return () => {
    active = false;
    stop();
  };
};

// ledConnectedAtom writes the "power LED connected" switch into the shared
// state once the board has taken it, so every view shows it at once.
export const ledConnectedAtom = atom(
  (get) => get(baseGpioAtom)?.ledConnected ?? false,
  (get, set, connected: boolean) => {
    writeVersion += 1;
    const prev = get(baseGpioAtom);
    set(baseGpioAtom, {
      pwr: prev?.pwr ?? false,
      hasHdd: prev?.hasHdd ?? false,
      hdd: prev?.hdd ?? false,
      ledConnected: connected
    });
  }
);
