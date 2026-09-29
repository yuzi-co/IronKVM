import { useEffect, useState } from 'react';

import * as api from '@/api/vm';

import { PowerLed } from './power-led.tsx';

// PowerLedSetting is the power menu's LED line and "Power LED connected"
// switch, for the settings pages that depend on it (watchdog, IPMI, Redfish).
// The setting lives on the device, so this and the power menu stay in step
// through the same API; each reads it on its own.
// framed draws it as a box of its own, for pages where it stands apart from
// the rows around it; the Device page lists it as one of its rows instead.
export const PowerLedSetting = ({ framed = true }: { framed?: boolean }) => {
  const [isPowerOn, setIsPowerOn] = useState(false);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    let active = true;

    async function getLed() {
      try {
        const rsp = await api.getGpio();
        if (active && rsp.code === 0) {
          setIsPowerOn(rsp.data.pwr);
          setConnected(rsp.data.ledConnected);
        }
      } catch (err) {
        console.log(err);
      }
    }

    getLed();
    const interval = setInterval(getLed, 5000);

    return () => {
      active = false;
      clearInterval(interval);
    };
  }, []);

  const led = <PowerLed isPowerOn={isPowerOn} connected={connected} setConnected={setConnected} />;
  // PowerLed pads itself for the box; unframed, the row lines up with the
  // rows around it.
  if (!framed) return <div className="-mx-1">{led}</div>;

  return <div className="rounded-lg border border-neutral-700/60 px-2 py-3">{led}</div>;
};
