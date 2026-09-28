import { useEffect, useState } from 'react';

import * as api from '@/api/vm';

import { PowerLed } from './power-led.tsx';

// PowerLedSetting is the power menu's LED line and "Power LED connected"
// switch, for the settings pages that depend on it (watchdog, IPMI, Redfish).
// The setting lives on the device, so this and the power menu stay in step
// through the same API; each reads it on its own.
export const PowerLedSetting = () => {
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

  return (
    <div className="rounded-lg border border-neutral-700/60 px-2 py-3">
      <PowerLed isPowerOn={isPowerOn} connected={connected} setConnected={setConnected} />
    </div>
  );
};
