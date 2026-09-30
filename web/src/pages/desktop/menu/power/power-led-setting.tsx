import { useAtom, useAtomValue } from 'jotai';

import { gpioAtom, ledConnectedAtom } from '@/jotai/power.ts';

import { PowerLed } from './power-led.tsx';

// PowerLedSetting is the power menu's LED line and "Power LED connected"
// switch, for the settings pages that depend on it (watchdog, IPMI, Redfish).
// The setting lives on the device; this and the power menu read it from one
// shared poll, and a change made here shows in the menu at once.
// framed draws it as a box of its own, for pages where it stands apart from
// the rows around it; the Device page lists it as one of its rows instead.
export const PowerLedSetting = ({ framed = true }: { framed?: boolean }) => {
  const isPowerOn = useAtomValue(gpioAtom)?.pwr ?? false;
  const [connected, setConnected] = useAtom(ledConnectedAtom);

  const led = <PowerLed isPowerOn={isPowerOn} connected={connected} setConnected={setConnected} />;
  // PowerLed pads itself for the box; unframed, the row lines up with the
  // rows around it.
  if (!framed) return <div className="-mx-1">{led}</div>;

  return <div className="rounded-lg border border-neutral-700/60 px-2 py-3">{led}</div>;
};
